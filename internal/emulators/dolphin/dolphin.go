// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
//
// Shader support uses crt-lottes-fast from:
// https://github.com/dolphin-emu/dolphin/pull/12014
package dolphin

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

//go:embed crt_lottes_fast.glsl
var crtShader []byte

const crtShaderFile = "crt_lottes_fast.glsl"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDDolphin,
		Name:            "Dolphin",
		Systems:         []model.SystemID{model.SystemIDGameCube, model.SystemIDWii},
		Package:         model.AppImageRef("dolphin"),
		ProvisionGroups: buildDolphinProvisionGroups(),
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "dolphin",
			GenericName: "GameCube/Wii Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if len(opts.LaunchArgs) > 0 {
					cmd += " " + strings.Join(opts.LaunchArgs, " ")
				}
				if opts.Fullscreen {
					cmd += " -C Dolphin.Display.Fullscreen=True"
				}
				cmd += " -e %ROM%"
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

type Config struct{}

var configTarget = model.ConfigTarget{
	RelPath: "dolphin-emu/Dolphin.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var gfxTarget = model.ConfigTarget{
	RelPath: "dolphin-emu/GFX.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var gcPadTarget = model.ConfigTarget{
	RelPath: "dolphin-emu/GCPadNew.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var gcPadProfileTarget = model.ConfigTarget{
	RelPath: "dolphin-emu/Profiles/GCPad/Kyaraben.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var hotkeysTarget = model.ConfigTarget{
	RelPath: "dolphin-emu/Hotkeys.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store

	patches := []model.ConfigPatch{
		{
			Target: configTarget,
			Entries: []model.ConfigEntry{
				{Path: []string{"General", "ISOPath0"}, Value: store.SystemRomsDir(model.SystemIDGameCube)},
				{Path: []string{"General", "ISOPath1"}, Value: store.SystemRomsDir(model.SystemIDWii)},
				{Path: []string{"General", "ISOPaths"}, Value: "2"},
				{Path: []string{"General", "DumpPath"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDDolphin)},
				{Path: []string{"Interface", "ConfirmStop"}, Value: "False"},
				{Path: []string{"AutoUpdate", "UpdateTrack"}, Value: ""},
				{Path: []string{"GBA", "BIOS"}, Value: store.SystemBiosDir(model.SystemIDGBA) + "/gba_bios.bin"},
				{Path: []string{"GBA", "SavesPath"}, Value: store.SystemSavesDir(model.SystemIDGBA)},
				{Path: []string{"GBA", "SavesInRomPath"}, Value: "0"},
				{Path: []string{"Core", "SIDevice0"}, Value: "6", DefaultOnly: true},
				{Path: []string{"Core", "SIDevice1"}, Value: "0", DefaultOnly: true},
				{Path: []string{"Core", "SIDevice2"}, Value: "0", DefaultOnly: true},
				{Path: []string{"Core", "SIDevice3"}, Value: "0", DefaultOnly: true},
			},
		},
		{
			Target:  gfxTarget,
			Entries: gfxEntries(ctx.Shaders),
		},
	}

	if cc := ctx.ControllerConfig; cc != nil {
		patches = append(patches,
			model.ConfigPatch{Target: gcPadTarget, Entries: gcPadEntries(cc)},
			model.ConfigPatch{Target: hotkeysTarget, Entries: dolphinHotkeyEntries(cc)},
			// Profile file so users can reload kyaraben bindings from the UI.
			model.ConfigPatch{
				Target:         gcPadProfileTarget,
				Entries:        gcPadProfileEntries(cc),
				ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
			},
		)
	}

	dataDir, err := ctx.BaseDirResolver.UserDataDir()
	if err != nil {
		return model.GenerateResult{}, err
	}
	dolphinDir := filepath.Join(dataDir, "dolphin-emu")

	symlinks := []model.SymlinkSpec{
		{Source: filepath.Join(dolphinDir, "GC"), Target: store.SystemSavesDir(model.SystemIDGameCube)},
		{Source: filepath.Join(dolphinDir, "Wii"), Target: store.SystemSavesDir(model.SystemIDWii)},
		{Source: filepath.Join(dolphinDir, "StateSaves"), Target: store.EmulatorStatesDir(model.EmulatorIDDolphin)},
		{Source: filepath.Join(dolphinDir, "ScreenShots"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDDolphin)},
	}

	embeddedFiles, err := shaderFiles(ctx.BaseDirResolver, ctx.Shaders)
	if err != nil {
		return model.GenerateResult{}, err
	}

	return model.GenerateResult{
		Patches:       patches,
		Symlinks:      symlinks,
		EmbeddedFiles: embeddedFiles,
	}, nil
}

