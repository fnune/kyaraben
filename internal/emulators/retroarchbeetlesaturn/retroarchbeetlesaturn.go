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
		Provisions: []model.Provision{
			{
				ID:          "saturn-bios",
				Kind:        model.ProvisionBIOS,
				Filename:    "sega_101.bin",
				Description: "Saturn BIOS (Japan)",
				Required:    true,
				MD5Hash:     "85ec9ca47d8f6807718151cbcca8b964",
			},
			{
				ID:          "saturn-bios-us",
				Kind:        model.ProvisionBIOS,
				Filename:    "mpr-17933.bin",
				Description: "Saturn BIOS (US/EU)",
				Required:    true,
				MD5Hash:     "3240872c70984b6cbfda1586cab68dbe",
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
