package melonds

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorMelonDS,
		Name:    "melonDS",
		Systems: []model.SystemID{model.SystemNDS},
		Package: model.VersionedAppImageRef("melonds"),
		Provisions: []model.Provision{
			{
				ID:          "nds-bios-arm7",
				Kind:        model.ProvisionBIOS,
				Filename:    "bios7.bin",
				Description: "Nintendo DS ARM7 BIOS",
				Required:    false,
				MD5Hash:     "df692a80a5b1bc90728bc3dfc76cd948",
			},
			{
				ID:          "nds-bios-arm9",
				Kind:        model.ProvisionBIOS,
				Filename:    "bios9.bin",
				Description: "Nintendo DS ARM9 BIOS",
				Required:    false,
				MD5Hash:     "a392174eb3e572fed6447e956bde4b25",
			},
			{
				ID:          "nds-firmware",
				Kind:        model.ProvisionFirmware,
				Filename:    "firmware.bin",
				Description: "Nintendo DS Firmware",
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
			Binary:      "melonDS",
			GenericName: "Nintendo DS Emulator",
			Categories:  []string{"Game", "Emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "melonDS/melonDS.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"BIOS9Path"}, Value: store.SystemBiosDir(model.SystemNDS) + "/bios9.bin"},
			{Path: []string{"BIOS7Path"}, Value: store.SystemBiosDir(model.SystemNDS) + "/bios7.bin"},
			{Path: []string{"FirmwarePath"}, Value: store.SystemBiosDir(model.SystemNDS) + "/firmware.bin"},
			{Path: []string{"SaveFilePath"}, Value: store.SystemSavesDir(model.SystemNDS)},
			{Path: []string{"SavestatePath"}, Value: store.EmulatorStatesDir(model.EmulatorMelonDS)},
			{Path: []string{"ScreenshotPath"}, Value: store.SystemScreenshotsDir(model.SystemNDS)},
			{Path: []string{"LastROMFolder"}, Value: store.SystemRomsDir(model.SystemNDS)},
		},
	}}, nil
}
