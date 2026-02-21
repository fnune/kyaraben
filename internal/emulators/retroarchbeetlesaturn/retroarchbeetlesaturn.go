// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package retroarchbeetlesaturn

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRetroArchBeetleSaturn,
		Name:    "Beetle Saturn (RetroArch)",
		Systems: []model.SystemID{model.SystemIDSaturn},
		Package: model.AppImageRef("retroarch"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "BIOS required (no HLE fallback available)",
			Provisions:  saturnBIOSProvisions,
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
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchBeetleSaturn, ctx.Store, ctx.BaseDirResolver)
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
	libretroCoreName = "mednafen_saturn_libretro"
	shortCoreName    = "mednafen_saturn"
)

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(shortCoreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"system_directory"}, Value: store.SystemBiosDir(model.SystemIDSaturn)},
		},
	}
}

var saturnBIOSProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionBIOS, "sega_101.bin", "Japan", []string{"85ec9ca47d8f6807718151cbcca8b964"}),
	model.HashedProvision(model.ProvisionBIOS, "sega_100.bin", "Japan", []string{"af5828fdff51384f99b3c4926be27762"}),
	model.HashedProvision(model.ProvisionBIOS, "sega_100a.bin", "Japan alternate", []string{"f273555d7d91e8a5a6bfd9bcf066331c"}),
	model.HashedProvision(model.ProvisionBIOS, "sega1003.bin", "Japan v1.003", []string{"74570fed4d44b2682b560c8cd44b8b6a"}),
	model.HashedProvision(model.ProvisionBIOS, "mpr-17933.bin", "US/EU", []string{"3240872c70984b6cbfda1586cab68dbe"}),
	model.HashedProvision(model.ProvisionBIOS, "mpr-18100.bin", "US/EU v1.01", []string{"cb2cebc1b6e573b7c44523d037edcd45"}),
	model.HashedProvision(model.ProvisionBIOS, "saturn_bios.bin", "Generic", []string{"af5828fdff51384f99b3c4926be27762"}),
	model.HashedProvision(model.ProvisionBIOS, "hisaturn.bin", "Hi-Saturn Japan", []string{"3ea3202e2634cb47cb90f3a05c015010"}),
	model.HashedProvision(model.ProvisionBIOS, "vsaturn.bin", "V-Saturn Japan", []string{"ac4e4b6522e200c0d23d371a8cecbfd3"}),
	model.HashedProvision(model.ProvisionBIOS, "mpr-18811-mx.ic1", "KOF95 cartridge", []string{"255113ba943c92a54facd25a10fd780c"}),
	model.HashedProvision(model.ProvisionBIOS, "mpr-19367-mx.ic1", "Ultraman cartridge", []string{"1cd19988d1d72a3e7caa0b73234c96b4"}),
}
