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
				cmd := opts.BinaryPath + " -batch"
				if opts.Fullscreen {
					cmd += " -fullscreen"
				}
				cmd += " -- %ROM%"
				return cmd
			},
		},
		PathUsage:         model.StandardPathUsage(),
		SupportedSettings: []string{model.SettingShaders},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "duckstation/settings.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserData,
}

var ProfileTarget = model.ConfigTarget{
	RelPath: "duckstation/inputprofiles/kyaraben-steamdeck.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserData,
}

const profileName = "kyaraben-steamdeck"

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store
	entries := []model.ConfigEntry{
		model.Entry(model.None, model.Path("Main", "SettingsVersion"), "3"),
		model.Entry(model.None, model.Path("Main", "ConfirmPowerOff"), "false"),
		model.Entry(model.None, model.Path("Main", "SaveStateOnExit"), "false"),
		model.Entry(model.None, model.Path("AutoUpdater", "CheckAtStartup"), "false"),
		model.Entry(model.Store, model.Path("BIOS", "SearchDirectory"), store.SystemBiosDir(model.SystemIDPSX)),
		model.Entry(model.Store, model.Path("MemoryCards", "Directory"), store.SystemSavesDir(model.SystemIDPSX)),
		model.Entry(model.Store, model.Path("Folders", "SaveStates"), store.EmulatorStatesDir(model.EmulatorIDDuckStation)),
		model.Entry(model.Store, model.Path("Folders", "Screenshots"), store.EmulatorScreenshotsDir(model.EmulatorIDDuckStation)),
		model.Entry(model.Store, model.Path("GameList", "RecursivePaths"), store.SystemRomsDir(model.SystemIDPSX)),
	}

	switch ctx.Shaders {
	case model.ShadersOn:
		entries = append(entries,
			model.Entry(model.None, model.Path("PostProcessing", "Enabled"), "true"),
			model.Entry(model.None, model.Path("PostProcessing", "StageCount"), "1"),
			model.Entry(model.None, model.Path("PostProcessing/Stage1", "ShaderName"), "crt-lottes"),
		)
	case model.ShadersOff:
		entries = append(entries,
			model.Entry(model.None, model.Path("PostProcessing", "Enabled"), "false"),
		)
	}

	patches := []model.ConfigPatch{{Target: configTarget, Entries: entries}}
	if cc := ctx.ControllerConfig; cc != nil {
		// Point to the profile preset (DefaultOnly so users can switch away).
		entries = append(entries, model.Default(model.None, model.Path("ControllerPorts", "InputProfileName"), profileName))

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
	fb := cc.FaceButtons(model.SystemIDPSX)
	south, east, west, north := fb.South, fb.East, fb.West, fb.North

	for i := 0; i < 4; i++ {
		section := fmt.Sprintf("Pad%d", i+1)
		entries = append(entries,
			model.Entry(model.None, model.Path(section, "Type"), "AnalogController"),
			model.Entry(model.None, model.Path(section, "Cross"), sdlRef(i, south)),
			model.Entry(model.None, model.Path(section, "Circle"), sdlRef(i, east)),
			model.Entry(model.None, model.Path(section, "Square"), sdlRef(i, west)),
			model.Entry(model.None, model.Path(section, "Triangle"), sdlRef(i, north)),
			model.Entry(model.None, model.Path(section, "L1"), sdlRef(i, model.ButtonLeftShoulder)),
			model.Entry(model.None, model.Path(section, "R1"), sdlRef(i, model.ButtonRightShoulder)),
			model.Entry(model.None, model.Path(section, "L2"), sdlAxisRef(i, "LeftTrigger", true)),
			model.Entry(model.None, model.Path(section, "R2"), sdlAxisRef(i, "RightTrigger", true)),
			model.Entry(model.None, model.Path(section, "L3"), sdlRef(i, model.ButtonLeftStick)),
			model.Entry(model.None, model.Path(section, "R3"), sdlRef(i, model.ButtonRightStick)),
			model.Entry(model.None, model.Path(section, "Up"), sdlRef(i, model.ButtonDPadUp)),
			model.Entry(model.None, model.Path(section, "Down"), sdlRef(i, model.ButtonDPadDown)),
			model.Entry(model.None, model.Path(section, "Left"), sdlRef(i, model.ButtonDPadLeft)),
			model.Entry(model.None, model.Path(section, "Right"), sdlRef(i, model.ButtonDPadRight)),
			model.Entry(model.None, model.Path(section, "LLeft"), sdlAxisRef(i, "LeftX", false)),
			model.Entry(model.None, model.Path(section, "LRight"), sdlAxisRef(i, "LeftX", true)),
			model.Entry(model.None, model.Path(section, "LUp"), sdlAxisRef(i, "LeftY", false)),
			model.Entry(model.None, model.Path(section, "LDown"), sdlAxisRef(i, "LeftY", true)),
			model.Entry(model.None, model.Path(section, "RLeft"), sdlAxisRef(i, "RightX", false)),
			model.Entry(model.None, model.Path(section, "RRight"), sdlAxisRef(i, "RightX", true)),
			model.Entry(model.None, model.Path(section, "RUp"), sdlAxisRef(i, "RightY", false)),
			model.Entry(model.None, model.Path(section, "RDown"), sdlAxisRef(i, "RightY", true)),
			model.Entry(model.None, model.Path(section, "Start"), sdlRef(i, model.ButtonStart)),
			model.Entry(model.None, model.Path(section, "Select"), sdlRef(i, model.ButtonBack)),
			model.Entry(model.None, model.Path(section, "SmallMotor"), fmt.Sprintf("SDL-%d/SmallMotor", i)),
			model.Entry(model.None, model.Path(section, "LargeMotor"), fmt.Sprintf("SDL-%d/LargeMotor", i)),
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
		{"PowerOff", hk.Quit},
	}

	var entries []model.ConfigEntry
	for _, m := range mappings {
		if len(m.binding.Buttons) > 0 {
			entries = append(entries, model.Entry(model.None, model.Path(section, m.key), hotkeyRef(m.binding)))
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
