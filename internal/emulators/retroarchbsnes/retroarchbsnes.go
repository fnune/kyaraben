package retroarchbsnes

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:         model.EmulatorIDRetroArchBsnes,
		Name:       "RetroArch (bsnes)",
		Systems:    []model.SystemID{model.SystemIDSNES},
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

const coreName = "bsnes_libretro"

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(coreName),
		Entries: []model.ConfigEntry{
			// Per-core save directory for individual sync capability
			{Path: []string{"savefile_directory"}, Value: store.EmulatorSavesDir(model.EmulatorIDRetroArchBsnes)},
			{Path: []string{"savestate_directory"}, Value: store.EmulatorStatesDir(model.EmulatorIDRetroArchBsnes)},
			{Path: []string{"screenshot_directory"}, Value: store.SystemScreenshotsDir(model.SystemIDSNES)},
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemIDSNES)},
		},
	}
}
