// Package retroarch provides shared configuration for RetroArch cores.
// All RetroArch cores use the same retroarch.cfg for base settings.
// Per-core overrides handle core-specific paths like BIOS directories.
// Symlinks redirect RetroArch's sorted directories to kyaraben's store.
// See: https://docs.libretro.com/guides/change-directories/
package retroarch

import (
	"fmt"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

const (
	kokoAioVersion     = "NG-1.9.85"
	kokoAioSHA256      = "c9aefcdc47156c1ed89f203ccf1628a370c64337195b78e48ab80738776f4cdb"
	kokoAioDir         = "koko-aio-slang"
	kokoAioStripPrefix = "koko-aio-slang-" + kokoAioVersion
)

var HotkeyMappings = model.HotkeyMappings{
	SaveState:        &model.HotkeyKey{Key: "input_save_state_btn"},
	LoadState:        &model.HotkeyKey{Key: "input_load_state_btn"},
	NextSlot:         &model.HotkeyKey{Key: "input_state_slot_increase_btn"},
	PrevSlot:         &model.HotkeyKey{Key: "input_state_slot_decrease_btn"},
	FastForward:      &model.HotkeyKey{Key: "input_toggle_fast_forward_btn"},
	Rewind:           &model.HotkeyKey{Key: "input_rewind_btn"},
	Pause:            &model.HotkeyKey{Key: "input_pause_toggle_btn"},
	Screenshot:       &model.HotkeyKey{Key: "input_screenshot_btn"},
	Quit:             &model.HotkeyKey{Key: "input_exit_emulator_btn"},
	ToggleFullscreen: &model.HotkeyKey{Key: "input_toggle_fullscreen_btn"},
	OpenMenu:         &model.HotkeyKey{Key: "input_menu_toggle_btn"},
}

var SharedLauncher = model.LauncherInfo{
	Binary:      "retroarch",
	DisplayName: "RetroArch",
	GenericName: "Multi-system Emulator",
	Categories:  []string{"Game", "Emulator"},
}

// LauncherWithCore returns a copy of SharedLauncher with CoreName and RomCommand
// set to load the given libretro core via the -L flag.
func LauncherWithCore(coreName string) model.LauncherInfo {
	l := SharedLauncher
	l.CoreName = coreName
	l.RomCommand = func(opts model.RomLaunchOptions) string {
		cmd := opts.BinaryPath
		if opts.Fullscreen {
			cmd += " -f"
		}
		cmd += " -L " + coreName + " %ROM%"
		return cmd
	}
	return l
}

var MainConfigTarget = model.ConfigTarget{
	RelPath: "retroarch/retroarch.cfg",
	Format:  model.ConfigFormatCFG,
	BaseDir: model.ConfigBaseDirUserConfig,
}

// PresetConfig holds graphics preset and resume settings for config generation.
type PresetConfig struct {
	Preset             string
	Bezels             bool
	TargetDevice       string
	Resume             string
	SystemDisplayTypes map[model.SystemID]model.DisplayType
}

// SharedConfig generates the base RetroArch configuration shared by all cores.
// Enables per-core sorting so RetroArch creates subdirectories like saves/bsnes/.
// We symlink these sorted directories to kyaraben's store locations.
// Screenshots go directly to a shared retroarch directory (no per-core sorting).
// See: https://docs.libretro.com/guides/change-directories/
func SharedConfig(store model.StoreReader, cc *model.ControllerConfig, pc *PresetConfig) model.ConfigPatch {
	entries := []model.ConfigEntry{
		model.Entry(model.Store, model.Path("libretro_directory"), store.CoresDir()),
		model.Entry(model.Store, model.Path("screenshot_directory"), store.EmulatorScreenshotsDir(model.EmulatorIDRetroArchBsnes)),
		model.Entry(model.None, model.Path("video_driver"), "vulkan"),
		model.Entry(model.None, model.Path("sort_savefiles_enable"), "true"),
		model.Entry(model.None, model.Path("sort_savestates_enable"), "true"),
		model.Entry(model.None, model.Path("sort_savefiles_by_content_enable"), "false"),
		model.Entry(model.None, model.Path("sort_savestates_by_content_enable"), "false"),
		model.Entry(model.Store, model.Path("rgui_browser_directory"), store.RomsDir()),
		model.Default(model.None, model.Path("menu_driver"), "ozone"),
		model.Entry(model.None, model.Path("menu_show_load_content_animation"), "false"),
		model.Default(model.None, model.Path("notification_show_config_override_load"), "false"),
		model.Default(model.None, model.Path("notification_show_remap_load"), "false"),
		model.Default(model.None, model.Path("notification_show_autoconfig"), "false"),
		model.Entry(model.None, model.Path("quit_press_twice"), "false"),
		model.Default(model.None, model.Path("input_player1_analog_dpad_mode"), "1"),
		model.Default(model.None, model.Path("input_player2_analog_dpad_mode"), "1"),
		model.Default(model.None, model.Path("input_player3_analog_dpad_mode"), "1"),
		model.Default(model.None, model.Path("input_player4_analog_dpad_mode"), "1"),
	}

	if cc != nil {
		entries = append(entries, controllerEntries(cc)...)
	}

	if pc != nil {
		entries = append(entries, presetEntries(pc)...)
	}

	if pc != nil && pc.Resume == model.EmulatorResumeOn {
		entries = append(entries,
			model.Entry(model.Resume, model.Path("savestate_auto_save"), "true"),
			model.Entry(model.Resume, model.Path("savestate_auto_load"), "true"),
		)
	} else if pc != nil && pc.Resume == model.EmulatorResumeOff {
		entries = append(entries,
			model.Entry(model.Resume, model.Path("savestate_auto_save"), "false"),
			model.Entry(model.Resume, model.Path("savestate_auto_load"), "false"),
		)
	}

	return model.ConfigPatch{Target: MainConfigTarget, Entries: entries}
}

func presetEntries(pc *PresetConfig) []model.ConfigEntry {
	var entries []model.ConfigEntry

	switch pc.Preset {
	case model.PresetModernPixels:
		entries = append(entries,
			model.Entry(model.Preset, model.Path("video_scale_integer"), "true"),
			model.Entry(model.Preset, model.Path("video_shader_enable"), "false"),
			model.Entry(model.Preset, model.Path("video_smooth"), "false"),
		)
	case model.PresetUpscaled:
		entries = append(entries,
			model.Entry(model.Preset, model.Path("video_scale_integer"), "false"),
			model.Entry(model.Preset, model.Path("video_shader_enable"), "false"),
			model.Entry(model.Preset, model.Path("video_smooth"), "true"),
		)
	case model.PresetPseudoAuthentic:
		entries = append(entries,
			model.Entry(model.Preset, model.Path("video_scale_integer"), "false"),
			model.Entry(model.Preset, model.Path("video_shader_enable"), "true"),
			model.Entry(model.Preset, model.Path("video_smooth"), "false"),
			model.Entry(model.Preset, model.Path("video_allow_rotate"), "true"),
			model.Entry(model.Preset, model.Path("aspect_ratio_index"), "24"),
		)
	}

	return entries
}

func controllerEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	entries := []model.ConfigEntry{
		model.Entry(model.None, model.Path("input_joypad_driver"), "sdl2"),
		model.Entry(model.None, model.Path("input_autodetect_enable"), "true"),
	}

	// RetroArch's libretro "RetroPad" is based on SNES layout (A=east, B=south).
	// The NintendoConfirm setting affects all RetroArch cores since most are Nintendo.
	fb := cc.FaceButtons(model.SystemIDSNES)
	for i := 1; i <= 4; i++ {
		prefix := fmt.Sprintf("input_player%d_", i)
		entries = append(entries,
			model.Entry(model.Nintendo, model.Path(prefix+"a_btn"), fmt.Sprintf("%d", model.SDLButtonIndex[fb.East])),
			model.Entry(model.Nintendo, model.Path(prefix+"b_btn"), fmt.Sprintf("%d", model.SDLButtonIndex[fb.South])),
			model.Entry(model.Nintendo, model.Path(prefix+"x_btn"), fmt.Sprintf("%d", model.SDLButtonIndex[fb.North])),
			model.Entry(model.Nintendo, model.Path(prefix+"y_btn"), fmt.Sprintf("%d", model.SDLButtonIndex[fb.West])),
		)
	}

	hk := cc.Hotkeys
	type mapping struct {
		key     string
		binding model.HotkeyBinding
	}
	mappings := []mapping{
		{"input_save_state_btn", hk.SaveState},
		{"input_load_state_btn", hk.LoadState},
		{"input_state_slot_increase_btn", hk.NextSlot},
		{"input_state_slot_decrease_btn", hk.PrevSlot},
		{"input_toggle_fast_forward_btn", hk.FastForward},
		{"input_rewind_btn", hk.Rewind},
		{"input_pause_toggle_btn", hk.Pause},
		{"input_screenshot_btn", hk.Screenshot},
		{"input_exit_emulator_btn", hk.Quit},
		{"input_toggle_fullscreen_btn", hk.ToggleFullscreen},
		{"input_menu_toggle_btn", hk.OpenMenu},
	}

	// RetroArch's hotkey system: the first button in the chord is the enable_hotkey,
	// the last button is the action. All hotkeys must use the same enable_hotkey.
	var enableBtnSet bool
	for _, m := range mappings {
		if len(m.binding.Buttons) < 2 {
			continue
		}
		if !enableBtnSet {
			enableBtn := m.binding.Buttons[0]
			entries = append(entries, model.Entry(model.Hotkeys, model.Path("input_enable_hotkey_btn"), fmt.Sprintf("%d", cc.SDLIndex(enableBtn))))
			enableBtnSet = true
		}
		actionBtn := m.binding.Buttons[len(m.binding.Buttons)-1]
		entries = append(entries, model.Entry(model.Hotkeys, model.Path(m.key), fmt.Sprintf("%d", cc.SDLIndex(actionBtn))))
	}

	return entries
}

