package ppsspp

import (
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDPPSSPP,
		Name:            "PPSSPP",
		Systems:         []model.SystemID{model.SystemIDPSP},
		Package:         model.AppImageRef("ppsspp"),
		ProvisionGroups: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "ppsspp",
			GenericName: "PlayStation Portable Emulator",
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

var configTarget = model.ConfigTarget{
	RelPath: "ppsspp/PSP/SYSTEM/ppsspp.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"General", "CurrentDirectory"}, Value: store.SystemRomsDir(model.SystemIDPSP)},
		},
	}}, nil
}

func (c *Config) Symlinks(store model.StoreReader, resolver model.BaseDirResolver) ([]model.SymlinkSpec, error) {
	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return nil, err
	}
	pspDir := filepath.Join(configDir, "ppsspp", "PSP")

	return []model.SymlinkSpec{
		{Source: filepath.Join(pspDir, "SAVEDATA"), Target: store.SystemSavesDir(model.SystemIDPSP)},
		{Source: filepath.Join(pspDir, "PPSSPP_STATE"), Target: store.EmulatorStatesDir(model.EmulatorIDPPSSPP)},
		{Source: filepath.Join(pspDir, "SCREENSHOT"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDPPSSPP)},
	}, nil
}
