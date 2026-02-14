package azahar

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDAzahar,
		Name:    "Azahar",
		Systems: []model.SystemID{model.SystemIDN3DS},
		Package: model.AppImageRef("azahar"),
		ProvisionGroups: []model.ProvisionGroup{
			{
				MinRequired: 1,
				Message:     "Encryption keys required for encrypted games",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionKeys, "aes_keys.txt", "AES keys").WithImportViaUI(),
				},
			},
			{
				MinRequired: 0,
				Message:     "Seed database (optional, for some encrypted games)",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionKeys, "seeddb.bin", "Seed DB").WithImportViaUI(),
				},
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "azahar",
			GenericName: "Nintendo 3DS Emulator",
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
		PathUsage: model.PathUsage{
			UsesScreenshotsDir: true,
			OpaqueContents:     "NAND, SDMC (saves, installed titles)",
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "azahar/qt-config.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// Azahar (Citra fork) uses Qt INI config
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"Data%20Storage", "nand_directory"}, Value: store.EmulatorOpaqueDir(model.EmulatorIDAzahar) + "/nand"},
			{Path: []string{"Data%20Storage", "sdmc_directory"}, Value: store.EmulatorOpaqueDir(model.EmulatorIDAzahar) + "/sdmc"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: store.SystemRomsDir(model.SystemIDN3DS)},
			{Path: []string{"UI", "Screenshots\\screenshot_path"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDAzahar)},
		},
	}}, nil
}
