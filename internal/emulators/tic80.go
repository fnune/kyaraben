package emulators

import (
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

// TIC80Config generates TIC-80 configuration.
// TIC-80 is primarily used for E2E testing since it has no provisions.
type TIC80Config struct{}

func (t *TIC80Config) EmulatorID() model.EmulatorID {
	return model.EmulatorTIC80
}

func (t *TIC80Config) ConfigPaths() []string {
	// TIC-80 doesn't have a standard config file path that we manage
	return []string{}
}

func (t *TIC80Config) Generate(userStore *store.UserStore, systems []model.SystemID) ([]model.ConfigPatch, error) {
	// TIC-80 stores data in ~/.local/share/com.nesbox.tic/
	// For now, we don't manage its config - it works out of the box
	// Users can place .tic carts in the roms/tic80 directory
	return []model.ConfigPatch{}, nil
}