var coreConfigDirNames = map[string]string{
	"bsnes":             "bsnes",
	"mesen":             "Mesen",
	"genesis_plus_gx":   "Genesis Plus GX",
	"mupen64plus_next":  "Mupen64Plus-Next",
	"mednafen_saturn":   "Beetle Saturn",
	"mednafen_pce_fast": "Beetle PCE Fast",
	"mednafen_ngp":      "Beetle NeoPop",
	"mgba":              "mGBA",
	"melondsds":         "melonDS DS",
	"citra":             "Citra",
	"fbneo":             "FinalBurn Neo",
	"stella":            "Stella",
	"vice_x64sc":        "VICE x64sc",
}

func CoreOverrideTarget(shortName string) model.ConfigTarget {
	configDirName := shortName
	if displayName, ok := coreConfigDirNames[shortName]; ok {
		configDirName = displayName
	}
	return model.ConfigTarget{
		RelPath: "retroarch/config/" + configDirName + "/" + configDirName + ".cfg",
		Format:  model.ConfigFormatCFG,
		BaseDir: model.ConfigBaseDirUserConfig,
	}
}

var coreShortNames = map[model.EmulatorID]string{
	model.EmulatorIDRetroArchBsnes:         "bsnes",
	model.EmulatorIDRetroArchMesen:         "mesen",
	model.EmulatorIDRetroArchGenesisPlusGX: "genesis_plus_gx",
	model.EmulatorIDRetroArchMupen64Plus:   "mupen64plus_next",
	model.EmulatorIDRetroArchBeetleSaturn:  "mednafen_saturn",
	model.EmulatorIDRetroArchBeetlePCE:     "mednafen_pce_fast",
	model.EmulatorIDRetroArchBeetleNGP:     "mednafen_ngp",
	model.EmulatorIDRetroArchMGBA:          "mgba",
	model.EmulatorIDRetroArchMelonDS:       "melondsds",
	model.EmulatorIDRetroArchCitra:         "citra",
	model.EmulatorIDRetroArchFBNeo:         "fbneo",
	model.EmulatorIDRetroArchStella:        "stella",
	model.EmulatorIDRetroArchVICE:          "vice_x64sc",
}

