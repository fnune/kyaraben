package psvita

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDPSVita,
		Name:         "PlayStation Vita",
		Description:  "Handheld console by Sony (2011)",
		Manufacturer: model.ManufacturerSony,
		Label:        "Vita",
		Extensions:   []string{".vpk"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDVita3K
}
