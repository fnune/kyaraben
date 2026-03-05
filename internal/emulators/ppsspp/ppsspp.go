package ppsspp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

var hotkeyMappings = model.HotkeyMappings{
	SaveState:   &model.HotkeyKey{Key: "Save State"},
	LoadState:   &model.HotkeyKey{Key: "Load State"},
	NextSlot:    &model.HotkeyKey{Key: "Next Slot"},
	PrevSlot:    &model.HotkeyKey{Key: "Previous Slot"},
	FastForward: &model.HotkeyKey{Key: "SpeedToggle"},
	Rewind:      &model.HotkeyKey{Key: "Rewind"},
	Pause:       &model.HotkeyKey{Key: "Pause"},
	Screenshot:  &model.HotkeyKey{Key: "Screenshot"},
	Quit:        &model.HotkeyKey{Key: "Exit App"},
}

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
		SupportedSettings: []string{model.SettingShaders, model.SettingResumeAutoload},
		SupportedHotkeys:  hotkeyMappings.SupportedHotkeys(),
		ResumeRecommended: false,
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
		model.Default(model.Store, model.Path("General", "CurrentDirectory"), store.SystemRomsDir(model.SystemIDPSP)),
		model.Default(model.None, model.Path("General", "AskForExitConfirmationAfterSeconds"), "0"),
		model.Default(model.None, model.Path("General", "FirstRun"), "False"),
	}

	switch ctx.Shaders {
	case model.EmulatorShadersOn:
		entries = append(entries,
			model.Entry(model.None, model.Path("Graphics", "PostShaderNames"), "LCDPersistence"),
		)
	case model.EmulatorShadersOff:
		entries = append(entries,
			model.Entry(model.None, model.Path("Graphics", "PostShaderNames"), "Off"),
		)
	}

	switch ctx.Resume {
	case model.EmulatorResumeOn:
		entries = append(entries, model.Entry(model.Resume, model.Path("General", "AutoLoadSaveState"), "1"))
	case model.EmulatorResumeOff:
		entries = append(entries, model.Entry(model.Resume, model.Path("General", "AutoLoadSaveState"), "0"))
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
		model.Entry(model.None, model.Path(section, "Cross"), ppssppRef(fb.South)),
		model.Entry(model.None, model.Path(section, "Circle"), ppssppRef(fb.East)),
		model.Entry(model.None, model.Path(section, "Square"), ppssppRef(fb.West)),
		model.Entry(model.None, model.Path(section, "Triangle"), ppssppRef(fb.North)),
		model.Entry(model.None, model.Path(section, "Start"), ppssppRef(model.ButtonStart)),
		model.Entry(model.None, model.Path(section, "Select"), ppssppRef(model.ButtonBack)),
		model.Entry(model.None, model.Path(section, "L"), ppssppRef(model.ButtonLeftShoulder)),
		model.Entry(model.None, model.Path(section, "R"), ppssppRef(model.ButtonRightShoulder)),
		model.Entry(model.None, model.Path(section, "Up"), ppssppRef(model.ButtonDPadUp)),
		model.Entry(model.None, model.Path(section, "Down"), ppssppRef(model.ButtonDPadDown)),
		model.Entry(model.None, model.Path(section, "Left"), ppssppRef(model.ButtonDPadLeft)),
		model.Entry(model.None, model.Path(section, "Right"), ppssppRef(model.ButtonDPadRight)),
		model.Entry(model.None, model.Path(section, "An.Up"), "10-4003"),
		model.Entry(model.None, model.Path(section, "An.Down"), "10-4002"),
		model.Entry(model.None, model.Path(section, "An.Left"), "10-4001"),
		model.Entry(model.None, model.Path(section, "An.Right"), "10-4000"),
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
			entries = append(entries, model.Entry(model.Hotkeys, model.Path(section, m.key), ppssppHotkeyRef(m.binding)))
		}
	}

	return entries
}
