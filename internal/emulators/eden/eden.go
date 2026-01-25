package eden

import (
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorEden,
		Name:    "Eden",
		Systems: []model.SystemID{model.SystemSwitch},
		Package: model.GitHubAppImageRef(
			"eden",
			"eden-emulator",
			"Releases",
			"v0.0.4",
			map[string]string{
				"x86_64":  "Eden-Linux-v0.0.4-amd64-clang-pgo.AppImage",
				"aarch64": "Eden-Linux-v0.0.4-aarch64-clang-pgo.AppImage",
			},
			// TODO: These hashes need to be computed by downloading the AppImages
			// and running: sha256sum <file>
			// Then convert to base32 for Nix: nix-hash --to-base32 --type sha256 <hex-hash>
			map[string]string{
				"x86_64":  "0000000000000000000000000000000000000000000000000000",
				"aarch64": "0000000000000000000000000000000000000000000000000000",
			},
		),
		// Eden requires firmware and keys which must be installed via the Eden UI.
		// Kyaraben can verify their presence but cannot automatically provision them.
		// See: https://eden-emu.dev/
		Provisions: []model.Provision{
			{
				ID:          "switch-keys",
				Kind:        model.ProvisionKeys,
				Filename:    "prod.keys",
				Description: "Nintendo Switch encryption keys (install via Eden UI)",
				Required:    true,
			},
			{
				ID:          "switch-firmware",
				Kind:        model.ProvisionFirmware,
				Filename:    "firmware",
				Description: "Nintendo Switch firmware (install via Eden UI)",
				Required:    false,
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
			model.StatePersistent, // NAND, shader cache, etc.
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
	opaqueDir := store.EmulatorOpaqueDir(model.EmulatorEden)
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
			{Path: []string{"UI", "Screenshots\\screenshot_path"}, Value: store.SystemScreenshotsDir(model.SystemSwitch)},

			// Game directories - point to ROMs folder
			// Note: Eden uses a complex gamedirs format, this sets the first entry
			{Path: []string{"UI", "Paths\\gamedirs\\size"}, Value: "1"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\deep_scan"}, Value: "false"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\expanded"}, Value: "true"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: store.SystemRomsDir(model.SystemSwitch)},
		},
	}}, nil
}
