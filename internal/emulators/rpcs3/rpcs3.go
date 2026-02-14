package rpcs3

import (
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRPCS3,
		Name:    "RPCS3",
		Systems: []model.SystemID{model.SystemIDPS3},
		Package: model.AppImageRef("rpcs3"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "Firmware required (provides system libraries and OS)",
			Provisions: []model.Provision{
				model.FileProvision(model.ProvisionFirmware, "PS3UPDAT.PUP", "Official firmware").WithImportViaUI(),
			},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "rpcs3",
			GenericName: "PlayStation 3 Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " --fullscreen"
				}
				cmd += " %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			UsesSavesDir:       true,
			UsesStatesDir:      true,
			UsesScreenshotsDir: true,
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var vfsTarget = model.ConfigTarget{
	RelPath: "rpcs3/vfs.yml",
	Format:  model.ConfigFormatYAML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var guiTarget = model.ConfigTarget{
	RelPath: "rpcs3/GuiConfigs/CurrentSettings.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{
		{
			Target: vfsTarget,
			Entries: []model.ConfigEntry{
				{Path: []string{"/dev_hdd0/"}, Value: store.SystemSavesDir(model.SystemIDPS3) + "/"},
				{Path: []string{"/games/"}, Value: store.SystemRomsDir(model.SystemIDPS3)},
			},
		},
		{
			Target: guiTarget,
			Entries: []model.ConfigEntry{
				{Path: []string{"main_window", "infoBoxEnabledWelcome"}, Value: "false"},
				{Path: []string{"main_window", "confirmationBoxExitGame"}, Value: "false"},
				{Path: []string{"Meta", "checkUpdateStart"}, Value: "false"},
			},
		},
	}, nil
}

func (c *Config) Symlinks(store model.StoreReader, resolver model.BaseDirResolver) ([]model.SymlinkSpec, error) {
	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return nil, err
	}
	rpcs3Dir := filepath.Join(configDir, "rpcs3")

	return []model.SymlinkSpec{
		{Source: filepath.Join(rpcs3Dir, "savestates"), Target: store.EmulatorStatesDir(model.EmulatorIDRPCS3)},
		{Source: filepath.Join(rpcs3Dir, "screenshots"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDRPCS3)},
	}, nil
}
