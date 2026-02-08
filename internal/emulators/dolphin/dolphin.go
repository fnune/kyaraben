// BIOS hash data compiled from:
// - Libretro documentation (https://docs.libretro.com)
// - Dolphin Emulator forums and documentation
package dolphin

import (
	"path/filepath"

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
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "usa/ipl.bin",
					Description: "USA",
					Hashes: []string{
						"6dac1f2a14f659a1a7fbf749892b4e41",
						"019e39822a9ca3029124f74dd4d55ac4",
					},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "eur/ipl.bin",
					Description: "Europe",
					Hashes: []string{
						"db92574caab77a7ec99d4605fd6f2450",
						"0cdda509e2da83c85bfe423dd87346cc",
					},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "jap/ipl.bin",
					Description: "Japan",
					Hashes: []string{
						"fc924a7c879b661abc37cec4f018fdf3",
						"81df278301dc7bdf57bb760d7393ab4d",
					},
				},
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
				if opts.Fullscreen {
					cmd += " -C Dolphin.Display.Fullscreen=True"
				}
				cmd += " -e %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			UsesScreenshotsDir: true,
			OpaqueContents:     "config, GC memory cards, Wii NAND, savestates",
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

// LaunchArgs implements model.LaunchArgsProvider.
// Dolphin's -u flag sets the user directory, which contains config, saves, and state.
// This avoids conflicts with system-wide Dolphin installations and allows kyaraben
// to fully manage Dolphin's data within the opaque directory.
func (c *Config) LaunchArgs(store model.StoreReader) []string {
	return []string{"-u", store.EmulatorOpaqueDir(model.EmulatorIDDolphin)}
}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// With -u flag, Dolphin stores everything in the user directory:
	// - Config at <user_dir>/Config/Dolphin.ini
	// - GC saves at <user_dir>/GC/
	// - Wii NAND at <user_dir>/Wii/
	// - Screenshots at <user_dir>/ScreenShots/
	//
	// We only need to configure ROM paths and screenshot location.
	opaqueDir := store.EmulatorOpaqueDir(model.EmulatorIDDolphin)

	configTarget := model.ConfigTarget{
		RelPath: filepath.Join(opaqueDir, "Config", "Dolphin.ini"),
		Format:  model.ConfigFormatINI,
		BaseDir: model.ConfigBaseDirOpaqueDir,
	}

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
