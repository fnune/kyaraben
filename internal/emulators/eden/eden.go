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

	symlinks := []model.SymlinkSpec{
		{Source: filepath.Join(edenDir, "keys"), Target: biosDir},
		{Source: filepath.Join(edenDir, "nand", "system", "Contents", "registered"), Target: biosDir},
		{Source: filepath.Join(edenDir, "screenshots"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDEden)},
		{Source: filepath.Join(edenDir, "nand", "user", "save"), Target: store.SystemSavesDir(model.SystemIDSwitch)},
	}

	return model.GenerateResult{
		Patches:  patches,
		Symlinks: symlinks,
	}, nil
}

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

// profileEntries returns entries for the Kyaraben.ini profile file.
// This profile is fully managed (FileRegion) and can be reloaded by users
// at any time to restore kyaraben bindings.
func profileEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	south, east, west, north := cc.FaceButtons()
	guid := model.SteamDeckGUID

	faceMap := map[string]model.SDLButton{
		"a": east,
		"b": south,
		"x": north,
		"y": west,
	}

	return []model.ConfigEntry{
		{Path: []string{"Controls", "type"}, Value: "0"},
		bindingEntry([]string{"Controls", "button_a"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, model.SDLButtonIndex[faceMap["a"]]))),
		bindingEntry([]string{"Controls", "button_b"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, model.SDLButtonIndex[faceMap["b"]]))),
		bindingEntry([]string{"Controls", "button_x"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, model.SDLButtonIndex[faceMap["x"]]))),
		bindingEntry([]string{"Controls", "button_y"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, model.SDLButtonIndex[faceMap["y"]]))),
		bindingEntry([]string{"Controls", "button_lstick"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, model.SDLButtonIndex[model.ButtonLeftStick]))),
		bindingEntry([]string{"Controls", "button_rstick"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, model.SDLButtonIndex[model.ButtonRightStick]))),
		bindingEntry([]string{"Controls", "button_l"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, model.SDLButtonIndex[model.ButtonLeftShoulder]))),
		bindingEntry([]string{"Controls", "button_r"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, model.SDLButtonIndex[model.ButtonRightShoulder]))),
		bindingEntry([]string{"Controls", "button_zl"}, fmt.Sprintf(`"%s"`, edenAxisRef(guid, 0, model.AxisLeftTrigger))),
		bindingEntry([]string{"Controls", "button_zr"}, fmt.Sprintf(`"%s"`, edenAxisRef(guid, 0, model.AxisRightTrigger))),
		bindingEntry([]string{"Controls", "button_plus"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, model.SDLButtonIndex[model.ButtonStart]))),
		bindingEntry([]string{"Controls", "button_minus"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, 0, model.SDLButtonIndex[model.ButtonBack]))),
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
	south, east, west, north := cc.FaceButtons()

	guid := model.SteamDeckGUID

	faceMap := map[string]model.SDLButton{
		"a": east,
		"b": south,
		"x": north,
		"y": west,
	}

	for i := 0; i < 2; i++ {
		prefix := fmt.Sprintf("player_%d_", i)
		entries = append(entries,
			model.ConfigEntry{Path: []string{"Controls", prefix + "connected"}, Value: "true", DefaultOnly: i > 0},
			model.ConfigEntry{Path: []string{"Controls", prefix + "profile_name"}, Value: "Kyaraben", DefaultOnly: true},
			model.ConfigEntry{Path: []string{"Controls", prefix + "type"}, Value: "0", DefaultOnly: true},
			defaultBindingEntry([]string{"Controls", prefix + "button_a"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[faceMap["a"]]))),
			defaultBindingEntry([]string{"Controls", prefix + "button_b"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[faceMap["b"]]))),
			defaultBindingEntry([]string{"Controls", prefix + "button_x"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[faceMap["x"]]))),
			defaultBindingEntry([]string{"Controls", prefix + "button_y"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[faceMap["y"]]))),
			defaultBindingEntry([]string{"Controls", prefix + "button_lstick"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonLeftStick]))),
			defaultBindingEntry([]string{"Controls", prefix + "button_rstick"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonRightStick]))),
			defaultBindingEntry([]string{"Controls", prefix + "button_l"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonLeftShoulder]))),
			defaultBindingEntry([]string{"Controls", prefix + "button_r"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonRightShoulder]))),
			defaultBindingEntry([]string{"Controls", prefix + "button_zl"}, fmt.Sprintf(`"%s"`, edenAxisRef(guid, i, model.AxisLeftTrigger))),
			defaultBindingEntry([]string{"Controls", prefix + "button_zr"}, fmt.Sprintf(`"%s"`, edenAxisRef(guid, i, model.AxisRightTrigger))),
			defaultBindingEntry([]string{"Controls", prefix + "button_plus"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonStart]))),
			defaultBindingEntry([]string{"Controls", prefix + "button_minus"}, fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonBack]))),
			defaultBindingEntry([]string{"Controls", prefix + "button_dleft"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "left"))),
			defaultBindingEntry([]string{"Controls", prefix + "button_dright"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "right"))),
			defaultBindingEntry([]string{"Controls", prefix + "button_dup"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "up"))),
			defaultBindingEntry([]string{"Controls", prefix + "button_ddown"}, fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "down"))),
			defaultBindingEntry([]string{"Controls", prefix + "lstick"}, fmt.Sprintf(`"%s"`, edenStickRef(guid, i, model.AxisLeftX, model.AxisLeftY))),
			defaultBindingEntry([]string{"Controls", prefix + "rstick"}, fmt.Sprintf(`"%s"`, edenStickRef(guid, i, model.AxisRightX, model.AxisRightY))),
		)
	}

	return entries
}

func defaultBindingEntry(path []string, value string) model.ConfigEntry {
	return model.ConfigEntry{
		Path:         path,
		Value:        value,
		DefaultOnly:  true,
		EqualityFunc: configformat.BindingValuesEqual,
	}
}

// edenButtonName maps SDL button names to Eden's Switch-style button names.
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
		return "Lstick"
	case model.ButtonRightStick:
		return "Rstick"
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
	section := "Shortcuts"

	type mapping struct {
		key     string
		binding model.HotkeyBinding
	}
	mappings := []mapping{
		{`Main%20Window\Continue\Pause%20Emulation\Controller_KeySeq`, hk.Pause},
		{`Main%20Window\Exit%20Eden\Controller_KeySeq`, hk.Quit},
		{`Main%20Window\Capture%20Screenshot\Controller_KeySeq`, hk.Screenshot},
		{`Main%20Window\Fullscreen\Controller_KeySeq`, hk.ToggleFullscreen},
		{`Main%20Window\Toggle%20Framerate%20Limit\Controller_KeySeq`, hk.FastForward},
	}

	var entries []model.ConfigEntry
	for _, m := range mappings {
		if len(m.binding.Buttons) > 0 {
			entries = append(entries, model.ConfigEntry{
				Path:  []string{section, m.key},
				Value: edenHotkeyRef(m.binding),
			})
		}
	}
	return entries
}
