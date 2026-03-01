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
		SupportedSettings:  []string{model.SettingShaders},
		ShadersRecommended: true,
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
				model.Entry(model.Store, model.Path("General", "ISOPath0"), store.SystemRomsDir(model.SystemIDGameCube)),
				model.Entry(model.Store, model.Path("General", "ISOPath1"), store.SystemRomsDir(model.SystemIDWii)),
				model.Entry(model.None, model.Path("General", "ISOPaths"), "2"),
				model.Entry(model.Store, model.Path("General", "DumpPath"), store.EmulatorScreenshotsDir(model.EmulatorIDDolphin)),
				model.Entry(model.None, model.Path("Interface", "ConfirmStop"), "False"),
				model.Entry(model.None, model.Path("AutoUpdate", "UpdateTrack"), ""),
				model.Entry(model.Store, model.Path("GBA", "BIOS"), store.SystemBiosDir(model.SystemIDGBA)+"/gba_bios.bin"),
				model.Entry(model.Store, model.Path("GBA", "SavesPath"), store.SystemSavesDir(model.SystemIDGBA)),
				model.Entry(model.None, model.Path("GBA", "SavesInRomPath"), "0"),
				model.Default(model.None, model.Path("Core", "SIDevice0"), "6"),
				model.Default(model.None, model.Path("Core", "SIDevice1"), "0"),
				model.Default(model.None, model.Path("Core", "SIDevice2"), "0"),
				model.Default(model.None, model.Path("Core", "SIDevice3"), "0"),
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

	newEntry := func(path []string, value string) model.ConfigEntry {
		if defaultOnly {
			return model.Default(model.Nintendo, path, value)
		}
		return model.Entry(model.Nintendo, path, value)
	}

	return []model.ConfigEntry{
		newEntry(model.Path(section, "Buttons/A"), fmt.Sprintf("`Button %s`", string(faceMap["A"]))),
		newEntry(model.Path(section, "Buttons/B"), fmt.Sprintf("`Button %s`", string(faceMap["B"]))),
		newEntry(model.Path(section, "Buttons/X"), fmt.Sprintf("`Button %s`", string(faceMap["X"]))),
		newEntry(model.Path(section, "Buttons/Y"), fmt.Sprintf("`Button %s`", string(faceMap["Y"]))),
		newEntry(model.Path(section, "Buttons/Z"), "`Shoulder R`"),
		newEntry(model.Path(section, "Buttons/Start"), "Start"),
		newEntry(model.Path(section, "Main Stick/Up"), "`Left Y+`"),
		newEntry(model.Path(section, "Main Stick/Down"), "`Left Y-`"),
		newEntry(model.Path(section, "Main Stick/Left"), "`Left X-`"),
		newEntry(model.Path(section, "Main Stick/Right"), "`Left X+`"),
		newEntry(model.Path(section, "C-Stick/Up"), "`Right Y+`"),
		newEntry(model.Path(section, "C-Stick/Down"), "`Right Y-`"),
		newEntry(model.Path(section, "C-Stick/Left"), "`Right X-`"),
		newEntry(model.Path(section, "C-Stick/Right"), "`Right X+`"),
		newEntry(model.Path(section, "Triggers/L"), "`Trigger L`"),
		newEntry(model.Path(section, "Triggers/R"), "`Trigger R`"),
		newEntry(model.Path(section, "Triggers/L-Analog"), "`Trigger L`"),
		newEntry(model.Path(section, "Triggers/R-Analog"), "`Trigger R`"),
		newEntry(model.Path(section, "D-Pad/Down"), "`Pad S`"),
		newEntry(model.Path(section, "D-Pad/Left"), "`Pad W`"),
		newEntry(model.Path(section, "D-Pad/Right"), "`Pad E`"),
		newEntry(model.Path(section, "Rumble/Motor"), "`Motor L`|`Motor R`"),
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
		entries = append(entries, model.Default(model.None, model.Path(section, "Device"), device))
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
		model.Default(model.None, model.Path(section, "Device"), "SDL/0/Steam Deck Controller"),
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
			entries = append(entries, model.Default(model.Nintendo, model.Path(section, m.key), value))
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

func gfxEntries(shaders string) []model.ConfigEntry {
	entries := []model.ConfigEntry{
		model.Default(model.None, model.Path("Settings", "InternalResolution"), "2"),
	}
	switch shaders {
	case model.EmulatorShadersOn:
		entries = append(entries, model.Entry(model.Shaders, model.Path("Enhancements", "PostProcessingShader"), "crt_lottes_fast"))
	case model.EmulatorShadersOff:
		entries = append(entries, model.Entry(model.Shaders, model.Path("Enhancements", "PostProcessingShader"), ""))
	}
	return entries
}

func shaderFiles(resolver model.BaseDirResolver, shaders string) ([]model.EmbeddedFile, error) {
	if shaders != model.EmulatorShadersOn {
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
