package azahar

import (
	"fmt"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDAzahar,
		Name:            "Azahar",
		Systems:         []model.SystemID{model.SystemIDN3DS},
		Package:         model.AppImageRef("azahar"),
		ProvisionGroups: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "azahar",
			GenericName: "Nintendo 3DS Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " -f"
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
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "azahar-emu/qt-config.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store

	entries := []model.ConfigEntry{
		{Path: []string{"Data%20Storage", "use_custom_storage"}, Value: "true"},
		{Path: []string{"Data%20Storage", "use_custom_storage\\default"}, Value: "false"},
		{Path: []string{"Data%20Storage", "sdmc_directory"}, Value: store.SystemSavesDir(model.SystemIDN3DS) + "/"},
		{Path: []string{"Data%20Storage", "sdmc_directory\\default"}, Value: "false"},
		{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: "INSTALLED"},
		{Path: []string{"UI", "Paths\\gamedirs\\2\\path"}, Value: "SYSTEM"},
		{Path: []string{"UI", "Paths\\gamedirs\\3\\path"}, Value: store.SystemRomsDir(model.SystemIDN3DS)},
		{Path: []string{"UI", "Paths\\gamedirs\\size"}, Value: "3"},
		{Path: []string{"UI", "Paths\\screenshotPath"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDAzahar)},
		{Path: []string{"UI", "Paths\\screenshotPath\\default"}, Value: "false"},
	}

	var ownedRegions []model.OwnedRegion
	if cc := ctx.ControllerConfig; cc != nil {
		entries = append(entries, profileEntries(cc)...)
		ownedRegions = append(ownedRegions, model.OwnedRegion{
			Section:   "Controls",
			KeyPrefix: `profiles\1\`,
		})
	}

	patches := []model.ConfigPatch{{Target: configTarget, Entries: entries, OwnedRegions: ownedRegions}}

	dataDir, err := ctx.BaseDirResolver.UserDataDir()
	if err != nil {
		return model.GenerateResult{}, err
	}
	azaharDir := filepath.Join(dataDir, "azahar-emu")

	symlinks := []model.SymlinkSpec{
		{Source: filepath.Join(azaharDir, "states"), Target: store.EmulatorStatesDir(model.EmulatorIDAzahar)},
	}

	return model.GenerateResult{
		Patches:  patches,
		Symlinks: symlinks,
	}, nil
}

func azaharButtonRef(guid string, port, button int) string {
	return fmt.Sprintf("button:%d,engine:sdl,guid:%s,port:%d", button, guid, port)
}

func azaharAxisRef(guid string, port, axis int, direction string) string {
	return fmt.Sprintf("axis:%d,direction:%s,engine:sdl,guid:%s,port:%d,threshold:0.5", axis, direction, guid, port)
}

func azaharHatRef(guid string, port, hat int, direction string) string {
	return fmt.Sprintf("direction:%s,engine:sdl,guid:%s,hat:%d,port:%d", direction, guid, hat, port)
}

func azaharStickRef(guid string, port, axisX, axisY int) string {
	return fmt.Sprintf("axis_x:%d,axis_y:%d,deadzone:0.100000,engine:sdl,guid:%s,port:%d", axisX, axisY, guid, port)
}

func profileEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	south, east, west, north := cc.FaceButtons()
	section := "Controls"
	guid := model.SteamDeckGUID

	// 3DS maps: A=east, B=south, X=north, Y=west (Nintendo layout).
	faceMap := map[string]model.SDLButton{
		"a": east,
		"b": south,
		"x": north,
		"y": west,
	}

	prefix := "profiles\\1\\"
	return []model.ConfigEntry{
		{Path: []string{section, "profile"}, Value: "1", Unmanaged: true},
		{Path: []string{section, "profiles\\size"}, Value: "1", Unmanaged: true},
		{Path: []string{section, prefix + "name"}, Value: "kyaraben-steamdeck"},
		{Path: []string{section, prefix + "button_a"}, Value: fmt.Sprintf(`"%s"`, azaharButtonRef(guid, 0, model.SDLButtonIndex[faceMap["a"]]))},
		{Path: []string{section, prefix + "button_b"}, Value: fmt.Sprintf(`"%s"`, azaharButtonRef(guid, 0, model.SDLButtonIndex[faceMap["b"]]))},
		{Path: []string{section, prefix + "button_x"}, Value: fmt.Sprintf(`"%s"`, azaharButtonRef(guid, 0, model.SDLButtonIndex[faceMap["x"]]))},
		{Path: []string{section, prefix + "button_y"}, Value: fmt.Sprintf(`"%s"`, azaharButtonRef(guid, 0, model.SDLButtonIndex[faceMap["y"]]))},
		{Path: []string{section, prefix + "button_l"}, Value: fmt.Sprintf(`"%s"`, azaharButtonRef(guid, 0, model.SDLButtonIndex[model.ButtonLeftShoulder]))},
		{Path: []string{section, prefix + "button_r"}, Value: fmt.Sprintf(`"%s"`, azaharButtonRef(guid, 0, model.SDLButtonIndex[model.ButtonRightShoulder]))},
		{Path: []string{section, prefix + "button_zl"}, Value: fmt.Sprintf(`"%s"`, azaharAxisRef(guid, 0, model.AxisLeftTrigger, "+"))},
		{Path: []string{section, prefix + "button_zr"}, Value: fmt.Sprintf(`"%s"`, azaharAxisRef(guid, 0, model.AxisRightTrigger, "+"))},
		{Path: []string{section, prefix + "button_start"}, Value: fmt.Sprintf(`"%s"`, azaharButtonRef(guid, 0, model.SDLButtonIndex[model.ButtonStart]))},
		{Path: []string{section, prefix + "button_select"}, Value: fmt.Sprintf(`"%s"`, azaharButtonRef(guid, 0, model.SDLButtonIndex[model.ButtonBack]))},
		{Path: []string{section, prefix + "button_up"}, Value: fmt.Sprintf(`"%s"`, azaharHatRef(guid, 0, 0, "up"))},
		{Path: []string{section, prefix + "button_down"}, Value: fmt.Sprintf(`"%s"`, azaharHatRef(guid, 0, 0, "down"))},
		{Path: []string{section, prefix + "button_left"}, Value: fmt.Sprintf(`"%s"`, azaharHatRef(guid, 0, 0, "left"))},
		{Path: []string{section, prefix + "button_right"}, Value: fmt.Sprintf(`"%s"`, azaharHatRef(guid, 0, 0, "right"))},
		{Path: []string{section, prefix + "circle_pad"}, Value: fmt.Sprintf(`"%s"`, azaharStickRef(guid, 0, model.AxisLeftX, model.AxisLeftY))},
		{Path: []string{section, prefix + "c_stick"}, Value: fmt.Sprintf(`"%s"`, azaharStickRef(guid, 0, model.AxisRightX, model.AxisRightY))},
	}
}
