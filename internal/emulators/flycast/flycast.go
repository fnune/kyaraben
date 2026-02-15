// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package flycast

import (
	"fmt"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDFlycast,
		Name:    "Flycast",
		Systems: []model.SystemID{model.SystemIDDreamcast},
		Package: model.AppImageRef("flycast"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 0,
			Message:     "BIOS files (optional, enables boot animation)",
			Provisions:  dreamcastBIOSProvisions,
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "flycast",
			GenericName: "Sega Dreamcast Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " -config window:fullscreen=yes"
				}
				cmd += " %ROM%"
				return cmd
			},
		},
		PathUsage: model.StandardPathUsage(),
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "flycast/emu.cfg",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

var mappingTarget = model.ConfigTarget{
	RelPath: "flycast/mappings/SDL_Gamepad.cfg",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store
	biosDir := store.SystemBiosDir(model.SystemIDDreamcast)
	savesDir := store.SystemSavesDir(model.SystemIDDreamcast)

	patches := []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"config", "Flycast.DataPath"}, Value: biosDir},
			{Path: []string{"config", "Dreamcast.BiosPath"}, Value: biosDir},
			{Path: []string{"config", "Dreamcast.ContentPath"}, Value: store.SystemRomsDir(model.SystemIDDreamcast)},
			{Path: []string{"config", "Dreamcast.SavePath"}, Value: savesDir},
			{Path: []string{"config", "Dreamcast.VMUPath"}, Value: savesDir},
			{Path: []string{"config", "Dreamcast.SavestatePath"}, Value: store.EmulatorStatesDir(model.EmulatorIDFlycast)},
		},
	}}

	if cc := ctx.ControllerConfig; cc != nil {
		patches = append(patches, model.ConfigPatch{
			Target:  mappingTarget,
			Entries: mappingEntries(cc),
		})
	}

	return model.GenerateResult{Patches: patches}, nil
}

func mappingEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	south, east, west, north := cc.FaceButtons()

	// Flycast uses axis:action format for analog and button:action for digital.
	// Dreamcast: A=south, B=east, X=west, Y=north.
	return []model.ConfigEntry{
		{Path: []string{"digital", "bind0"}, Value: fmt.Sprintf("%d:btn_a", model.SDLButtonIndex[south])},
		{Path: []string{"digital", "bind1"}, Value: fmt.Sprintf("%d:btn_b", model.SDLButtonIndex[east])},
		{Path: []string{"digital", "bind2"}, Value: fmt.Sprintf("%d:btn_x", model.SDLButtonIndex[west])},
		{Path: []string{"digital", "bind3"}, Value: fmt.Sprintf("%d:btn_y", model.SDLButtonIndex[north])},
		{Path: []string{"digital", "bind4"}, Value: fmt.Sprintf("%d:btn_menu", model.SDLButtonIndex[model.ButtonBack])},
		{Path: []string{"digital", "bind5"}, Value: fmt.Sprintf("%d:btn_start", model.SDLButtonIndex[model.ButtonStart])},
		{Path: []string{"digital", "bind6"}, Value: fmt.Sprintf("%d:btn_dpad1_up", model.SDLButtonIndex[model.ButtonDPadUp])},
		{Path: []string{"digital", "bind7"}, Value: fmt.Sprintf("%d:btn_dpad1_down", model.SDLButtonIndex[model.ButtonDPadDown])},
		{Path: []string{"digital", "bind8"}, Value: fmt.Sprintf("%d:btn_dpad1_left", model.SDLButtonIndex[model.ButtonDPadLeft])},
		{Path: []string{"digital", "bind9"}, Value: fmt.Sprintf("%d:btn_dpad1_right", model.SDLButtonIndex[model.ButtonDPadRight])},
		{Path: []string{"analog", "bind0"}, Value: "0-:btn_analog_left"},
		{Path: []string{"analog", "bind1"}, Value: "0+:btn_analog_right"},
		{Path: []string{"analog", "bind2"}, Value: "1-:btn_analog_up"},
		{Path: []string{"analog", "bind3"}, Value: "1+:btn_analog_down"},
		{Path: []string{"analog", "bind4"}, Value: "2+:btn_trigger_left"},
		{Path: []string{"analog", "bind5"}, Value: "5+:btn_trigger_right"},
		{Path: []string{"emulator", "dead_zone"}, Value: "10"},
	}
}

var dreamcastBIOSProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionBIOS, "dc_boot.bin", "Boot ROM", []string{"e10c53c2f8b90bab96ead2d368858623", "d407fcf70b56acb84b8c77c93b0e5327", "93a9766f14159b403178ac77417c6b68"}),
	model.HashedProvision(model.ProvisionBIOS, "dc_flash.bin", "Flash ROM", []string{"0a93f7940c455905bea6e392dfde92a4"}),
	model.HashedProvision(model.ProvisionBIOS, "flash.bin", "Flash ROM (alternate name)", []string{"0a93f7940c455905bea6e392dfde92a4"}),
}
