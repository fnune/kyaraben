package flycast

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDFlycast,
		Name:    "Flycast",
		Systems: []model.SystemID{model.SystemIDDreamcast},
		Package: model.AppImageRef("flycast"),
		Provisions: []model.Provision{
			{
				ID:          "dc-bios",
				Kind:        model.ProvisionBIOS,
				Filename:    "dc_boot.bin",
				Description: "Boot",
				Required:    false,
				MD5Hash:     "e10c53c2f8b90bab96ead2d368858623",
			},
			{
				ID:          "dc-flash",
				Kind:        model.ProvisionBIOS,
				Filename:    "dc_flash.bin",
				Description: "Flash",
				Required:    false,
				MD5Hash:     "0a93f7940c455905bea6e392dfde92a4",
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "flycast",
			GenericName: "Sega Dreamcast Emulator",
			Categories:  []string{"Game", "Emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "flycast/emu.cfg",
	Format:  model.ConfigFormatCFG,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"config", "Dreamcast.ContentPath"}, Value: store.SystemRomsDir(model.SystemIDDreamcast)},
			{Path: []string{"config", "SavestatesPath"}, Value: store.EmulatorStatesDir(model.EmulatorIDFlycast)},
			{Path: []string{"config", "ScreenshotsPath"}, Value: store.SystemScreenshotsDir(model.SystemIDDreamcast)},
		},
	}}, nil
}
