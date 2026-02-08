package eden

import (
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDEden,
		Name:    "Eden",
		Systems: []model.SystemID{model.SystemIDSwitch},
		Package: model.AppImageRef("eden"),
		// Eden requires firmware and keys which must be installed via the Eden UI.
		// Kyaraben can verify their presence but cannot automatically provision them.
		// See: https://eden-emu.dev/
		ProvisionGroups: []model.ProvisionGroup{
			{
				MinRequired: 1,
				Message:     "Keys required",
				Provisions: []model.Provision{{
					Kind:        model.ProvisionKeys,
					Filename:    "prod.keys",
					ImportViaUI: true,
				}},
			},
			{
				MinRequired: 0,
				Message:     "Firmware (optional)",
				Provisions: []model.Provision{{
					Kind:        model.ProvisionFirmware,
					Filename:    "firmware",
					ImportViaUI: true,
				}},
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
			model.StatePersistent, // NAND, shader cache, etc.
		},
		Launcher: model.LauncherInfo{
			Binary:      "eden",
			GenericName: "Nintendo Switch Emulator",
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
			UsesScreenshotsDir: true,
			OpaqueContents:     "NAND, SDMC, keys, saves, savestates, shader cache",
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

// LaunchArgs implements model.LaunchArgsProvider.
// Eden's -r flag sets the root data directory where all data is stored:
// config, nand, sdmc, keys, etc.
func (c *Config) LaunchArgs(store model.StoreReader) []string {
	return []string{"-r", store.EmulatorOpaqueDir(model.EmulatorIDEden)}
}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// With -r flag, Eden stores everything in the root data directory:
	// - Config at <root>/config/qt-config.ini
	// - NAND at <root>/nand/
	// - SDMC at <root>/sdmc/
	// - Keys at <root>/keys/
	//
	// We only need to configure ROM paths and screenshot location.
	opaqueDir := store.EmulatorOpaqueDir(model.EmulatorIDEden)

	configTarget := model.ConfigTarget{
		RelPath: filepath.Join(opaqueDir, "config", "qt-config.ini"),
		Format:  model.ConfigFormatINI,
		BaseDir: model.ConfigBaseDirOpaqueDir,
	}

	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"UI", "Screenshots\\screenshot_path"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDEden)},
			{Path: []string{"UI", "Paths\\gamedirs\\size"}, Value: "1"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\deep_scan"}, Value: "false"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\expanded"}, Value: "true"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: store.SystemRomsDir(model.SystemIDSwitch)},
		},
	}}, nil
}
