package ps3

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemPS3,
		Name:        "PlayStation 3",
		Description: "Home console by Sony (2006)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorRPCS3
}
