package eden

import "github.com/fnune/kyaraben/internal/model"

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
// Config is at ~/.config/eden/qt-config.ini (or similar)
// See FILESYSTEM.md for the opaque directory pattern.
var configTarget = model.ConfigTarget{
	RelPath: "eden/qt-config.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// Eden uses an opaque data directory at ~/Emulation/opaque/eden/
	// We configure Eden to use this as its NAND/user directory.
	// The ROMs directory is set separately.
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			// Set the game directory for ROM browsing
			{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: store.SystemRomsDir(model.SystemSwitch)},
			// Set NAND directory to opaque location
			{Path: []string{"Data%20Storage", "nand_directory"}, Value: store.EmulatorOpaqueDir(model.EmulatorEden)},
			// Screenshots go to the structured location
			{Path: []string{"UI", "Screenshots\\screenshot_path"}, Value: store.SystemScreenshotsDir(model.SystemSwitch)},
		},
	}}, nil
}
