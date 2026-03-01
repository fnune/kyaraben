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
		Launcher:  retroarch.LauncherWithCore(libretroCoreName),
		PathUsage: model.StandardPathUsage(),
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

	patches := retroarch.CorePatches(model.EmulatorIDRetroArchVICE, ctx.Store, ctx.ControllerConfig)
	patches = append(patches, model.ConfigPatch{
		Target: coreOptionsTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"vice_drive_true_emulation"}, Value: "disabled", DefaultOnly: true},
			{Path: []string{"vice_autoloadwarp"}, Value: "enabled", DefaultOnly: true},
		},
	})

	return model.GenerateResult{
		Patches:  patches,
		Symlinks: symlinks,
	}, nil
}

const libretroCoreName = "vice_x64sc_libretro"
