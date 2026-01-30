package retroarchppsspp

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorRetroArchPPSSPP,
		Name:    "RetroArch (PPSSPP)",
		Systems: []model.SystemID{model.SystemPSP},
		Package: model.NixpkgsOverlayRef(
			"retroarch-ppsspp",
			`pkgs.wrapRetroArch { cores = with pkgs.libretro; [ ppsspp ]; }`,
		),
		// PPSSPP is an HLE emulator - no BIOS required.
		// See: https://docs.libretro.com/library/ppsspp/
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

const coreName = "ppsspp_libretro"

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(coreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"savefile_directory"}, Value: store.SystemSavesDir(model.SystemPSP)},
			{Path: []string{"savestate_directory"}, Value: store.EmulatorStatesDir(model.EmulatorRetroArchPPSSPP)},
			{Path: []string{"screenshot_directory"}, Value: store.SystemScreenshotsDir(model.SystemPSP)},
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemPSP)},
		},
	}
}
