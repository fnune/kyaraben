package ngp

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDNGP,
		Name:         "Neo Geo Pocket Color",
		Description:  "Handheld console by SNK (1999)",
		Manufacturer: model.ManufacturerSNK,
		Label:        "NGP",
		Extensions:   []string{".ngp", ".ngc", ".ngpc", ".npc", ".7z", ".zip"},
		DisplayType:  model.DisplayTypeLCD,
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchBeetleNGP
}
