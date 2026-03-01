// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package flycast

import (
	"fmt"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDFlycast,
		Name:    "Flycast",
		Systems: []model.SystemID{model.SystemIDDreamcast},
		Package: model.AppImageRef("flycast"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 0,
			Message:     "BIOS files (optional, enables boot animation)",
			Provisions:  dreamcastBIOSProvisions,
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "flycast",
			GenericName: "Sega Dreamcast Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " -config window:fullscreen=yes"
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
	RelPath: "flycast/emu.cfg",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

var mappingTarget = model.ConfigTarget{
	RelPath: "flycast/mappings/SDL_Steam Deck Controller.cfg",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store
	biosDir := store.SystemBiosDir(model.SystemIDDreamcast)
	savesDir := store.SystemSavesDir(model.SystemIDDreamcast)

	patches := []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"config", "Flycast.DataPath"}, Value: biosDir},
			{Path: []string{"config", "Dreamcast.BiosPath"}, Value: biosDir},
			{Path: []string{"config", "Dreamcast.ContentPath"}, Value: store.SystemRomsDir(model.SystemIDDreamcast)},
			{Path: []string{"config", "Dreamcast.SavePath"}, Value: savesDir},
			{Path: []string{"config", "Dreamcast.VMUPath"}, Value: savesDir},
			{Path: []string{"config", "Dreamcast.SavestatePath"}, Value: store.EmulatorStatesDir(model.EmulatorIDFlycast)},
		},
	}}

	if cc := ctx.ControllerConfig; cc != nil {
		patches = append(patches, model.ConfigPatch{
			Target:  mappingTarget,
			Entries: mappingEntries(cc),
		})
	}

	return model.GenerateResult{Patches: patches}, nil
}

// Raw joystick button indices for Xbox 360/Steam Deck controller.
// Flycast uses raw indices, not SDL GameController indices.
var rawJoystickIndex = map[model.SDLButton]int{
	model.ButtonA:             0,
	model.ButtonB:             1,
	model.ButtonX:             2,
	model.ButtonY:             3,
	model.ButtonLeftShoulder:  4,
	model.ButtonRightShoulder: 5,
	model.ButtonBack:          6,
	model.ButtonStart:         7,
	model.ButtonGuide:         8,
	model.ButtonLeftStick:     9,
	model.ButtonRightStick:    10,
}

// DPad uses HAT indices in Flycast (256 + direction).
const (
	hatUp    = 256
	hatDown  = 257
	hatLeft  = 258
	hatRight = 259
)

func mappingEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	fb := cc.FaceButtons(model.SystemIDDreamcast)

	// Flycast uses axis:action format for analog and button:action for digital.
	// Dreamcast: A=south, B=east, X=west, Y=north.
	// Dreamcast has 6 face buttons: A, B, X, Y plus Z and C (digital triggers behind A/B).
	// L/R shoulders map to Z/C since Steam Deck has analog triggers for L/R.
	entries := []model.ConfigEntry{
		{Path: []string{"digital", "bind0"}, Value: fmt.Sprintf("%d:btn_a", rawJoystickIndex[fb.South])},
		{Path: []string{"digital", "bind1"}, Value: fmt.Sprintf("%d:btn_b", rawJoystickIndex[fb.East])},
		{Path: []string{"digital", "bind2"}, Value: fmt.Sprintf("%d:btn_x", rawJoystickIndex[fb.West])},
		{Path: []string{"digital", "bind3"}, Value: fmt.Sprintf("%d:btn_y", rawJoystickIndex[fb.North])},
		{Path: []string{"digital", "bind4"}, Value: fmt.Sprintf("%d:btn_z", rawJoystickIndex[model.ButtonLeftShoulder])},
		{Path: []string{"digital", "bind5"}, Value: fmt.Sprintf("%d:btn_c", rawJoystickIndex[model.ButtonRightShoulder])},
		{Path: []string{"digital", "bind6"}, Value: fmt.Sprintf("%d:btn_start", rawJoystickIndex[model.ButtonStart])},
		{Path: []string{"digital", "bind7"}, Value: fmt.Sprintf("%d:btn_dpad2_up", rawJoystickIndex[model.ButtonGuide])},
		{Path: []string{"digital", "bind8"}, Value: fmt.Sprintf("%d:btn_dpad1_up", hatUp)},
		{Path: []string{"digital", "bind9"}, Value: fmt.Sprintf("%d:btn_dpad1_down", hatDown)},
		{Path: []string{"digital", "bind10"}, Value: fmt.Sprintf("%d:btn_dpad1_left", hatLeft)},
		{Path: []string{"digital", "bind11"}, Value: fmt.Sprintf("%d:btn_dpad1_right", hatRight)},
		{Path: []string{"analog", "bind0"}, Value: "0-:btn_analog_left"},
		{Path: []string{"analog", "bind1"}, Value: "0+:btn_analog_right"},
		{Path: []string{"analog", "bind2"}, Value: "1-:btn_analog_up"},
		{Path: []string{"analog", "bind3"}, Value: "1+:btn_analog_down"},
		{Path: []string{"analog", "bind4"}, Value: "2+:btn_trigger_left"},
		{Path: []string{"analog", "bind5"}, Value: "3-:axis2_left"},
		{Path: []string{"analog", "bind6"}, Value: "3+:axis2_right"},
		{Path: []string{"analog", "bind7"}, Value: "4-:axis2_up"},
		{Path: []string{"analog", "bind8"}, Value: "4+:axis2_down"},
		{Path: []string{"analog", "bind9"}, Value: "5+:btn_trigger_right"},
		{Path: []string{"emulator", "dead_zone"}, Value: "10"},
		{Path: []string{"emulator", "mapping_name"}, Value: "Steam Deck Controller"},
		{Path: []string{"emulator", "rumble_power"}, Value: "100"},
		{Path: []string{"emulator", "saturation"}, Value: "100"},
		{Path: []string{"emulator", "triggers"}, Value: "2,5"},
		{Path: []string{"emulator", "version"}, Value: "4"},
	}

	entries = append(entries, hotkeyEntries(cc)...)
	return entries
}

func hotkeyEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	hk := cc.Hotkeys
	var entries []model.ConfigEntry
	bindNum := 0

	type mapping struct {
		action  string
		binding model.HotkeyBinding
	}
	mappings := []mapping{
		{"btn_screenshot", hk.Screenshot},
		{"btn_fforward", hk.FastForward},
		{"btn_jump_state", hk.LoadState},
		{"btn_quick_save", hk.SaveState},
		{"btn_escape", hk.Quit},
	}

	for _, m := range mappings {
		if len(m.binding.Buttons) >= 2 {
			buttonIndices := make([]string, len(m.binding.Buttons))
			for i, b := range m.binding.Buttons {
				buttonIndices[i] = fmt.Sprintf("%d", rawJoystickIndex[b])
			}
			// Format: button1,button2:action:sequential (0=simultaneous, 1=sequential)
			value := fmt.Sprintf("%s:%s:0", join(buttonIndices, ","), m.action)
			entries = append(entries, model.ConfigEntry{
				Path:  []string{"combo", fmt.Sprintf("bind%d", bindNum)},
				Value: value,
			})
			bindNum++
		}
	}
	return entries
}

func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

var dreamcastBIOSProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionBIOS, "dc_boot.bin", "Boot ROM", []string{"e10c53c2f8b90bab96ead2d368858623", "d407fcf70b56acb84b8c77c93b0e5327", "93a9766f14159b403178ac77417c6b68"}),
	model.HashedProvision(model.ProvisionBIOS, "dc_flash.bin", "Flash ROM", []string{"0a93f7940c455905bea6e392dfde92a4"}),
	model.HashedProvision(model.ProvisionBIOS, "flash.bin", "Flash ROM (alternate name)", []string{"0a93f7940c455905bea6e392dfde92a4"}),
}
