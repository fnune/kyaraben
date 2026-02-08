package melonds

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDMelonDS,
		Name:    "melonDS",
		Systems: []model.SystemID{model.SystemIDNDS},
		Package: model.AppImageRef("melonds"),
		Provisions: []model.Provision{
			{
				ID:          "nds-bios-arm7",
				Kind:        model.ProvisionBIOS,
				Filename:    "bios7.bin",
				Description: "ARM7",
				Required:    false,
				MD5Hash:     "df692a80a5b1bc90728bc3dfc76cd948",
			},
			{
				ID:          "nds-bios-arm9",
				Kind:        model.ProvisionBIOS,
				Filename:    "bios9.bin",
				Description: "ARM9",
				Required:    false,
				MD5Hash:     "a392174eb3e572fed6447e956bde4b25",
			},
			{
				ID:          "nds-firmware",
				Kind:        model.ProvisionFirmware,
				Filename:    "firmware.bin",
				Description: "Firmware",
				Required:    false,
				MD5Hash:     "e45033d9b0fa6b0de071292bba7c9d13",
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "melonds",
			GenericName: "Nintendo DS Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " -f"
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
	RelPath: "melonDS/melonDS.toml",
	Format:  model.ConfigFormatTOML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"DS", "BIOS9Path"}, Value: store.SystemBiosDir(model.SystemIDNDS) + "/bios9.bin"},
			{Path: []string{"DS", "BIOS7Path"}, Value: store.SystemBiosDir(model.SystemIDNDS) + "/bios7.bin"},
			{Path: []string{"DS", "FirmwarePath"}, Value: store.SystemBiosDir(model.SystemIDNDS) + "/firmware.bin"},
			{Path: []string{"Instance0", "SaveFilePath"}, Value: store.SystemSavesDir(model.SystemIDNDS)},
			{Path: []string{"Instance0", "SavestatePath"}, Value: store.EmulatorStatesDir(model.EmulatorIDMelonDS)},
			{Path: []string{"Instance0", "ScreenshotPath"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDMelonDS)},
			{Path: []string{"Instance0", "LastROMFolder"}, Value: store.SystemRomsDir(model.SystemIDNDS)},
		},
	}}, nil
}