func gcPadBindingEntries(cc *model.ControllerConfig, section string, defaultOnly bool) []model.ConfigEntry {
	fb := cc.FaceButtons(model.SystemIDGameCube)

	// Dolphin maps SDL buttons to its own descriptive names.
	// GameCube A=east (green), B=south (red), X=north, Y=west.
	// FaceButtons returns which SDL button is at each physical position.
	faceMap := map[string]model.SDLButton{
		"A": fb.East,
		"B": fb.South,
		"X": fb.North,
		"Y": fb.West,
	}

	return []model.ConfigEntry{
		{Path: []string{section, "Buttons/A"}, Value: fmt.Sprintf("`Button %s`", string(faceMap["A"])), DefaultOnly: defaultOnly},
		{Path: []string{section, "Buttons/B"}, Value: fmt.Sprintf("`Button %s`", string(faceMap["B"])), DefaultOnly: defaultOnly},
		{Path: []string{section, "Buttons/X"}, Value: fmt.Sprintf("`Button %s`", string(faceMap["X"])), DefaultOnly: defaultOnly},
		{Path: []string{section, "Buttons/Y"}, Value: fmt.Sprintf("`Button %s`", string(faceMap["Y"])), DefaultOnly: defaultOnly},
		{Path: []string{section, "Buttons/Z"}, Value: "`Shoulder R`", DefaultOnly: defaultOnly},
		{Path: []string{section, "Buttons/Start"}, Value: "Start", DefaultOnly: defaultOnly},
		{Path: []string{section, "Main Stick/Up"}, Value: "`Left Y+`", DefaultOnly: defaultOnly},
		{Path: []string{section, "Main Stick/Down"}, Value: "`Left Y-`", DefaultOnly: defaultOnly},
		{Path: []string{section, "Main Stick/Left"}, Value: "`Left X-`", DefaultOnly: defaultOnly},
		{Path: []string{section, "Main Stick/Right"}, Value: "`Left X+`", DefaultOnly: defaultOnly},
		{Path: []string{section, "C-Stick/Up"}, Value: "`Right Y+`", DefaultOnly: defaultOnly},
		{Path: []string{section, "C-Stick/Down"}, Value: "`Right Y-`", DefaultOnly: defaultOnly},
		{Path: []string{section, "C-Stick/Left"}, Value: "`Right X-`", DefaultOnly: defaultOnly},
		{Path: []string{section, "C-Stick/Right"}, Value: "`Right X+`", DefaultOnly: defaultOnly},
		{Path: []string{section, "Triggers/L"}, Value: "`Trigger L`", DefaultOnly: defaultOnly},
		{Path: []string{section, "Triggers/R"}, Value: "`Trigger R`", DefaultOnly: defaultOnly},
		{Path: []string{section, "Triggers/L-Analog"}, Value: "`Trigger L`", DefaultOnly: defaultOnly},
		{Path: []string{section, "Triggers/R-Analog"}, Value: "`Trigger R`", DefaultOnly: defaultOnly},
		{Path: []string{section, "D-Pad/Down"}, Value: "`Pad S`", DefaultOnly: defaultOnly},
		{Path: []string{section, "D-Pad/Left"}, Value: "`Pad W`", DefaultOnly: defaultOnly},
		{Path: []string{section, "D-Pad/Right"}, Value: "`Pad E`", DefaultOnly: defaultOnly},
		{Path: []string{section, "Rumble/Motor"}, Value: "`Motor L`|`Motor R`", DefaultOnly: defaultOnly},
	}
}

