package lynx

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDLynx,
		Name:         "Atari Lynx",
		Description:  "Handheld console by Atari (1989)",
		Manufacturer: model.ManufacturerAtari,
		Label:        "LYNX",
		Extensions:   []string{".lnx", ".o", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchBeetleLynx
}
