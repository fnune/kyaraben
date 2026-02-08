package cemu

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDCemu,
		Name:    "Cemu",
		Systems: []model.SystemID{model.SystemIDWiiU},
		Package: model.AppImageRef("cemu"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "Encryption keys required",
			Provisions: []model.Provision{{
				Kind:        model.ProvisionKeys,
				Filename:    "keys.txt",
				Description: "Wii U keys",
				ImportViaUI: true,
			}},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
		},
		Launcher: model.LauncherInfo{
			Binary:      "cemu",
			GenericName: "Wii U Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " -f"
				}
				cmd += " -g %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			OpaqueContents: "MLC (saves, updates, DLC)",
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "Cemu/settings.xml",
	Format:  model.ConfigFormatXML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

// LaunchArgs implements model.LaunchArgsProvider.
// Cemu's -mlc flag sets the MLC directory which stores saves, updates, and DLC.
// This is separate from the config file location.
func (c *Config) LaunchArgs(store model.StoreReader) []string {
	return []string{"-mlc", store.EmulatorOpaqueDir(model.EmulatorIDCemu)}
}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// Cemu uses XML config for settings. MLC path is set via -mlc CLI flag,
	// but we still configure ROM paths via the settings file.
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"content", "GamePaths", "Entry"}, Value: store.SystemRomsDir(model.SystemIDWiiU)},
		},
	}}, nil
}
