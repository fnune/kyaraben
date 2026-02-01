package retroarchgenesisplusgx

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:         model.EmulatorIDRetroArchGenesisPlusGX,
		Name:       "RetroArch (Genesis Plus GX)",
		Systems:    []model.SystemID{model.SystemIDGenesis},
		Package:    model.AppImageRef("retroarch"),
		Provisions: []model.Provision{},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: retroarch.SharedLauncher,
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

const coreName = "genesis_plus_gx_libretro"

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(coreName),
		Entries: []model.ConfigEntry{
			// Per-core save directory for individual sync capability
			{Path: []string{"savefile_directory"}, Value: store.EmulatorSavesDir(model.EmulatorIDRetroArchGenesisPlusGX)},
			{Path: []string{"savestate_directory"}, Value: store.EmulatorStatesDir(model.EmulatorIDRetroArchGenesisPlusGX)},
			{Path: []string{"screenshot_directory"}, Value: store.SystemScreenshotsDir(model.SystemIDGenesis)},
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemIDGenesis)},
		},
	}
}
