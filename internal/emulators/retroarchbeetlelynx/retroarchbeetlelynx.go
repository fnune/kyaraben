// BIOS hash data compiled from:
// - Libretro documentation (https://docs.libretro.com/library/beetle_lynx/)
package retroarchbeetlelynx

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRetroArchBeetleLynx,
		Name:    "Beetle Lynx (RetroArch)",
		Systems: []model.SystemID{model.SystemIDLynx},
		Package: model.AppImageRef("retroarch"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "BIOS required",
			Provisions:  lynxBIOSProvisions,
		}},
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
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchBeetleLynx, ctx.Store, ctx.BaseDirResolver)
	if err != nil {
		return model.GenerateResult{}, err
	}
	return model.GenerateResult{
		Patches: []model.ConfigPatch{
			retroarch.SharedConfig(ctx.Store, ctx.ControllerConfig),
			coreOverrideConfig(ctx.Store),
		},
		Symlinks: symlinks,
	}, nil
}

const (
	libretroCoreName = "mednafen_lynx_libretro"
	shortCoreName    = "mednafen_lynx"
)

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(shortCoreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"system_directory"}, Value: store.SystemBiosDir(model.SystemIDLynx)},
		},
	}
}

var lynxBIOSProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionBIOS, "lynxboot.img", "Lynx boot ROM", []string{"fcd403db69f54290b51035d82f835e7b"}),
}