var coreToSystem = map[model.EmulatorID]model.SystemID{
	model.EmulatorIDRetroArchBsnes:         model.SystemIDSNES,
	model.EmulatorIDRetroArchMesen:         model.SystemIDNES,
	model.EmulatorIDRetroArchGenesisPlusGX: model.SystemIDGenesis,
	model.EmulatorIDRetroArchMupen64Plus:   model.SystemIDN64,
	model.EmulatorIDRetroArchBeetleSaturn:  model.SystemIDSaturn,
	model.EmulatorIDRetroArchBeetlePCE:     model.SystemIDPCEngine,
	model.EmulatorIDRetroArchBeetleNGP:     model.SystemIDNGP,
	model.EmulatorIDRetroArchMGBA:          model.SystemIDGBA,
	model.EmulatorIDRetroArchMelonDS:       model.SystemIDNDS,
	model.EmulatorIDRetroArchCitra:         model.SystemIDN3DS,
	model.EmulatorIDRetroArchFBNeo:         model.SystemIDArcade,
	model.EmulatorIDRetroArchStella:        model.SystemIDAtari2600,
	model.EmulatorIDRetroArchVICE:          model.SystemIDC64,
}

var coreNeedsBiosDir = map[model.EmulatorID]bool{
	model.EmulatorIDRetroArchBeetleSaturn:  true,
	model.EmulatorIDRetroArchBeetlePCE:     true,
	model.EmulatorIDRetroArchGenesisPlusGX: true,
}

