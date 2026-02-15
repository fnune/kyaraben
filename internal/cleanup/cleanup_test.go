package cleanup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestCollectConfigDirs(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	mgbaDir := filepath.Join(configDir, "mgba")
	duckstationDir := filepath.Join(configDir, "duckstation")
	if err := os.MkdirAll(mgbaDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(duckstationDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", configDir)

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

	dirs := CollectConfigDirs(configs)

	if len(dirs) != 2 {
		t.Errorf("Expected 2 dirs, got %d", len(dirs))
	}

	found := make(map[string]bool)
	for _, d := range dirs {
		found[d] = true
	}

	if !found[mgbaDir] {
		t.Errorf("Expected %s in results", mgbaDir)
	}
	if !found[duckstationDir] {
		t.Errorf("Expected %s in results", duckstationDir)
	}
}

func TestCollectConfigDirsDeduplicates(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	retroarchDir := filepath.Join(configDir, "retroarch")
	if err := os.MkdirAll(retroarchDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", configDir)

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

	dirs := CollectConfigDirs(configs)

	if len(dirs) != 1 {
		t.Errorf("Expected 1 dir (deduplicated), got %d: %v", len(dirs), dirs)
	}
}

func TestCollectConfigDirsSkipsNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", configDir)

	configs := []model.ManagedConfig{
		{
			EmulatorIDs: []model.EmulatorID{model.EmulatorIDMGBA},
			Target: model.ConfigTarget{
				RelPath: "nonexistent/config.ini",
				BaseDir: model.ConfigBaseDirUserConfig,
			},
		},
	}

	dirs := CollectConfigDirs(configs)

	if len(dirs) != 0 {
		t.Errorf("Expected 0 dirs for nonexistent path, got %d", len(dirs))
	}
}

func TestRemoveConfigDirs(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	mgbaDir := filepath.Join(configDir, "mgba")
	if err := os.MkdirAll(mgbaDir, 0755); err != nil {
		t.Fatal(err)
	}
	configFile := filepath.Join(mgbaDir, "config.ini")
	if err := os.WriteFile(configFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", configDir)

	configs := []model.ManagedConfig{
		{
			EmulatorIDs: []model.EmulatorID{model.EmulatorIDMGBA},
			Target: model.ConfigTarget{
				RelPath: "mgba/config.ini",
				BaseDir: model.ConfigBaseDirUserConfig,
			},
		},
	}

	removed := RemoveConfigDirs(configs)

	if len(removed) != 1 {
		t.Errorf("Expected 1 removed dir, got %d", len(removed))
	}

	if _, err := os.Stat(mgbaDir); !os.IsNotExist(err) {
		t.Error("mgba directory should have been removed")
	}
}

func TestRemoveConfigDirsHandlesReadOnlyFiles(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	emuDir := filepath.Join(configDir, "emulator")
	if err := os.MkdirAll(emuDir, 0755); err != nil {
		t.Fatal(err)
	}
	readOnlyFile := filepath.Join(emuDir, "readonly.cfg")
	if err := os.WriteFile(readOnlyFile, []byte("test"), 0444); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", configDir)

	configs := []model.ManagedConfig{
		{
			EmulatorIDs: []model.EmulatorID{"test-emu"},
			Target: model.ConfigTarget{
				RelPath: "emulator/config.ini",
				BaseDir: model.ConfigBaseDirUserConfig,
			},
		},
	}

	removed := RemoveConfigDirs(configs)

	if len(removed) != 1 {
		t.Errorf("Expected 1 removed dir, got %d", len(removed))
	}

	if _, err := os.Stat(emuDir); !os.IsNotExist(err) {
		t.Error("emulator directory should have been removed despite read-only file")
	}
}
