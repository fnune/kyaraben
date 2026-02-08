// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package duckstation

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDDuckStation,
		Name:    "DuckStation",
		Systems: []model.SystemID{model.SystemIDPSX},
		Package: model.AppImageRef("duckstation"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "At least one BIOS required",
			Provisions: []model.Provision{
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph5501.bin",
					Description: "USA v3.0",
					Hashes:      []string{"490f666e1afb15b7362b406ed1cea246"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph5500.bin",
					Description: "Japan v3.0",
					Hashes:      []string{"8dd7d5296a650fac7319bce665a6a53c"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph5502.bin",
					Description: "Europe v3.0",
					Hashes:      []string{"32736f17079d0b2b7024407c39bd3050"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph1001.bin",
					Description: "USA v2.0",
					Hashes:      []string{"924e392ed05558ffdb115408c263dccf"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph7001.bin",
					Description: "USA v4.1",
					Hashes:      []string{"1e68c231d0896b7eadcad1d7d8e76129"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph101.bin",
					Description: "USA v4.4",
					Hashes:      []string{"6e3735ff4c7dc899ee98981385f6f3d0"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "PSXONPSP660.bin",
					Description: "from PSP firmware",
					Hashes:      []string{"c53ca5908936d412331790f4426c6c33"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "ps1_rom.bin",
					Description: "from PS3 firmware",
					Hashes:      []string{"81bbe60ba7a3d1cea1d48c14cbcc647b"},
				},
			},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "duckstation",
			GenericName: "PlayStation Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " -fullscreen"
				}
				cmd += " %ROM%"
				return cmd
			},
		},
		PathUsage: model.StandardPathUsage(),
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
			{Path: []string{"Main", "SettingsVersion"}, Value: "3"},
			{Path: []string{"AutoUpdater", "CheckAtStartup"}, Value: "false"},
			{Path: []string{"BIOS", "SearchDirectory"}, Value: store.SystemBiosDir(model.SystemIDPSX)},
			{Path: []string{"MemoryCards", "Directory"}, Value: store.SystemSavesDir(model.SystemIDPSX)},
			{Path: []string{"Folders", "SaveStates"}, Value: store.EmulatorStatesDir(model.EmulatorIDDuckStation)},
			{Path: []string{"Folders", "Screenshots"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDDuckStation)},
			{Path: []string{"GameList", "RecursivePaths"}, Value: store.SystemRomsDir(model.SystemIDPSX)},
		},
	}}, nil
}