// gcPadEntries returns entries for GCPadNew.ini. DefaultOnly so user
// customizations are preserved; users can reload kyaraben bindings via
// the Kyaraben profile.
func gcPadEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	var entries []model.ConfigEntry
	for i := 0; i < 4; i++ {
		section := fmt.Sprintf("GCPad%d", i+1)
		device := fmt.Sprintf("SDL/%d/Steam Deck Controller", i)
		entries = append(entries, model.ConfigEntry{Path: []string{section, "Device"}, Value: device, DefaultOnly: true})
		entries = append(entries, gcPadBindingEntries(cc, section, true)...)
	}
	return entries
}

func gcPadProfileEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	return gcPadBindingEntries(cc, "Profile", false)
}

// dolphinButtonName maps SDL button names to Dolphin's naming scheme.
func dolphinButtonName(b model.SDLButton) string {
	switch b {
	case model.ButtonA:
		return "`Button S`"
	case model.ButtonB:
		return "`Button E`"
	case model.ButtonX:
		return "`Button W`"
	case model.ButtonY:
		return "`Button N`"
	case model.ButtonBack:
		return "Back"
	case model.ButtonStart:
		return "Start"
	case model.ButtonLeftShoulder:
		return "`Shoulder L`"
	case model.ButtonRightShoulder:
		return "`Shoulder R`"
	case model.ButtonLeftTrigger:
		return "`Trigger L`"
	case model.ButtonRightTrigger:
		return "`Trigger R`"
	case model.ButtonLeftStick:
		return "`Thumb L`"
	case model.ButtonRightStick:
		return "`Thumb R`"
	case model.ButtonDPadUp:
		return "`Pad N`"
	case model.ButtonDPadDown:
		return "`Pad S`"
	case model.ButtonDPadLeft:
		return "`Pad W`"
	case model.ButtonDPadRight:
		return "`Pad E`"
	default:
		return string(b)
	}
}

func dolphinHotkeyChord(binding model.HotkeyBinding) string {
	if len(binding.Buttons) == 0 {
		return ""
	}
	if len(binding.Buttons) == 1 {
		return dolphinButtonName(binding.Buttons[0])
	}
	parts := make([]string, len(binding.Buttons))
	for i, b := range binding.Buttons {
		parts[i] = dolphinButtonName(b)
	}
	return "@(" + strings.Join(parts, "+") + ")"
}

// dolphinHotkeyEntries returns entries for Hotkeys.ini. DefaultOnly so user
// customizations are preserved.
func dolphinHotkeyEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	hk := cc.Hotkeys
	section := "Hotkeys"

	entries := []model.ConfigEntry{
		{Path: []string{section, "Device"}, Value: "SDL/0/Steam Deck Controller", DefaultOnly: true},
	}

	type mapping struct {
		key     string
		binding model.HotkeyBinding
		toggle  bool
	}
	mappings := []mapping{
		{key: "Save State/Save to Selected Slot", binding: hk.SaveState},
		{key: "Load State/Load from Selected Slot", binding: hk.LoadState},
		{key: "Other State Hotkeys/Increase Selected State Slot", binding: hk.NextSlot},
		{key: "Other State Hotkeys/Decrease Selected State Slot", binding: hk.PrevSlot},
		{key: "Emulation Speed/Disable Emulation Speed Limit", binding: hk.FastForward, toggle: true},
		{key: "General/Toggle Pause", binding: hk.Pause},
		{key: "General/Take Screenshot", binding: hk.Screenshot},
		{key: "General/Exit", binding: hk.Quit},
		{key: "General/Toggle Fullscreen", binding: hk.ToggleFullscreen},
	}

	for _, m := range mappings {
		if len(m.binding.Buttons) > 0 {
			value := dolphinHotkeyChord(m.binding)
			if m.toggle {
				value = "toggle(" + value + ")"
			}
			entries = append(entries, model.ConfigEntry{
				Path:        []string{section, m.key},
				Value:       value,
				DefaultOnly: true,
			})
		}
	}
	return entries
}

