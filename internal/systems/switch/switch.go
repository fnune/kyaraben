package switchsys

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemSwitch,
		Name:        "Nintendo Switch",
		Description: "Hybrid console by Nintendo (2017)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorEden
}
