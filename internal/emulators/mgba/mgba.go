// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package mgba

import (
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDMGBA,
		Name:            "mGBA",
		Systems:         []model.SystemID{model.SystemIDGB, model.SystemIDGBC, model.SystemIDGBA},
		Package:         model.AppImageRef("mgba"),
		ProvisionGroups: buildMgbaProvisionGroups(),
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "mgba",
			GenericName: "Game Boy Advance Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " -f"
				}
				if opts.SavesDir != "" {
					cmd += " -C savegamePath=" + opts.SavesDir
				}
				cmd += " %ROM%"
				return cmd
			},
		},
		PathUsage: model.StandardPathUsage(),
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "mgba/config.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store
	biosPaths := mgbaBiosPaths(store)
	return model.GenerateResult{
		Patches: []model.ConfigPatch{{
			Target: configTarget,
			Entries: []model.ConfigEntry{
				{Path: []string{"bios"}, Value: biosPaths["bios"]},
				{Path: []string{"gb.bios"}, Value: biosPaths["gb.bios"]},
				{Path: []string{"gbc.bios"}, Value: biosPaths["gbc.bios"]},
				{Path: []string{"sgb.bios"}, Value: biosPaths["sgb.bios"]},
				{Path: []string{"gba.bios"}, Value: biosPaths["gba.bios"]},
				{Path: []string{"ports.qt", "bios"}, Value: biosPaths["bios"]},
				{Path: []string{"ports.qt", "gb.bios"}, Value: biosPaths["gb.bios"]},
				{Path: []string{"ports.qt", "gbc.bios"}, Value: biosPaths["gbc.bios"]},
				{Path: []string{"ports.qt", "sgb.bios"}, Value: biosPaths["sgb.bios"]},
				{Path: []string{"ports.qt", "gba.bios"}, Value: biosPaths["gba.bios"]},
				{Path: []string{"ports.qt", "useBios"}, Value: "1"},
				{Path: []string{"ports.qt", "savegamePath"}, Value: store.SystemSavesDir(model.SystemIDGBA)},
				{Path: []string{"ports.qt", "savestatePath"}, Value: store.EmulatorStatesDir(model.EmulatorIDMGBA)},
				{Path: []string{"ports.qt", "screenshotPath"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDMGBA)},
				{Path: []string{"ports.qt", "showLibrary"}, Value: "1"},
			},
		}},
	}, nil
}

func mgbaBiosPaths(store model.StoreReader) map[string]string {
	return map[string]string{
		"bios":     filepath.Join(store.SystemBiosDir(model.SystemIDGBA), "gba_bios.bin"),
		"gb.bios":  filepath.Join(store.SystemBiosDir(model.SystemIDGB), "gb_bios.bin"),
		"gbc.bios": filepath.Join(store.SystemBiosDir(model.SystemIDGBC), "gbc_bios.bin"),
		"sgb.bios": filepath.Join(store.SystemBiosDir(model.SystemIDGB), "sgb_bios.bin"),
		"gba.bios": filepath.Join(store.SystemBiosDir(model.SystemIDGBA), "gba_bios.bin"),
	}
}

const (
	// md5 hashes for the BIOS files tracked by kyaraben and mGBA.
	gbBiosHash  = "32fbbd84168d3482956eb3c5051637f5"
	gbcBiosHash = "dbfce9db9deaa2567f6a84fde55f9680"
	sgbBiosHash = "d574d4f9c12f305074798f54c091a8b4"
	gbaBiosHash = "a860e8c0b6d573d191e4ec7db1b1e4f6"
)

func buildMgbaProvisionGroups() []model.ProvisionGroup {
	return []model.ProvisionGroup{
		{
			MinRequired: 0,
			Message:     "Game Boy BIOS (optional)",
			BaseDir: func(store model.StoreReader, sys model.SystemID) string {
				return store.SystemBiosDir(model.SystemIDGB)
			},
			Provisions: []model.Provision{
				model.HashedProvision(model.ProvisionBIOS, "gb_bios.bin", "Game Boy BIOS", []string{gbBiosHash}).ForSystems(model.SystemIDGB),
			},
		},
		{
			MinRequired: 0,
			Message:     "Game Boy Color BIOS (optional)",
			BaseDir: func(store model.StoreReader, sys model.SystemID) string {
				return store.SystemBiosDir(model.SystemIDGBC)
			},
			Provisions: []model.Provision{
				model.HashedProvision(model.ProvisionBIOS, "gbc_bios.bin", "Game Boy Color BIOS", []string{gbcBiosHash}).ForSystems(model.SystemIDGBC),
			},
		},
		{
			MinRequired: 0,
			Message:     "Super Game Boy BIOS (optional)",
			BaseDir: func(store model.StoreReader, sys model.SystemID) string {
				return store.SystemBiosDir(model.SystemIDGB)
			},
			Provisions: []model.Provision{
				model.HashedProvision(model.ProvisionBIOS, "sgb_bios.bin", "Super Game Boy BIOS", []string{sgbBiosHash}).ForSystems(model.SystemIDGB),
			},
		},
		{
			MinRequired: 0,
			Message:     "Game Boy Advance BIOS (optional, enables boot animation and improves compatibility)",
			BaseDir: func(store model.StoreReader, sys model.SystemID) string {
				return store.SystemBiosDir(model.SystemIDGBA)
			},
			Provisions: []model.Provision{
				model.HashedProvision(model.ProvisionBIOS, "gba_bios.bin", "", []string{gbaBiosHash}).ForSystems(model.SystemIDGBA),
			},
		},
	}
}
