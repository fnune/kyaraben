package dreamcast

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDDreamcast,
		Name:         "Sega Dreamcast",
		Description:  "Home console by Sega (1998)",
		Manufacturer: model.ManufacturerSega,
		Label:        "DC",
		Extensions:   []string{".gdi", ".cdi", ".chd", ".cue", ".iso", ".m3u", ".7z", ".zip"},
		DisplayType:  model.DisplayTypeCRT,
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDFlycast
}
