package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type UserStore struct {
	Root string
}

func NewUserStore(root string) *UserStore {
	return &UserStore{Root: root}
}

func (s *UserStore) Directories() []string {
	return []string{"roms", "bios", "saves", "states", "screenshots"}
}

func (s *UserStore) RomsDir() string        { return filepath.Join(s.Root, "roms") }
func (s *UserStore) BiosDir() string        { return filepath.Join(s.Root, "bios") }
func (s *UserStore) SavesDir() string       { return filepath.Join(s.Root, "saves") }
func (s *UserStore) StatesDir() string      { return filepath.Join(s.Root, "states") }
func (s *UserStore) ScreenshotsDir() string { return filepath.Join(s.Root, "screenshots") }

func (s *UserStore) SystemRomsDir(sys model.SystemID) string {
	return filepath.Join(s.RomsDir(), string(sys))
}
func (s *UserStore) SystemBiosDir(sys model.SystemID) string {
	return filepath.Join(s.BiosDir(), string(sys))
}
func (s *UserStore) SystemSavesDir(sys model.SystemID) string {
	return filepath.Join(s.SavesDir(), string(sys))
}
func (s *UserStore) SystemStatesDir(sys model.SystemID) string {
	return filepath.Join(s.StatesDir(), string(sys))
}
func (s *UserStore) SystemScreenshotsDir(sys model.SystemID) string {
	return filepath.Join(s.ScreenshotsDir(), string(sys))
}

func (s *UserStore) Initialize() error {
	for _, dir := range s.Directories() {
		path := filepath.Join(s.Root, dir)
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

func (s *UserStore) Exists() bool {
	info, err := os.Stat(s.Root)
	return err == nil && info.IsDir()
}

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
