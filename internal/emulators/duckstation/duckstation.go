package duckstation

import "github.com/fnune/kyaraben/internal/model"

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
		Launcher: model.LauncherInfo{
			Binary:      "duckstation-qt",
			GenericName: "PlayStation Emulator",
			Categories:  []string{"Game", "Emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "duckstation/settings.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"BIOS", "SearchDirectory"}, Value: store.SystemBiosDir(model.SystemPSX)},
			{Path: []string{"MemoryCards", "Directory"}, Value: store.SystemSavesDir(model.SystemPSX)},
			{Path: []string{"Folders", "SaveStates"}, Value: store.EmulatorStatesDir(model.EmulatorDuckStation)},
			{Path: []string{"Folders", "Screenshots"}, Value: store.SystemScreenshotsDir(model.SystemPSX)},
			{Path: []string{"GameList", "RecursivePaths"}, Value: store.SystemRomsDir(model.SystemPSX)},
		},
	}}, nil
}