// coreUsesHWRendering lists cores that use hardware GPU rendering internally.
// These cores are incompatible with koko-aio shaders which add additional GPU load.
var coreUsesHWRendering = map[model.EmulatorID]bool{
	model.EmulatorIDRetroArchMupen64Plus: true,
	model.EmulatorIDRetroArchCitra:       true,
	model.EmulatorIDRetroArchMelonDS:     true,
}

func CoreShortName(emuID model.EmulatorID) string {
	return coreShortNames[emuID]
}

// CorePatches returns the base RetroArch config plus a per-core override for
// cores that need additional settings like system_directory for BIOS files.
func CorePatches(emuID model.EmulatorID, store model.StoreReader, cc *model.ControllerConfig, pc *PresetConfig, resolver model.BaseDirResolver) []model.ConfigPatch {
	patches := []model.ConfigPatch{SharedConfig(store, cc, pc)}

	shortName := CoreShortName(emuID)
	if shortName == "" {
		return patches
	}

	var entries []model.ConfigEntry
	if coreNeedsBiosDir[emuID] {
		systemID := coreToSystem[emuID]
		entries = append(entries, model.Entry(model.Store, model.Path("system_directory"), store.SystemBiosDir(systemID)))
	}

	if pc != nil && pc.Preset != model.PresetManual && pc.Preset != "" {
		configDirName := shortName
		if displayName, ok := coreConfigDirNames[shortName]; ok {
			configDirName = displayName
		}
		if pc.Preset == model.PresetPseudoAuthentic && !coreUsesHWRendering[emuID] {
			configDir, err := resolver.UserConfigDir()
			if err == nil {
				presetPath := filepath.Join(configDir, "retroarch", "config", configDirName, configDirName+".slangp")
				entries = append(entries,
					model.Entry(model.Preset, model.Path("video_shader_enable"), "true"),
					model.Entry(model.Preset, model.Path("video_shader"), presetPath),
				)
			}
		} else {
			entries = append(entries,
				model.Entry(model.Preset, model.Path("video_shader_enable"), "false"),
				model.Entry(model.Preset, model.Path("video_shader"), ""),
				model.Entry(model.Preset, model.Path("aspect_ratio_index"), "22"),
			)
		}
	}

	if len(entries) > 0 {
		patches = append(patches, model.ConfigPatch{
			Target:  CoreOverrideTarget(shortName),
			Entries: entries,
		})
	}

	if pc != nil && pc.Preset == model.PresetPseudoAuthentic && !coreUsesHWRendering[emuID] {
		shaderPatch := coreShaderPatch(emuID, pc.SystemDisplayTypes, resolver)
		if shaderPatch != nil {
			patches = append(patches, *shaderPatch)
		}
	}

	return patches
}

var kokoAioPresets = map[model.SystemID]string{
	model.SystemIDGB:       "Presets_Handhelds-ng/GameboyMono.slangp",
	model.SystemIDGBC:      "Presets_Handhelds-ng/GameboyColor.slangp",
	model.SystemIDGBA:      "Presets_Handhelds-ng/GameboyAdvance.slangp",
	model.SystemIDNGP:      "Presets_Handhelds-ng/Generic-Handheld-RGB.slangp",
	model.SystemIDGameGear: "Presets_Handhelds-ng/GameGear.slangp",
	model.SystemIDNDS:      "Presets_Handhelds-ng/Generic-Handheld-RGB.slangp",
	model.SystemIDN3DS:     "Presets_Handhelds-ng/Generic-Handheld-RGB.slangp",
	model.SystemIDPSP:      "Presets_Handhelds-ng/PSP.slangp",
}

