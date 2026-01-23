package sync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
)

type Process struct {
	cmd        *exec.Cmd
	config     model.SyncConfig
	binaryPath string
	configDir  string
	dataDir    string
	apiKey     string
}

func NewProcess(config model.SyncConfig, binaryPath string) (*Process, error) {
	configDir, err := SyncthingConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting syncthing config dir: %w", err)
	}

	dataDir, err := SyncthingDataDir()
	if err != nil {
		return nil, fmt.Errorf("getting syncthing data dir: %w", err)
	}

	return &Process{
		config:     config,
		binaryPath: binaryPath,
		configDir:  configDir,
		dataDir:    dataDir,
	}, nil
}

func SyncthingConfigDir() (string, error) {
	configDir, err := paths.KyarabenConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "syncthing"), nil
}

func SyncthingDataDir() (string, error) {
	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, "syncthing"), nil
}

func (p *Process) EnsureDirectories() error {
	if err := os.MkdirAll(p.configDir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	if err := os.MkdirAll(p.dataDir, 0700); err != nil {
		return fmt.Errorf("creating data dir: %w", err)
	}
	return nil
}

func (p *Process) Start(ctx context.Context) error {
	if err := p.EnsureDirectories(); err != nil {
		return err
	}

	apiKey, err := p.ensureAPIKey()
	if err != nil {
		return fmt.Errorf("ensuring API key: %w", err)
	}
	p.apiKey = apiKey

	args := []string{
		"serve",
		"--home", p.configDir,
		"--no-browser",
		"--gui-address", fmt.Sprintf("127.0.0.1:%d", p.config.Syncthing.GUIPort),
		"--gui-apikey", p.apiKey,
	}

	p.cmd = exec.CommandContext(ctx, p.binaryPath, args...)
	p.cmd.Env = append(os.Environ(),
		fmt.Sprintf("STNODEFAULTFOLDER=1"),
		fmt.Sprintf("STNOUPGRADE=1"),
	)

	log.Info("Starting syncthing: %s %v", p.binaryPath, args)

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("starting syncthing: %w", err)
	}

	if err := p.waitForReady(ctx); err != nil {
		p.Stop()
		return fmt.Errorf("waiting for syncthing: %w", err)
	}

	log.Info("Syncthing started on port %d", p.config.Syncthing.GUIPort)
	return nil
}

func (p *Process) waitForReady(ctx context.Context) error {
	client := NewClient(p.config)
	client.SetAPIKey(p.apiKey)

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if client.IsRunning(ctx) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("syncthing did not become ready within 30 seconds")
}

func (p *Process) Stop() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	log.Info("Stopping syncthing")

	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("sending SIGTERM: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := p.cmd.Process.Wait()
		done <- err
	}()

	select {
	case <-done:
		log.Info("Syncthing stopped")
		return nil
	case <-time.After(10 * time.Second):
		log.Info("Syncthing did not stop gracefully, killing")
		return p.cmd.Process.Kill()
	}
}

func (p *Process) APIKey() string {
	return p.apiKey
}

func (p *Process) ensureAPIKey() (string, error) {
	keyFile := filepath.Join(p.configDir, "api-key")

	data, err := os.ReadFile(keyFile)
	if err == nil && len(data) > 0 {
		return string(data), nil
	}

	key := generateAPIKey()
	if err := os.WriteFile(keyFile, []byte(key), 0600); err != nil {
		return "", fmt.Errorf("writing api key: %w", err)
	}

	return key, nil
}

func generateAPIKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
