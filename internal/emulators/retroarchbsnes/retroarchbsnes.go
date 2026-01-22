package retroarchbsnes

import (
	"os"
	"path/filepath"

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
		ConfigPaths: []string{
			"~/.config/retroarch/retroarch.cfg",
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

func (c *Config) ConfigPaths() []string {
	configDir, _ := os.UserConfigDir()
	return []string{
		filepath.Join(configDir, "retroarch", "retroarch.cfg"),
	}
}

func (c *Config) Generate(store model.StoreReader, systems []model.SystemID) ([]model.ConfigPatch, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "retroarch", "retroarch.cfg")

	var primarySystem model.SystemID
	for _, sys := range systems {
		if sys == model.SystemSNES {
			primarySystem = sys
			break
		}
	}

	entries := []model.ConfigEntry{
		{Key: "system_directory", Value: quote(store.BiosDir())},
		{Key: "savefile_directory", Value: quote(store.SystemSavesDir(primarySystem))},
		{Key: "savestate_directory", Value: quote(store.SystemStatesDir(primarySystem))},
		{Key: "screenshot_directory", Value: quote(store.SystemScreenshotsDir(primarySystem))},
		{Key: "rgui_browser_directory", Value: quote(store.SystemRomsDir(primarySystem))},
		{Key: "sort_savefiles_enable", Value: "false"},
		{Key: "sort_savestates_enable", Value: "false"},
		{Key: "sort_screenshots_enable", Value: "false"},
	}

	patch := model.ConfigPatch{
		Config: model.EmulatorConfig{
			Path:       configPath,
			Format:     model.ConfigFormatCFG,
			EmulatorID: model.EmulatorRetroArchBsnes,
		},
		Entries: entries,
	}

	return []model.ConfigPatch{patch}, nil
}

func quote(s string) string {
	return `"` + s + `"`
}
