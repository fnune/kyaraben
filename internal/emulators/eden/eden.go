package eden

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/configformat"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDEden,
		Name:    "Eden",
		Systems: []model.SystemID{model.SystemIDSwitch},
		Package: model.AppImageRef("eden"),
		// Eden copies keys to ~/.local/share/eden/keys/ on import (just a file copy).
		// See: https://gitlab.com/codxjb/eden/-/blob/master/src/yuzu/main.cpp#L4353
		// Firmware is copied to nand/system/Contents/registered/ as *.nca files.
		// See: https://gitlab.com/codxjb/eden/-/blob/master/src/yuzu/main.cpp#L4216
		ProvisionGroups: []model.ProvisionGroup{
			{
				MinRequired: 1,
				Message:     "Decryption keys required",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionKeys, "prod.keys", "Production keys"),
				},
			},
			{
				MinRequired: 0,
				Message:     "Title keys (optional, for DLC and updates)",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionKeys, "title.keys", "Title keys"),
				},
			},
			{
				MinRequired: 0,
				Message:     "Firmware (optional, enables system applets)",
				Provisions: []model.Provision{
					model.PatternProvision(model.ProvisionFirmware, "*.nca", "firmware NCA", "Switch firmware"),
				},
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "eden",
			GenericName: "Nintendo Switch Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if len(opts.LaunchArgs) > 0 {
					cmd += " " + strings.Join(opts.LaunchArgs, " ")
				}
				if opts.Fullscreen {
					cmd += " -f"
				}
				cmd += " -g %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			UsesBiosDir:        true,
			UsesSavesDir:       true,
			UsesScreenshotsDir: true,
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

var configTarget = model.ConfigTarget{
	RelPath: "eden/qt-config.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var profileTarget = model.ConfigTarget{
	RelPath: "eden/input/Kyaraben.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store

	entries := []model.ConfigEntry{
		{Path: []string{"UI", "Screenshots\\screenshot_path"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDEden)},
		{Path: []string{"UI", "Paths\\gamedirs\\size"}, Value: "1"},
		{Path: []string{"UI", "Paths\\gamedirs\\1\\deep_scan"}, Value: "false"},
		{Path: []string{"UI", "Paths\\gamedirs\\1\\expanded"}, Value: "true"},
		{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: store.SystemRomsDir(model.SystemIDSwitch)},
	}

	entries = append(entries, performanceEntries()...)

	var patches []model.ConfigPatch

	if cc := ctx.ControllerConfig; cc != nil {
		entries = append(entries, qtConfigControllerEntries(cc)...)
		entries = append(entries, hotkeyEntries(cc)...)
		patches = append(patches, model.ConfigPatch{
			Target:         profileTarget,
			Entries:        profileEntries(cc),
			ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
		})
	}

	patches = append(patches, model.ConfigPatch{Target: configTarget, Entries: entries})

	dataDir, err := ctx.BaseDirResolver.UserDataDir()
	if err != nil {
		return model.GenerateResult{}, err
	}
	edenDir := filepath.Join(dataDir, "eden")
	biosDir := store.SystemBiosDir(model.SystemIDSwitch)

	savesDir := store.SystemSavesDir(model.SystemIDSwitch)
	symlinks := []model.SymlinkSpec{
		{Source: filepath.Join(edenDir, "keys"), Target: biosDir},
		{Source: filepath.Join(edenDir, "nand", "system", "Contents", "registered"), Target: biosDir},
		{Source: filepath.Join(edenDir, "screenshots"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDEden)},
		{Source: filepath.Join(edenDir, "nand", "user", "save"), Target: filepath.Join(savesDir, "user")},
		{Source: filepath.Join(edenDir, "nand", "system", "save"), Target: filepath.Join(savesDir, "system")},
	}

	return model.GenerateResult{
		Patches:  patches,
		Symlinks: symlinks,
	}, nil
}

