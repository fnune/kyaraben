package cemu

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDCemu,
		Name:    "Cemu",
		Systems: []model.SystemID{model.SystemIDWiiU},
		Package: model.AppImageRef("cemu"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 0,
			Message:     "Only needed for encrypted game dumps",
			Provisions: []model.Provision{
				model.FileProvision(model.ProvisionKeys, "keys.txt", "Wii U keys").WithImportViaUI(),
			},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
		},
		Launcher: model.LauncherInfo{
			Binary:      "cemu",
			GenericName: "Wii U Emulator",
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
			UsesSavesDir:       true,
			UsesScreenshotsDir: true,
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "Cemu/settings.xml",
	Format:  model.ConfigFormatXML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var controllerProfileTarget = model.ConfigTarget{
	RelPath: "Cemu/controllerProfiles/controller0.xml",
	Format:  model.ConfigFormatRaw,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store

	patches := []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"content", "GamePaths", "Entry"}, Value: store.SystemRomsDir(model.SystemIDWiiU)},
			{Path: []string{"content", "check_update"}, Value: "false"},
			{Path: []string{"content", "Graphic", "api"}, Value: "1", DefaultOnly: true},
			{Path: []string{"content", "Graphic", "VSync"}, Value: "0", DefaultOnly: true},
			{Path: []string{"content", "Graphic", "AsyncCompile"}, Value: "true", DefaultOnly: true},
			{Path: []string{"content", "Graphic", "GX2DrawdoneSync"}, Value: "true", DefaultOnly: true},
			{Path: []string{"content", "Graphic", "Notification", "ShaderCompiling"}, Value: "false", DefaultOnly: true},
		},
	}}

	if cc := ctx.ControllerConfig; cc != nil {
		patches = append(patches, model.ConfigPatch{
			Target: controllerProfileTarget,
			Entries: []model.ConfigEntry{
				{Value: generateControllerXML(cc), DefaultOnly: true},
			},
		})
	}

	dataDir, err := ctx.BaseDirResolver.UserDataDir()
	if err != nil {
		return model.GenerateResult{}, err
	}
	cemuDir := filepath.Join(dataDir, "Cemu")

	symlinks := []model.SymlinkSpec{
		{Source: filepath.Join(cemuDir, "mlc01", "usr", "save", "00050000"), Target: store.SystemSavesDir(model.SystemIDWiiU)},
		{Source: filepath.Join(cemuDir, "screenshots"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDCemu)},
	}

	return model.GenerateResult{
		Patches:  patches,
		Symlinks: symlinks,
	}, nil
}

// Cemu VPAD button mapping IDs (from VPADController.h)
const (
	vpadA           = 1
	vpadB           = 2
	vpadX           = 3
	vpadY           = 4
	vpadL           = 5
	vpadR           = 6
	vpadZL          = 7
	vpadZR          = 8
	vpadPlus        = 9
	vpadMinus       = 10
	vpadUp          = 11
	vpadDown        = 12
	vpadLeft        = 13
	vpadRight       = 14
	vpadStickL      = 15
	vpadStickR      = 16
	vpadStickLUp    = 17
	vpadStickLDown  = 18
	vpadStickLLeft  = 19
	vpadStickLRight = 20
	vpadStickRUp    = 21
	vpadStickRDown  = 22
	vpadStickRLeft  = 23
	vpadStickRRight = 24
	vpadHome        = 27
)

// Cemu SDL button IDs (from Controller.h Buttons2 enum)
const (
	sdlAxisXP     = 38 // left stick X positive (right)
	sdlAxisYP     = 39 // left stick Y positive (down)
	sdlRotationXP = 40 // right stick X positive (right)
	sdlRotationYP = 41 // right stick Y positive (down)
	sdlTriggerXP  = 42 // left trigger
	sdlTriggerYP  = 43 // right trigger
	sdlAxisXN     = 44 // left stick X negative (left)
	sdlAxisYN     = 45 // left stick Y negative (up)
	sdlRotationXN = 46 // right stick X negative (left)
	sdlRotationYN = 47 // right stick Y negative (up)
)

// cemuSteamDeckGUID is the Steam Virtual Gamepad GUID as reported by SDL.
// Unlike Eden which normalizes CRC bytes, Cemu uses the exact SDL GUID.
// This matches EmuDeck's configuration.
const cemuSteamDeckGUID = "030079f6de280000ff11000001000000"

func generateControllerXML(cc *model.ControllerConfig) string {
	fb := cc.FaceButtons(model.SystemIDWiiU)
	south, east, west, north := fb.South, fb.East, fb.West, fb.North

	// Map face buttons to SDL button indices
	// SDL: A=0, B=1, X=2, Y=3
	faceButtonIndex := map[model.SDLButton]int{
		model.ButtonA: 0,
		model.ButtonB: 1,
		model.ButtonX: 2,
		model.ButtonY: 3,
	}

	// VPAD A=east, B=south, X=north, Y=west (like Nintendo layout)
	mappings := []struct {
		vpad   int
		button int
	}{
		{vpadA, faceButtonIndex[east]},
		{vpadB, faceButtonIndex[south]},
		{vpadX, faceButtonIndex[north]},
		{vpadY, faceButtonIndex[west]},
		{vpadL, model.SDLButtonIndex[model.ButtonLeftShoulder]},
		{vpadR, model.SDLButtonIndex[model.ButtonRightShoulder]},
		{vpadZL, sdlTriggerXP},
		{vpadZR, sdlTriggerYP},
		{vpadPlus, model.SDLButtonIndex[model.ButtonStart]},
		{vpadMinus, model.SDLButtonIndex[model.ButtonBack]},
		{vpadHome, model.SDLButtonIndex[model.ButtonGuide]},
		{vpadUp, 11},    // SDL DPad Up
		{vpadDown, 12},  // SDL DPad Down
		{vpadLeft, 13},  // SDL DPad Left
		{vpadRight, 14}, // SDL DPad Right
		{vpadStickL, model.SDLButtonIndex[model.ButtonLeftStick]},
		{vpadStickR, model.SDLButtonIndex[model.ButtonRightStick]},
		{vpadStickLUp, sdlAxisYN},
		{vpadStickLDown, sdlAxisYP},
		{vpadStickLLeft, sdlAxisXN},
		{vpadStickLRight, sdlAxisXP},
		{vpadStickRUp, sdlRotationYN},
		{vpadStickRDown, sdlRotationYP},
		{vpadStickRLeft, sdlRotationXN},
		{vpadStickRRight, sdlRotationXP},
	}

	var mappingXML strings.Builder
	for _, m := range mappings {
		mappingXML.WriteString(fmt.Sprintf("\t\t\t<entry>\n\t\t\t\t<mapping>%d</mapping>\n\t\t\t\t<button>%d</button>\n\t\t\t</entry>\n", m.vpad, m.button))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<emulated_controller>
	<type>Wii U GamePad</type>
	<controller>
		<api>SDLController</api>
		<uuid>0_%s</uuid>
		<display_name>Steam Virtual Gamepad</display_name>
		<rumble>1</rumble>
		<axis>
			<deadzone>0.25</deadzone>
			<range>1</range>
		</axis>
		<rotation>
			<deadzone>0.25</deadzone>
			<range>1</range>
		</rotation>
		<trigger>
			<deadzone>0.25</deadzone>
			<range>1</range>
		</trigger>
		<mappings>
%s		</mappings>
	</controller>
</emulated_controller>
`, cemuSteamDeckGUID, mappingXML.String())
}
