// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package pcsx2

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDPCSX2,
		Name:    "PCSX2",
		Systems: []model.SystemID{model.SystemIDPS2},
		Package: model.AppImageRef("pcsx2"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "At least one BIOS required",
			Provisions: []model.Provision{
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph39001.bin",
					Description: "USA v1.60",
					Hashes:      []string{"d5ce2c7d119f563ce04bc04dbc3a323e"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph70012.bin",
					Description: "USA v2.00",
					Hashes:      []string{"d333558cc14561c1fdc334c75d5f37b7"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph77001.bin",
					Description: "USA v2.20",
					Hashes:      []string{"af60e6d1a939019d55e5b330d24b1c25"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph70004.bin",
					Description: "Europe v2.00",
					Hashes:      []string{"d333558cc14561c1fdc334c75d5f37b7"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph39004.bin",
					Description: "Europe v1.60",
					Hashes:      []string{"1ad977bb539fc9448a08ab276a836bbc"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph10000.bin",
					Description: "Japan v1.00",
					Hashes:      []string{"b7fa11e87d51752a98b38e3e691cbf17"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph50000.bin",
					Description: "Japan v1.70",
					Hashes:      []string{"d3f1853a16c2ec18f3cd1ae655213308"},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "scph70000.bin",
					Description: "Japan v2.00",
					Hashes:      []string{"0bf988e9c7aaa4c051805b0fa6eb3387"},
				},
			},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "pcsx2",
			GenericName: "PlayStation 2 Emulator",
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
	RelPath: "PCSX2/inis/PCSX2.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"UI", "SettingsVersion"}, Value: "1"},
			{Path: []string{"UI", "SetupWizardIncomplete"}, Value: "false"},
			{Path: []string{"Folders", "Bios"}, Value: store.SystemBiosDir(model.SystemIDPS2)},
			{Path: []string{"Folders", "MemoryCards"}, Value: store.SystemSavesDir(model.SystemIDPS2)},
			{Path: []string{"Folders", "Savestates"}, Value: store.EmulatorStatesDir(model.EmulatorIDPCSX2)},
			{Path: []string{"Folders", "Snapshots"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDPCSX2)},
			{Path: []string{"GameList", "RecursivePaths"}, Value: store.SystemRomsDir(model.SystemIDPS2)},
		},
	}}, nil
}
