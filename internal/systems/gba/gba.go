package gba

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDGBA,
		Name:         "Game Boy Advance",
		Description:  "32-bit handheld by Nintendo (2001)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "GBA",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDMGBA
}
