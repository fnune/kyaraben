package tic80

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestDefinitionSystem(t *testing.T) {
	def := Definition{}
	sys := def.System()

	if sys.ID != model.SystemTIC80 {
		t.Errorf("ID: got %s, want %s", sys.ID, model.SystemTIC80)
	}

	if sys.Name == "" {
		t.Error("Name should not be empty")
	}

	if sys.Description == "" {
		t.Error("Description should not be empty")
	}

	if sys.Hidden {
		t.Error("TIC-80 should not be hidden")
	}
}

func TestDefinitionDefaultEmulatorID(t *testing.T) {
	def := Definition{}
	emuID := def.DefaultEmulatorID()

	if emuID != model.EmulatorTIC80 {
		t.Errorf("DefaultEmulatorID: got %s, want %s", emuID, model.EmulatorTIC80)
	}
}
