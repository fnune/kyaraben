package azahar

import (
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDAzahar,
		Name:    "Azahar",
		Systems: []model.SystemID{model.SystemIDN3DS},
		Package: model.AppImageRef("azahar"),
		ProvisionGroups: nil,
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
			UsesSavesDir:       true,
			UsesStatesDir:      true,
			UsesScreenshotsDir: true,
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "azahar-emu/qt-config.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"Data%20Storage", "use_custom_storage"}, Value: "true"},
			{Path: []string{"Data%20Storage", "use_custom_storage\\default"}, Value: "false"},
			{Path: []string{"Data%20Storage", "sdmc_directory"}, Value: store.SystemSavesDir(model.SystemIDN3DS) + "/"},
			{Path: []string{"Data%20Storage", "sdmc_directory\\default"}, Value: "false"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: "INSTALLED"},
			{Path: []string{"UI", "Paths\\gamedirs\\2\\path"}, Value: "SYSTEM"},
			{Path: []string{"UI", "Paths\\gamedirs\\3\\path"}, Value: store.SystemRomsDir(model.SystemIDN3DS)},
			{Path: []string{"UI", "Paths\\gamedirs\\size"}, Value: "3"},
			{Path: []string{"UI", "Paths\\screenshotPath"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDAzahar)},
			{Path: []string{"UI", "Paths\\screenshotPath\\default"}, Value: "false"},
		},
	}}, nil
}

func (c *Config) Symlinks(store model.StoreReader, resolver model.BaseDirResolver) ([]model.SymlinkSpec, error) {
	dataDir, err := resolver.UserDataDir()
	if err != nil {
		return nil, err
	}
	azaharDir := filepath.Join(dataDir, "azahar-emu")

	return []model.SymlinkSpec{
		{Source: filepath.Join(azaharDir, "states"), Target: store.EmulatorStatesDir(model.EmulatorIDAzahar)},
	}, nil
}
