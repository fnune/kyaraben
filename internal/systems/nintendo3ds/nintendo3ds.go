package nintendo3ds

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemID3DS,
		Name:         "Nintendo 3DS",
		Description:  "Handheld console by Nintendo (2011)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "3DS",
		Extensions:   []string{".3ds", ".3dsx", ".cia", ".cxi"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDAzahar
}