const kokoAioCRTPreset = "Presets-ng/Base.slangp"

const kyarabenShaderOverridesBase = `
DO_IN_GLOW = "0.0"
DO_HALO = "0.0"
DO_BLOOM = "0.0"
DO_CURVATURE = "0.0"
DO_BG_IMAGE = "0.0"
DO_VIGNETTE = "0.0"
DO_SPOT = "0.0"
DO_DYNZOOM = "0.0"
DO_GLOBAL_SHZO = "0.0"
PIXELGRID_BASAL_GRID = "0.0"
DO_BEZEL = "0.0"
`

const kyarabenShaderOverridesCRT = kyarabenShaderOverridesBase + `
DO_AMBILIGHT = "1.0"
AMBI_STEPS = "5.0"
AMBI_FALLOFF = "0.2"
AMBI_STRETCH = "0.12"
AMBI_POWER = "0.2"
`

const kyarabenShaderOverridesLCD = kyarabenShaderOverridesBase + `
DO_AMBILIGHT = "1.0"
AMBI_FALLOFF = "0.2"
AMBI_STRETCH = "0.12"
AMBI_POWER = "0.15"
`

const kyarabenShaderOverridesGB = kyarabenShaderOverridesBase + `
DO_AMBILIGHT = "1.0"
AMBI_FALLOFF = "0.2"
AMBI_STRETCH = "0.12"
AMBI_POWER = "0.15"
COLOR_MONO_HUE1 = "0.12"
ADAPTIVE_BLACK = "0.3"
`

const kyarabenShaderOverridesGBC = kyarabenShaderOverridesBase + `
DO_AMBILIGHT = "1.0"
AMBI_FALLOFF = "0.3"
AMBI_STRETCH = "0.25"
AMBI_POWER = "0.6"
`

const kyarabenShaderOverridesGBA = kyarabenShaderOverridesBase + `
DO_AMBILIGHT = "0.0"
`

const kyarabenShaderOverridesNGP = kyarabenShaderOverridesBase + `
DO_AMBILIGHT = "1.0"
AMBI_FALLOFF = "0.2"
AMBI_STRETCH = "0.5"
AMBI_POWER = "0.3"
ASPECT_X = "20.0"
ASPECT_Y = "19.0"
`

func shaderOverridesForSystem(systemID model.SystemID, displayType model.DisplayType) string {
	switch systemID {
	case model.SystemIDGB:
		return kyarabenShaderOverridesGB
	case model.SystemIDGBC:
		return kyarabenShaderOverridesGBC
	case model.SystemIDGBA:
		return kyarabenShaderOverridesGBA
	case model.SystemIDNGP:
		return kyarabenShaderOverridesNGP
	default:
		if displayType == model.DisplayTypeLCD {
			return kyarabenShaderOverridesLCD
		}
		return kyarabenShaderOverridesCRT
	}
}

func coreShaderPatch(emuID model.EmulatorID, displayTypes map[model.SystemID]model.DisplayType, resolver model.BaseDirResolver) *model.ConfigPatch {
	shortName := CoreShortName(emuID)
	if shortName == "" {
		return nil
	}

	configDirName := shortName
	if displayName, ok := coreConfigDirNames[shortName]; ok {
		configDirName = displayName
	}

	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return nil
	}

	systemID := coreToSystem[emuID]
	displayType := displayTypes[systemID]
	kokoAioPath := filepath.Join(configDir, "retroarch", "shaders", kokoAioDir)

	presetFile := kokoAioPresets[systemID]
	if presetFile == "" {
		presetFile = kokoAioCRTPreset
	}

	overrides := shaderOverridesForSystem(systemID, displayType)

	target := model.ConfigTarget{
		RelPath: "retroarch/config/" + configDirName + "/" + configDirName + ".slangp",
		Format:  model.ConfigFormatRaw,
		BaseDir: model.ConfigBaseDirUserConfig,
	}

	slangpContent := fmt.Sprintf("#reference \"%s\"\n%s", filepath.Join(kokoAioPath, presetFile), overrides)

	entries := []model.ConfigEntry{
		{Value: slangpContent},
	}

	return &model.ConfigPatch{
		Target:         target,
		Entries:        entries,
		ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
	}
}

