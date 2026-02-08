package gb

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDGB,
		Name:         "Game Boy",
		Description:  "8-bit handheld console by Nintendo (1989)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "GB",
		Extensions:   []string{".gb", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDMGBA
}
