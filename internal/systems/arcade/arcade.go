package arcade

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDArcade,
		Name:         "Arcade",
		Description:  "Arcade game systems",
		Manufacturer: model.ManufacturerOther,
		Label:        "ARCADE",
		Extensions:   []string{".zip", ".7z"},
		DisplayType:  model.DisplayTypeCRT,
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchFBNeo
}
