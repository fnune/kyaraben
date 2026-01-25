package snes

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestDefinitionSystem(t *testing.T) {
	def := Definition{}
	sys := def.System()

	if sys.ID != model.SystemSNES {
		t.Errorf("ID: got %s, want %s", sys.ID, model.SystemSNES)
	}

	if sys.Name == "" {
		t.Error("Name should not be empty")
	}

	if sys.Description == "" {
		t.Error("Description should not be empty")
	}

	if sys.Hidden {
		t.Error("SNES should not be hidden")
	}
}

func TestDefinitionDefaultEmulatorID(t *testing.T) {
	def := Definition{}
	emuID := def.DefaultEmulatorID()

	if emuID != model.EmulatorRetroArchBsnes {
		t.Errorf("DefaultEmulatorID: got %s, want %s", emuID, model.EmulatorRetroArchBsnes)
	}
}
