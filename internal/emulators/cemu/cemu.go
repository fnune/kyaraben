package cemu

import (
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

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
			Provisions: []model.Provision{
				model.FileProvision(model.ProvisionKeys, "keys.txt", "Wii U keys").WithImportViaUI(),
			},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
		},
		Launcher: model.LauncherInfo{
			Binary:      "cemu",
			GenericName: "Wii U Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if len(opts.LaunchArgs) > 0 {
					cmd += " " + strings.Join(opts.LaunchArgs, " ")
				}
				if opts.Fullscreen {
					cmd += " -f"
				}
				cmd += " -g %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			UsesSavesDir:       true,
			UsesScreenshotsDir: true,
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

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"content", "GamePaths", "Entry"}, Value: store.SystemRomsDir(model.SystemIDWiiU)},
			{Path: []string{"content", "check_update"}, Value: "false"},
		},
	}}, nil
}

func (c *Config) Symlinks(store model.StoreReader, resolver model.BaseDirResolver) ([]model.SymlinkSpec, error) {
	dataDir, err := resolver.UserDataDir()
	if err != nil {
		return nil, err
	}
	cemuDir := filepath.Join(dataDir, "Cemu")

	return []model.SymlinkSpec{
		{Source: filepath.Join(cemuDir, "mlc01", "usr", "save", "00050000"), Target: store.SystemSavesDir(model.SystemIDWiiU)},
		{Source: filepath.Join(cemuDir, "screenshots"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDCemu)},
	}, nil
}
