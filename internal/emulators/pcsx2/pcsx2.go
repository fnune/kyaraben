// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package pcsx2

import (
	"fmt"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDPCSX2,
		Name:    "PCSX2",
		Systems: []model.SystemID{model.SystemIDPS2},
		Package: model.AppImageRef("pcsx2"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "BIOS required (no HLE fallback available)",
			Provisions:  ps2BIOSProvisions,
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "pcsx2",
			GenericName: "PlayStation 2 Emulator",
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
	RelPath: "PCSX2/inis/PCSX2.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var profileTarget = model.ConfigTarget{
	RelPath: "PCSX2/inputprofiles/Kyaraben.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store
	entries := []model.ConfigEntry{
		{Path: []string{"UI", "SettingsVersion"}, Value: "1"},
		{Path: []string{"UI", "SetupWizardIncomplete"}, Value: "false"},
		{Path: []string{"UI", "ConfirmShutdown"}, Value: "false"},
		{Path: []string{"UI", "StartFullscreen"}, Value: "true"},
		{Path: []string{"Folders", "Bios"}, Value: store.SystemBiosDir(model.SystemIDPS2)},
		{Path: []string{"Folders", "MemoryCards"}, Value: store.SystemSavesDir(model.SystemIDPS2)},
		{Path: []string{"Folders", "Savestates"}, Value: store.EmulatorStatesDir(model.EmulatorIDPCSX2)},
		{Path: []string{"Folders", "Snapshots"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDPCSX2)},
		{Path: []string{"GameList", "RecursivePaths"}, Value: store.SystemRomsDir(model.SystemIDPS2)},
	}

	switch ctx.Shaders {
	case model.ShadersOn:
		entries = append(entries, model.ConfigEntry{Path: []string{"EmuCore/GS", "TVShader"}, Value: "5"})
	case model.ShadersOff:
		entries = append(entries, model.ConfigEntry{Path: []string{"EmuCore/GS", "TVShader"}, Value: "0"})
	}

	patches := []model.ConfigPatch{{Target: configTarget, Entries: entries}}

	if cc := ctx.ControllerConfig; cc != nil {
		// Write pad/hotkey to main config as DefaultOnly (applied on first run, user can change).
		for _, e := range padEntries(cc) {
			e.DefaultOnly = true
			entries = append(entries, e)
		}
		for _, e := range hotkeyEntries(cc) {
			e.DefaultOnly = true
			entries = append(entries, e)
		}
		patches[0].Entries = entries

		// Also create a profile file users can reapply (fully managed).
		profileEntries := []model.ConfigEntry{
			{Path: []string{"Pad", "UseProfileHotkeyBindings"}, Value: "true"},
		}
		profileEntries = append(profileEntries, padEntries(cc)...)
		profileEntries = append(profileEntries, hotkeyEntries(cc)...)
		patches = append(patches, model.ConfigPatch{
			Target:         profileTarget,
			Entries:        profileEntries,
			ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
		})
	}

	return model.GenerateResult{Patches: patches}, nil
}

var sdl3FaceButtons = map[model.SDLButton]string{
	model.ButtonA: "FaceSouth",
	model.ButtonB: "FaceEast",
	model.ButtonX: "FaceWest",
	model.ButtonY: "FaceNorth",
}

func sdlRef(playerIdx int, button model.SDLButton) string {
	name := string(button)
	if sdl3Name, ok := sdl3FaceButtons[button]; ok {
		name = sdl3Name
	}
	return fmt.Sprintf("SDL-%d/%s", playerIdx, name)
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
	fb := cc.FaceButtons(model.SystemIDPS2)
	south, east, west, north := fb.South, fb.East, fb.West, fb.North

	for i := 0; i < 4; i++ {
		section := fmt.Sprintf("Pad%d", i+1)
		entries = append(entries,
			model.ConfigEntry{Path: []string{section, "Type"}, Value: "DualShock2"},
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
		{"SaveStateToSlot", hk.SaveState},
		{"LoadStateFromSlot", hk.LoadState},
		{"NextSaveStateSlot", hk.NextSlot},
		{"PreviousSaveStateSlot", hk.PrevSlot},
		{"ToggleTurbo", hk.FastForward},
		{"ToggleSlowMotion", hk.Rewind},
		{"TogglePause", hk.Pause},
		{"Screenshot", hk.Screenshot},
		{"ShutdownVM", hk.Quit},
		{"ToggleFullscreen", hk.ToggleFullscreen},
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

var ps2BIOSProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionBIOS, "ps2-0100j-20000117.bin", "Japan v1.00", []string{"acf4730ceb38ac9d8c7d8e21f2614600"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0100jd-20000117.bin", "Japan v1.00 debug", []string{"32f2e4d5ff5ee11072a6bc45530f5765"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0101j-20000217.bin", "Japan v1.01", []string{"b1459d7446c69e3e97e6ace3ae23dd1c"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0101jd-20000217.bin", "Japan v1.01 debug", []string{"acf9968c8f596d2b15f42272082513d1"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0101xd-20000224.bin", "Export/TOOL debug", []string{"d3f1853a16c2ec18f3cd1ae655213308"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0110a-20000727.bin", "Asia v1.10", []string{"a20c97c02210f16678ca3010127caf36"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0110ad-20000727.bin", "Asia v1.10 debug", []string{"63e6fd9b3c72e0d7b920e80cf76645cd"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0120a-20000902.bin", "Asia v1.20", []string{"8db2fbbac7413bf3e7154c1e0715e565"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0120e-20000902.bin", "Europe v1.20", []string{"b7fa11e87d51752a98b38e3e691cbf17"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0120ed-20000902.bin", "Europe v1.20 debug", []string{"91c87cb2f2eb6ce529a2360f80ce2457"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0120ed-20000902-20030110.bin", "Europe v1.20 debug", []string{"3016b3dd42148a67e2c048595ca4d7ce"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0120j-20001027-185015.bin", "Japan v1.20", []string{"f63bc530bd7ad7c026fcd6f7bd0d9525"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0120j-20001027-191435.bin", "Japan v1.20", []string{"cee06bd68c333fc5768244eae77e4495"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0150a-20001228.bin", "Asia v1.50", []string{"8accc3c49ac45f5ae2c5db0adc854633"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0150ad-20001228-20030520.bin", "Asia v1.50 debug", []string{"0bf988e9c7aaa4c051805b0fa6eb3387"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0150e-20001228.bin", "Europe v1.50", []string{"838544f12de9b0abc90811279ee223c8"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0150ed-20001228-20030520.bin", "Europe v1.50 debug", []string{"6f9a6feb749f0533aaae2cc45090b0ed"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0150j-20010118.bin", "Japan v1.50", []string{"815ac991d8bc3b364696bead3457de7d"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0150jd-20010118.bin", "Japan v1.50 debug", []string{"bb6bbc850458fff08af30e969ffd0175"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0160a-20010427.bin", "Asia v1.60", []string{"b107b5710042abe887c0f6175f6e94bb"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0160e-20010704.bin", "Europe v1.60", []string{"491209dd815ceee9de02dbbc408c06d6"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0160e-20011004.bin", "Europe v1.60", []string{"8359638e857c8bc18c3c18ac17d9cc3c"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0160h-20010730.bin", "China/HK v1.60", []string{"352d2ff9b3f68be7e6fa7e6dd8389346"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0160h-20020426.bin", "China/HK v1.60", []string{"315a4003535dfda689752cb25f24785c"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0160j-20010427.bin", "Japan v1.60", []string{"ab55cceea548303c22c72570cfd4dd71"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0160j-20020426.bin", "Japan v1.60", []string{"72da56fccb8fcd77bba16d1b6f479914"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0170ad-20030325.bin", "Asia v1.70 debug", []string{"eb960de68f0c0f7f9fa083e9f79d0360"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0170e-20030227.bin", "Europe v1.70", []string{"666018ffec65c5c7e04796081295c6c7"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0170ed-20030227.bin", "Europe v1.70 debug", []string{"666018ffec65c5c7e04796081295c6c7"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0170j-20030206.bin", "Japan v1.70", []string{"312ad4816c232a9606e56f946bc0678a"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0180cd-20030224.bin", "Korea/China debug", []string{"240d4c5ddd4b54069bdc4a3cd2faf99d"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0180j-20031028.bin", "Japan v1.80", []string{"1c6cd089e6c83da618fbf2a081eb4888"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0190a-20031128.bin", "Asia v1.90", []string{"4587a1dd58d6805265bfc68f0fef2732"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0190ad-20031128.bin", "Asia v1.90 debug", []string{"30ab5f78d80431a862283642b71f8fb3"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0190e-20031031.bin", "Europe v1.90", []string{"d1c4ac413fdf57cd93585ff19cf50df0"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0190ed-20031031.bin", "Europe v1.90 debug", []string{"0e440a76cd950cbbb1d576242e0429b0"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0190h-20030916.bin", "China/HK v1.90", []string{"05359e5a7c88a69eb25bf5f4b85c3a0b"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0190j-20030707.bin", "Japan v1.90", []string{"0270d4ca21f18f0f220f131cf9136a14"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0190jd-20030707.bin", "Japan v1.90 debug", []string{"760e38eb17cded716bc756a8965643be"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0190r-20030916.bin", "Russia v1.90", []string{"67d613747c39b50f2ce3fbbed7e8ebe4"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0190u-20031017.bin", "USA v1.90", []string{"e973c9c5f78aaff915f9fe09a33bce8e"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0200a-20040419.bin", "Asia v2.00", []string{"6654c39431dfe83692dd26695d836de7"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0200a-20040614.bin", "USA v2.00", []string{"d333558cc14561c1fdc334c75d5f37b7"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0200ad-20040419.bin", "Asia v2.00 debug", []string{"8d83489e564f72a5b31bfa56bcf6e108"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0200e-20040614.bin", "Europe v2.00", []string{"dc752f160044f2ed5fc1f4964db2a095"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0200e-20041115.bin", "Europe v2.00", []string{"dc996660a859a8c23a952f633a409556"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0200h-20040419.bin", "China/HK v2.00", []string{"b82f12f8b186cbf8e2e3fdc337044684"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0200j-20040323.bin", "Japan v2.00", []string{"2253c3d364ff49c234a38f3d631fc31a"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0200j-20040614.bin", "Japan v2.00", []string{"0eee5d1c779aa50e94edd168b4ebf42e"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0200u-20040419.bin", "USA v2.00", []string{"99c5cc568676fa3d4c4d2c4a492d1ec5"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0201a-20040915.bin", "Asia v2.01", []string{"d5bb76a9e7c11f59a4263f92f3053c75"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0201h-20040915.bin", "China/HK v2.01", []string{"546e7d6f258d2c229812f9158cfc204d"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0201j-20040819.bin", "Japan v2.01", []string{"8d6e89165d77a6eca1528b0811aabc7a"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0201u-20040915.bin", "USA v2.01", []string{"6bf1fce90e751b6df66aed444057ff94"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0202a-20050121.bin", "Asia v2.02", []string{"62fab6930f454c651eecadf42db4927f"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0202e-20050121.bin", "Europe v2.02", []string{"b411fffe806a458518a874db389e7271"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0202h-20050121.bin", "China/HK v2.02", []string{"f4b982aa5a5c1adf7e9f7dc3cba7e3a4"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0202j-20041222.bin", "Japan v2.02", []string{"f854a87b6be3e88841e2de3c8d2d101e"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0202u-20050121.bin", "USA v2.02", []string{"7c6358d74fd121ca3df740f9db3195a4"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0203a-20050304.bin", "Asia v2.03", []string{"34d8e1f5ba1fa8b41c1ec2ed2b920eac"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0203e-20050304.bin", "Europe v2.03", []string{"810f42b2fcf7d95dfdc9e5f629d88cc4"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0205u-20051020.bin", "USA v2.05", []string{"11d888baf64c6c4baa2f673f0aa8ce43"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0230a-20080220.bin", "USA v2.30", []string{"21038400dc633070a78ad53090c53017"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0230e-20080220.bin", "Europe v2.30", []string{"dc69f0643a3030aaa4797501b483d6c4"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0230j-20080220.bin", "Japan v2.30", []string{"30d56e79d89fbddf10938fa67fe3f34e"}),
	model.HashedProvision(model.ProvisionBIOS, "ps2-0250e-20100415.bin", "Europe v2.50", []string{"93ea3bcee4252627919175ff1b16a1d9"}),
	model.HashedProvision(model.ProvisionBIOS, "scph10000.bin", "Japan v1.00 SCPH-10000", []string{"b7fa11e87d51752a98b38e3e691cbf17"}),
	model.HashedProvision(model.ProvisionBIOS, "scph39001.bin", "USA v1.60 SCPH-39001", []string{"d5ce2c7d119f563ce04bc04dbc3a323e"}),
	model.HashedProvision(model.ProvisionBIOS, "scph39004.bin", "Europe v1.60 SCPH-39004", []string{"1ad977bb539fc9448a08ab276a836bbc"}),
	model.HashedProvision(model.ProvisionBIOS, "scph50000.bin", "Japan v1.70 SCPH-50000", []string{"d3f1853a16c2ec18f3cd1ae655213308"}),
	model.HashedProvision(model.ProvisionBIOS, "scph70000.bin", "Japan v2.00 SCPH-70000", []string{"0bf988e9c7aaa4c051805b0fa6eb3387"}),
	model.HashedProvision(model.ProvisionBIOS, "scph70004.bin", "Europe v2.00 SCPH-70004", []string{"d333558cc14561c1fdc334c75d5f37b7"}),
	model.HashedProvision(model.ProvisionBIOS, "scph70012.bin", "USA v2.00 SCPH-70012", []string{"d333558cc14561c1fdc334c75d5f37b7"}),
	model.HashedProvision(model.ProvisionBIOS, "scph77001.bin", "USA v2.20 SCPH-77001", []string{"af60e6d1a939019d55e5b330d24b1c25"}),
}
