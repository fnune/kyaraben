package psx

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestDefinitionSystem(t *testing.T) {
	def := Definition{}
	sys := def.System()

	if sys.ID != model.SystemPSX {
		t.Errorf("ID: got %s, want %s", sys.ID, model.SystemPSX)
	}

	if sys.Name == "" {
		t.Error("Name should not be empty")
	}

	if sys.Description == "" {
		t.Error("Description should not be empty")
	}

	if sys.Hidden {
		t.Error("PSX should not be hidden")
	}
}

func TestDefinitionDefaultEmulatorID(t *testing.T) {
	def := Definition{}
	emuID := def.DefaultEmulatorID()

	if emuID != model.EmulatorDuckStation {
		t.Errorf("DefaultEmulatorID: got %s, want %s", emuID, model.EmulatorDuckStation)
	}
}
