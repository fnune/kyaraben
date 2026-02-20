package gamegear

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDGameGear,
		Name:         "Sega Game Gear",
		Description:  "8-bit handheld console by Sega (1990)",
		Manufacturer: model.ManufacturerSega,
		Label:        "GG",
		Extensions:   []string{".gg", ".bin", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchGenesisPlusGX
}
