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
		ProvisionGroups: []model.ProvisionGroup{
			{
				MinRequired: 1,
				Message:     "BIOS required (no HLE fallback available)",
				Provisions:  pcfxBIOSProvisions,
			},
			{
				MinRequired: 0,
				Message:     "Optional firmware for extended hardware support",
				Provisions:  pcfxOptionalProvisions,
			},
		},
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

var pcfxOptionalProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionFirmware, "pcfxv101.bin", "PC-FX BIOS v1.01", []string{"e2fb7c7220e3a7838c2dd7e401a7f3d8"}),
	model.HashedProvision(model.ProvisionFirmware, "pcfxga.rom", "PC-FX graphics accelerator BIOS", []string{"5885bc9a64bf80d4530b9b9b978ff587"}),
	model.HashedProvision(model.ProvisionFirmware, "fx-scsi.rom", "PC-FX SCSI controller BIOS", []string{"430e9745f9235c515bc8e652d6ca3004"}),
}
