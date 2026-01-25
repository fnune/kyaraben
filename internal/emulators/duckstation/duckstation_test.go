package duckstation

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

func TestDefinitionEmulator(t *testing.T) {
	def := Definition{}
	emu := def.Emulator()

	if emu.ID != model.EmulatorDuckStation {
		t.Errorf("ID: got %s, want %s", emu.ID, model.EmulatorDuckStation)
	}

	if len(emu.Systems) != 1 || emu.Systems[0] != model.SystemPSX {
		t.Errorf("Systems: got %v, want [%s]", emu.Systems, model.SystemPSX)
	}

	if emu.Name == "" {
		t.Error("Name should not be empty")
	}

	if len(emu.Provisions) == 0 {
		t.Error("DuckStation should have BIOS provisions")
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
