package emulators

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

func TestRetroArchConfigGenerate(t *testing.T) {
	userStore := &store.UserStore{Root: "/home/user/Emulation"}
	gen := &RetroArchConfig{}

	patches, err := gen.Generate(userStore, []model.SystemID{model.SystemSNES})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Config.Format != model.ConfigFormatCFG {
		t.Errorf("expected format CFG, got %s", patch.Config.Format)
	}

	if patch.Config.EmulatorID != model.EmulatorRetroArchBsnes {
		t.Errorf("expected emulator ID %s, got %s", model.EmulatorRetroArchBsnes, patch.Config.EmulatorID)
	}

	expectedKeys := map[string]bool{
		"system_directory":       false,
		"savefile_directory":     false,
		"savestate_directory":    false,
		"screenshot_directory":   false,
		"rgui_browser_directory": false,
		"sort_savefiles_enable":  false,
		"sort_savestates_enable": false,
		"sort_screenshots_enable": false,
	}

	for _, entry := range patch.Entries {
		if _, ok := expectedKeys[entry.Key]; ok {
			expectedKeys[entry.Key] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("missing expected key: %s", key)
		}
	}
}

func TestRetroArchConfigPaths(t *testing.T) {
	gen := &RetroArchConfig{}
	paths := gen.ConfigPaths()

	if len(paths) != 1 {
		t.Fatalf("expected 1 config path, got %d", len(paths))
	}

	if paths[0] == "" {
		t.Error("config path should not be empty")
	}
}

func TestDuckStationConfigGenerate(t *testing.T) {
	userStore := &store.UserStore{Root: "/home/user/Emulation"}
	gen := &DuckStationConfig{}

	patches, err := gen.Generate(userStore, []model.SystemID{model.SystemPSX})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Config.Format != model.ConfigFormatINI {
		t.Errorf("expected format INI, got %s", patch.Config.Format)
	}

	if patch.Config.EmulatorID != model.EmulatorDuckStation {
		t.Errorf("expected emulator ID %s, got %s", model.EmulatorDuckStation, patch.Config.EmulatorID)
	}

	expectedSections := map[string]bool{
		"BIOS":        false,
		"MemoryCards": false,
		"Folders":     false,
		"GameList":    false,
	}

	for _, entry := range patch.Entries {
		if _, ok := expectedSections[entry.Section]; ok {
			expectedSections[entry.Section] = true
		}
	}

	for section, found := range expectedSections {
		if !found {
			t.Errorf("missing expected section: %s", section)
		}
	}
}

func TestDuckStationConfigPaths(t *testing.T) {
	gen := &DuckStationConfig{}
	paths := gen.ConfigPaths()

	if len(paths) != 1 {
		t.Fatalf("expected 1 config path, got %d", len(paths))
	}

	if paths[0] == "" {
		t.Error("config path should not be empty")
	}
}

func TestGetConfigGenerator(t *testing.T) {
	tests := []struct {
		emuID    model.EmulatorID
		expected bool
	}{
		{model.EmulatorRetroArchBsnes, true},
		{model.EmulatorDuckStation, true},
		{model.EmulatorTIC80, true},
		{model.EmulatorID("unknown"), false},
	}

	for _, test := range tests {
		gen := GetConfigGenerator(test.emuID)
		if test.expected && gen == nil {
			t.Errorf("expected generator for %s, got nil", test.emuID)
		}
		if !test.expected && gen != nil {
			t.Errorf("expected nil for %s, got generator", test.emuID)
		}
	}
}

func TestRetroArchConfigPathsContainUserStorePaths(t *testing.T) {
	userStore := &store.UserStore{Root: "/tmp/test-emulation"}
	gen := &RetroArchConfig{}

	patches, err := gen.Generate(userStore, []model.SystemID{model.SystemSNES})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	patch := patches[0]
	foundUserStorePath := false

	for _, entry := range patch.Entries {
		if entry.Key == "system_directory" && entry.Value != "" {
			foundUserStorePath = true
			break
		}
	}

	if !foundUserStorePath {
		t.Error("expected config entries to reference UserStore paths")
	}
}

func TestDuckStationConfigPathsContainUserStorePaths(t *testing.T) {
	userStore := &store.UserStore{Root: "/tmp/test-emulation"}
	gen := &DuckStationConfig{}

	patches, err := gen.Generate(userStore, []model.SystemID{model.SystemPSX})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	patch := patches[0]
	foundUserStorePath := false

	for _, entry := range patch.Entries {
		if entry.Section == "BIOS" && entry.Key == "SearchDirectory" && entry.Value != "" {
			foundUserStorePath = true
			break
		}
	}

	if !foundUserStorePath {
		t.Error("expected config entries to reference UserStore paths")
	}
}
