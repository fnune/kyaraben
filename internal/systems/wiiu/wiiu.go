package wiiu

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemWiiU,
		Name:        "Wii U",
		Description: "Home console by Nintendo (2012)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorCemu
}