func performanceEntries() []model.ConfigEntry {
	settings := []struct {
		section string
		key     string
		value   string
	}{
		{"Core", "use_multi_core", "true"},
		{"Cpu", "cpu_accuracy", "0"},
		{"Renderer", "backend", "1"},
		{"Renderer", "gpu_accuracy", "0"},
		{"Renderer", "use_asynchronous_gpu_emulation", "true"},
		{"Renderer", "use_asynchronous_shaders", "true"},
		{"Renderer", "use_disk_shader_cache", "true"},
		{"Renderer", "fast_gpu_time", "1"},
		{"Renderer", "resolution_setup", "2"},
		{"Renderer", "scaling_filter", "5"},
		{"Renderer", "fsr_sharpening_slider", "25"},
		{"Renderer", "use_vsync", "2"},
		{"Renderer", "fullscreen_mode", "1"},
		{"Renderer", "fps_cap", "1000"},
		{"System", "use_docked_mode", "1"},
	}

	var entries []model.ConfigEntry
	for _, s := range settings {
		entries = append(entries,
			model.ConfigEntry{Path: []string{s.section, s.key}, Value: s.value, DefaultOnly: true},
			model.ConfigEntry{Path: []string{s.section, s.key + `\default`}, Value: "false", DefaultOnly: true},
		)
	}
	return entries
}

// Steam Deck raw joystick button indices. Eden uses raw SDL joystick API
// (not GameController), so button indices differ from SDL GameController standard.
// Indices correspond to physical positions, matching EmuDeck's configuration.
const (
	steamDeckButtonA      = 0 // south position
	steamDeckButtonB      = 1 // east position
	steamDeckButtonX      = 2 // west position
	steamDeckButtonY      = 3 // north position
	steamDeckButtonL      = 4
	steamDeckButtonR      = 5
	steamDeckButtonMinus  = 6
	steamDeckButtonPlus   = 7
	steamDeckButtonHome   = 8
	steamDeckButtonLStick = 9
	steamDeckButtonRStick = 10
)

// Eden (yuzu-based) embeds GUID in every binding.
// Key ordering must match Eden's native format to avoid config churn when Eden
// rewrites its config on close. Eden uses: engine,port,guid,<binding-specific>
func edenButtonRef(guid string, port, button int) string {
	return fmt.Sprintf("engine:sdl,port:%d,guid:%s,button:%d", port, guid, button)
}

func edenAxisRef(guid string, port, axis int) string {
	return fmt.Sprintf("engine:sdl,port:%d,guid:%s,axis:%d,threshold:0.500000", port, guid, axis)
}

func edenHatRef(guid string, port, hat int, direction string) string {
	return fmt.Sprintf("engine:sdl,port:%d,guid:%s,direction:%s,hat:%d", port, guid, direction, hat)
}

func edenStickRef(guid string, port, axisX, axisY int) string {
	return fmt.Sprintf("engine:sdl,port:%d,guid:%s,axis_x:%d,axis_y:%d,deadzone:0.100000", port, guid, axisX, axisY)
}

// bindingEntry creates a ConfigEntry for Eden binding values with semantic
// comparison enabled. Eden's key ordering in binding strings is nondeterministic.
func bindingEntry(path []string, value string) model.ConfigEntry {
	return model.ConfigEntry{
		Path:         path,
		Value:        value,
		EqualityFunc: configformat.BindingValuesEqual,
	}
}

// steamDeckFaceButton returns the Steam Deck raw joystick button index for a face button.
func steamDeckFaceButton(btn model.SDLButton) int {
	switch btn {
	case model.ButtonA:
		return steamDeckButtonA
	case model.ButtonB:
		return steamDeckButtonB
	case model.ButtonX:
		return steamDeckButtonX
	case model.ButtonY:
		return steamDeckButtonY
	default:
		return steamDeckButtonA
	}
}

