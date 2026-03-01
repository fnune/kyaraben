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

type Collection struct {
	fs       vfs.FS
	paths    *paths.Paths
	path     string
	resolved string
}

func NewCollection(fs vfs.FS, p *paths.Paths, path string) (*Collection, error) {
	resolved, err := expandPath(path)
	if err != nil {
		return nil, err
	}
	return &Collection{fs: fs, paths: p, path: path, resolved: resolved}, nil
}

func NewDefaultCollection(path string) (*Collection, error) {
	return NewCollection(vfs.OSFS, paths.DefaultPaths(), path)
}

func (s *Collection) Path() string {
	return s.path
}

func (s *Collection) Root() string {
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

func (s *Collection) Directories() []string {
	return []string{"roms", "bios", "saves", "states", "screenshots"}
}

func (s *Collection) RomsDir() string        { return filepath.Join(s.resolved, "roms") }
func (s *Collection) BiosDir() string        { return filepath.Join(s.resolved, "bios") }
func (s *Collection) SavesDir() string       { return filepath.Join(s.resolved, "saves") }
func (s *Collection) StatesDir() string      { return filepath.Join(s.resolved, "states") }
func (s *Collection) ScreenshotsDir() string { return filepath.Join(s.resolved, "screenshots") }

func (s *Collection) SystemRomsDir(sys model.SystemID) string {
	return filepath.Join(s.RomsDir(), string(sys))
}
func (s *Collection) SystemBiosDir(sys model.SystemID) string {
	return filepath.Join(s.BiosDir(), string(sys))
}
func (s *Collection) SystemSavesDir(sys model.SystemID) string {
	return filepath.Join(s.SavesDir(), string(sys))
}
func (s *Collection) EmulatorSavesDir(emu model.EmulatorID) string {
	return filepath.Join(s.SavesDir(), string(emu))
}
func (s *Collection) EmulatorStatesDir(emu model.EmulatorID) string {
	return filepath.Join(s.StatesDir(), string(emu))
}
func (s *Collection) EmulatorScreenshotsDir(emu model.EmulatorID) string {
	name := string(emu)
	if strings.HasPrefix(name, "retroarch:") {
		name = "retroarch"
	}
	return filepath.Join(s.ScreenshotsDir(), name)
}

func (s *Collection) CoresDir() string {
	dir, err := s.paths.CoresDir()
	if err != nil {
		return ""
	}
	return dir
}

func (s *Collection) FrontendGamelistDir(fe model.FrontendID, sys model.SystemID) string {
	return filepath.Join(s.resolved, "frontends", string(fe), "gamelists", string(sys))
}

func (s *Collection) FrontendMediaDir(fe model.FrontendID, sys model.SystemID) string {
	return filepath.Join(s.resolved, "frontends", string(fe), "media", string(sys))
}

func (s *Collection) FrontendMediaBaseDir(fe model.FrontendID) string {
	return filepath.Join(s.resolved, "frontends", string(fe), "media")
}

func (s *Collection) Initialize() error {
	for _, dir := range s.Directories() {
		path := filepath.Join(s.resolved, dir)
		if err := vfs.MkdirAll(s.fs, path, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	return nil
}

func (s *Collection) InitializeForEmulator(sys model.SystemID, emu model.EmulatorID, pathUsage model.PathUsage) error {
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

func (s *Collection) Exists() bool {
	info, err := s.fs.Stat(s.resolved)
	return err == nil && info.IsDir()
}

func (s *Collection) IsInitialized() bool {
	for _, dir := range s.Directories() {
		path := filepath.Join(s.resolved, dir)
		info, err := s.fs.Stat(path)
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}
