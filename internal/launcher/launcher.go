package launcher

import (
	"fmt"
	"os"
	"path/filepath"
)

type Manager struct {
	profileDir string
}

func NewManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	return &Manager{
		profileDir: filepath.Join(homeDir, ".local", "share", "kyaraben"),
	}
}

func (m *Manager) ProfileDir() string {
	return m.profileDir
}

func (m *Manager) CurrentLink() string {
	return filepath.Join(m.profileDir, "current")
}

func (m *Manager) Link(storePath string) error {
	if err := os.MkdirAll(m.profileDir, 0755); err != nil {
		return fmt.Errorf("creating profile directory: %w", err)
	}

	currentLink := m.CurrentLink()

	if _, err := os.Lstat(currentLink); err == nil {
		if err := os.Remove(currentLink); err != nil {
			return fmt.Errorf("removing old symlink: %w", err)
		}
	}

	if err := os.Symlink(storePath, currentLink); err != nil {
		return fmt.Errorf("creating symlink: %w", err)
	}

	return nil
}

func (m *Manager) Unlink() error {
	currentLink := m.CurrentLink()
	if _, err := os.Lstat(currentLink); err == nil {
		return os.Remove(currentLink)
	}
	return nil
}
