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
		PathUsage:         model.StandardPathUsage(),
		SupportedSettings: []string{model.SettingResumeAutosave, model.SettingResumeAutoload},
		ResumeRecommended: true,
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

	entries := []model.ConfigEntry{
		model.Entry(model.Store, model.Path("config", "Flycast.DataPath"), biosDir),
		model.Entry(model.Store, model.Path("config", "Dreamcast.BiosPath"), biosDir),
		model.Entry(model.Store, model.Path("config", "Dreamcast.ContentPath"), store.SystemRomsDir(model.SystemIDDreamcast)),
		model.Entry(model.Store, model.Path("config", "Dreamcast.SavePath"), savesDir),
		model.Entry(model.Store, model.Path("config", "Dreamcast.VMUPath"), savesDir),
		model.Entry(model.Store, model.Path("config", "Dreamcast.SavestatePath"), store.EmulatorStatesDir(model.EmulatorIDFlycast)),
	}

	switch ctx.Resume {
	case model.EmulatorResumeOn:
		entries = append(entries,
			model.Entry(model.Resume, model.Path("config", "Dreamcast.AutoSaveState"), "yes"),
			model.Entry(model.Resume, model.Path("config", "Dreamcast.AutoLoadState"), "yes"),
		)
	case model.EmulatorResumeOff:
		entries = append(entries,
			model.Entry(model.Resume, model.Path("config", "Dreamcast.AutoSaveState"), "no"),
			model.Entry(model.Resume, model.Path("config", "Dreamcast.AutoLoadState"), "no"),
		)
	}

	patches := []model.ConfigPatch{{
		Target:  configTarget,
		Entries: entries,
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
		model.Entry(model.Nintendo, model.Path("digital", "bind0"), fmt.Sprintf("%d:btn_a", rawJoystickIndex[fb.South])),
		model.Entry(model.Nintendo, model.Path("digital", "bind1"), fmt.Sprintf("%d:btn_b", rawJoystickIndex[fb.East])),
		model.Entry(model.Nintendo, model.Path("digital", "bind2"), fmt.Sprintf("%d:btn_x", rawJoystickIndex[fb.West])),
		model.Entry(model.Nintendo, model.Path("digital", "bind3"), fmt.Sprintf("%d:btn_y", rawJoystickIndex[fb.North])),
		model.Entry(model.None, model.Path("digital", "bind4"), fmt.Sprintf("%d:btn_z", rawJoystickIndex[model.ButtonLeftShoulder])),
		model.Entry(model.None, model.Path("digital", "bind5"), fmt.Sprintf("%d:btn_c", rawJoystickIndex[model.ButtonRightShoulder])),
		model.Entry(model.None, model.Path("digital", "bind6"), fmt.Sprintf("%d:btn_start", rawJoystickIndex[model.ButtonStart])),
		model.Entry(model.None, model.Path("digital", "bind7"), fmt.Sprintf("%d:btn_dpad2_up", rawJoystickIndex[model.ButtonGuide])),
		model.Entry(model.None, model.Path("digital", "bind8"), fmt.Sprintf("%d:btn_dpad1_up", hatUp)),
		model.Entry(model.None, model.Path("digital", "bind9"), fmt.Sprintf("%d:btn_dpad1_down", hatDown)),
		model.Entry(model.None, model.Path("digital", "bind10"), fmt.Sprintf("%d:btn_dpad1_left", hatLeft)),
		model.Entry(model.None, model.Path("digital", "bind11"), fmt.Sprintf("%d:btn_dpad1_right", hatRight)),
		model.Entry(model.None, model.Path("analog", "bind0"), "0-:btn_analog_left"),
		model.Entry(model.None, model.Path("analog", "bind1"), "0+:btn_analog_right"),
		model.Entry(model.None, model.Path("analog", "bind2"), "1-:btn_analog_up"),
		model.Entry(model.None, model.Path("analog", "bind3"), "1+:btn_analog_down"),
		model.Entry(model.None, model.Path("analog", "bind4"), "2+:btn_trigger_left"),
		model.Entry(model.None, model.Path("analog", "bind5"), "3-:axis2_left"),
		model.Entry(model.None, model.Path("analog", "bind6"), "3+:axis2_right"),
		model.Entry(model.None, model.Path("analog", "bind7"), "4-:axis2_up"),
		model.Entry(model.None, model.Path("analog", "bind8"), "4+:axis2_down"),
		model.Entry(model.None, model.Path("analog", "bind9"), "5+:btn_trigger_right"),
		model.Entry(model.None, model.Path("emulator", "dead_zone"), "10"),
		model.Entry(model.None, model.Path("emulator", "mapping_name"), "Steam Deck Controller"),
		model.Entry(model.None, model.Path("emulator", "rumble_power"), "100"),
		model.Entry(model.None, model.Path("emulator", "saturation"), "100"),
		model.Entry(model.None, model.Path("emulator", "triggers"), "2,5"),
		model.Entry(model.None, model.Path("emulator", "version"), "4"),
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
			entries = append(entries, model.Entry(model.Hotkeys, model.Path("combo", fmt.Sprintf("bind%d", bindNum)), value))
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
