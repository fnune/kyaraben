// BIOS hash data compiled from:
// - Libretro documentation (https://docs.libretro.com)
package mgba

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDMGBA,
		Name:    "mGBA",
		Systems: []model.SystemID{model.SystemIDGB, model.SystemIDGBC, model.SystemIDGBA},
		Package: model.AppImageRef("mgba"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 0,
			Message:     "BIOS (optional, mGBA has HLE fallback)",
			Provisions: []model.Provision{{
				Kind:     model.ProvisionBIOS,
				Filename: "gba_bios.bin",
				Hashes:   []string{"a860e8c0b6d573d191e4ec7db1b1e4f6"},
			}},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "mgba",
			GenericName: "Game Boy Advance Emulator",
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
	RelPath: "mgba/config.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"ports.qt", "bios"}, Value: store.SystemBiosDir(model.SystemIDGBA) + "/gba_bios.bin"},
			{Path: []string{"ports.qt", "savegamePath"}, Value: store.SystemSavesDir(model.SystemIDGBA)},
			{Path: []string{"ports.qt", "savestatePath"}, Value: store.EmulatorStatesDir(model.EmulatorIDMGBA)},
			{Path: []string{"ports.qt", "screenshotPath"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDMGBA)},
			{Path: []string{"ports.qt", "showLibrary"}, Value: "1"},
		},
	}}, nil
}
