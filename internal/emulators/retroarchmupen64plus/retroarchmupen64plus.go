package retroarchmupen64plus

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:         model.EmulatorIDRetroArchMupen64Plus,
		Name:       "RetroArch (Mupen64Plus-Next)",
		Systems:    []model.SystemID{model.SystemIDN64},
		Package:    model.AppImageRef("retroarch"),
		Provisions: []model.Provision{},
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

const coreName = "mupen64plus_next_libretro"

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(coreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"savefile_directory"}, Value: store.SystemSavesDir(model.SystemIDN64)},
			{Path: []string{"savestate_directory"}, Value: store.EmulatorStatesDir(model.EmulatorIDRetroArchMupen64Plus)},
			{Path: []string{"screenshot_directory"}, Value: store.SystemScreenshotsDir(model.SystemIDN64)},
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemIDN64)},
		},
	}
}