func CoreShaderDownloads(emuID model.EmulatorID, resolver model.BaseDirResolver, pc *PresetConfig) ([]model.InitialDownload, error) {
	if pc == nil || pc.Preset != model.PresetPseudoAuthentic {
		return nil, nil
	}

	systemID := coreToSystem[emuID]
	if systemID == "" {
		return nil, nil
	}

	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting config dir: %w", err)
	}

	shaderDir := filepath.Join(configDir, "retroarch", "shaders")

	return []model.InitialDownload{{
		URL:         "https://github.com/kokoko3k/koko-aio-slang/archive/refs/tags/" + kokoAioVersion + ".tar.gz",
		SHA256:      kokoAioSHA256,
		ArchiveType: "tar.gz",
		ExtractDir:  filepath.Join(shaderDir, kokoAioDir),
		StripPrefix: kokoAioStripPrefix,
	}}, nil
}

func CoreSymlinks(emuID model.EmulatorID, store model.StoreReader, resolver model.BaseDirResolver) ([]model.SymlinkSpec, error) {
	shortName := CoreShortName(emuID)
	if shortName == "" {
		return nil, nil
	}

	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting config dir: %w", err)
	}

	systemID := coreToSystem[emuID]
	return []model.SymlinkSpec{
		{
			Source: filepath.Join(configDir, "retroarch", "saves", shortName),
			Target: store.SystemSavesDir(systemID),
		},
		{
			Source: filepath.Join(configDir, "retroarch", "states", shortName),
			Target: store.EmulatorStatesDir(emuID),
		},
	}, nil
}

// CoreOptionsTarget returns the config target for a core's .opt file.
func CoreOptionsTarget(shortName string) model.ConfigTarget {
	configDirName := shortName
	if displayName, ok := coreConfigDirNames[shortName]; ok {
		configDirName = displayName
	}
	return model.ConfigTarget{
		RelPath: "retroarch/config/" + configDirName + "/" + configDirName + ".opt",
		Format:  model.ConfigFormatCFG,
		BaseDir: model.ConfigBaseDirUserConfig,
	}
}

// CoreOptionsPatch returns a config patch for core-specific options.
// This configures features like color correction and interframe blending
// per RGC recommendations for authentic display emulation.
func CoreOptionsPatch(emuID model.EmulatorID, pc *PresetConfig) *model.ConfigPatch {
	if pc == nil {
		return nil
	}

	shortName := CoreShortName(emuID)
	if shortName == "" {
		return nil
	}

	var entries []model.ConfigEntry

	if pc.Preset == model.PresetPseudoAuthentic {
		switch emuID {
		case model.EmulatorIDRetroArchMGBA:
			entries = append(entries,
				model.Entry(model.Preset, model.Path("mgba_color_correction"), "Auto"),
				model.Entry(model.Preset, model.Path("mgba_interframe_blending"), "mix_smart"),
				model.Entry(model.Preset, model.Path("mgba_gb_colors"), "GB Pocket"),
			)
		case model.EmulatorIDRetroArchGenesisPlusGX:
			entries = append(entries,
				model.Entry(model.Preset, model.Path("genesis_plus_gx_blargg_ntsc_filter"), "S-Video"),
			)
		case model.EmulatorIDRetroArchMesen:
			entries = append(entries,
				model.Entry(model.Preset, model.Path("mesen_palette"), "PVM Style (by FirebrandX)"),
			)
		}
	}

	if pc.Bezels && emuID == model.EmulatorIDRetroArchMGBA {
		entries = append(entries, model.Entry(model.Preset, model.Path("mgba_sgb_borders"), "OFF"))
	}

	if len(entries) == 0 {
		return nil
	}

	return &model.ConfigPatch{
		Target:  CoreOptionsTarget(shortName),
		Entries: entries,
	}
}

// ContentDirOverrideTarget returns the config target for a per-content-directory override.
// This allows different settings when loading ROMs from different system directories.
func ContentDirOverrideTarget(shortName string, systemID model.SystemID) model.ConfigTarget {
	configDirName := shortName
	if displayName, ok := coreConfigDirNames[shortName]; ok {
		configDirName = displayName
	}
	contentDir := string(systemID)
	return model.ConfigTarget{
		RelPath: "retroarch/config/" + configDirName + "/" + contentDir + ".cfg",
		Format:  model.ConfigFormatCFG,
		BaseDir: model.ConfigBaseDirUserConfig,
	}
}

