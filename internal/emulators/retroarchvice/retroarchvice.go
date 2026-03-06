package retroarchvice

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDRetroArchVICE,
		Name:            "VICE (RetroArch)",
		Systems:         []model.SystemID{model.SystemIDC64},
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

var coreOptionsTarget = model.ConfigTarget{
	RelPath: "retroarch/config/VICE x64sc/VICE x64sc.opt",
	Format:  model.ConfigFormatCFG,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchVICE, ctx.Store, ctx.BaseDirResolver)
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
	downloads, err := retroarch.CoreShaderDownloads(model.EmulatorIDRetroArchVICE, ctx.BaseDirResolver, pc)
	if err != nil {
		return model.GenerateResult{}, err
	}
	embeddedFiles, err := retroarch.CoreEmbeddedFiles(model.EmulatorIDRetroArchVICE, pc, ctx.BaseDirResolver)
	if err != nil {
		return model.GenerateResult{}, err
	}

	patches := retroarch.CorePatches(model.EmulatorIDRetroArchVICE, ctx.Store, ctx.ControllerConfig, pc, ctx.BaseDirResolver)
	patches = append(patches, model.ConfigPatch{
		Target: coreOptionsTarget,
		Entries: []model.ConfigEntry{
			model.Default(model.None, model.Path("vice_drive_true_emulation"), "disabled"),
			model.Default(model.None, model.Path("vice_autoloadwarp"), "enabled"),
		},
	})
	if overlayPatch := retroarch.OverlayPatch(model.EmulatorIDRetroArchVICE, pc, ctx.BaseDirResolver); overlayPatch != nil {
		patches = append(patches, *overlayPatch)
	}

	return model.GenerateResult{
		Patches:          patches,
		Symlinks:         symlinks,
		InitialDownloads: downloads,
		EmbeddedFiles:    embeddedFiles,
	}, nil
}

const libretroCoreName = "vice_x64sc_libretro"
