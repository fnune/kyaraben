package retroarchmame

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRetroArchMAME,
		Name:    "MAME (RetroArch)",
		Systems: []model.SystemID{model.SystemIDArcade, model.SystemIDNeoGeo},
		Package: model.AppImageRef("retroarch"),
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

var coreOptionsTarget = model.ConfigTarget{
	RelPath: "retroarch/config/MAME/MAME.opt",
	Format:  model.ConfigFormatCFG,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchMAME, ctx.Store, ctx.BaseDirResolver)
	if err != nil {
		return model.GenerateResult{}, err
	}
	pc := &retroarch.PresetConfig{
		Preset:             ctx.Preset,
		TargetDevice:       ctx.TargetDevice,
		Resume:             ctx.Resume,
		SystemDisplayTypes: ctx.SystemDisplayTypes,
	}
	downloads, err := retroarch.CoreShaderDownloads(model.EmulatorIDRetroArchMAME, ctx.BaseDirResolver, pc)
	if err != nil {
		return model.GenerateResult{}, err
	}
	systems := []model.SystemID{model.SystemIDArcade}
	embeddedFiles, err := retroarch.CoreEmbeddedFiles(model.EmulatorIDRetroArchMAME, systems, pc, ctx.BaseDirResolver)
	if err != nil {
		return model.GenerateResult{}, err
	}

	patches := retroarch.CorePatches(model.EmulatorIDRetroArchMAME, ctx.Store, ctx.ControllerConfig, pc, ctx.BaseDirResolver)
	patches = append(patches, model.ConfigPatch{
		Target: coreOptionsTarget,
		Entries: []model.ConfigEntry{
			// Leaving rotation to RetroArch lets video_allow_rotate decide, the same
			// switch FBNeo honours, so one setting controls vertical games on both cores.
			model.Default(model.None, model.Path("mame_rotation_mode"), "libretro"),
		},
	})
	patches = append(patches, retroarch.ContentDirBiosPatches(model.EmulatorIDRetroArchMAME, Definition{}.Emulator().Systems, ctx.Store)...)

	return model.GenerateResult{
		Patches:          patches,
		Symlinks:         symlinks,
		InitialDownloads: downloads,
		EmbeddedFiles:    embeddedFiles,
	}, nil
}

const libretroCoreName = "mame_libretro"
