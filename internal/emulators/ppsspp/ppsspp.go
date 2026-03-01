package ppsspp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDPPSSPP,
		Name:            "PPSSPP",
		Systems:         []model.SystemID{model.SystemIDPSP},
		Package:         model.AppImageRef("ppsspp"),
		ProvisionGroups: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "ppsspp",
			GenericName: "PlayStation Portable Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " --fullscreen"
				}
				cmd += " %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			UsesSavesDir:       true,
			UsesStatesDir:      true,
			UsesScreenshotsDir: true,
		},
		SupportedSettings: []string{model.SettingShaders},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "ppsspp/PSP/SYSTEM/ppsspp.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var controlsTarget = model.ConfigTarget{
	RelPath: "ppsspp/PSP/SYSTEM/controls.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store

	entries := []model.ConfigEntry{
		{Path: []string{"General", "CurrentDirectory"}, Value: store.SystemRomsDir(model.SystemIDPSP)},
		{Path: []string{"General", "AskForExitConfirmationAfterSeconds"}, Value: "0"},
		{Path: []string{"General", "FirstRun"}, Value: "False"},
	}

	if ctx.Shaders != nil {
		if *ctx.Shaders {
			entries = append(entries,
				model.ConfigEntry{Path: []string{"Graphics", "PostShaderNames"}, Value: "LCDPersistence"},
			)
		} else {
			entries = append(entries,
				model.ConfigEntry{Path: []string{"Graphics", "PostShaderNames"}, Value: "Off"},
			)
		}
	}

	patches := []model.ConfigPatch{{
		Target:  configTarget,
		Entries: entries,
	}}

	if cc := ctx.ControllerConfig; cc != nil {
		patches = append(patches, model.ConfigPatch{
			Target:  controlsTarget,
			Entries: padEntries(cc),
		})
	}

	configDir, err := ctx.BaseDirResolver.UserConfigDir()
	if err != nil {
		return model.GenerateResult{}, err
	}
	pspDir := filepath.Join(configDir, "ppsspp", "PSP")

	symlinks := []model.SymlinkSpec{
		{Source: filepath.Join(pspDir, "SAVEDATA"), Target: store.SystemSavesDir(model.SystemIDPSP)},
		{Source: filepath.Join(pspDir, "PPSSPP_STATE"), Target: store.EmulatorStatesDir(model.EmulatorIDPPSSPP)},
		{Source: filepath.Join(pspDir, "SCREENSHOT"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDPPSSPP)},
	}

	return model.GenerateResult{
		Patches:  patches,
		Symlinks: symlinks,
	}, nil
}

// PPSSPP uses device-keycode format where device 10 is gamepad.
// Keycodes from EmuDeck: A=189, B=190, X=191, Y=188, Start=197, Select=196,
// L=193, R=192, DPadUp=19, DPadDown=20, DPadLeft=21, DPadRight=22.
// Analog: An.Up=4003, An.Down=4002, An.Left=4001, An.Right=4000.
var ppssppKeycode = map[model.SDLButton]int{
	model.ButtonA:             189,
	model.ButtonB:             190,
	model.ButtonX:             191,
	model.ButtonY:             188,
	model.ButtonStart:         197,
	model.ButtonBack:          196,
	model.ButtonLeftShoulder:  193,
	model.ButtonRightShoulder: 192,
	model.ButtonDPadUp:        19,
	model.ButtonDPadDown:      20,
	model.ButtonDPadLeft:      21,
	model.ButtonDPadRight:     22,
}

func ppssppRef(button model.SDLButton) string {
	return fmt.Sprintf("10-%d", ppssppKeycode[button])
}

func ppssppHotkeyRef(binding model.HotkeyBinding) string {
	parts := make([]string, len(binding.Buttons))
	for i, b := range binding.Buttons {
		parts[i] = ppssppRef(b)
	}
	return strings.Join(parts, ":")
}

func padEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	fb := cc.FaceButtons(model.SystemIDPSP)
	section := "ControlMapping"

	// PSP maps: Cross=south, Circle=east, Square=west, Triangle=north
	entries := []model.ConfigEntry{
		{Path: []string{section, "Cross"}, Value: ppssppRef(fb.South)},
		{Path: []string{section, "Circle"}, Value: ppssppRef(fb.East)},
		{Path: []string{section, "Square"}, Value: ppssppRef(fb.West)},
		{Path: []string{section, "Triangle"}, Value: ppssppRef(fb.North)},
		{Path: []string{section, "Start"}, Value: ppssppRef(model.ButtonStart)},
		{Path: []string{section, "Select"}, Value: ppssppRef(model.ButtonBack)},
		{Path: []string{section, "L"}, Value: ppssppRef(model.ButtonLeftShoulder)},
		{Path: []string{section, "R"}, Value: ppssppRef(model.ButtonRightShoulder)},
		{Path: []string{section, "Up"}, Value: ppssppRef(model.ButtonDPadUp)},
		{Path: []string{section, "Down"}, Value: ppssppRef(model.ButtonDPadDown)},
		{Path: []string{section, "Left"}, Value: ppssppRef(model.ButtonDPadLeft)},
		{Path: []string{section, "Right"}, Value: ppssppRef(model.ButtonDPadRight)},
		{Path: []string{section, "An.Up"}, Value: "10-4003"},
		{Path: []string{section, "An.Down"}, Value: "10-4002"},
		{Path: []string{section, "An.Left"}, Value: "10-4001"},
		{Path: []string{section, "An.Right"}, Value: "10-4000"},
	}

	hk := cc.Hotkeys
	type mapping struct {
		key     string
		binding model.HotkeyBinding
	}
	mappings := []mapping{
		{"Save State", hk.SaveState},
		{"Load State", hk.LoadState},
		{"Next Slot", hk.NextSlot},
		{"Previous Slot", hk.PrevSlot},
		{"SpeedToggle", hk.FastForward},
		{"Rewind", hk.Rewind},
		{"Pause", hk.Pause},
		{"Screenshot", hk.Screenshot},
		{"Exit App", hk.Quit},
	}
	for _, m := range mappings {
		if len(m.binding.Buttons) > 0 {
			entries = append(entries, model.ConfigEntry{
				Path:  []string{section, m.key},
				Value: ppssppHotkeyRef(m.binding),
			})
		}
	}

	return entries
}
