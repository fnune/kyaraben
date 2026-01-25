package tic80emu

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

func TestDefinitionEmulator(t *testing.T) {
	def := Definition{}
	emu := def.Emulator()

	if emu.ID != model.EmulatorTIC80 {
		t.Errorf("ID: got %s, want %s", emu.ID, model.EmulatorTIC80)
	}

	if len(emu.Systems) != 1 || emu.Systems[0] != model.SystemTIC80 {
		t.Errorf("Systems: got %v, want [%s]", emu.Systems, model.SystemTIC80)
	}

	if emu.Name == "" {
		t.Error("Name should not be empty")
	}

	if len(emu.Provisions) != 0 {
		t.Error("TIC-80 should have no provisions")
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

	patches, err := gen.Generate(userStore, []model.SystemID{model.SystemTIC80})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(patches) != 0 {
		t.Errorf("expected 0 patches (TIC-80 has no config), got %d", len(patches))
	}
}

func TestConfigPaths(t *testing.T) {
	gen := &Config{}
	paths := gen.ConfigPaths()

	if len(paths) != 0 {
		t.Errorf("expected 0 config paths, got %d", len(paths))
	}
}
