package nintendo3ds

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.System3DS,
		Name:        "Nintendo 3DS",
		Description: "Handheld console by Nintendo (2011)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorAzahar
}
