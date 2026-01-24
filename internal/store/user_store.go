package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

// UserStore manages the user's emulation directory structure.
type UserStore struct {
	Root string
}

// NewUserStore creates a new UserStore at the given path.
func NewUserStore(root string) *UserStore {
	return &UserStore{Root: root}
}

// Directories returns the standard subdirectory names.
func (s *UserStore) Directories() []string {
	return []string{"roms", "bios", "saves", "states", "screenshots"}
}

// RomsDir returns the path to the roms directory.
func (s *UserStore) RomsDir() string {
	return filepath.Join(s.Root, "roms")
}

// BiosDir returns the path to the bios directory.
func (s *UserStore) BiosDir() string {
	return filepath.Join(s.Root, "bios")
}

// SavesDir returns the path to the saves directory.
func (s *UserStore) SavesDir() string {
	return filepath.Join(s.Root, "saves")
}

// StatesDir returns the path to the states directory.
func (s *UserStore) StatesDir() string {
	return filepath.Join(s.Root, "states")
}

// ScreenshotsDir returns the path to the screenshots directory.
func (s *UserStore) ScreenshotsDir() string {
	return filepath.Join(s.Root, "screenshots")
}

// SystemRomsDir returns the path to a system's roms directory.
func (s *UserStore) SystemRomsDir(sys model.SystemID) string {
	return filepath.Join(s.RomsDir(), string(sys))
}

// SystemBiosDir returns the path to a system's bios directory.
func (s *UserStore) SystemBiosDir(sys model.SystemID) string {
	return filepath.Join(s.BiosDir(), string(sys))
}

// SystemSavesDir returns the path to a system's saves directory.
func (s *UserStore) SystemSavesDir(sys model.SystemID) string {
	return filepath.Join(s.SavesDir(), string(sys))
}

// SystemStatesDir returns the path to a system's states directory.
func (s *UserStore) SystemStatesDir(sys model.SystemID) string {
	return filepath.Join(s.StatesDir(), string(sys))
}

// SystemScreenshotsDir returns the path to a system's screenshots directory.
func (s *UserStore) SystemScreenshotsDir(sys model.SystemID) string {
	return filepath.Join(s.ScreenshotsDir(), string(sys))
}

// Initialize creates the base directory structure.
func (s *UserStore) Initialize() error {
	for _, dir := range s.Directories() {
		path := filepath.Join(s.Root, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	return nil
}

// InitializeSystem creates the directory structure for a specific system.
func (s *UserStore) InitializeSystem(sys model.SystemID) error {
	dirs := []string{
		s.SystemRomsDir(sys),
		s.SystemBiosDir(sys),
		s.SystemSavesDir(sys),
		s.SystemStatesDir(sys),
		s.SystemScreenshotsDir(sys),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	return nil
}

// Exists checks if the user store root exists.
func (s *UserStore) Exists() bool {
	info, err := os.Stat(s.Root)
	return err == nil && info.IsDir()
}

// IsInitialized checks if the base directory structure exists.
func (s *UserStore) IsInitialized() bool {
	for _, dir := range s.Directories() {
		path := filepath.Join(s.Root, dir)
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}
