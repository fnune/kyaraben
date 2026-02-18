// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package duckstation

import (
	"fmt"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDDuckStation,
		Name:    "DuckStation",
		Systems: []model.SystemID{model.SystemIDPSX},
		Package: model.AppImageRef("duckstation"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "BIOS required (HLE fallback has limited compatibility)",
			Provisions:  psxBIOSProvisions,
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "duckstation",
			GenericName: "PlayStation Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " -fullscreen"
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
	RelPath: "duckstation/settings.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var ProfileTarget = model.ConfigTarget{
	RelPath: "duckstation/inputprofiles/kyaraben-steamdeck.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

const profileName = "kyaraben-steamdeck"

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store
	entries := []model.ConfigEntry{
		{Path: []string{"Main", "SettingsVersion"}, Value: "3"},
		{Path: []string{"AutoUpdater", "CheckAtStartup"}, Value: "false"},
		{Path: []string{"BIOS", "SearchDirectory"}, Value: store.SystemBiosDir(model.SystemIDPSX)},
		{Path: []string{"MemoryCards", "Directory"}, Value: store.SystemSavesDir(model.SystemIDPSX)},
		{Path: []string{"Folders", "SaveStates"}, Value: store.EmulatorStatesDir(model.EmulatorIDDuckStation)},
		{Path: []string{"Folders", "Screenshots"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDDuckStation)},
		{Path: []string{"GameList", "RecursivePaths"}, Value: store.SystemRomsDir(model.SystemIDPSX)},
	}

	patches := []model.ConfigPatch{{Target: configTarget, Entries: entries}}
	if cc := ctx.ControllerConfig; cc != nil {
		// Point to the profile preset (DefaultOnly so users can switch away).
		entries = append(entries, model.ConfigEntry{
			Path:        []string{"ControllerPorts", "InputProfileName"},
			Value:       profileName,
			DefaultOnly: true,
		})

		// Write bindings directly to main config (fully managed).
		entries = append(entries, padEntries(cc)...)
		entries = append(entries, hotkeyEntries(cc)...)
		patches[0].Entries = entries

		// Also write a profile file as a reusable preset (fully managed).
		profileEntries := padEntries(cc)
		profileEntries = append(profileEntries, hotkeyEntries(cc)...)
		patches = append(patches, model.ConfigPatch{
			Target:         ProfileTarget,
			Entries:        profileEntries,
			ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
		})
	}

	return model.GenerateResult{
		Patches: patches,
	}, nil
}

func sdlRef(playerIdx int, button model.SDLButton) string {
	return fmt.Sprintf("SDL-%d/%s", playerIdx, string(button))
}

func sdlAxisRef(playerIdx int, axis string, positive bool) string {
	sign := "-"
	if positive {
		sign = "+"
	}
	return fmt.Sprintf("SDL-%d/%s%s", playerIdx, sign, axis)
}

func padEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	var entries []model.ConfigEntry
	south, east, west, north := cc.FaceButtons()

	for i := 0; i < 4; i++ {
		section := fmt.Sprintf("Pad%d", i+1)
		entries = append(entries,
			model.ConfigEntry{Path: []string{section, "Type"}, Value: "AnalogController"},
			model.ConfigEntry{Path: []string{section, "Cross"}, Value: sdlRef(i, south)},
			model.ConfigEntry{Path: []string{section, "Circle"}, Value: sdlRef(i, east)},
			model.ConfigEntry{Path: []string{section, "Square"}, Value: sdlRef(i, west)},
			model.ConfigEntry{Path: []string{section, "Triangle"}, Value: sdlRef(i, north)},
			model.ConfigEntry{Path: []string{section, "L1"}, Value: sdlRef(i, model.ButtonLeftShoulder)},
			model.ConfigEntry{Path: []string{section, "R1"}, Value: sdlRef(i, model.ButtonRightShoulder)},
			model.ConfigEntry{Path: []string{section, "L2"}, Value: sdlAxisRef(i, "LeftTrigger", true)},
			model.ConfigEntry{Path: []string{section, "R2"}, Value: sdlAxisRef(i, "RightTrigger", true)},
			model.ConfigEntry{Path: []string{section, "L3"}, Value: sdlRef(i, model.ButtonLeftStick)},
			model.ConfigEntry{Path: []string{section, "R3"}, Value: sdlRef(i, model.ButtonRightStick)},
			model.ConfigEntry{Path: []string{section, "Up"}, Value: sdlRef(i, model.ButtonDPadUp)},
			model.ConfigEntry{Path: []string{section, "Down"}, Value: sdlRef(i, model.ButtonDPadDown)},
			model.ConfigEntry{Path: []string{section, "Left"}, Value: sdlRef(i, model.ButtonDPadLeft)},
			model.ConfigEntry{Path: []string{section, "Right"}, Value: sdlRef(i, model.ButtonDPadRight)},
			model.ConfigEntry{Path: []string{section, "LLeft"}, Value: sdlAxisRef(i, "LeftX", false)},
			model.ConfigEntry{Path: []string{section, "LRight"}, Value: sdlAxisRef(i, "LeftX", true)},
			model.ConfigEntry{Path: []string{section, "LUp"}, Value: sdlAxisRef(i, "LeftY", false)},
			model.ConfigEntry{Path: []string{section, "LDown"}, Value: sdlAxisRef(i, "LeftY", true)},
			model.ConfigEntry{Path: []string{section, "RLeft"}, Value: sdlAxisRef(i, "RightX", false)},
			model.ConfigEntry{Path: []string{section, "RRight"}, Value: sdlAxisRef(i, "RightX", true)},
			model.ConfigEntry{Path: []string{section, "RUp"}, Value: sdlAxisRef(i, "RightY", false)},
			model.ConfigEntry{Path: []string{section, "RDown"}, Value: sdlAxisRef(i, "RightY", true)},
			model.ConfigEntry{Path: []string{section, "Start"}, Value: sdlRef(i, model.ButtonStart)},
			model.ConfigEntry{Path: []string{section, "Select"}, Value: sdlRef(i, model.ButtonBack)},
			model.ConfigEntry{Path: []string{section, "SmallMotor"}, Value: fmt.Sprintf("SDL-%d/SmallMotor", i)},
			model.ConfigEntry{Path: []string{section, "LargeMotor"}, Value: fmt.Sprintf("SDL-%d/LargeMotor", i)},
		)
	}
	return entries
}

