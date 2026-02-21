// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package melonds

import (
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDMelonDS,
		Name:    "melonDS",
		Systems: []model.SystemID{model.SystemIDNDS},
		Package: model.AppImageRef("melonds"),
		ProvisionGroups: []model.ProvisionGroup{
			{
				MinRequired: 0,
				Message:     "Native BIOS (optional, improves accuracy and enables DS menu)",
				Provisions: []model.Provision{
					model.HashedProvision(model.ProvisionBIOS, "bios7.bin", "ARM7", []string{"df692a80a5b1bc90728bc3dfc76cd948"}),
					model.HashedProvision(model.ProvisionBIOS, "bios9.bin", "ARM9", []string{"a392174eb3e572fed6447e956bde4b25"}),
				},
			},
			{
				MinRequired: 0,
				Message:     "Native firmware (optional, required for DSi mode and WiFi)",
				Provisions: []model.Provision{
					model.HashedProvision(model.ProvisionFirmware, "firmware.bin", "", []string{"e45033d9b0fa6b0de071292bba7c9d13"}),
				},
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "melonds",
			GenericName: "Nintendo DS Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " -f"
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
	RelPath: "melonDS/melonDS.toml",
	Format:  model.ConfigFormatTOML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store
	entries := []model.ConfigEntry{
		{Path: []string{"DS", "BIOS9Path"}, Value: store.SystemBiosDir(model.SystemIDNDS) + "/bios9.bin"},
		{Path: []string{"DS", "BIOS7Path"}, Value: store.SystemBiosDir(model.SystemIDNDS) + "/bios7.bin"},
		{Path: []string{"DS", "FirmwarePath"}, Value: store.SystemBiosDir(model.SystemIDNDS) + "/firmware.bin"},
		{Path: []string{"Instance0", "SaveFilePath"}, Value: store.SystemSavesDir(model.SystemIDNDS)},
		{Path: []string{"Instance0", "SavestatePath"}, Value: store.EmulatorStatesDir(model.EmulatorIDMelonDS)},
		{Path: []string{"Instance0", "ScreenshotPath"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDMelonDS)},
		{Path: []string{"Instance0", "LastROMFolder"}, Value: store.SystemRomsDir(model.SystemIDNDS)},
	}

	// Controller config disabled: melonDS standalone uses raw joystick indices
	// and single-button hotkeys. Plan is to migrate to RetroArch melonDS core.
	_ = ctx.ControllerConfig

	return model.GenerateResult{
		Patches: []model.ConfigPatch{{Target: configTarget, Entries: entries}},
	}, nil
}
