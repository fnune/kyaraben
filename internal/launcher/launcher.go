package launcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/logging"
)

var log = logging.New("launcher")

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
	log.Info("Linking profile: %s -> %s", m.CurrentLink(), storePath)

	if err := os.MkdirAll(m.profileDir, 0755); err != nil {
		return fmt.Errorf("creating profile directory: %w", err)
	}

	currentLink := m.CurrentLink()

	if _, err := os.Lstat(currentLink); err == nil {
		log.Debug("Removing old symlink: %s", currentLink)
		if err := os.Remove(currentLink); err != nil {
			return fmt.Errorf("removing old symlink: %w", err)
		}
	}

	if err := os.Symlink(storePath, currentLink); err != nil {
		return fmt.Errorf("creating symlink: %w", err)
	}

	if target, err := os.Readlink(currentLink); err == nil {
		log.Info("Created symlink: %s -> %s", currentLink, target)
		if entries, err := os.ReadDir(filepath.Join(target, "bin")); err == nil {
			binaries := make([]string, 0, len(entries))
			for _, e := range entries {
				binaries = append(binaries, e.Name())
			}
			log.Info("Available binaries: %v", binaries)
		}
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
