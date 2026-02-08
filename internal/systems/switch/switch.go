package switchsys

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDSwitch,
		Name:         "Nintendo Switch",
		Description:  "Hybrid console by Nintendo (2017)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "NSW",
		Extensions:   []string{".nsp", ".xci", ".nca", ".nro", ".nso", ".nsz", ".xcz"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDEden
}
