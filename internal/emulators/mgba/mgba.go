package mgba

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorMGBA,
		Name:    "mGBA",
		Systems: []model.SystemID{model.SystemGBA},
		Package: model.AppImageRef("mgba"),
		Provisions: []model.Provision{
			{
				ID:          "gba-bios",
				Kind:        model.ProvisionBIOS,
				Filename:    "gba_bios.bin",
				Description: "Game Boy Advance BIOS",
				Required:    false, // mGBA has HLE, BIOS is optional
				MD5Hash:     "a860e8c0b6d573d191e4ec7db1b1e4f6",
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "mgba",
			GenericName: "Game Boy Advance Emulator",
			Categories:  []string{"Game", "Emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "mgba/config.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"ports.qt", "savegamePath"}, Value: store.SystemSavesDir(model.SystemGBA)},
			{Path: []string{"ports.qt", "savestatePath"}, Value: store.EmulatorStatesDir(model.EmulatorMGBA)},
			{Path: []string{"ports.qt", "screenshotPath"}, Value: store.SystemScreenshotsDir(model.SystemGBA)},
		},
	}}, nil
}
