package retroarchstella

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDRetroArchStella,
		Name:            "Stella (RetroArch)",
		Systems:         []model.SystemID{model.SystemIDAtari2600},
		Package:         model.AppImageRef("retroarch"),
		ProvisionGroups: nil,
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
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchStella, ctx.Store, ctx.BaseDirResolver)
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
	downloads, err := retroarch.CoreShaderDownloads(model.EmulatorIDRetroArchStella, ctx.BaseDirResolver, pc)
	if err != nil {
		return model.GenerateResult{}, err
	}
	embeddedFiles, err := retroarch.CoreEmbeddedFiles(model.EmulatorIDRetroArchStella, pc, ctx.BaseDirResolver)
	if err != nil {
		return model.GenerateResult{}, err
	}

	patches := retroarch.CorePatches(model.EmulatorIDRetroArchStella, ctx.Store, ctx.ControllerConfig, pc, ctx.BaseDirResolver)
	if overlayPatch := retroarch.OverlayPatch(model.EmulatorIDRetroArchStella, pc, ctx.BaseDirResolver); overlayPatch != nil {
		patches = append(patches, *overlayPatch)
	}

	return model.GenerateResult{
		Patches:          patches,
		Symlinks:         symlinks,
		InitialDownloads: downloads,
		EmbeddedFiles:    embeddedFiles,
	}, nil
}

const libretroCoreName = "stella_libretro"
