package gamecube

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:          model.SystemGameCube,
		Name:        "GameCube",
		Description: "Home console by Nintendo (2001)",
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorDolphin
}