type iplRegionSpec struct {
	Name   string
	Dir    string
	File   string
	Hashes []string
}

var iplRegionSpecs = []iplRegionSpec{
	{
		Name:   "USA",
		Dir:    "USA",
		File:   "IPL.bin",
		Hashes: []string{"6dac1f2a14f659a1a7fbf749892b4e41", "019e39822a9ca3029124f74dd4d55ac4"},
	},
	{
		Name:   "Europe",
		Dir:    "EUR",
		File:   "IPL.bin",
		Hashes: []string{"db92574caab77a7ec99d4605fd6f2450", "0cdda509e2da83c85bfe423dd87346cc"},
	},
	{
		Name:   "Japan",
		Dir:    "JAP",
		File:   "IPL.bin",
		Hashes: []string{"fc924a7c879b661abc37cec4f018fdf3", "81df278301dc7bdf57bb760d7393ab4d"},
	},
}

func buildDolphinProvisionGroups() []model.ProvisionGroup {
	groups := make([]model.ProvisionGroup, 0, len(iplRegionSpecs)+1)
	for _, region := range iplRegionSpecs {
		region := region
		groups = append(groups, model.ProvisionGroup{
			MinRequired: 0,
			Message:     fmt.Sprintf("GameCube IPL (%s)", region.Name),
			BaseDir: func(store model.StoreReader, sys model.SystemID) string {
				return filepath.Join(store.SystemSavesDir(model.SystemIDGameCube), region.Dir)
			},
			Provisions: []model.Provision{
				model.HashedProvision(model.ProvisionBIOS, region.File, region.Name, region.Hashes).ForSystems(model.SystemIDGameCube),
			},
		})
	}

	groups = append(groups, model.ProvisionGroup{
		MinRequired: 0,
		Message:     "Game Boy Advance BIOS (optional, shared with mGBA)",
		BaseDir: func(store model.StoreReader, sys model.SystemID) string {
			return store.SystemBiosDir(model.SystemIDGBA)
		},
		Provisions: []model.Provision{
			model.HashedProvision(model.ProvisionBIOS, "gba_bios.bin", "Game Boy Advance BIOS", []string{"a860e8c0b6d573d191e4ec7db1b1e4f6"}).ForSystems(model.SystemIDGameCube),
		},
	})

	return groups
}

func gfxEntries(shaders *bool) []model.ConfigEntry {
	entries := []model.ConfigEntry{
		{Path: []string{"Settings", "InternalResolution"}, Value: "2", DefaultOnly: true},
	}
	if shaders == nil {
		return entries
	}
	if *shaders {
		entries = append(entries, model.ConfigEntry{
			Path: []string{"Enhancements", "PostProcessingShader"}, Value: "crt_lottes_fast",
		})
	} else {
		entries = append(entries, model.ConfigEntry{
			Path: []string{"Enhancements", "PostProcessingShader"}, Value: "",
		})
	}
	return entries
}

func shaderFiles(resolver model.BaseDirResolver, shaders *bool) ([]model.EmbeddedFile, error) {
	if shaders == nil || !*shaders {
		return nil, nil
	}

	dataDir, err := resolver.UserDataDir()
	if err != nil {
		return nil, fmt.Errorf("getting data dir: %w", err)
	}

	shaderDir := filepath.Join(dataDir, "dolphin-emu", "Shaders")
	return []model.EmbeddedFile{{
		Content:  crtShader,
		DestPath: filepath.Join(shaderDir, crtShaderFile),
	}}, nil
}
