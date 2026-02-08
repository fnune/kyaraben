package genesis

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDGenesis,
		Name:         "Sega Genesis",
		Description:  "16-bit home console by Sega (1988)",
		Manufacturer: model.ManufacturerSega,
		Label:        "GEN",
		Extensions:   []string{".gen", ".md", ".smd", ".bin", ".chd", ".cue", ".iso", ".m3u", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchGenesisPlusGX
}
