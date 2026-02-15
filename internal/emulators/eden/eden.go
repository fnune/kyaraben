package eden

import (
	"fmt"
	"path/filepath"
	"strings"

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

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store

	entries := []model.ConfigEntry{
		{Path: []string{"UI", "Screenshots\\screenshot_path"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDEden)},
		{Path: []string{"UI", "Paths\\gamedirs\\size"}, Value: "1"},
		{Path: []string{"UI", "Paths\\gamedirs\\1\\deep_scan"}, Value: "false"},
		{Path: []string{"UI", "Paths\\gamedirs\\1\\expanded"}, Value: "true"},
		{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: store.SystemRomsDir(model.SystemIDSwitch)},
	}

	if cc := ctx.ControllerConfig; cc != nil {
		entries = append(entries, playerEntries(cc)...)
	}

	patches := []model.ConfigPatch{{Target: configTarget, Entries: entries}}

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
// rewrites its config on close.
func edenButtonRef(guid string, port, button int) string {
	return fmt.Sprintf("button:%d,guid:%s,port:%d,engine:sdl", button, guid, port)
}

func edenAxisRef(guid string, port, axis int) string {
	return fmt.Sprintf("threshold:0.500000,axis:%d,guid:%s,port:%d,engine:sdl", axis, guid, port)
}

func edenHatRef(guid string, port, hat int, direction string) string {
	return fmt.Sprintf("hat:%d,direction:%s,guid:%s,port:%d,engine:sdl", hat, direction, guid, port)
}

func edenStickRef(guid string, port, axisX, axisY int) string {
	return fmt.Sprintf("deadzone:0.100000,axis_y:%d,axis_x:%d,guid:%s,port:%d,engine:sdl", axisY, axisX, guid, port)
}

func playerEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	var entries []model.ConfigEntry
	south, east, west, north := cc.FaceButtons()

	guid := model.SteamDeckGUID

	// Switch maps: A=east, B=south, X=north, Y=west in Nintendo layout.
	// Eden is a Switch emulator, so Switch A/B/X/Y are the console buttons.
	// With standard layout, physical south -> Switch B, physical east -> Switch A, etc.
	faceMap := map[string]model.SDLButton{
		"a": east,
		"b": south,
		"x": north,
		"y": west,
	}

	for i := 0; i < 2; i++ {
		prefix := fmt.Sprintf("player_%d_", i)
		// Player 0 is always connected. Player 1+ use DefaultOnly so Eden can
		// manage connection state based on actual controllers.
		connectedEntry := model.ConfigEntry{Path: []string{"Controls", prefix + "connected"}, Value: "true"}
		if i > 0 {
			connectedEntry.DefaultOnly = true
		}
		entries = append(entries,
			connectedEntry,
			model.ConfigEntry{Path: []string{"Controls", prefix + "type"}, Value: "0"},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_a"}, Value: fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[faceMap["a"]]))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_b"}, Value: fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[faceMap["b"]]))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_x"}, Value: fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[faceMap["x"]]))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_y"}, Value: fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[faceMap["y"]]))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_lstick"}, Value: fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonLeftStick]))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_rstick"}, Value: fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonRightStick]))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_l"}, Value: fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonLeftShoulder]))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_r"}, Value: fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonRightShoulder]))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_zl"}, Value: fmt.Sprintf(`"%s"`, edenAxisRef(guid, i, model.AxisLeftTrigger))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_zr"}, Value: fmt.Sprintf(`"%s"`, edenAxisRef(guid, i, model.AxisRightTrigger))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_plus"}, Value: fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonStart]))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_minus"}, Value: fmt.Sprintf(`"%s"`, edenButtonRef(guid, i, model.SDLButtonIndex[model.ButtonBack]))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_dleft"}, Value: fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "left"))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_dright"}, Value: fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "right"))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_dup"}, Value: fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "up"))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "button_ddown"}, Value: fmt.Sprintf(`"%s"`, edenHatRef(guid, i, 0, "down"))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "lstick"}, Value: fmt.Sprintf(`"%s"`, edenStickRef(guid, i, model.AxisLeftX, model.AxisLeftY))},
			model.ConfigEntry{Path: []string{"Controls", prefix + "rstick"}, Value: fmt.Sprintf(`"%s"`, edenStickRef(guid, i, model.AxisRightX, model.AxisRightY))},
		)
	}

	return entries
}
