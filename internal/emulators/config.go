package emulators

import (
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

type ConfigGenerator interface {
	EmulatorID() model.EmulatorID
	Generate(userStore *store.UserStore, systems []model.SystemID) ([]model.ConfigPatch, error)
	ConfigPaths() []string
}

func GetConfigGenerator(emuID model.EmulatorID) ConfigGenerator {
	switch emuID {
	case model.EmulatorRetroArchBsnes:
		return &RetroArchConfig{}
	case model.EmulatorDuckStation:
		return &DuckStationConfig{}
	case model.EmulatorTIC80:
		return &TIC80Config{}
	default:
		return nil
	}
}
