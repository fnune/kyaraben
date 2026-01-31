package psvita

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemPSVita,
		Name:        "PlayStation Vita",
		Description: "Handheld console by Sony (2011)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorVita3K
}
