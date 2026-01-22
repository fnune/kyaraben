package e2etest

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemE2ETest,
		Name:        "E2E Test",
		Description: "Hidden system for CI testing (uses hello from nixpkgs)",
		Hidden:      true,
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorE2ETest
}
