package c64

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDC64,
		Name:         "Commodore 64",
		Description:  "8-bit home computer by Commodore (1982)",
		Manufacturer: model.ManufacturerCommodore,
		Label:        "C64",
		Extensions: []string{
			".crt", ".d64", ".d71", ".d80", ".d81", ".d82",
			".g41", ".g64", ".nib", ".prg", ".p00", ".t64",
			".tap", ".x64", ".vsf", ".m3u", ".7z", ".zip",
		},
		DisplayType: model.DisplayTypeCRT,
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchVICE
}
