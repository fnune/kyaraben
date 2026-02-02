package emulators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/emulators/duckstation"
	"github.com/fnune/kyaraben/internal/emulators/flycast"
	"github.com/fnune/kyaraben/internal/emulators/melonds"
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbsnes"
	"github.com/fnune/kyaraben/internal/emulators/rpcs3"
	"github.com/fnune/kyaraben/internal/emulators/vita3k"
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
	return filepath.Join(f.root, "saves", string(sys))
}

func (f *fakeStoreReader) EmulatorStatesDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "states", string(emu))
}

func (f *fakeStoreReader) SystemScreenshotsDir(sys model.SystemID) string {
	return filepath.Join(f.root, "screenshots", string(sys))
}

func (f *fakeStoreReader) SystemRomsDir(sys model.SystemID) string {
	return filepath.Join(f.root, "roms", string(sys))
}

func (f *fakeStoreReader) EmulatorOpaqueDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "opaque", string(emu))
}

func TestDuckStationGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := duckstation.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
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

func TestRetroArchCoresGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}

	tests := []struct {
		name       string
		gen        model.ConfigGenerator
		system     model.SystemID
		coreName   string
		wantRomDir string
	}{
		{
			name:       "bsnes",
			gen:        retroarchbsnes.Definition{}.ConfigGenerator(),
			system:     model.SystemSNES,
			coreName:   "bsnes_libretro",
			wantRomDir: "/emulation/roms/snes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches, err := tt.gen.Generate(store)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if len(patches) != 2 {
				t.Fatalf("expected 2 patches (shared + override), got %d", len(patches))
			}

			// First patch: shared retroarch.cfg (only system_directory)
			shared := patches[0]
			if shared.Target.RelPath != "retroarch/retroarch.cfg" {
				t.Errorf("expected shared config path, got %s", shared.Target.RelPath)
			}
			sharedKeys := collectKeys(shared.Entries)
			if !sharedKeys["system_directory"] {
				t.Error("shared config missing system_directory")
			}

			// Second patch: per-core override with all paths
			override := patches[1]
			expectedOverridePath := retroarch.CoreOverrideTarget(tt.coreName).RelPath
			if override.Target.RelPath != expectedOverridePath {
				t.Errorf("expected override path %q, got %q", expectedOverridePath, override.Target.RelPath)
			}

			// Check all path settings are in override
			overrideKeys := collectKeys(override.Entries)
			for _, key := range []string{"savefile_directory", "savestate_directory", "screenshot_directory", "rgui_browser_directory"} {
				if !overrideKeys[key] {
					t.Errorf("override missing key %q", key)
				}
			}

			// Verify ROM browser points to correct system
			for _, entry := range override.Entries {
				if entry.Key() == "rgui_browser_directory" && !strings.Contains(entry.Value, tt.wantRomDir) {
					t.Errorf("rgui_browser_directory %q doesn't contain %s", entry.Value, tt.wantRomDir)
				}
			}
		})
	}
}

func collectKeys(entries []model.ConfigEntry) map[string]bool {
	keys := make(map[string]bool)
	for _, e := range entries {
		keys[e.Key()] = true
	}
	return keys
}

func TestMelonDSGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := melonds.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
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

	if !strings.Contains(patch.Target.RelPath, "melonDS") {
		t.Errorf("expected RelPath to contain 'melonDS', got %s", patch.Target.RelPath)
	}

	expectedKeys := map[string]bool{
		"BIOS9Path":      false,
		"BIOS7Path":      false,
		"SaveFilePath":   false,
		"SavestatePath":  false,
		"ScreenshotPath": false,
		"LastROMFolder":  false,
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

func TestFlycastGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := flycast.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
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

	if !strings.Contains(patch.Target.RelPath, "flycast") {
		t.Errorf("expected RelPath to contain 'flycast', got %s", patch.Target.RelPath)
	}

	foundStorePath := false
	for _, entry := range patch.Entries {
		if strings.Contains(entry.Value, "/emulation") {
			foundStorePath = true
			break
		}
	}

	if !foundStorePath {
		t.Error("no entry contains the store path")
	}
}

func TestVita3KGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := vita3k.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatYAML {
		t.Errorf("expected YAML format, got %s", patch.Target.Format)
	}

	if !strings.Contains(patch.Target.RelPath, "Vita3K") {
		t.Errorf("expected RelPath to contain 'Vita3K', got %s", patch.Target.RelPath)
	}

	foundPrefPath := false
	for _, entry := range patch.Entries {
		if entry.Key() == "pref-path" {
			foundPrefPath = true
			if !strings.Contains(entry.Value, "vita3k") {
				t.Errorf("pref-path should contain 'vita3k', got %s", entry.Value)
			}
		}
	}

	if !foundPrefPath {
		t.Error("expected pref-path entry not found")
	}
}

func TestRPCS3Generate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := rpcs3.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatYAML {
		t.Errorf("expected YAML format, got %s", patch.Target.Format)
	}

	if !strings.Contains(patch.Target.RelPath, "rpcs3") {
		t.Errorf("expected RelPath to contain 'rpcs3', got %s", patch.Target.RelPath)
	}

	foundEmulatorDir := false
	for _, entry := range patch.Entries {
		if strings.Contains(entry.Key(), "EmulatorDir") {
			foundEmulatorDir = true
			if !strings.Contains(entry.Value, "rpcs3") {
				t.Errorf("EmulatorDir should contain 'rpcs3', got %s", entry.Value)
			}
		}
	}

	if !foundEmulatorDir {
		t.Error("expected EmulatorDir entry not found")
	}
}

func TestGeneratedEntriesContainStorePaths(t *testing.T) {
	store := &fakeStoreReader{root: "/test/emulation"}
	gen := duckstation.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
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
