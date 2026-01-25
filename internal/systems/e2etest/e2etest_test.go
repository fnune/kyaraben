package e2etest

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestDefinitionSystem(t *testing.T) {
	def := Definition{}
	sys := def.System()

	if sys.ID != model.SystemE2ETest {
		t.Errorf("ID: got %s, want %s", sys.ID, model.SystemE2ETest)
	}

	if !sys.Hidden {
		t.Error("E2E test system should be hidden")
	}
}

func TestDefinitionDefaultEmulatorID(t *testing.T) {
	def := Definition{}
	emuID := def.DefaultEmulatorID()

	if emuID != model.EmulatorE2ETest {
		t.Errorf("DefaultEmulatorID: got %s, want %s", emuID, model.EmulatorE2ETest)
	}
}
