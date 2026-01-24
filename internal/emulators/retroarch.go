package emulators

import (
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

// RetroArchConfig generates RetroArch configuration.
type RetroArchConfig struct{}

func (r *RetroArchConfig) EmulatorID() model.EmulatorID {
	return model.EmulatorRetroArchBsnes
}

func (r *RetroArchConfig) ConfigPaths() []string {
	configDir, _ := os.UserConfigDir()
	return []string{
		filepath.Join(configDir, "retroarch", "retroarch.cfg"),
	}
}

func (r *RetroArchConfig) Generate(userStore *store.UserStore, systems []model.SystemID) ([]model.ConfigPatch, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "retroarch", "retroarch.cfg")

	// Find the first system that RetroArch handles (for saves/states paths)
	// In practice, RetroArch uses one directory for all cores
	var primarySystem model.SystemID
	for _, sys := range systems {
		if sys == model.SystemSNES {
			primarySystem = sys
			break
		}
	}

	entries := []model.ConfigEntry{
		{Key: "system_directory", Value: quote(userStore.BiosDir())},
		{Key: "savefile_directory", Value: quote(userStore.SystemSavesDir(primarySystem))},
		{Key: "savestate_directory", Value: quote(userStore.SystemStatesDir(primarySystem))},
		{Key: "screenshot_directory", Value: quote(userStore.SystemScreenshotsDir(primarySystem))},
		{Key: "rgui_browser_directory", Value: quote(userStore.SystemRomsDir(primarySystem))},
		// Disable RetroArch's own directory sorting to use our structure
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
