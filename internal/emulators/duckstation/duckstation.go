// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
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
			Message:     "BIOS required (HLE fallback has limited compatibility)",
			Provisions:  psxBIOSProvisions,
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

var psxBIOSProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionBIOS, "scph1000.bin", "Japan v1.0", []string{"239665b1a3dade1b5a52c06338011044"}),
	model.HashedProvision(model.ProvisionBIOS, "scph1001.bin", "USA v2.0", []string{"924e392ed05558ffdb115408c263dccf"}),
	model.HashedProvision(model.ProvisionBIOS, "scph1002.bin", "Europe v2.0", []string{"54847e693405ffeb0359c6287434cbef"}),
	model.HashedProvision(model.ProvisionBIOS, "scph100.bin", "Japan early", []string{"8abc1b549a4a80954addc48ef02c4521"}),
	model.HashedProvision(model.ProvisionBIOS, "scph101.bin", "USA v4.4", []string{"6e3735ff4c7dc899ee98981385f6f3d0"}),
	model.HashedProvision(model.ProvisionBIOS, "scph102A.bin", "Europe v4.4", []string{"b10f5e0e3d9eb60e5159690680b1e774"}),
	model.HashedProvision(model.ProvisionBIOS, "scph102B.bin", "Europe v4.4", []string{"de93caec13d1a141a40a79f5c86168d6"}),
	model.HashedProvision(model.ProvisionBIOS, "scph102C.bin", "Europe v4.4", []string{"de93caec13d1a141a40a79f5c86168d6"}),
	model.HashedProvision(model.ProvisionBIOS, "scph3000.bin", "Japan v3.0", []string{"849515939161e62f6b866f6853006780"}),
	model.HashedProvision(model.ProvisionBIOS, "scph3500.bin", "Japan v3.5", []string{"cba733ceeff5aef5c32254f1d617fa62"}),
	model.HashedProvision(model.ProvisionBIOS, "scph5000.bin", "Japan v5.0", []string{"eb201d2d98251a598af467d4347bb62f"}),
	model.HashedProvision(model.ProvisionBIOS, "scph5500.bin", "Japan v3.0", []string{"8dd7d5296a650fac7319bce665a6a53c"}),
	model.HashedProvision(model.ProvisionBIOS, "scph5501.bin", "USA v3.0", []string{"490f666e1afb15b7362b406ed1cea246"}),
	model.HashedProvision(model.ProvisionBIOS, "scph5502.bin", "Europe v3.0", []string{"32736f17079d0b2b7024407c39bd3050"}),
	model.HashedProvision(model.ProvisionBIOS, "scph7001.bin", "USA v4.1", []string{"1e68c231d0896b7eadcad1d7d8e76129"}),
	model.HashedProvision(model.ProvisionBIOS, "scph7002.bin", "Europe v4.1", []string{"b9d9a0286c33dc6b7237bb13cd46fdee"}),
	model.HashedProvision(model.ProvisionBIOS, "scph7003.bin", "USA v4.1", []string{"490f666e1afb15b7362b406ed1cea246"}),
	model.HashedProvision(model.ProvisionBIOS, "scph7502.bin", "Europe v4.1", []string{"b9d9a0286c33dc6b7237bb13cd46fdee"}),
	model.HashedProvision(model.ProvisionBIOS, "scph9002(7502).bin", "Europe v4.1", []string{"b9d9a0286c33dc6b7237bb13cd46fdee"}),
	model.HashedProvision(model.ProvisionBIOS, "PSXONPSP660.bin", "PSP firmware", []string{"c53ca5908936d412331790f4426c6c33"}),
	model.HashedProvision(model.ProvisionBIOS, "psxonpsp660.bin", "PSP firmware", []string{"c53ca5908936d412331790f4426c6c33"}),
	model.HashedProvision(model.ProvisionBIOS, "ps1_rom.bin", "PS3 firmware", []string{"81bbe60ba7a3d1cea1d48c14cbcc647b"}),
}