// ContentDirOptionsTarget returns the config target for per-content-directory core options.
func ContentDirOptionsTarget(shortName string, systemID model.SystemID) model.ConfigTarget {
	configDirName := shortName
	if displayName, ok := coreConfigDirNames[shortName]; ok {
		configDirName = displayName
	}
	contentDir := string(systemID)
	return model.ConfigTarget{
		RelPath: "retroarch/config/" + configDirName + "/" + contentDir + ".opt",
		Format:  model.ConfigFormatCFG,
		BaseDir: model.ConfigBaseDirUserConfig,
	}
}

// ContentDirOptionsPatches creates per-content-directory core options.
// This allows multi-system cores (like mGBA) to use different options for each system.
func ContentDirOptionsPatches(emuID model.EmulatorID, systems []model.SystemID, pc *PresetConfig) []model.ConfigPatch {
	if pc == nil || pc.Preset != model.PresetPseudoAuthentic {
		return nil
	}

	shortName := CoreShortName(emuID)
	if shortName == "" {
		return nil
	}

	var patches []model.ConfigPatch

	if emuID == model.EmulatorIDRetroArchMGBA {
		for _, systemID := range systems {
			if systemID == model.SystemIDGBA {
				patches = append(patches, model.ConfigPatch{
					Target: ContentDirOptionsTarget(shortName, systemID),
					Entries: []model.ConfigEntry{
						model.Entry(model.Preset, model.Path("mgba_color_correction"), "OFF"),
					},
				})
			}
		}
	}

	return patches
}

// OverlayPatches is deprecated - koko-aio shader handles bezels internally.
// Kept for API compatibility with existing emulator generators.
func OverlayPatches(_ model.EmulatorID, _ []model.SystemID, _ *PresetConfig, _ model.BaseDirResolver) []model.ConfigPatch {
	return nil
}

// CoreEmbeddedFiles is deprecated - koko-aio shader handles bezels internally.
// Kept for API compatibility with existing emulator generators.
func CoreEmbeddedFiles(_ []model.SystemID, _ *PresetConfig, _ model.BaseDirResolver) ([]model.EmbeddedFile, error) {
	return nil, nil
}

// ContentDirShaderPatches creates per-content-directory shader presets.
// This allows multi-system cores (like mGBA) to use different shaders for each system.
func ContentDirShaderPatches(emuID model.EmulatorID, systems []model.SystemID, pc *PresetConfig, resolver model.BaseDirResolver) []model.ConfigPatch {
	if pc == nil || pc.Preset != model.PresetPseudoAuthentic {
		return nil
	}

	shortName := CoreShortName(emuID)
	if shortName == "" {
		return nil
	}

	configDirName := shortName
	if displayName, ok := coreConfigDirNames[shortName]; ok {
		configDirName = displayName
	}

	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return nil
	}

	kokoAioPath := filepath.Join(configDir, "retroarch", "shaders", kokoAioDir)

	var patches []model.ConfigPatch
	for _, systemID := range systems {
		presetFile := kokoAioPresets[systemID]
		if presetFile == "" {
			presetFile = kokoAioCRTPreset
		}

		displayType := pc.SystemDisplayTypes[systemID]
		overrides := shaderOverridesForSystem(systemID, displayType)

		target := model.ConfigTarget{
			RelPath: "retroarch/config/" + configDirName + "/" + string(systemID) + ".slangp",
			Format:  model.ConfigFormatRaw,
			BaseDir: model.ConfigBaseDirUserConfig,
		}

		slangpContent := fmt.Sprintf("#reference \"%s\"\n%s", filepath.Join(kokoAioPath, presetFile), overrides)

		patches = append(patches, model.ConfigPatch{
			Target:         target,
			Entries:        []model.ConfigEntry{{Value: slangpContent}},
			ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
		})

		overrideTarget := ContentDirOverrideTarget(shortName, systemID)
		overrideEntries := []model.ConfigEntry{
			model.Entry(model.Preset, model.Path("video_shader_enable"), "true"),
			model.Entry(model.Preset, model.Path("video_shader"), filepath.Join(configDir, target.RelPath)),
		}
		patches = append(patches, model.ConfigPatch{
			Target:  overrideTarget,
			Entries: overrideEntries,
		})
	}

	return patches
}
