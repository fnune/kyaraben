package wonderswan

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDWonderSwan,
		Name:         "WonderSwan",
		Description:  "Handheld console by Bandai (1999)",
		Manufacturer: model.ManufacturerBandai,
		Label:        "WS",
		Extensions:   []string{".ws", ".wsc", ".pc2", ".7z", ".zip"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchBeetleWSwan
}
