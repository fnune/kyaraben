package psx

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDPSX,
		Name:         "PlayStation",
		Description:  "32-bit home console by Sony (1994)",
		Manufacturer: model.ManufacturerSony,
		Label:        "PSX",
		Extensions:   []string{".bin", ".cue", ".chd", ".iso", ".img", ".m3u"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDDuckStation
}
