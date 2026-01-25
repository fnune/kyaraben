package psx

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemPSX,
		Name:        "PlayStation",
		Description: "32-bit home console by Sony (1994)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorDuckStation
}
