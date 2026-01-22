package emulators

import (
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

// ConfigGenerator generates configuration for a specific emulator.
type ConfigGenerator interface {
	// EmulatorID returns which emulator this generator is for.
	EmulatorID() model.EmulatorID

	// Generate produces the configuration patches needed.
	Generate(userStore *store.UserStore, systems []model.SystemID) ([]model.ConfigPatch, error)

	// ConfigPaths returns the paths to config files this generator manages.
	ConfigPaths() []string
}

// GetConfigGenerator returns the appropriate generator for an emulator.
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
