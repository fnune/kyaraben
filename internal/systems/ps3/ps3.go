package ps3

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDPS3,
		Name:         "PlayStation 3",
		Description:  "Home console by Sony (2006)",
		Manufacturer: model.ManufacturerSony,
		Label:        "PS3",
		Extensions:   []string{".ps3", ".ps3dir"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRPCS3
}
