package tic80

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemTIC80,
		Name:        "TIC-80",
		Description: "Fantasy console for making and playing tiny games",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorTIC80
}
