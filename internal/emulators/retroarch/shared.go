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
	slangShadersCommit = "09c7468812414b682703719c1726bb6263ec5530"
	crtShaderFile      = "crt-mattias.slang"
	crtShaderSHA256    = "8219c0dcadfec5db9fb4b9c53cb695cecaa8e1b0fb5993fe3d5a7a7584fe4fed"
	lcdShaderFile      = "lcd-grid-v2.slang"
	lcdShaderSHA256    = "dc77b29530e5b771f12a5a2dd68c390805139bc25d84a30dead021fd38e581c5"
)

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

// ShaderConfig holds shader and resume settings for config generation.
type ShaderConfig struct {
	Shaders            string
	Resume             string
	SystemDisplayTypes map[model.SystemID]model.DisplayType
}

// SharedConfig generates the base RetroArch configuration shared by all cores.
// Enables per-core sorting so RetroArch creates subdirectories like saves/bsnes/.
// We symlink these sorted directories to kyaraben's store locations.
// Screenshots go directly to a shared retroarch directory (no per-core sorting).
// See: https://docs.libretro.com/guides/change-directories/
func SharedConfig(store model.StoreReader, cc *model.ControllerConfig, sc *ShaderConfig) model.ConfigPatch {
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

	if sc != nil && sc.Shaders == model.EmulatorShadersOn {
		entries = append(entries, model.Entry(model.Shaders, model.Path("video_shader_enable"), "true"))
	}

	if sc != nil && sc.Resume == model.EmulatorResumeOn {
		entries = append(entries,
			model.Entry(model.Resume, model.Path("savestate_auto_save"), "true"),
			model.Entry(model.Resume, model.Path("savestate_auto_load"), "true"),
		)
	} else if sc != nil && sc.Resume == model.EmulatorResumeOff {
		entries = append(entries,
			model.Entry(model.Resume, model.Path("savestate_auto_save"), "false"),
			model.Entry(model.Resume, model.Path("savestate_auto_load"), "false"),
		)
	}

	return model.ConfigPatch{Target: MainConfigTarget, Entries: entries}
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
			entries = append(entries, model.Entry(model.Nintendo, model.Path("input_enable_hotkey_btn"), fmt.Sprintf("%d", cc.SDLIndex(enableBtn))))
			enableBtnSet = true
		}
		actionBtn := m.binding.Buttons[len(m.binding.Buttons)-1]
		entries = append(entries, model.Entry(model.Nintendo, model.Path(m.key), fmt.Sprintf("%d", cc.SDLIndex(actionBtn))))
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

func CoreShortName(emuID model.EmulatorID) string {
	return coreShortNames[emuID]
}

// CorePatches returns the base RetroArch config plus a per-core override for
// cores that need additional settings like system_directory for BIOS files.
func CorePatches(emuID model.EmulatorID, store model.StoreReader, cc *model.ControllerConfig, sc *ShaderConfig, resolver model.BaseDirResolver) []model.ConfigPatch {
	patches := []model.ConfigPatch{SharedConfig(store, cc, sc)}

	shortName := CoreShortName(emuID)
	if shortName == "" {
		return patches
	}

	var entries []model.ConfigEntry
	if coreNeedsBiosDir[emuID] {
		systemID := coreToSystem[emuID]
		entries = append(entries, model.Entry(model.Store, model.Path("system_directory"), store.SystemBiosDir(systemID)))
	}

	if sc != nil && sc.Shaders != model.EmulatorShadersManual && sc.Shaders != "" {
		configDirName := shortName
		if displayName, ok := coreConfigDirNames[shortName]; ok {
			configDirName = displayName
		}
		if sc.Shaders == model.EmulatorShadersOn {
			configDir, err := resolver.UserConfigDir()
			if err == nil {
				presetPath := filepath.Join(configDir, "retroarch", "config", configDirName, configDirName+".slangp")
				entries = append(entries,
					model.Entry(model.Shaders, model.Path("video_shader_enable"), "true"),
					model.Entry(model.Shaders, model.Path("video_shader"), presetPath),
				)
			}
		} else {
			entries = append(entries,
				model.Entry(model.Shaders, model.Path("video_shader_enable"), "false"),
				model.Entry(model.Shaders, model.Path("video_shader"), ""),
			)
		}
	}

	if len(entries) > 0 {
		patches = append(patches, model.ConfigPatch{
			Target:  CoreOverrideTarget(shortName),
			Entries: entries,
		})
	}

	if sc != nil && sc.Shaders == model.EmulatorShadersOn {
		shaderPatch := coreShaderPatch(emuID, sc.SystemDisplayTypes, resolver)
		if shaderPatch != nil {
			patches = append(patches, *shaderPatch)
		}
	}

	return patches
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

	target := model.ConfigTarget{
		RelPath: "retroarch/config/" + configDirName + "/" + configDirName + ".slangp",
		Format:  model.ConfigFormatCFG,
		BaseDir: model.ConfigBaseDirUserConfig,
	}

	shaderDir := filepath.Join(configDir, "retroarch", "shaders", "kyaraben")
	var entries []model.ConfigEntry

	if displayType == model.DisplayTypeLCD {
		entries = []model.ConfigEntry{
			model.Entry(model.Shaders, model.Path("shaders"), "1"),
			model.Entry(model.Shaders, model.Path("shader0"), filepath.Join(shaderDir, lcdShaderFile)),
			model.Entry(model.Shaders, model.Path("scale_type0"), "viewport"),
			model.Entry(model.Shaders, model.Path("filter_linear0"), "true"),
		}
	} else {
		entries = []model.ConfigEntry{
			model.Entry(model.Shaders, model.Path("shaders"), "1"),
			model.Entry(model.Shaders, model.Path("shader0"), filepath.Join(shaderDir, crtShaderFile)),
			model.Entry(model.Shaders, model.Path("filter_linear0"), "false"),
		}
	}

	return &model.ConfigPatch{
		Target:         target,
		Entries:        entries,
		ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
	}
}

func CoreShaderDownloads(emuID model.EmulatorID, resolver model.BaseDirResolver, sc *ShaderConfig) ([]model.InitialDownload, error) {
	if sc == nil || sc.Shaders != model.EmulatorShadersOn {
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

	shaderDir := filepath.Join(configDir, "retroarch", "shaders", "kyaraben")
	displayType := sc.SystemDisplayTypes[systemID]

	if displayType == model.DisplayTypeLCD {
		return []model.InitialDownload{{
			URL:      "https://raw.githubusercontent.com/libretro/slang-shaders/" + slangShadersCommit + "/handheld/shaders/lcd-cgwg/lcd-grid-v2.slang",
			SHA256:   lcdShaderSHA256,
			DestPath: filepath.Join(shaderDir, lcdShaderFile),
		}}, nil
	}

	return []model.InitialDownload{{
		URL:      "https://raw.githubusercontent.com/libretro/slang-shaders/" + slangShadersCommit + "/crt/shaders/crt-mattias.slang",
		SHA256:   crtShaderSHA256,
		DestPath: filepath.Join(shaderDir, crtShaderFile),
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
