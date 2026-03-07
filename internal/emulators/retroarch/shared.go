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
	"github.com/fnune/kyaraben/internal/shaders"
)

const shadersDir = "retroarch/shaders/kyaraben"

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
	case model.PresetClean:
		entries = append(entries,
			model.Entry(model.Preset, model.Path("video_scale_integer"), "false"),
			model.Entry(model.Preset, model.Path("video_shader_enable"), "false"),
			model.Entry(model.Preset, model.Path("video_smooth"), "false"),
			model.Entry(model.Preset, model.Path("aspect_ratio_index"), "22"),
		)
	case model.PresetRetro:
		entries = append(entries,
			model.Entry(model.Preset, model.Path("video_scale_integer"), "false"),
			model.Entry(model.Preset, model.Path("video_shader_enable"), "true"),
			model.Entry(model.Preset, model.Path("video_smooth"), "false"),
			model.Entry(model.Preset, model.Path("aspect_ratio_index"), "22"),
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

// CoreInfo contains metadata for a RetroArch libretro core.
type CoreInfo struct {
	// ShortName is the core's filename stem (e.g., "genesis_plus_gx" for genesis_plus_gx_libretro.so).
	ShortName string
	// LibraryName is the value from retro_system_info.library_name that the core reports.
	// RetroArch uses this for per-core save/state directories when sort_savefiles_enable is true.
	// See: https://github.com/libretro/RetroArch/blob/master/runloop.c (runloop_path_set_redirect, ~line 8303)
	// Values sourced from: https://github.com/libretro/libretro-core-info (corename field in .info files)
	LibraryName     string
	SystemID        model.SystemID
	NeedsBiosDir    bool
	UsesHWRendering bool
}

var coreRegistry = map[model.EmulatorID]CoreInfo{
	model.EmulatorIDRetroArchBsnes: {
		ShortName:   "bsnes",
		LibraryName: "bsnes",
		SystemID:    model.SystemIDSNES,
	},
	model.EmulatorIDRetroArchSnes9x: {
		ShortName:   "snes9x",
		LibraryName: "Snes9x",
		SystemID:    model.SystemIDSNES,
	},
	model.EmulatorIDRetroArchMesen: {
		ShortName:   "mesen",
		LibraryName: "Mesen",
		SystemID:    model.SystemIDNES,
	},
	model.EmulatorIDRetroArchGenesisPlusGX: {
		ShortName:    "genesis_plus_gx",
		LibraryName:  "Genesis Plus GX",
		SystemID:     model.SystemIDGenesis,
		NeedsBiosDir: true,
	},
	model.EmulatorIDRetroArchMupen64Plus: {
		ShortName:       "mupen64plus_next",
		LibraryName:     "Mupen64Plus-Next",
		SystemID:        model.SystemIDN64,
		UsesHWRendering: true,
	},
	model.EmulatorIDRetroArchBeetleSaturn: {
		ShortName:    "mednafen_saturn",
		LibraryName:  "Beetle Saturn",
		SystemID:     model.SystemIDSaturn,
		NeedsBiosDir: true,
	},
	model.EmulatorIDRetroArchBeetlePCE: {
		ShortName:    "mednafen_pce_fast",
		LibraryName:  "Beetle PCE Fast",
		SystemID:     model.SystemIDPCEngine,
		NeedsBiosDir: true,
	},
	model.EmulatorIDRetroArchBeetleNGP: {
		ShortName:   "mednafen_ngp",
		LibraryName: "Beetle NeoPop",
		SystemID:    model.SystemIDNGP,
	},
	model.EmulatorIDRetroArchMGBA: {
		ShortName:   "mgba",
		LibraryName: "mGBA",
		SystemID:    model.SystemIDGBA,
	},
	model.EmulatorIDRetroArchMelonDS: {
		ShortName:       "melondsds",
		LibraryName:     "melonDS DS",
		SystemID:        model.SystemIDNDS,
		UsesHWRendering: true,
	},
	model.EmulatorIDRetroArchCitra: {
		ShortName:       "citra",
		LibraryName:     "Citra",
		SystemID:        model.SystemIDN3DS,
		UsesHWRendering: true,
	},
	model.EmulatorIDRetroArchFBNeo: {
		ShortName:   "fbneo",
		LibraryName: "FinalBurn Neo",
		SystemID:    model.SystemIDArcade,
	},
	model.EmulatorIDRetroArchStella: {
		ShortName:   "stella",
		LibraryName: "Stella",
		SystemID:    model.SystemIDAtari2600,
	},
	model.EmulatorIDRetroArchVICE: {
		ShortName:   "vice_x64sc",
		LibraryName: "VICE x64sc",
		SystemID:    model.SystemIDC64,
	},
}

// IsRetroArchCore returns true if the emulator ID is a registered RetroArch core.
func IsRetroArchCore(emuID model.EmulatorID) bool {
	_, ok := coreRegistry[emuID]
	return ok
}

// MustGetCoreInfo returns the CoreInfo for a RetroArch core.
// Panics if the emulator ID is not a registered RetroArch core.
func MustGetCoreInfo(emuID model.EmulatorID) CoreInfo {
	info, ok := coreRegistry[emuID]
	if !ok {
		panic(fmt.Sprintf("retroarch: no CoreInfo registered for %s", emuID))
	}
	return info
}

// CoreShortName returns the core's short name, or empty string if not a RetroArch core.
func CoreShortName(emuID model.EmulatorID) string {
	if info, ok := coreRegistry[emuID]; ok {
		return info.ShortName
	}
	return ""
}

func CoreOverrideTarget(shortName string) model.ConfigTarget {
	for _, info := range coreRegistry {
		if info.ShortName == shortName {
			return model.ConfigTarget{
				RelPath: "retroarch/config/" + info.LibraryName + "/" + info.LibraryName + ".cfg",
				Format:  model.ConfigFormatCFG,
				BaseDir: model.ConfigBaseDirUserConfig,
			}
		}
	}
	panic(fmt.Sprintf("retroarch: no CoreInfo registered for short name %q", shortName))
}

// CorePatches returns the base RetroArch config plus a per-core override for
// cores that need additional settings like system_directory for BIOS files.
func CorePatches(emuID model.EmulatorID, store model.StoreReader, cc *model.ControllerConfig, pc *PresetConfig, resolver model.BaseDirResolver) []model.ConfigPatch {
	patches := []model.ConfigPatch{SharedConfig(store, cc, pc)}

	if !IsRetroArchCore(emuID) {
		return patches
	}
	info := MustGetCoreInfo(emuID)

	var entries []model.ConfigEntry
	if info.NeedsBiosDir {
		entries = append(entries, model.Entry(model.Store, model.Path("system_directory"), store.SystemBiosDir(info.SystemID)))
	}

	if len(entries) > 0 {
		patches = append(patches, model.ConfigPatch{
			Target:  CoreOverrideTarget(info.ShortName),
			Entries: entries,
		})
	}

	return patches
}

func isMultiSystemCore(emuID model.EmulatorID) bool {
	switch emuID {
	case model.EmulatorIDRetroArchMGBA:
		return true
	default:
		return false
	}
}

// multiSystemCoreSystems returns the systems a multi-system core supports.
func multiSystemCoreSystems(emuID model.EmulatorID) []model.SystemID {
	switch emuID {
	case model.EmulatorIDRetroArchMGBA:
		return []model.SystemID{model.SystemIDGB, model.SystemIDGBC, model.SystemIDGBA}
	default:
		return nil
	}
}

func CoreShaderDownloads(emuID model.EmulatorID, resolver model.BaseDirResolver, pc *PresetConfig) ([]model.InitialDownload, error) {
	if pc == nil || pc.Preset != model.PresetRetro {
		return nil, nil
	}

	if !IsRetroArchCore(emuID) {
		return nil, nil
	}

	var systems []model.SystemID
	if isMultiSystemCore(emuID) {
		systems = multiSystemCoreSystems(emuID)
	} else {
		info := MustGetCoreInfo(emuID)
		systems = []model.SystemID{info.SystemID}
	}

	shaderTypesNeeded := make(map[shaders.Type]bool)
	for _, systemID := range systems {
		shaderTypesNeeded[shaders.TypeForSystem(systemID)] = true
	}

	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting config dir: %w", err)
	}

	downloaded := make(map[string]bool)
	var downloads []model.InitialDownload
	for shaderType := range shaderTypesNeeded {
		files := shaders.DownloadFiles(shaderType)
		for _, f := range files {
			if downloaded[f.Path] {
				continue
			}
			downloaded[f.Path] = true

			var destPath string
			if shaderType == shaders.TypeGameboy {
				destPath = filepath.Join(configDir, shadersDir, f.Path[len("handheld/shaders/"):])
			} else {
				destPath = filepath.Join(configDir, shadersDir, filepath.Base(f.Path))
			}
			downloads = append(downloads, model.InitialDownload{
				URL:      shaders.SlangShadersBaseURL + "/" + f.Path,
				SHA256:   f.SHA256,
				DestPath: destPath,
			})
		}
	}

	return downloads, nil
}

