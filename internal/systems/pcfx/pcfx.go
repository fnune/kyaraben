package pcfx

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDPCFX,
		Name:         "PC-FX",
		Description:  "32-bit home console by NEC (1994)",
		Manufacturer: model.ManufacturerNEC,
		Label:        "PCFX",
		Extensions:   []string{".cue", ".ccd", ".chd", ".iso", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchBeetlePCFX
}
