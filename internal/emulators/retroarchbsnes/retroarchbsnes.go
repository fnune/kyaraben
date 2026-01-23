package retroarchbsnes

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorRetroArchBsnes,
		Name:    "RetroArch (bsnes)",
		Systems: []model.SystemID{model.SystemSNES},
		Package: model.NixpkgsOverlayRef(
			"retroarch-bsnes",
			`pkgs.retroarch.override { cores = with pkgs.libretro; [ bsnes ]; }`,
		),
		Provisions: []model.Provision{},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader, systems []model.SystemID) ([]model.ConfigPatch, error) {
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
			{Path: []string{"savefile_directory"}, Value: quote(store.SystemSavesDir(model.SystemSNES))},
			{Path: []string{"savestate_directory"}, Value: quote(store.SystemStatesDir(model.SystemSNES))},
			{Path: []string{"screenshot_directory"}, Value: quote(store.SystemScreenshotsDir(model.SystemSNES))},
			{Path: []string{"rgui_browser_directory"}, Value: quote(store.SystemRomsDir(model.SystemSNES))},
		},
	}
}

func quote(s string) string {
	return `"` + s + `"`
}
