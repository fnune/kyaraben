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

var SharedLauncher = model.LauncherInfo{
	Binary:      "retroarch",
	DisplayName: "RetroArch",
	GenericName: "Multi-system Emulator",
	Categories:  []string{"Game", "Emulator"},
}

// LauncherWithCore returns a copy of SharedLauncher with RomCommand set to load
// the given libretro core via the -L flag when launching games.
func LauncherWithCore(coreName string) model.LauncherInfo {
	l := SharedLauncher
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

// SharedConfig generates the base RetroArch configuration shared by all cores.
// Enables per-core sorting so RetroArch creates subdirectories like saves/bsnes/.
// We symlink these sorted directories to kyaraben's store locations.
// Screenshots go directly to a shared retroarch directory (no per-core sorting).
// See: https://docs.libretro.com/guides/change-directories/
func SharedConfig(store model.StoreReader, cc *model.ControllerConfig) model.ConfigPatch {
	entries := []model.ConfigEntry{
		{Path: []string{"libretro_directory"}, Value: store.CoresDir()},
		{Path: []string{"screenshot_directory"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDRetroArchBsnes)},
		{Path: []string{"sort_savefiles_enable"}, Value: "true"},
		{Path: []string{"sort_savestates_enable"}, Value: "true"},
		{Path: []string{"sort_savefiles_by_content_enable"}, Value: "false"},
		{Path: []string{"sort_savestates_by_content_enable"}, Value: "false"},
		{Path: []string{"rgui_browser_directory"}, Value: store.RomsDir()},
		{Path: []string{"menu_driver"}, Value: "ozone", DefaultOnly: true},
		{Path: []string{"menu_show_load_content_animation"}, Value: "false"},
		{Path: []string{"notification_show_config_override_load"}, Value: "false", DefaultOnly: true},
		{Path: []string{"notification_show_remap_load"}, Value: "false", DefaultOnly: true},
		{Path: []string{"notification_show_autoconfig"}, Value: "false", DefaultOnly: true},
		{Path: []string{"quit_press_twice"}, Value: "false"},
		{Path: []string{"input_player1_analog_dpad_mode"}, Value: "1", DefaultOnly: true},
		{Path: []string{"input_player2_analog_dpad_mode"}, Value: "1", DefaultOnly: true},
		{Path: []string{"input_player3_analog_dpad_mode"}, Value: "1", DefaultOnly: true},
		{Path: []string{"input_player4_analog_dpad_mode"}, Value: "1", DefaultOnly: true},
	}

	if cc != nil {
		entries = append(entries, controllerEntries(cc)...)
	}

	return model.ConfigPatch{Target: MainConfigTarget, Entries: entries}
}

func controllerEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	entries := []model.ConfigEntry{
		{Path: []string{"input_joypad_driver"}, Value: "sdl2"},
		{Path: []string{"input_autodetect_enable"}, Value: "true"},
	}

	south, east, west, north := cc.FaceButtons()
	for i := 1; i <= 4; i++ {
		prefix := fmt.Sprintf("input_player%d_", i)
		entries = append(entries,
			model.ConfigEntry{Path: []string{prefix + "a_btn"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[east])},
			model.ConfigEntry{Path: []string{prefix + "b_btn"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[south])},
			model.ConfigEntry{Path: []string{prefix + "x_btn"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[north])},
			model.ConfigEntry{Path: []string{prefix + "y_btn"}, Value: fmt.Sprintf("%d", model.SDLButtonIndex[west])},
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
			entries = append(entries, model.ConfigEntry{
				Path:  []string{"input_enable_hotkey_btn"},
				Value: fmt.Sprintf("%d", cc.SDLIndex(enableBtn)),
			})
			enableBtnSet = true
		}
		actionBtn := m.binding.Buttons[len(m.binding.Buttons)-1]
		entries = append(entries, model.ConfigEntry{
			Path:  []string{m.key},
			Value: fmt.Sprintf("%d", cc.SDLIndex(actionBtn)),
		})
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
}

func CoreShortName(emuID model.EmulatorID) string {
	return coreShortNames[emuID]
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
