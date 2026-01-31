package wii

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemWii,
		Name:        "Wii",
		Description: "Home console by Nintendo (2006)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorDolphin
}
