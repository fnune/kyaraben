package service

import (
	"context"
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

func NewManager(cfg Config) *Manager {
	stConfig := syncthing.DefaultConfig()
	stConfig.GUIPort = cfg.GUIPort
	stConfig.BaseURL = fmt.Sprintf("http://localhost:%d", cfg.GUIPort)

	return &Manager{
		process:   NewProcessManager(cfg.DataDir),
		autostart: NewAutostartManager(cfg.UserdataPath, cfg.Platform, cfg.PakPath, cfg.LogsPath),
		client:    syncthing.NewClient(stConfig),
		config:    cfg,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	if pid, running := m.process.GetRunningPID(); running {
		if m.client.IsRunning(ctx) {
			return nil
		}
		_ = m.process.StopProcess()
		_ = m.process.RemovePID()
		_ = pid
	}

	configDir := m.process.HomePath()
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

	cmd := exec.Command(m.config.SyncthingPath,
		"--home="+configDir,
		"--no-browser",
		"--no-default-folder",
		fmt.Sprintf("--gui-address=127.0.0.1:%d", m.config.GUIPort),
	)
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Start(); err != nil {
		_ = out.Close()
		return fmt.Errorf("start syncthing: %w", err)
	}
	_ = out.Close()

	if err := m.process.WritePID(cmd.Process.Pid); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("write pid: %w", err)
	}

	if err := m.waitForReady(ctx); err != nil {
		_ = m.Stop()
		return fmt.Errorf("syncthing not ready: %w", err)
	}

	return nil
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
