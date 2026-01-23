package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type UserStore struct {
	path     string
	resolved string
}

func NewUserStore(path string) (*UserStore, error) {
	resolved, err := expandPath(path)
	if err != nil {
		return nil, err
	}
	return &UserStore{path: path, resolved: resolved}, nil
}

func (s *UserStore) Path() string {
	return s.path
}

func (s *UserStore) Root() string {
	return s.resolved
}

func expandPath(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	return path, nil
}

func (s *UserStore) Directories() []string {
	return []string{"roms", "bios", "saves", "states", "screenshots"}
}

func (s *UserStore) RomsDir() string        { return filepath.Join(s.resolved, "roms") }
func (s *UserStore) BiosDir() string        { return filepath.Join(s.resolved, "bios") }
func (s *UserStore) SavesDir() string       { return filepath.Join(s.resolved, "saves") }
func (s *UserStore) StatesDir() string      { return filepath.Join(s.resolved, "states") }
func (s *UserStore) ScreenshotsDir() string { return filepath.Join(s.resolved, "screenshots") }

func (s *UserStore) SystemRomsDir(sys model.SystemID) string {
	return filepath.Join(s.RomsDir(), string(sys))
}
func (s *UserStore) SystemBiosDir(sys model.SystemID) string {
	return filepath.Join(s.BiosDir(), string(sys))
}
func (s *UserStore) SystemSavesDir(sys model.SystemID) string {
	return filepath.Join(s.SavesDir(), string(sys))
}
func (s *UserStore) EmulatorStatesDir(emu model.EmulatorID) string {
	return filepath.Join(s.StatesDir(), string(emu))
}
func (s *UserStore) SystemScreenshotsDir(sys model.SystemID) string {
	return filepath.Join(s.ScreenshotsDir(), string(sys))
}

func (s *UserStore) Initialize() error {
	for _, dir := range s.Directories() {
		path := filepath.Join(s.resolved, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	return nil
}

func (s *UserStore) InitializeSystem(sys model.SystemID) error {
	dirs := []string{
		s.SystemRomsDir(sys),
		s.SystemBiosDir(sys),
		s.SystemSavesDir(sys),
		s.SystemScreenshotsDir(sys),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	return nil
}

func (s *UserStore) InitializeEmulator(emu model.EmulatorID) error {
	dir := s.EmulatorStatesDir(emu)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}
	return nil
}

func (s *UserStore) Exists() bool {
	info, err := os.Stat(s.resolved)
	return err == nil && info.IsDir()
}

func (s *UserStore) IsInitialized() bool {
	for _, dir := range s.Directories() {
		path := filepath.Join(s.resolved, dir)
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}