func CoreSymlinks(emuID model.EmulatorID, store model.StoreReader, resolver model.BaseDirResolver) ([]model.SymlinkSpec, error) {
	if !IsRetroArchCore(emuID) {
		return nil, nil
	}
	info := MustGetCoreInfo(emuID)

	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting config dir: %w", err)
	}

	return []model.SymlinkSpec{
		{
			Source: filepath.Join(configDir, "retroarch", "saves", info.LibraryName),
			Target: store.SystemSavesDir(info.SystemID),
		},
		{
			Source: filepath.Join(configDir, "retroarch", "states", info.LibraryName),
			Target: store.EmulatorStatesDir(emuID),
		},
	}, nil
}

// CoreOptionsTarget returns the config target for a core's .opt file.
func CoreOptionsTarget(emuID model.EmulatorID) model.ConfigTarget {
	info := MustGetCoreInfo(emuID)
	return model.ConfigTarget{
		RelPath: "retroarch/config/" + info.LibraryName + "/" + info.LibraryName + ".opt",
		Format:  model.ConfigFormatCFG,
		BaseDir: model.ConfigBaseDirUserConfig,
	}
}

// CoreOptionsPatch returns a config patch for core-specific options.
// For multi-system cores like mGBA, base options are set here and per-system
// overrides are handled by ContentDirOptionsPatches.
func CoreOptionsPatch(emuID model.EmulatorID, pc *PresetConfig) *model.ConfigPatch {
	if pc == nil || !IsRetroArchCore(emuID) {
		return nil
	}

	var entries []model.ConfigEntry

	if pc.Bezels && emuID == model.EmulatorIDRetroArchMGBA {
		entries = append(entries, model.Entry(model.Preset, model.Path("mgba_sgb_borders"), "OFF"))
	}

	if len(entries) == 0 {
		return nil
	}

	return &model.ConfigPatch{
		Target:  CoreOptionsTarget(emuID),
		Entries: entries,
	}
}

