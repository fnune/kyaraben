// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
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
		Name:    "RetroArch (Beetle Saturn)",
		Systems: []model.SystemID{model.SystemIDSaturn},
		Package: model.AppImageRef("retroarch"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "At least one BIOS required",
			Provisions: []model.Provision{
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "sega_101.bin",
					Description: "Japan",
					Hashes: []string{
						"85ec9ca47d8f6807718151cbcca8b964",
						"f273555d7d91e8a5a6bfd9bcf066331c",
					},
				},
				{
					Kind:        model.ProvisionBIOS,
					Filename:    "mpr-17933.bin",
					Description: "US/EU",
					Hashes: []string{
						"3240872c70984b6cbfda1586cab68dbe",
						"ac4e4b6522e200c0d23d371a8cecbfd3",
					},
				},
			},
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

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{
		retroarch.SharedConfig(store),
		coreOverrideConfig(store),
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
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemIDSaturn)},
		},
	}
}
