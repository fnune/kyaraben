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
		Launcher: retroarch.LauncherWithCore(coreName),
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

const coreName = "mednafen_saturn_libretro"

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(coreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"savefile_directory"}, Value: store.SystemSavesDir(model.SystemIDSaturn)},
			{Path: []string{"savestate_directory"}, Value: store.EmulatorStatesDir(model.EmulatorIDRetroArchBeetleSaturn)},
			{Path: []string{"screenshot_directory"}, Value: store.SystemScreenshotsDir(model.SystemIDSaturn)},
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemIDSaturn)},
		},
	}
}
