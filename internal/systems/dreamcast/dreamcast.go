package dreamcast

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemDreamcast,
		Name:        "Sega Dreamcast",
		Description: "Home console by Sega (1998)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorFlycast
}
