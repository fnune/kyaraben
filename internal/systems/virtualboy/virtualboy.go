package virtualboy

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDVirtualBoy,
		Name:         "Virtual Boy",
		Description:  "32-bit tabletop console by Nintendo (1995)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "VB",
		Extensions:   []string{".vb", ".vboy", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchBeetleVB
}
