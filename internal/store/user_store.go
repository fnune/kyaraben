package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
)

type UserStore struct {
	fs       vfs.FS
	paths    *paths.Paths
	path     string
	resolved string
}

func NewUserStore(fs vfs.FS, p *paths.Paths, path string) (*UserStore, error) {
	resolved, err := expandPath(path)
	if err != nil {
		return nil, err
	}
	return &UserStore{fs: fs, paths: p, path: path, resolved: resolved}, nil
}

func NewDefaultUserStore(path string) (*UserStore, error) {
	return NewUserStore(vfs.OSFS, paths.DefaultPaths(), path)
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
func (s *UserStore) EmulatorSavesDir(emu model.EmulatorID) string {
	return filepath.Join(s.SavesDir(), string(emu))
}
func (s *UserStore) EmulatorStatesDir(emu model.EmulatorID) string {
	return filepath.Join(s.StatesDir(), string(emu))
}
func (s *UserStore) EmulatorScreenshotsDir(emu model.EmulatorID) string {
	name := string(emu)
	if strings.HasPrefix(name, "retroarch:") {
		name = "retroarch"
	}
	return filepath.Join(s.ScreenshotsDir(), name)
}

func (s *UserStore) CoresDir() string {
	dir, err := s.paths.CoresDir()
	if err != nil {
		return ""
	}
	return dir
}

func (s *UserStore) Initialize() error {
	for _, dir := range s.Directories() {
		path := filepath.Join(s.resolved, dir)
		if err := vfs.MkdirAll(s.fs, path, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	return nil
}

func (s *UserStore) InitializeForEmulator(sys model.SystemID, emu model.EmulatorID, pathUsage model.PathUsage) error {
	dirs := []string{s.SystemRomsDir(sys)}

	if pathUsage.UsesBiosDir {
		dirs = append(dirs, s.SystemBiosDir(sys))
	}
	if pathUsage.UsesSavesDir {
		dirs = append(dirs, s.SystemSavesDir(sys))
	}
	if pathUsage.UsesStatesDir {
		dirs = append(dirs, s.EmulatorStatesDir(emu))
	}
	if pathUsage.UsesScreenshotsDir {
		dirs = append(dirs, s.EmulatorScreenshotsDir(emu))
	}

	for _, dir := range dirs {
		if err := vfs.MkdirAll(s.fs, dir, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	return nil
}

func (s *UserStore) Exists() bool {
	info, err := s.fs.Stat(s.resolved)
	return err == nil && info.IsDir()
}

func (s *UserStore) IsInitialized() bool {
	for _, dir := range s.Directories() {
		path := filepath.Join(s.resolved, dir)
		info, err := s.fs.Stat(path)
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}
