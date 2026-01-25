package retroarchbsnes

import "github.com/fnune/kyaraben/internal/model"

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

var configTarget = model.ConfigTarget{
	RelPath: "retroarch/retroarch.cfg",
	Format:  model.ConfigFormatCFG,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader, systems []model.SystemID) ([]model.ConfigPatch, error) {
	var primarySystem model.SystemID
	for _, sys := range systems {
		if sys == model.SystemSNES {
			primarySystem = sys
			break
		}
	}

	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Key: "system_directory", Value: quote(store.BiosDir())},
			{Key: "savefile_directory", Value: quote(store.SystemSavesDir(primarySystem))},
			{Key: "savestate_directory", Value: quote(store.SystemStatesDir(primarySystem))},
			{Key: "screenshot_directory", Value: quote(store.SystemScreenshotsDir(primarySystem))},
			{Key: "rgui_browser_directory", Value: quote(store.SystemRomsDir(primarySystem))},
			{Key: "sort_savefiles_enable", Value: "false"},
			{Key: "sort_savestates_enable", Value: "false"},
			{Key: "sort_screenshots_enable", Value: "false"},
		},
	}}, nil
}

func quote(s string) string {
	return `"` + s + `"`
}
