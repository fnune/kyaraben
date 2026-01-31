package ps2

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDPS2,
		Name:         "PlayStation 2",
		Description:  "128-bit home console by Sony (2000)",
		Manufacturer: model.ManufacturerSony,
		Label:        "PS2",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDPCSX2
}
