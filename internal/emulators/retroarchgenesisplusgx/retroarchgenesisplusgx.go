package retroarchgenesisplusgx

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDRetroArchGenesisPlusGX,
		Name:            "RetroArch (Genesis Plus GX)",
		Systems:         []model.SystemID{model.SystemIDGenesis},
		Package:         model.AppImageRef("retroarch"),
		ProvisionGroups: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: retroarch.LauncherWithCore(libretroCoreName),
		PathUsage: model.PathUsage{
			UsesBiosDir:        false,
			UsesSavesDir:       true,
			UsesStatesDir:      true,
			UsesScreenshotsDir: true,
		},
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

func (c *Config) Symlinks(store model.StoreReader, resolver model.BaseDirResolver) ([]model.SymlinkSpec, error) {
	return retroarch.CoreSymlinks(model.EmulatorIDRetroArchGenesisPlusGX, store, resolver)
}

const (
	libretroCoreName = "genesis_plus_gx_libretro"
	shortCoreName    = "genesis_plus_gx"
)

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(shortCoreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemIDGenesis)},
		},
	}
}
