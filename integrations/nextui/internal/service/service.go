package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fnune/kyaraben/internal/syncthing"
)

type Manager struct {
	process   *ProcessManager
	autostart *AutostartManager
	client    syncthing.SyncClient
	config    Config
}

type Config struct {
	DataDir       string
	PakPath       string
	UserdataPath  string
	Platform      string
	LogsPath      string
	SyncthingPath string
	GUIPort       int
}

func NewManager(cfg Config, process *ProcessManager, autostart *AutostartManager, client syncthing.SyncClient) *Manager {
	return &Manager{
		process:   process,
		autostart: autostart,
		client:    client,
		config:    cfg,
	}
}

func NewDefaultManager(cfg Config) *Manager {
	stConfig := syncthing.DefaultConfig()
	stConfig.GUIPort = cfg.GUIPort
	stConfig.BaseURL = fmt.Sprintf("http://localhost:%d", cfg.GUIPort)

	return NewManager(
		cfg,
		NewProcessManager(cfg.DataDir),
		NewAutostartManager(cfg.UserdataPath, cfg.Platform, cfg.PakPath, cfg.LogsPath),
		syncthing.NewClient(stConfig),
	)
}

func (m *Manager) Start(ctx context.Context) error {
	fmt.Fprintf(os.Stderr, "service.Start: checking existing process\n")
	if pid, running := m.process.GetRunningPID(); running {
		fmt.Fprintf(os.Stderr, "service.Start: found PID %d, loading API key\n", pid)
		if apiKey, err := m.loadAPIKey(); err == nil && apiKey != "" {
			m.client.SetAPIKey(apiKey)
			fmt.Fprintf(os.Stderr, "service.Start: checking if responsive\n")
			if m.client.IsRunning(ctx) {
				fmt.Fprintf(os.Stderr, "service.Start: already running\n")
				return nil
			}
		}
		fmt.Fprintf(os.Stderr, "service.Start: not responsive, stopping\n")
		_ = m.process.StopProcess()
		_ = m.process.RemovePID()
	}

	configDir := m.process.HomePath()
	fmt.Fprintf(os.Stderr, "service.Start: creating config dir %s\n", configDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	logFile := filepath.Join(m.config.LogsPath, "kyaraben-syncthing.log")
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		return fmt.Errorf("create logs dir: %w", err)
	}

	out, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "service.Start: launching syncthing at %s\n", m.config.SyncthingPath)
	cmd := exec.Command(m.config.SyncthingPath,
		"--home="+configDir,
		"--no-browser",
		"--no-upgrade",
		fmt.Sprintf("--gui-address=0.0.0.0:%d", m.config.GUIPort),
	)
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Start(); err != nil {
		_ = out.Close()
		return fmt.Errorf("start syncthing: %w", err)
	}
	_ = out.Close()
	fmt.Fprintf(os.Stderr, "service.Start: syncthing launched with PID %d\n", cmd.Process.Pid)

	if err := m.process.WritePID(cmd.Process.Pid); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("write pid: %w", err)
	}

	fmt.Fprintf(os.Stderr, "service.Start: waiting for API key\n")
	if err := m.waitForAPIKey(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "service.Start: waitForAPIKey failed: %v\n", err)
		_ = m.Stop()
		return fmt.Errorf("syncthing config not ready: %w", err)
	}

	fmt.Fprintf(os.Stderr, "service.Start: waiting for syncthing to be ready\n")
	if err := m.waitForReady(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "service.Start: waitForReady failed: %v\n", err)
		_ = m.Stop()
		return fmt.Errorf("syncthing not ready: %w", err)
	}
	fmt.Fprintf(os.Stderr, "service.Start: syncthing ready\n")

	if err := m.setDeviceName(ctx, friendlyDeviceName(m.config.Platform)); err != nil {
		fmt.Fprintf(os.Stderr, "service.Start: failed to set device name: %v\n", err)
	}

	if err := m.client.DisableUsageReporting(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "service.Start: failed to disable usage reporting: %v\n", err)
	}

	if err := m.client.AllowInsecureAdmin(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "service.Start: failed to allow insecure admin: %v\n", err)
	}

	return nil
}

func (m *Manager) setDeviceName(ctx context.Context, name string) error {
	deviceID, err := m.client.GetDeviceID(ctx)
	if err != nil {
		return err
	}
	return m.client.AddDeviceWithAddresses(ctx, deviceID, name, []string{"dynamic"})
}

var platformNames = map[string]string{
	"tg5040":      "trimui-brick",
	"trimuismart": "trimui-smart-pro",
	"rg35xxplus":  "rg35xx-plus",
	"rg35xx":      "rg35xx",
	"miyoomini":   "miyoo-mini",
	"rgb30":       "rgb30",
}

func friendlyDeviceName(platform string) string {
	if name, ok := platformNames[platform]; ok {
		return name
	}
	return platform
}

func (m *Manager) Stop() error {
	return m.process.StopProcess()
}

func (m *Manager) IsRunning(ctx context.Context) bool {
	_, hasPID := m.process.GetRunningPID()
	if !hasPID {
		return false
	}
	return m.client.IsRunning(ctx)
}

func (m *Manager) waitForAPIKey(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			apiKey, err := m.loadAPIKey()
			if err == nil && apiKey != "" {
				fmt.Fprintf(os.Stderr, "service.Start: got API key\n")
				m.client.SetAPIKey(apiKey)
				return nil
			}
		}
	}
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

func (m *Manager) EnableAutostart() error {
	return m.autostart.Enable()
}

func (m *Manager) DisableAutostart() error {
	return m.autostart.Disable()
}

func (m *Manager) IsAutostartEnabled() bool {
	return m.autostart.IsEnabled()
}

func (m *Manager) Client() syncthing.SyncClient {
	return m.client
}

func (m *Manager) ProcessManager() *ProcessManager {
	return m.process
}

func (m *Manager) loadAPIKey() (string, error) {
	configPath := filepath.Join(m.process.HomePath(), "config.xml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("read config: %w", err)
	}

	var config struct {
		GUI struct {
			APIKey string `xml:"apikey"`
		} `xml:"gui"`
	}
	if err := xml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("parse config: %w", err)
	}

	return config.GUI.APIKey, nil
}
