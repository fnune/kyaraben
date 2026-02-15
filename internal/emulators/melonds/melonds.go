// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package melonds

import (
	"fmt"

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

	if cc := ctx.ControllerConfig; cc != nil {
		entries = append(entries, padEntries(cc)...)
		entries = append(entries, hotkeyEntries(cc)...)
	}

	return model.GenerateResult{
		Patches: []model.ConfigPatch{{Target: configTarget, Entries: entries}},
	}, nil
}

// melonDS uses special hat encoding for D-pad directions.
const (
	hatUp    = 257
	hatRight = 258
	hatDown  = 260
	hatLeft  = 264
)

func padEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	south, east, west, north := cc.FaceButtons()
	section := "Instance0"
	return []model.ConfigEntry{
		{Path: []string{section, "Joy_A"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[south])},
		{Path: []string{section, "Joy_B"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[east])},
		{Path: []string{section, "Joy_X"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[west])},
		{Path: []string{section, "Joy_Y"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[north])},
		{Path: []string{section, "Joy_Select"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[model.ButtonBack])},
		{Path: []string{section, "Joy_Start"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[model.ButtonStart])},
		{Path: []string{section, "Joy_L"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[model.ButtonLeftShoulder])},
		{Path: []string{section, "Joy_R"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[model.ButtonRightShoulder])},
		{Path: []string{section, "Joy_Up"}, Value: fmt.Sprintf("%d", hatUp)},
		{Path: []string{section, "Joy_Down"}, Value: fmt.Sprintf("%d", hatDown)},
		{Path: []string{section, "Joy_Left"}, Value: fmt.Sprintf("%d", hatLeft)},
		{Path: []string{section, "Joy_Right"}, Value: fmt.Sprintf("%d", hatRight)},
		{Path: []string{section, "JoystickID"}, Value: "0"},
	}
}

func hotkeyEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	hk := cc.Hotkeys
	section := "Instance0"
	var entries []model.ConfigEntry

	type mapping struct {
		key     string
		binding model.HotkeyBinding
	}
	mappings := []mapping{
		{"HKJoy_FastForward", hk.FastForward},
		{"HKJoy_Pause", hk.Pause},
		{"HKJoy_FullscreenToggle", hk.ToggleFullscreen},
	}
	for _, m := range mappings {
		if len(m.binding.Buttons) > 0 {
			entries = append(entries, model.ConfigEntry{
				Path:  []string{section, m.key},
				Value: fmt.Sprintf("%d", model.SDLButtonIndex[m.binding.Buttons[len(m.binding.Buttons)-1]]),
			})
		}
	}
	return entries
}
