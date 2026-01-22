package e2etestemu

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

func TestDefinitionEmulator(t *testing.T) {
	def := Definition{}
	emu := def.Emulator()

	if emu.ID != model.EmulatorE2ETest {
		t.Errorf("ID: got %s, want %s", emu.ID, model.EmulatorE2ETest)
	}

	if len(emu.Systems) != 1 || emu.Systems[0] != model.SystemE2ETest {
		t.Errorf("Systems: got %v, want [%s]", emu.Systems, model.SystemE2ETest)
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
	userStore := store.NewUserStore("/tmp/test")
	gen := &Config{}

	patches, err := gen.Generate(userStore, []model.SystemID{model.SystemE2ETest})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(patches) != 0 {
		t.Errorf("expected 0 patches, got %d", len(patches))
	}
}

func TestConfigPaths(t *testing.T) {
	gen := &Config{}
	paths := gen.ConfigPaths()

	if len(paths) != 0 {
		t.Errorf("expected 0 config paths, got %d", len(paths))
	}
}