// ContentDirOverrideTarget returns the config target for a per-content-directory override.
// This allows different settings when loading ROMs from different system directories.
func ContentDirOverrideTarget(emuID model.EmulatorID, systemID model.SystemID) model.ConfigTarget {
	info := MustGetCoreInfo(emuID)
	return model.ConfigTarget{
		RelPath: "retroarch/config/" + info.LibraryName + "/" + string(systemID) + ".cfg",
		Format:  model.ConfigFormatCFG,
		BaseDir: model.ConfigBaseDirUserConfig,
	}
}

// ContentDirOptionsTarget returns the config target for per-content-directory core options.
func ContentDirOptionsTarget(emuID model.EmulatorID, systemID model.SystemID) model.ConfigTarget {
	info := MustGetCoreInfo(emuID)
	return model.ConfigTarget{
		RelPath: "retroarch/config/" + info.LibraryName + "/" + string(systemID) + ".opt",
		Format:  model.ConfigFormatCFG,
		BaseDir: model.ConfigBaseDirUserConfig,
	}
}

// ContentDirOptionsPatches creates per-content-directory core options.
// This allows multi-system cores (like mGBA) to use different options for each system.
// Color correction settings are applied regardless of preset for accurate colors.
func ContentDirOptionsPatches(emuID model.EmulatorID, systems []model.SystemID, pc *PresetConfig) []model.ConfigPatch {
	if pc == nil || !IsRetroArchCore(emuID) {
		return nil
	}

	var patches []model.ConfigPatch

	if emuID == model.EmulatorIDRetroArchMGBA {
		for _, systemID := range systems {
			var entries []model.ConfigEntry
			switch systemID {
			case model.SystemIDGB:
				if pc.Preset == model.PresetRetro {
					entries = append(entries,
						model.Entry(model.Preset, model.Path("mgba_gb_colors"), "Grayscale"),
					)
				}
			case model.SystemIDGBC:
				entries = append(entries,
					model.Entry(model.None, model.Path("mgba_color_correction"), "Game Boy Color"),
				)
			case model.SystemIDGBA:
				entries = append(entries,
					model.Entry(model.None, model.Path("mgba_color_correction"), "OFF"),
				)
			}
			if len(entries) > 0 {
				patches = append(patches, model.ConfigPatch{
					Target:  ContentDirOptionsTarget(emuID, systemID),
					Entries: entries,
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

// CoreEmbeddedFiles returns shader preset files for embedded installation.
// Creates the main .slangp presets in the shaders directory, plus reference
// presets in the core config directory for auto-loading.
func CoreEmbeddedFiles(emuID model.EmulatorID, systems []model.SystemID, pc *PresetConfig, resolver model.BaseDirResolver) ([]model.EmbeddedFile, error) {
	if pc == nil || pc.Preset != model.PresetRetro {
		return nil, nil
	}

	if !IsRetroArchCore(emuID) {
		return nil, nil
	}

	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting config dir: %w", err)
	}

	info := MustGetCoreInfo(emuID)
	var files []model.EmbeddedFile

	shaderTypesNeeded := make(map[shaders.Type]bool)
	for _, systemID := range systems {
		shaderTypesNeeded[shaders.TypeForSystem(systemID)] = true
	}

	for shaderType := range shaderTypesNeeded {
		content := shaders.PresetContent(shaderType)
		if content == "" {
			continue
		}
		presetPath := filepath.Join(configDir, shadersDir, string(shaderType)+".slangp")
		files = append(files, model.EmbeddedFile{
			Content:  []byte(content),
			DestPath: presetPath,
		})
	}

	if isMultiSystemCore(emuID) {
		for _, systemID := range systems {
			shaderType := shaders.TypeForSystem(systemID)
			refContent := fmt.Sprintf(`#reference "../../shaders/kyaraben/%s.slangp"`, shaderType)
			refPath := filepath.Join(configDir, "retroarch/config", info.LibraryName, string(systemID)+".slangp")
			files = append(files, model.EmbeddedFile{
				Content:  []byte(refContent),
				DestPath: refPath,
			})
		}
	} else {
		shaderType := shaders.TypeForSystem(info.SystemID)
		refContent := fmt.Sprintf(`#reference "../../shaders/kyaraben/%s.slangp"`, shaderType)
		refPath := filepath.Join(configDir, "retroarch/config", info.LibraryName, info.LibraryName+".slangp")
		files = append(files, model.EmbeddedFile{
			Content:  []byte(refContent),
			DestPath: refPath,
		})
	}

	return files, nil
}

// ContentDirShaderPatches is deprecated - shader presets are now created via
// CoreEmbeddedFiles using RetroArch's #reference directive.
// Kept for API compatibility with existing emulator generators.
func ContentDirShaderPatches(_ model.EmulatorID, _ []model.SystemID, _ *PresetConfig, _ model.BaseDirResolver) []model.ConfigPatch {
	return nil
}
