package cleanup

import (
	"path/filepath"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
)

type fakeResolver struct {
	configDir string
	homeDir   string
	dataDir   string
}

func (f fakeResolver) UserConfigDir() (string, error) { return f.configDir, nil }
func (f fakeResolver) UserHomeDir() (string, error)   { return f.homeDir, nil }
func (f fakeResolver) UserDataDir() (string, error) {
	if f.dataDir != "" {
		return f.dataDir, nil
	}
	return filepath.Join(f.homeDir, ".local", "share"), nil
}

func TestCollectConfigDirs(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/mgba":        &vfst.Dir{Perm: 0755},
		"/config/duckstation": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config", homeDir: "/home/user"}

	configs := []model.ManagedConfig{
		{
			EmulatorIDs: []model.EmulatorID{model.EmulatorIDMGBA},
			Target: model.ConfigTarget{
				RelPath: "mgba/config.ini",
				BaseDir: model.ConfigBaseDirUserConfig,
			},
		},
		{
			EmulatorIDs: []model.EmulatorID{model.EmulatorIDDuckStation},
			Target: model.ConfigTarget{
				RelPath: "duckstation/settings.ini",
				BaseDir: model.ConfigBaseDirUserConfig,
			},
		},
	}

	cleaner := New(fs, resolver)
	dirs := cleaner.CollectConfigDirs(configs)

	if len(dirs) != 2 {
		t.Errorf("Expected 2 dirs, got %d", len(dirs))
	}
}

func TestCollectConfigDirsDeduplicates(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/retroarch": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config", homeDir: "/home/user"}

	configs := []model.ManagedConfig{
		{
			EmulatorIDs: []model.EmulatorID{"retroarch:bsnes"},
			Target: model.ConfigTarget{
				RelPath: "retroarch/config/bsnes/bsnes.cfg",
				BaseDir: model.ConfigBaseDirUserConfig,
			},
		},
		{
			EmulatorIDs: []model.EmulatorID{"retroarch:mesen"},
			Target: model.ConfigTarget{
				RelPath: "retroarch/config/mesen/mesen.cfg",
				BaseDir: model.ConfigBaseDirUserConfig,
			},
		},
	}

	cleaner := New(fs, resolver)
	dirs := cleaner.CollectConfigDirs(configs)

	if len(dirs) != 1 {
		t.Errorf("Expected 1 dir (deduplicated), got %d: %v", len(dirs), dirs)
	}
}

func TestCollectConfigDirsSkipsNonexistent(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config", homeDir: "/home/user"}

	configs := []model.ManagedConfig{
		{
			EmulatorIDs: []model.EmulatorID{model.EmulatorIDMGBA},
			Target: model.ConfigTarget{
				RelPath: "nonexistent/config.ini",
				BaseDir: model.ConfigBaseDirUserConfig,
			},
		},
	}

	cleaner := New(fs, resolver)
	dirs := cleaner.CollectConfigDirs(configs)

	if len(dirs) != 0 {
		t.Errorf("Expected 0 dirs for nonexistent path, got %d", len(dirs))
	}
}

func TestRemoveConfigDirs(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/mgba/config.ini": "test content",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config", homeDir: "/home/user"}

	configs := []model.ManagedConfig{
		{
			EmulatorIDs: []model.EmulatorID{model.EmulatorIDMGBA},
			Target: model.ConfigTarget{
				RelPath: "mgba/config.ini",
				BaseDir: model.ConfigBaseDirUserConfig,
			},
		},
	}

	cleaner := New(fs, resolver)
	removed := cleaner.RemoveConfigDirs(configs)

	if len(removed) != 1 {
		t.Errorf("Expected 1 removed dir, got %d", len(removed))
	}

	vfst.RunTests(t, fs, "",
		vfst.TestPath("/config/mgba",
			vfst.TestDoesNotExist(),
		),
	)
}

func TestRemoveConfigDirsHandlesReadOnlyFiles(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/emulator/readonly.cfg": &vfst.File{
			Perm:     0444,
			Contents: []byte("test"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config", homeDir: "/home/user"}

	configs := []model.ManagedConfig{
		{
			EmulatorIDs: []model.EmulatorID{"test-emu"},
			Target: model.ConfigTarget{
				RelPath: "emulator/config.ini",
				BaseDir: model.ConfigBaseDirUserConfig,
			},
		},
	}

	cleaner := New(fs, resolver)
	removed := cleaner.RemoveConfigDirs(configs)

	if len(removed) != 1 {
		t.Errorf("Expected 1 removed dir, got %d", len(removed))
	}

	vfst.RunTests(t, fs, "",
		vfst.TestPath("/config/emulator",
			vfst.TestDoesNotExist(),
		),
	)
}
