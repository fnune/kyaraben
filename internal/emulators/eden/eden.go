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

// Eden uses an opaque data directory structure.
// Config is at ~/.config/eden/qt-config.ini
// Data directories (nand, sdmc) are configured to live in ~/Emulation/opaque/eden/
// See FILESYSTEM.md for the opaque directory pattern.
var configTarget = model.ConfigTarget{
	RelPath: "eden/qt-config.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// Eden uses an opaque data directory at ~/Emulation/opaque/eden/
	// Within this, we configure nand/ and sdmc/ subdirectories.
	opaqueDir := store.EmulatorOpaqueDir(model.EmulatorIDEden)
	nandDir := filepath.Join(opaqueDir, "nand")
	sdmcDir := filepath.Join(opaqueDir, "sdmc")

	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			// Data storage paths - NAND and SDMC in opaque directory
			{Path: []string{"Data%20Storage", "nand_directory"}, Value: nandDir},
			{Path: []string{"Data%20Storage", `nand_directory\default`}, Value: "false"},
			{Path: []string{"Data%20Storage", "sdmc_directory"}, Value: sdmcDir},
			{Path: []string{"Data%20Storage", `sdmc_directory\default`}, Value: "false"},

			// UI paths
			{Path: []string{"UI", "Screenshots\\screenshot_path"}, Value: store.SystemScreenshotsDir(model.SystemIDSwitch)},

			// Game directories - point to ROMs folder
			// Note: Eden uses a complex gamedirs format, this sets the first entry
			{Path: []string{"UI", "Paths\\gamedirs\\size"}, Value: "1"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\deep_scan"}, Value: "false"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\expanded"}, Value: "true"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: store.SystemRomsDir(model.SystemIDSwitch)},
		},
	}}, nil
}
