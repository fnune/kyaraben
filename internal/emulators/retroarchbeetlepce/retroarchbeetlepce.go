// BIOS hash data compiled from:
// - Libretro documentation (https://docs.libretro.com/library/beetle_pce_fast/)
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - retroarch_system (https://github.com/Abdess/retroarch_system)
package retroarchbeetlepce

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRetroArchBeetlePCE,
		Name:    "Beetle PCE FAST (RetroArch)",
		Systems: []model.SystemID{model.SystemIDPCEngine},
		Package: model.AppImageRef("retroarch"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 0,
			Message:     "BIOS required for CD games (HuCard games work without BIOS)",
			Provisions:  pceBIOSProvisions,
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher:          retroarch.LauncherWithCore(libretroCoreName),
		PathUsage:         model.StandardPathUsage(),
		SupportedSettings: []string{model.SettingPreset, model.SettingResumeAutosave, model.SettingResumeAutoload},
		SupportedHotkeys:  retroarch.HotkeyMappings.SupportedHotkeys(),
		ResumeRecommended: true,
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchBeetlePCE, ctx.Store, ctx.BaseDirResolver)
	if err != nil {
		return model.GenerateResult{}, err
	}
	pc := &retroarch.PresetConfig{
		Preset:             ctx.Preset,
		Bezels:             ctx.Bezels,
		TargetDevice:       ctx.TargetDevice,
		Resume:             ctx.Resume,
		SystemDisplayTypes: ctx.SystemDisplayTypes,
	}
	downloads, err := retroarch.CoreShaderDownloads(model.EmulatorIDRetroArchBeetlePCE, ctx.BaseDirResolver, pc)
	if err != nil {
		return model.GenerateResult{}, err
	}
	embeddedFiles, err := retroarch.CoreEmbeddedFiles(model.EmulatorIDRetroArchBeetlePCE, pc, ctx.BaseDirResolver)
	if err != nil {
		return model.GenerateResult{}, err
	}

	patches := retroarch.CorePatches(model.EmulatorIDRetroArchBeetlePCE, ctx.Store, ctx.ControllerConfig, pc, ctx.BaseDirResolver)
	if overlayPatch := retroarch.OverlayPatch(model.EmulatorIDRetroArchBeetlePCE, pc, ctx.BaseDirResolver); overlayPatch != nil {
		patches = append(patches, *overlayPatch)
	}

	return model.GenerateResult{
		Patches:          patches,
		Symlinks:         symlinks,
		InitialDownloads: downloads,
		EmbeddedFiles:    embeddedFiles,
	}, nil
}

const libretroCoreName = "mednafen_pce_fast_libretro"

var pceBIOSProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionBIOS, "syscard3.pce", "Super System Card 3.0 JP (required for CD games)", []string{"38179df8f4ac870017db21ebcbf53114"}),
	model.HashedProvision(model.ProvisionBIOS, "syscard3u.pce", "Super System Card 3.0 US", []string{"0754f903b52e3b3342202bdafb13efa5"}),
	model.HashedProvision(model.ProvisionBIOS, "syscard2.pce", "System Card 2.0 JP", []string{"3cdd6614a918616bfc41c862e889dd79"}),
	model.HashedProvision(model.ProvisionBIOS, "syscard2u.pce", "System Card 2.0 US", []string{"94279f315e8b52904f65ab3108542afe"}),
	model.HashedProvision(model.ProvisionBIOS, "syscard1.pce", "System Card 1.0", []string{"2b7ccb3d86baa18f6402c176f3065082"}),
	model.HashedProvision(model.ProvisionBIOS, "gexpress.pce", "Game Express CD Card", []string{"6d2cb14fc3e1f65ceb135633d1694122"}),
}
