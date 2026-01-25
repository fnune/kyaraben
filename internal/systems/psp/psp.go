package psp

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemPSP,
		Name:        "PlayStation Portable",
		Description: "Handheld console by Sony (2004)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorPPSSPP
}
