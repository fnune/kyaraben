package duckstation

import (
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorDuckStation,
		Name:    "DuckStation",
		Systems: []model.SystemID{model.SystemPSX},
		Package: model.NixpkgsRef("duckstation"),
		Provisions: []model.Provision{
			{
				ID:          "psx-bios-usa",
				Kind:        model.ProvisionBIOS,
				Filename:    "scph5501.bin",
				Description: "PlayStation BIOS (USA)",
				Required:    true,
				MD5Hash:     "490f666e1afb15b7362b406ed1cea246",
			},
			{
				ID:          "psx-bios-japan",
				Kind:        model.ProvisionBIOS,
				Filename:    "scph5500.bin",
				Description: "PlayStation BIOS (Japan)",
				Required:    false,
				MD5Hash:     "8dd7d5296a650fac7319bce665a6a53c",
			},
			{
				ID:          "psx-bios-europe",
				Kind:        model.ProvisionBIOS,
				Filename:    "scph5502.bin",
				Description: "PlayStation BIOS (Europe)",
				Required:    false,
				MD5Hash:     "32736f17079d0b2b7024407c39bd3050",
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		ConfigPaths: []string{
			"~/.config/duckstation/settings.ini",
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

func (c *Config) ConfigPaths() []string {
	configDir, _ := os.UserConfigDir()
	return []string{
		filepath.Join(configDir, "duckstation", "settings.ini"),
	}
}

func (c *Config) Generate(store model.StoreReader, systems []model.SystemID) ([]model.ConfigPatch, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "duckstation", "settings.ini")

	entries := []model.ConfigEntry{
		{Section: "BIOS", Key: "SearchDirectory", Value: store.SystemBiosDir(model.SystemPSX)},
		{Section: "MemoryCards", Key: "Directory", Value: store.SystemSavesDir(model.SystemPSX)},
		{Section: "Folders", Key: "SaveStates", Value: store.SystemStatesDir(model.SystemPSX)},
		{Section: "Folders", Key: "Screenshots", Value: store.SystemScreenshotsDir(model.SystemPSX)},
		{Section: "GameList", Key: "RecursivePaths", Value: store.SystemRomsDir(model.SystemPSX)},
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
