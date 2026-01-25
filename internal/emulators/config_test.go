package emulators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/emulators/duckstation"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbsnes"
	"github.com/fnune/kyaraben/internal/emulators/tic80emu"
	"github.com/fnune/kyaraben/internal/model"
)

type fakeStoreReader struct {
	root string
}

func (f *fakeStoreReader) BiosDir() string {
	return filepath.Join(f.root, "bios")
}

func (f *fakeStoreReader) SystemBiosDir(sys model.SystemID) string {
	return filepath.Join(f.root, "bios", string(sys))
}

func (f *fakeStoreReader) SystemSavesDir(sys model.SystemID) string {
	return filepath.Join(f.root, string(sys), "saves")
}

func (f *fakeStoreReader) SystemStatesDir(sys model.SystemID) string {
	return filepath.Join(f.root, string(sys), "states")
}

func (f *fakeStoreReader) SystemScreenshotsDir(sys model.SystemID) string {
	return filepath.Join(f.root, string(sys), "screenshots")
}

func (f *fakeStoreReader) SystemRomsDir(sys model.SystemID) string {
	return filepath.Join(f.root, string(sys), "roms")
}

func TestDuckStationGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := duckstation.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store, []model.SystemID{model.SystemPSX})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatINI {
		t.Errorf("expected INI format, got %s", patch.Target.Format)
	}

	if patch.Target.BaseDir != model.ConfigBaseDirUserConfig {
		t.Errorf("expected UserConfig base dir, got %s", patch.Target.BaseDir)
	}

	if !strings.Contains(patch.Target.RelPath, "duckstation") {
		t.Errorf("expected RelPath to contain 'duckstation', got %s", patch.Target.RelPath)
	}

	path, err := patch.Target.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	configDir, _ := os.UserConfigDir()
	expectedPath := filepath.Join(configDir, "duckstation", "settings.ini")
	if path != expectedPath {
		t.Errorf("resolved path = %q, want %q", path, expectedPath)
	}

	expectedKeys := map[string]bool{
		"SearchDirectory": false,
		"Directory":       false,
		"SaveStates":      false,
		"Screenshots":     false,
		"RecursivePaths":  false,
	}

	for _, entry := range patch.Entries {
		if _, ok := expectedKeys[entry.Key()]; ok {
			expectedKeys[entry.Key()] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("expected key %q not found in entries", key)
		}
	}
}

func TestRetroArchBsnesGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := retroarchbsnes.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store, []model.SystemID{model.SystemSNES})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatCFG {
		t.Errorf("expected CFG format, got %s", patch.Target.Format)
	}

	if patch.Target.BaseDir != model.ConfigBaseDirUserConfig {
		t.Errorf("expected UserConfig base dir, got %s", patch.Target.BaseDir)
	}

	expectedKeys := []string{
		"system_directory",
		"savefile_directory",
		"savestate_directory",
		"screenshot_directory",
		"rgui_browser_directory",
	}

	foundKeys := make(map[string]bool)
	for _, entry := range patch.Entries {
		foundKeys[entry.Key()] = true
	}

	for _, key := range expectedKeys {
		if !foundKeys[key] {
			t.Errorf("expected key %q not found in entries", key)
		}
	}
}

func TestTIC80Generate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := tic80emu.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store, []model.SystemID{model.SystemTIC80})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 0 {
		t.Errorf("expected nil or empty patches for TIC-80, got %d", len(patches))
	}
}

func TestGeneratedEntriesContainStorePaths(t *testing.T) {
	store := &fakeStoreReader{root: "/test/emulation"}
	gen := duckstation.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store, []model.SystemID{model.SystemPSX})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	patch := patches[0]

	foundStorePath := false
	for _, entry := range patch.Entries {
		if strings.Contains(entry.Value, "/test/emulation") {
			foundStorePath = true
			break
		}
	}

	if !foundStorePath {
		t.Error("no entry contains the store path")
	}
}

func TestConfigTargetResolveIntegration(t *testing.T) {
	targets := []model.ConfigTarget{
		{RelPath: "test/config.ini", Format: model.ConfigFormatINI, BaseDir: model.ConfigBaseDirUserConfig},
		{RelPath: "test/config.cfg", Format: model.ConfigFormatCFG, BaseDir: model.ConfigBaseDirUserData},
		{RelPath: ".testrc", Format: model.ConfigFormatINI, BaseDir: model.ConfigBaseDirHome},
	}

	for _, target := range targets {
		path, err := target.Resolve()
		if err != nil {
			t.Errorf("Resolve() for %s failed: %v", target.RelPath, err)
			continue
		}

		if !filepath.IsAbs(path) {
			t.Errorf("Resolve() returned non-absolute path: %s", path)
		}

		if !strings.HasSuffix(path, target.RelPath) {
			t.Errorf("Resolve() path %q doesn't end with RelPath %q", path, target.RelPath)
		}
	}
}
