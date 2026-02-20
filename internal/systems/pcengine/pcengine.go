package pcengine

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDPCEngine,
		Name:         "PC Engine / TurboGrafx-16",
		Description:  "8/16-bit home console by NEC (1987)",
		Manufacturer: model.ManufacturerNEC,
		Label:        "PCE",
		Extensions:   []string{".pce", ".sgx", ".cue", ".ccd", ".chd", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchBeetlePCE
}
