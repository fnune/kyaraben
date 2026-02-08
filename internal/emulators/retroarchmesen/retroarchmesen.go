package retroarchmesen

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:         model.EmulatorIDRetroArchMesen,
		Name:       "RetroArch (Mesen)",
		Systems:    []model.SystemID{model.SystemIDNES},
		Package:    model.AppImageRef("retroarch"),
		Provisions: []model.Provision{},
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
	libretroCoreName = "mesen_libretro"
	shortCoreName    = "mesen"
)

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(shortCoreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemIDNES)},
		},
	}
}
