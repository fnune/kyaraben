package retroarchbsnes

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

func TestDefinitionSystem(t *testing.T) {
	def := Definition{}
	emu := def.Emulator()

	if emu.ID != model.EmulatorRetroArchBsnes {
		t.Errorf("ID: got %s, want %s", emu.ID, model.EmulatorRetroArchBsnes)
	}

	if len(emu.Systems) != 1 || emu.Systems[0] != model.SystemSNES {
		t.Errorf("Systems: got %v, want [%s]", emu.Systems, model.SystemSNES)
	}

	if emu.Name == "" {
		t.Error("Name should not be empty")
	}
}

func TestDefinitionConfigGenerator(t *testing.T) {
	def := Definition{}
	gen := def.ConfigGenerator()

	if gen == nil {
		t.Fatal("ConfigGenerator should not be nil")
	}
}

func TestConfigGenerate(t *testing.T) {
	userStore := store.NewUserStore("/home/user/Emulation")
	gen := &Config{}

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
		"system_directory":        false,
		"savefile_directory":      false,
		"savestate_directory":     false,
		"screenshot_directory":    false,
		"rgui_browser_directory":  false,
		"sort_savefiles_enable":   false,
		"sort_savestates_enable":  false,
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

func TestConfigPaths(t *testing.T) {
	gen := &Config{}
	paths := gen.ConfigPaths()

	if len(paths) != 1 {
		t.Fatalf("expected 1 config path, got %d", len(paths))
	}

	if paths[0] == "" {
		t.Error("config path should not be empty")
	}
}
