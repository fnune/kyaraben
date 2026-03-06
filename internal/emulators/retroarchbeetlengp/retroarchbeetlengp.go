package retroarchbeetlengp

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDRetroArchBeetleNGP,
		Name:            "Beetle NeoPop (RetroArch)",
		Systems:         []model.SystemID{model.SystemIDNGP},
		Package:         model.AppImageRef("retroarch"),
		ProvisionGroups: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: retroarch.LauncherWithCore(libretroCoreName),
		PathUsage: model.PathUsage{
			UsesBiosDir:        false,
			UsesSavesDir:       true,
			UsesStatesDir:      true,
			UsesScreenshotsDir: true,
		},
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
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchBeetleNGP, ctx.Store, ctx.BaseDirResolver)
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
	downloads, err := retroarch.CoreShaderDownloads(model.EmulatorIDRetroArchBeetleNGP, ctx.BaseDirResolver, pc)
	if err != nil {
		return model.GenerateResult{}, err
	}
	return model.GenerateResult{
		Patches:          retroarch.CorePatches(model.EmulatorIDRetroArchBeetleNGP, ctx.Store, ctx.ControllerConfig, pc, ctx.BaseDirResolver),
		Symlinks:         symlinks,
		InitialDownloads: downloads,
	}, nil
}

const libretroCoreName = "mednafen_ngp_libretro"
