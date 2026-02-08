package psp

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDPSP,
		Name:         "PlayStation Portable",
		Description:  "Handheld console by Sony (2004)",
		Manufacturer: model.ManufacturerSony,
		Label:        "PSP",
		Extensions:   []string{".iso", ".cso", ".chd", ".pbp", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDPPSSPP
}