func hotkeyRef(binding model.HotkeyBinding) string {
	parts := make([]string, len(binding.Buttons))
	for i, b := range binding.Buttons {
		parts[i] = sdlRef(0, b)
	}
	return strings.Join(parts, " & ")
}

func hotkeyEntries(cc *model.ControllerConfig) []model.ConfigEntry {
	hk := cc.Hotkeys
	section := "Hotkeys"
	type mapping struct {
		key     string
		binding model.HotkeyBinding
	}
	mappings := []mapping{
		{"SaveSelectedSaveState", hk.SaveState},
		{"LoadSelectedSaveState", hk.LoadState},
		{"SelectNextSaveStateSlot", hk.NextSlot},
		{"SelectPreviousSaveStateSlot", hk.PrevSlot},
		{"ToggleFastForward", hk.FastForward},
		{"Rewind", hk.Rewind},
		{"TogglePause", hk.Pause},
		{"Screenshot", hk.Screenshot},
		{"OpenPauseMenu", hk.OpenMenu},
	}

	var entries []model.ConfigEntry
	for _, m := range mappings {
		if len(m.binding.Buttons) > 0 {
			entries = append(entries, model.ConfigEntry{
				Path:  []string{section, m.key},
				Value: hotkeyRef(m.binding),
			})
		}
	}
	return entries
}

var psxBIOSProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionBIOS, "scph1000.bin", "Japan v1.0", []string{"239665b1a3dade1b5a52c06338011044"}),
	model.HashedProvision(model.ProvisionBIOS, "scph1001.bin", "USA v2.0", []string{"924e392ed05558ffdb115408c263dccf"}),
	model.HashedProvision(model.ProvisionBIOS, "scph1002.bin", "Europe v2.0", []string{"54847e693405ffeb0359c6287434cbef"}),
	model.HashedProvision(model.ProvisionBIOS, "scph100.bin", "Japan early", []string{"8abc1b549a4a80954addc48ef02c4521"}),
	model.HashedProvision(model.ProvisionBIOS, "scph101.bin", "USA v4.4", []string{"6e3735ff4c7dc899ee98981385f6f3d0"}),
	model.HashedProvision(model.ProvisionBIOS, "scph102A.bin", "Europe v4.4", []string{"b10f5e0e3d9eb60e5159690680b1e774"}),
	model.HashedProvision(model.ProvisionBIOS, "scph102B.bin", "Europe v4.4", []string{"de93caec13d1a141a40a79f5c86168d6"}),
	model.HashedProvision(model.ProvisionBIOS, "scph102C.bin", "Europe v4.4", []string{"de93caec13d1a141a40a79f5c86168d6"}),
	model.HashedProvision(model.ProvisionBIOS, "scph3000.bin", "Japan v3.0", []string{"849515939161e62f6b866f6853006780"}),
	model.HashedProvision(model.ProvisionBIOS, "scph3500.bin", "Japan v3.5", []string{"cba733ceeff5aef5c32254f1d617fa62"}),
	model.HashedProvision(model.ProvisionBIOS, "scph5000.bin", "Japan v5.0", []string{"eb201d2d98251a598af467d4347bb62f"}),
	model.HashedProvision(model.ProvisionBIOS, "scph5500.bin", "Japan v3.0", []string{"8dd7d5296a650fac7319bce665a6a53c"}),
	model.HashedProvision(model.ProvisionBIOS, "scph5501.bin", "USA v3.0", []string{"490f666e1afb15b7362b406ed1cea246"}),
	model.HashedProvision(model.ProvisionBIOS, "scph5502.bin", "Europe v3.0", []string{"32736f17079d0b2b7024407c39bd3050"}),
	model.HashedProvision(model.ProvisionBIOS, "scph7001.bin", "USA v4.1", []string{"1e68c231d0896b7eadcad1d7d8e76129"}),
	model.HashedProvision(model.ProvisionBIOS, "scph7002.bin", "Europe v4.1", []string{"b9d9a0286c33dc6b7237bb13cd46fdee"}),
	model.HashedProvision(model.ProvisionBIOS, "scph7003.bin", "USA v4.1", []string{"490f666e1afb15b7362b406ed1cea246"}),
	model.HashedProvision(model.ProvisionBIOS, "scph7502.bin", "Europe v4.1", []string{"b9d9a0286c33dc6b7237bb13cd46fdee"}),
	model.HashedProvision(model.ProvisionBIOS, "scph9002(7502).bin", "Europe v4.1", []string{"b9d9a0286c33dc6b7237bb13cd46fdee"}),
	model.HashedProvision(model.ProvisionBIOS, "PSXONPSP660.bin", "PSP firmware", []string{"c53ca5908936d412331790f4426c6c33"}),
	model.HashedProvision(model.ProvisionBIOS, "psxonpsp660.bin", "PSP firmware", []string{"c53ca5908936d412331790f4426c6c33"}),
	model.HashedProvision(model.ProvisionBIOS, "ps1_rom.bin", "PS3 firmware", []string{"81bbe60ba7a3d1cea1d48c14cbcc647b"}),
}
