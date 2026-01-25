package snes

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemSNES,
		Name:        "Super Nintendo",
		Description: "16-bit home console by Nintendo (1990)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorRetroArchBsnes
}
