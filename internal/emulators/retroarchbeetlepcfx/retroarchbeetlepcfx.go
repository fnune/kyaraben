// BIOS hash data compiled from:
// - Libretro documentation (https://docs.libretro.com/library/beetle_pc_fx/)
package retroarchbeetlepcfx

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRetroArchBeetlePCFX,
		Name:    "Beetle PC-FX (RetroArch)",
		Systems: []model.SystemID{model.SystemIDPCFX},
		Package: model.AppImageRef("retroarch"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "BIOS required (no HLE fallback available)",
			Provisions:  pcfxBIOSProvisions,
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
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchBeetlePCFX, ctx.Store, ctx.BaseDirResolver)
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
	libretroCoreName = "mednafen_pcfx_libretro"
	shortCoreName    = "mednafen_pcfx"
)

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(shortCoreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"system_directory"}, Value: store.SystemBiosDir(model.SystemIDPCFX)},
		},
	}
}

var pcfxBIOSProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionBIOS, "pcfx.rom", "PC-FX BIOS v1.00", []string{"08e36edbea28a017f79f8d4f7ff9b6d7"}),
}
