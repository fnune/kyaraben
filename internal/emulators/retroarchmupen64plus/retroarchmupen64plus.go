package retroarchmupen64plus

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDRetroArchMupen64Plus,
		Name:            "Mupen64Plus-Next (RetroArch)",
		Systems:         []model.SystemID{model.SystemIDN64},
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
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var coreOptionsTarget = model.ConfigTarget{
	RelPath: "retroarch/config/Mupen64Plus-Next/Mupen64Plus-Next.opt",
	Format:  model.ConfigFormatCFG,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchMupen64Plus, ctx.Store, ctx.BaseDirResolver)
	if err != nil {
		return model.GenerateResult{}, err
	}
	sc := &retroarch.ShaderConfig{
		Shaders:            ctx.Shaders,
		SystemDisplayTypes: ctx.SystemDisplayTypes,
	}
	downloads, err := retroarch.CoreShaderDownloads(model.EmulatorIDRetroArchMupen64Plus, ctx.BaseDirResolver, sc)
	if err != nil {
		return model.GenerateResult{}, err
	}

	patches := retroarch.CorePatches(model.EmulatorIDRetroArchMupen64Plus, ctx.Store, ctx.ControllerConfig, sc, ctx.BaseDirResolver)
	patches = append(patches, model.ConfigPatch{
		Target: coreOptionsTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"mupen64plus-43screensize"}, Value: "1280x960", DefaultOnly: true},
			{Path: []string{"mupen64plus-169screensize"}, Value: "1920x1080", DefaultOnly: true},
		},
	})

	return model.GenerateResult{
		Patches:          patches,
		Symlinks:         symlinks,
		InitialDownloads: downloads,
	}, nil
}

const libretroCoreName = "mupen64plus_next_libretro"
