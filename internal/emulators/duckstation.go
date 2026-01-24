package emulators

import (
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

// DuckStationConfig generates DuckStation configuration.
type DuckStationConfig struct{}

func (d *DuckStationConfig) EmulatorID() model.EmulatorID {
	return model.EmulatorDuckStation
}

func (d *DuckStationConfig) ConfigPaths() []string {
	configDir, _ := os.UserConfigDir()
	return []string{
		filepath.Join(configDir, "duckstation", "settings.ini"),
	}
}

func (d *DuckStationConfig) Generate(userStore *store.UserStore, systems []model.SystemID) ([]model.ConfigPatch, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "duckstation", "settings.ini")

	entries := []model.ConfigEntry{
		// BIOS settings
		{Section: "BIOS", Key: "SearchDirectory", Value: userStore.SystemBiosDir(model.SystemPSX)},

		// Memory card settings
		{Section: "MemoryCards", Key: "Directory", Value: userStore.SystemSavesDir(model.SystemPSX)},

		// Save state settings
		{Section: "Folders", Key: "SaveStates", Value: userStore.SystemStatesDir(model.SystemPSX)},
		{Section: "Folders", Key: "Screenshots", Value: userStore.SystemScreenshotsDir(model.SystemPSX)},

		// Game list directory
		{Section: "GameList", Key: "RecursivePaths", Value: userStore.SystemRomsDir(model.SystemPSX)},
	}

	patch := model.ConfigPatch{
		Config: model.EmulatorConfig{
			Path:       configPath,
			Format:     model.ConfigFormatINI,
			EmulatorID: model.EmulatorDuckStation,
		},
		Entries: entries,
	}

	return []model.ConfigPatch{patch}, nil
}
