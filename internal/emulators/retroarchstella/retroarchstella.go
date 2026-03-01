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
		Launcher:  retroarch.LauncherWithCore(libretroCoreName),
		PathUsage: model.StandardPathUsage(),
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
	sc := &retroarch.ShaderConfig{
		Shaders:            ctx.Shaders,
		SystemDisplayTypes: ctx.SystemDisplayTypes,
	}
	downloads, err := retroarch.CoreShaderDownloads(model.EmulatorIDRetroArchStella, ctx.BaseDirResolver, sc)
	if err != nil {
		return model.GenerateResult{}, err
	}
	return model.GenerateResult{
		Patches:          retroarch.CorePatches(model.EmulatorIDRetroArchStella, ctx.Store, ctx.ControllerConfig, sc, ctx.BaseDirResolver),
		Symlinks:         symlinks,
		InitialDownloads: downloads,
	}, nil
}

const libretroCoreName = "stella_libretro"
