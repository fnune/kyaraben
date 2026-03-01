package atari2600

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDAtari2600,
		Name:         "Atari 2600",
		Description:  "8-bit home console by Atari (1977)",
		Manufacturer: model.ManufacturerAtari,
		Label:        "2600",
		Extensions:   []string{".a26", ".bin", ".7z", ".zip"},
		DisplayType:  model.DisplayTypeCRT,
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchStella
}
