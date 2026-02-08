package wii

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDWii,
		Name:         "Wii",
		Description:  "Home console by Nintendo (2006)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "Wii",
		Extensions:   []string{".iso", ".gcz", ".rvz", ".wbfs", ".wad", ".wia", ".ciso", ".dol", ".elf", ".m3u", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDDolphin
}
