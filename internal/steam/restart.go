package steam

import (
	"context"
	"os/exec"
	"path/filepath"
	"time"
)

func (m *Manager) Restart(ctx context.Context) error {
	if !m.IsAvailable() {
		return nil
	}

	steamBin := m.steamExecutable()
	if steamBin == "" {
		return nil
	}

	if !m.isRunning() {
		return nil
	}

	if err := m.shutdown(ctx, steamBin); err != nil {
		return err
	}

	if err := m.waitForExit(ctx); err != nil {
		return err
	}

	return m.start(steamBin)
}

func (m *Manager) steamExecutable() string {
	candidates := []string{
		filepath.Join(m.install.RootPath, "steam.sh"),
		"/usr/bin/steam",
		"/usr/local/bin/steam",
	}

	for _, path := range candidates {
		if info, err := m.fs.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}

func (m *Manager) isRunning() bool {
	cmd := exec.Command("pgrep", "-x", "steam")
	return cmd.Run() == nil
}

func (m *Manager) shutdown(ctx context.Context, steamBin string) error {
	cmd := exec.CommandContext(ctx, steamBin, "-shutdown")
	return cmd.Run()
}

func (m *Manager) waitForExit(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return nil
		case <-ticker.C:
			if !m.isRunning() {
				return nil
			}
		}
	}
}

func (m *Manager) start(steamBin string) error {
	cmd := exec.Command(steamBin)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}