// profileEntries returns entries for the Kyaraben.ini profile file.
// This profile is fully managed (FileRegion) and can be reloaded by users
// at any time to restore kyaraben bindings.
func profileEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	south, east, west, north := cc.FaceButtons(model.SystemIDSwitch)
	guid := model.SteamDeckGUID

	return []model.ConfigEntry{
		{Path: []string{"Controls", "type"}, Value: "0"},
		bindingEntry([]string{"Controls", "button_a"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, steamDeckFaceButton(east)))),
		bindingEntry([]string{"Controls", "button_b"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, steamDeckFaceButton(south)))),
		bindingEntry([]string{"Controls", "button_x"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, steamDeckFaceButton(north)))),
		bindingEntry([]string{"Controls", "button_y"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, steamDeckFaceButton(west)))),
		bindingEntry([]string{"Controls", "button_lstick"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, steamDeckButtonLStick))),
		bindingEntry([]string{"Controls", "button_rstick"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, steamDeckButtonRStick))),
		bindingEntry([]string{"Controls", "button_l"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, steamDeckButtonL))),
		bindingEntry([]string{"Controls", "button_r"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, steamDeckButtonR))),
		bindingEntry([]string{"Controls", "button_zl"}, fmt.Sprintf(`"%s"`, edenAxisRef(guid, 0, model.AxisLeftTrigger))),
		bindingEntry([]string{"Controls", "button_zr"}, fmt.Sprintf(`"%s"`, edenAxisRef(guid, 0, model.AxisRightTrigger))),
		bindingEntry([]string{"Controls", "button_plus"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, steamDeckButtonPlus))),
		bindingEntry([]string{"Controls", "button_minus"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, steamDeckButtonMinus))),
		bindingEntry([]string{"Controls", "button_dleft"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, 0, 0, "left"))),
		bindingEntry([]string{"Controls", "button_dright"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, 0, 0, "right"))),
		bindingEntry([]string{"Controls", "button_dup"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, 0, 0, "up"))),
		bindingEntry([]string{"Controls", "button_ddown"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, 0, 0, "down"))),
		bindingEntry([]string{"Controls", "lstick"}, fmt.Sprintf(`"%s"`, edenStickRef(guid, 0, model.AxisLeftX, model.AxisLeftY))),
		bindingEntry([]string{"Controls", "rstick"}, fmt.Sprintf(`"%s"`, edenStickRef(guid, 0, model.AxisRightX, model.AxisRightY))),
	}
}

// qtConfigControllerEntries returns entries for qt-config.ini.
// Bindings are DefaultOnly so user customizations are preserved; users can
// reload kyaraben bindings via the Kyaraben profile.
func qtConfigControllerEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	var entries []model.ConfigEntry
	south, east, west, north := cc.FaceButtons(model.SystemIDSwitch)

	guid := model.SteamDeckGUID

	for i := 0; i < 2; i++ {
		prefix := fmt.Sprintf("player_%d_", i)
		entries = append(entries,
			model.ConfigEntry{Path: []string{"Controls", prefix + "connected"}, Value: "true", DefaultOnly: i > 0},
			model.ConfigEntry{Path: []string{"Controls", prefix + "profile_name"}, Value: "Kyaraben", DefaultOnly: true},
			model.ConfigEntry{Path: []string{"Controls", prefix + "type"}, Value: "0", DefaultOnly: true},
		)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_a"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, steamDeckFaceButton(east))))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_b"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, steamDeckFaceButton(south))))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_x"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, steamDeckFaceButton(north))))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_y"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, steamDeckFaceButton(west))))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_lstick"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, steamDeckButtonLStick)))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_rstick"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, steamDeckButtonRStick)))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_l"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, steamDeckButtonL)))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_r"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, steamDeckButtonR)))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_zl"}, fmt.Sprintf(`"%s"`, edenAxisRef(guid, i, model.AxisLeftTrigger)))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_zr"}, fmt.Sprintf(`"%s"`, edenAxisRef(guid, i, model.AxisRightTrigger)))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_plus"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, steamDeckButtonPlus)))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_minus"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, steamDeckButtonMinus)))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_dleft"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "left")))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_dright"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "right")))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_dup"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "up")))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "button_ddown"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "down")))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "lstick"}, fmt.Sprintf(`"%s"`, edenStickRef(guid, i, model.AxisLeftX, model.AxisLeftY)))...)
		entries = append(entries, defaultBindingEntries([]string{"Controls", prefix + "rstick"}, fmt.Sprintf(`"%s"`, edenStickRef(guid, i, model.AxisRightX, model.AxisRightY)))...)
	}

	return entries
}

