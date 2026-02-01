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
		Provisions: []model.Provision{
			{
				ID:          "switch-keys",
				Kind:        model.ProvisionKeys,
				Filename:    "prod.keys",
				Description: "Keys",
				Required:    true,
			},
			{
				ID:          "switch-firmware",
				Kind:        model.ProvisionFirmware,
				Filename:    "firmware",
				Description: "Firmware",
				Required:    false,
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
		BaseDir: model.ConfigBaseDirAbsolute,
	}

	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			// Screenshots go to system screenshots directory
			{Path: []string{"UI", "Screenshots\\screenshot_path"}, Value: store.SystemScreenshotsDir(model.SystemIDSwitch)},

			// Game directories - point to ROMs folder
			{Path: []string{"UI", "Paths\\gamedirs\\size"}, Value: "1"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\deep_scan"}, Value: "false"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\expanded"}, Value: "true"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: store.SystemRomsDir(model.SystemIDSwitch)},
		},
		// Config lives inside opaque dir (via -r flag), so don't track in manifest
		Untracked: true,
	}}, nil
}
