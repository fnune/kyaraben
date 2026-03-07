package syncguest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fnune/kyaraben/internal/syncthing"
)

type Config struct {
	DataDir       string
	SyncthingPath string
	Syncthing     syncthing.Config
	RelayURLs     []string
}

func DefaultConfig(dataDir string) Config {
	return Config{
		DataDir:       dataDir,
		SyncthingPath: "syncthing",
		Syncthing:     syncthing.DefaultConfig(),
		RelayURLs:     syncthing.ProductionRelayURLs,
	}
}

type Manager struct {
	config      Config
	client      syncthing.SyncClient
	relayClient *syncthing.RelayClient
	process     *exec.Cmd
	logger      syncthing.Logger
}

func New(config Config) *Manager {
	stConfig := config.Syncthing
	stConfig.BaseURL = fmt.Sprintf("http://localhost:%d", config.Syncthing.GUIPort)

	return &Manager{
		config: config,
		client: syncthing.NewClient(stConfig),
		logger: noopLogger{},
	}
}

func NewWithClient(config Config, client syncthing.SyncClient) *Manager {
	return &Manager{
		config: config,
		client: client,
		logger: noopLogger{},
	}
}

type noopLogger struct{}

func (noopLogger) Debug(format string, args ...any) {}
func (noopLogger) Info(format string, args ...any)  {}
func (noopLogger) Error(format string, args ...any) {}

func (m *Manager) SetLogger(logger syncthing.Logger) {
	m.logger = logger
}

func (m *Manager) Client() syncthing.SyncClient {
	return m.client
}

func (m *Manager) GUIPort() int {
	return m.config.Syncthing.GUIPort
}

func (m *Manager) Start(ctx context.Context) error {
	if m.IsRunning(ctx) {
		return nil
	}

	configDir := filepath.Join(m.config.DataDir, "syncthing")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	args := []string{
		"--home=" + configDir,
		"--no-browser",
		"--no-default-folder",
		fmt.Sprintf("--gui-address=127.0.0.1:%d", m.config.Syncthing.GUIPort),
	}

	m.process = exec.Command(m.config.SyncthingPath, args...)
	m.process.Stdout = nil
	m.process.Stderr = nil

	if err := m.process.Start(); err != nil {
		return fmt.Errorf("start syncthing: %w", err)
	}

	if err := m.waitForReady(ctx); err != nil {
		_ = m.Stop()
		return fmt.Errorf("syncthing not ready: %w", err)
	}

	return nil
}

func (m *Manager) Stop() error {
	if m.process == nil || m.process.Process == nil {
		return nil
	}

	if err := m.process.Process.Signal(os.Interrupt); err != nil {
		return m.process.Process.Kill()
	}

	done := make(chan error, 1)
	go func() {
		done <- m.process.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		return m.process.Process.Kill()
	}
}

func (m *Manager) IsRunning(ctx context.Context) bool {
	return m.client.IsRunning(ctx)
}

func (m *Manager) waitForReady(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if m.client.IsRunning(ctx) {
				return nil
			}
		}
	}
}

func (m *Manager) GetDeviceID(ctx context.Context) (string, error) {
	return m.client.GetDeviceID(ctx)
}

func (m *Manager) ensureRelayClient() error {
	if m.relayClient != nil {
		return nil
	}

	urls := m.config.RelayURLs
	if len(urls) == 0 {
		urls = syncthing.ProductionRelayURLs
	}

	client, err := syncthing.NewRelayClient(urls)
	if err != nil {
		return err
	}

	m.relayClient = client
	return nil
}

func (m *Manager) CreatePairingSession(ctx context.Context) (*PairingSession, error) {
	if err := m.ensureRelayClient(); err != nil {
		return nil, fmt.Errorf("relay client: %w", err)
	}

	deviceID, err := m.client.GetDeviceID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get device ID: %w", err)
	}

	resp, err := m.relayClient.CreateSession(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &PairingSession{
		Code:      resp.Code,
		DeviceID:  deviceID,
		ExpiresIn: time.Duration(resp.ExpiresIn) * time.Second,
	}, nil
}

func (m *Manager) WaitForPeer(ctx context.Context, code string) (string, error) {
	if err := m.ensureRelayClient(); err != nil {
		return "", fmt.Errorf("relay client: %w", err)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			resp, err := m.relayClient.GetResponse(ctx, code)
			if err != nil {
				m.logger.Debug("poll for peer: %v", err)
				continue
			}
			if resp != nil && resp.Ready && resp.DeviceID != "" {
				return resp.DeviceID, nil
			}
		}
	}
}

func (m *Manager) JoinPairingSession(ctx context.Context, code string) (string, error) {
	if err := m.ensureRelayClient(); err != nil {
		return "", fmt.Errorf("relay client: %w", err)
	}

	session, err := m.relayClient.GetSession(ctx, code)
	if err != nil {
		return "", fmt.Errorf("get session: %w", err)
	}

	deviceID, err := m.client.GetDeviceID(ctx)
	if err != nil {
		return "", fmt.Errorf("get device ID: %w", err)
	}

	if err := m.relayClient.SubmitResponse(ctx, code, deviceID); err != nil {
		return "", fmt.Errorf("submit response: %w", err)
	}

	return session.DeviceID, nil
}

func (m *Manager) AddPeer(ctx context.Context, deviceID string) error {
	if err := m.client.AddDevice(ctx, deviceID, ""); err != nil {
		return fmt.Errorf("add device: %w", err)
	}

	if err := m.client.ShareFoldersWithDevice(ctx, deviceID); err != nil {
		return fmt.Errorf("share folders: %w", err)
	}

	return nil
}

func (m *Manager) ShareFoldersWithAllDevices(ctx context.Context) error {
	devices, err := m.client.GetConfiguredDevices(ctx)
	if err != nil {
		return fmt.Errorf("get devices: %w", err)
	}

	for _, d := range devices {
		if err := m.client.ShareFoldersWithDevice(ctx, d.ID); err != nil {
			return fmt.Errorf("share with %s: %w", d.ID, err)
		}
	}

	return nil
}

func (m *Manager) GetStatus(ctx context.Context) (*Status, error) {
	if !m.IsRunning(ctx) {
		return &Status{Running: false}, nil
	}

	deviceID, err := m.client.GetDeviceID(ctx)
	if err != nil {
		return nil, err
	}

	devices, err := m.client.GetConfiguredDevices(ctx)
	if err != nil {
		return nil, err
	}

	conns, err := m.client.GetConnections(ctx)
	if err != nil {
		return nil, err
	}

	progress, err := m.client.GetSyncProgress(ctx)
	if err != nil {
		return nil, err
	}

	status := &Status{
		Running:   true,
		DeviceID:  deviceID,
		Syncing:   progress.Percent < 100,
		Progress:  progress.Percent,
		NeedBytes: progress.NeedBytes,
	}

	for _, d := range devices {
		conn, connected := conns[d.ID]
		status.Peers = append(status.Peers, PeerStatus{
			ID:        d.ID,
			Name:      d.Name,
			Connected: connected && conn.Connected,
		})
		if connected && conn.Connected {
			status.ConnectedPeers++
		}
	}

	return status, nil
}

type PairingSession struct {
	Code      string
	DeviceID  string
	ExpiresIn time.Duration
}

type Status struct {
	Running        bool
	DeviceID       string
	Syncing        bool
	Progress       int
	NeedBytes      int64
	ConnectedPeers int
	Peers          []PeerStatus
}

type PeerStatus struct {
	ID        string
	Name      string
	Connected bool
}
