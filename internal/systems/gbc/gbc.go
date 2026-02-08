package gbc

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDGBC,
		Name:         "Game Boy Color",
		Description:  "8-bit handheld console by Nintendo (1998)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "GBC",
		Extensions:   []string{".gbc", ".gb", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDMGBA
}
