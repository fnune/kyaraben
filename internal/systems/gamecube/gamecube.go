package gamecube

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDGameCube,
		Name:         "GameCube",
		Description:  "Home console by Nintendo (2001)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "GC",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDDolphin
}
