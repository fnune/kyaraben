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
		// Wii system menu optional, GameCube IPL optional
		Provisions: nil,
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
				return opts.BinaryPath + " -e %ROM%"
			},
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
			{Path: []string{"General", "DumpPath"}, Value: store.SystemScreenshotsDir(model.SystemIDGameCube)},
		},
	}}, nil
}
