package nds

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDNDS,
		Name:         "Nintendo DS",
		Description:  "Dual-screen handheld by Nintendo (2004)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "NDS",
		Extensions:   []string{".nds", ".app", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchMelonDS
}