func defaultBindingEntries(path []string, value string) []model.ConfigEntry {
	defaultFlagPath := make([]string, len(path))
	copy(defaultFlagPath, path)
	defaultFlagPath[len(defaultFlagPath)-1] = defaultFlagPath[len(defaultFlagPath)-1] + `\default`

	return []model.ConfigEntry{
		{
			Path:         path,
			Value:        value,
			DefaultOnly:  true,
			EqualityFunc: configformat.BindingValuesEqual,
		},
		{
			Path:        defaultFlagPath,
			Value:       "false",
			DefaultOnly: true,
		},
	}
}

// edenButtonName maps SDL button names to Eden's Switch-style button names for hotkeys.
// Names match EmuDeck's configuration format.
func edenButtonName(b model.SDLButton) string {
	switch b {
	case model.ButtonA:
		return "A"
	case model.ButtonB:
		return "B"
	case model.ButtonX:
		return "X"
	case model.ButtonY:
		return "Y"
	case model.ButtonBack:
		return "Minus"
	case model.ButtonStart:
		return "Plus"
	case model.ButtonGuide:
		return "Home"
	case model.ButtonLeftShoulder:
		return "L"
	case model.ButtonRightShoulder:
		return "R"
	case model.ButtonLeftTrigger:
		return "ZL"
	case model.ButtonRightTrigger:
		return "ZR"
	case model.ButtonLeftStick:
		return "Left_Stick"
	case model.ButtonRightStick:
		return "Right_Stick"
	case model.ButtonDPadUp:
		return "Dpad_Up"
	case model.ButtonDPadDown:
		return "Dpad_Down"
	case model.ButtonDPadLeft:
		return "Dpad_Left"
	case model.ButtonDPadRight:
		return "Dpad_Right"
	default:
		return string(b)
	}
}

func edenHotkeyRef(binding model.HotkeyBinding) string {
	parts := make([]string, len(binding.Buttons))
	for i, b := range binding.Buttons {
		parts[i] = edenButtonName(b)
	}
	return strings.Join(parts, "+")
}

func hotkeyEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	hk := cc.Hotkeys

	type mapping struct {
		key     string
		binding model.HotkeyBinding
	}
	// Eden uses Shortcuts\ as part of the key name, not as a section header.
	mappings := []mapping{
		{`Shortcuts\Main%20Window\Continue\Pause%20Emulation\Controller_KeySeq`, hk.Pause},
		{`Shortcuts\Main%20Window\Exit%20Eden\Controller_KeySeq`, hk.Quit},
		{`Shortcuts\Main%20Window\Capture%20Screenshot\Controller_KeySeq`, hk.Screenshot},
		{`Shortcuts\Main%20Window\Fullscreen\Controller_KeySeq`, hk.ToggleFullscreen},
		{`Shortcuts\Main%20Window\Toggle%20Framerate%20Limit\Controller_KeySeq`, hk.FastForward},
	}

	var entries []model.ConfigEntry
	for _, m := range mappings {
		if len(m.binding.Buttons) > 0 {
			entries = append(entries,
				model.ConfigEntry{
					Path:  []string{m.key},
					Value: edenHotkeyRef(m.binding),
				},
				model.ConfigEntry{
					Path:  []string{m.key + `\default`},
					Value: "false",
				},
			)
		}
	}
	return entries
}
