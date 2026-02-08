package nes

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDNES,
		Name:         "Nintendo Entertainment System",
		Description:  "8-bit home console by Nintendo (1983)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "NES",
		Extensions:   []string{".nes", ".unf", ".unif", ".fds", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchMesen
}
