package saturn

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDSaturn,
		Name:         "Sega Saturn",
		Description:  "32-bit home console by Sega (1994)",
		Manufacturer: model.ManufacturerSega,
		Label:        "SAT",
		Extensions:   []string{".bin", ".cue", ".chd", ".iso", ".m3u", ".mds", ".ccd", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchBeetleSaturn
}
