package mastersystem

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDMasterSystem,
		Name:         "Sega Master System",
		Description:  "8-bit home console by Sega (1985)",
		Manufacturer: model.ManufacturerSega,
		Label:        "SMS",
		Extensions:   []string{".sms", ".bin", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchGenesisPlusGX
}
