// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package duckstation

import (
	"fmt"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/shaders"
)

type Definition struct{}

var hotkeyMappings = model.HotkeyMappings{
	SaveState:   &model.HotkeyKey{Key: "SaveSelectedSaveState"},
	LoadState:   &model.HotkeyKey{Key: "LoadSelectedSaveState"},
	NextSlot:    &model.HotkeyKey{Key: "SelectNextSaveStateSlot"},
	PrevSlot:    &model.HotkeyKey{Key: "SelectPreviousSaveStateSlot"},
	FastForward: &model.HotkeyKey{Key: "ToggleFastForward"},
	Rewind:      &model.HotkeyKey{Key: "Rewind"},
	Pause:       &model.HotkeyKey{Key: "TogglePause"},
	Screenshot:  &model.HotkeyKey{Key: "Screenshot"},
	OpenMenu:    &model.HotkeyKey{Key: "OpenPauseMenu"},
	Quit:        &model.HotkeyKey{Key: "PowerOff"},
}

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
				for _, arg := range opts.LaunchArgs {
					cmd += " " + arg
				}
				cmd += " -- %ROM%"
				return cmd
			},
		},
		PathUsage:          model.StandardPathUsage(),
		SupportedSettings:  []string{model.SettingPreset, model.SettingResumeAutosave, model.SettingResumeAutoload},
		SupportedHotkeys:   hotkeyMappings.SupportedHotkeys(),
		ShadersRecommended: true,
		ResumeRecommended:  true,
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

// DuckStation uses XDG_CONFIG_HOME if set, otherwise ~/.local/share.
// Since XDG_CONFIG_HOME may differ between when Kyaraben applies and when the user plays,
// we write to both locations to ensure DuckStation finds the config.
var configTargets = []model.ConfigTarget{
	{RelPath: "duckstation/settings.ini", Format: model.ConfigFormatINI, BaseDir: model.ConfigBaseDirUserConfig},
	{RelPath: "duckstation/settings.ini", Format: model.ConfigFormatINI, BaseDir: model.ConfigBaseDirUserData},
}

var profileTargets = []model.ConfigTarget{
	{RelPath: "duckstation/inputprofiles/kyaraben-steamdeck.ini", Format: model.ConfigFormatINI, BaseDir: model.ConfigBaseDirUserConfig},
	{RelPath: "duckstation/inputprofiles/kyaraben-steamdeck.ini", Format: model.ConfigFormatINI, BaseDir: model.ConfigBaseDirUserData},
}

const profileName = "kyaraben-steamdeck"

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store
	entries := []model.ConfigEntry{
		model.Entry(model.None, model.Path("Main", "SettingsVersion"), "3"),
		model.Entry(model.None, model.Path("Main", "SetupWizardIncomplete"), "false"),
		model.Entry(model.None, model.Path("Main", "ConfirmPowerOff"), "false"),
		model.Default(model.None, model.Path("Main", "NoDesktopFile"), "true"),
		model.Entry(model.None, model.Path("AutoUpdater", "CheckAtStartup"), "false"),
		model.Entry(model.Store, model.Path("BIOS", "SearchDirectory"), store.SystemBiosDir(model.SystemIDPSX)),
		model.Entry(model.Store, model.Path("MemoryCards", "Directory"), store.SystemSavesDir(model.SystemIDPSX)),
		model.Entry(model.Store, model.Path("Folders", "SaveStates"), store.EmulatorStatesDir(model.EmulatorIDDuckStation)),
		model.Entry(model.Store, model.Path("Folders", "Screenshots"), store.EmulatorScreenshotsDir(model.EmulatorIDDuckStation)),
		model.Entry(model.Store, model.Path("GameList", "RecursivePaths"), store.SystemRomsDir(model.SystemIDPSX)),
	}

	switch ctx.Preset {
	case model.PresetClean:
		entries = append(entries,
			model.Entry(model.Preset, model.Path("PostProcessing", "Enabled"), "false"),
		)
	case model.PresetRetro:
		crt := shaders.CRTLottes
		entries = append(entries,
			model.Entry(model.Preset, model.Path("PostProcessing", "Enabled"), "true"),
			model.Entry(model.Preset, model.Path("PostProcessing", "StageCount"), "1"),
			model.Entry(model.Preset, model.Path("PostProcessing/Stage1", "ShaderName"), "crt-lottes"),
			model.Entry(model.Preset, model.Path("PostProcessing/Stage1", "shadowMask"), fmt.Sprintf("%g", crt.ShadowMask)),
			model.Entry(model.Preset, model.Path("PostProcessing/Stage1", "warpX"), fmt.Sprintf("%g", crt.WarpX)),
			model.Entry(model.Preset, model.Path("PostProcessing/Stage1", "warpY"), fmt.Sprintf("%g", crt.WarpY)),
			model.Entry(model.Preset, model.Path("PostProcessing/Stage1", "bloomAmount"), fmt.Sprintf("%g", crt.BloomAmount)),
			model.Entry(model.Preset, model.Path("PostProcessing/Stage1", "brightBoost"), fmt.Sprintf("%g", crt.BrightBoost)),
		)
	}

	switch ctx.Resume {
	case model.EmulatorResumeOn:
		entries = append(entries, model.Entry(model.Resume, model.Path("Main", "SaveStateOnExit"), "true"))
	case model.EmulatorResumeOff:
		entries = append(entries, model.Entry(model.Resume, model.Path("Main", "SaveStateOnExit"), "false"))
	}

	if cc := ctx.ControllerConfig; cc != nil {
		entries = append(entries, model.Default(model.None, model.Path("ControllerPorts", "InputProfileName"), profileName))
		entries = append(entries, padEntries(cc)...)
		entries = append(entries, hotkeyEntries(cc)...)
	}

	var patches []model.ConfigPatch
	for _, target := range configTargets {
		patches = append(patches, model.ConfigPatch{Target: target, Entries: entries})
	}

	if ctx.ControllerConfig != nil {
		profileEntries := padEntries(ctx.ControllerConfig)
		profileEntries = append(profileEntries, hotkeyEntries(ctx.ControllerConfig)...)
		for _, target := range profileTargets {
			patches = append(patches, model.ConfigPatch{
				Target:         target,
				Entries:        profileEntries,
				ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
			})
		}
	}

	var launchArgs []string
	if ctx.Resume == model.EmulatorResumeOn {
		launchArgs = append(launchArgs, "-resume")
	}

	return model.GenerateResult{
		Patches:    patches,
		LaunchArgs: launchArgs,
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
			entries = append(entries, model.Entry(model.Hotkeys, model.Path(section, m.key), hotkeyRef(m.binding)))
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
