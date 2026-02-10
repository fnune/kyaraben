// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package dolphin

import (
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDDolphin,
		Name:    "Dolphin",
		Systems: []model.SystemID{model.SystemIDGameCube, model.SystemIDWii},
		Package: model.AppImageRef("dolphin"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 0,
			Message:     "GameCube IPL (optional, enables boot animation and system fonts)",
			Provisions: []model.Provision{
				model.HashedProvision(model.ProvisionBIOS, "usa/ipl.bin", "USA", []string{"6dac1f2a14f659a1a7fbf749892b4e41", "019e39822a9ca3029124f74dd4d55ac4"}),
				model.HashedProvision(model.ProvisionBIOS, "eur/ipl.bin", "Europe", []string{"db92574caab77a7ec99d4605fd6f2450", "0cdda509e2da83c85bfe423dd87346cc"}),
				model.HashedProvision(model.ProvisionBIOS, "jap/ipl.bin", "Japan", []string{"fc924a7c879b661abc37cec4f018fdf3", "81df278301dc7bdf57bb760d7393ab4d"}),
			},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "dolphin",
			GenericName: "GameCube/Wii Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if len(opts.LaunchArgs) > 0 {
					cmd += " " + strings.Join(opts.LaunchArgs, " ")
				}
				if opts.Fullscreen {
					cmd += " -C Dolphin.Display.Fullscreen=True"
				}
				cmd += " -e %ROM%"
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

type Config struct{}

var configTarget = model.ConfigTarget{
	RelPath: "dolphin-emu/Dolphin.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"General", "ISOPath0"}, Value: store.SystemRomsDir(model.SystemIDGameCube)},
			{Path: []string{"General", "ISOPath1"}, Value: store.SystemRomsDir(model.SystemIDWii)},
			{Path: []string{"General", "ISOPaths"}, Value: "2"},
			{Path: []string{"General", "DumpPath"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDDolphin)},
		},
	}}, nil
}

func (c *Config) Symlinks(store model.StoreReader, resolver model.BaseDirResolver) ([]model.SymlinkSpec, error) {
	dataDir, err := resolver.UserDataDir()
	if err != nil {
		return nil, err
	}
	dolphinDir := filepath.Join(dataDir, "dolphin-emu")

	return []model.SymlinkSpec{
		{Source: filepath.Join(dolphinDir, "GC"), Target: store.SystemSavesDir(model.SystemIDGameCube)},
		{Source: filepath.Join(dolphinDir, "Wii"), Target: store.SystemSavesDir(model.SystemIDWii)},
		{Source: filepath.Join(dolphinDir, "StateSaves"), Target: store.EmulatorStatesDir(model.EmulatorIDDolphin)},
		{Source: filepath.Join(dolphinDir, "ScreenShots"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDDolphin)},
	}, nil
}
