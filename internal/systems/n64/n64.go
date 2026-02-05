package n64

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDN64,
		Name:         "Nintendo 64",
		Description:  "64-bit home console by Nintendo (1996)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "N64",
		Extensions:   []string{".n64", ".v64", ".z64"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchMupen64Plus
}
