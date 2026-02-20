package esde

import (
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Frontend() model.Frontend {
	return model.Frontend{
		ID:      model.FrontendIDESDE,
		Name:    "EmulationStation DE",
		Package: model.AppImageRef("esde"),
		Launcher: model.LauncherInfo{
			Binary:      "esde",
			DisplayName: "EmulationStation DE",
			GenericName: "Game Frontend",
			Categories:  []string{"Game"},
			Keywords:    []string{"esde", "ES-DE", "frontend", "game launcher", "emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.FrontendConfigGenerator {
	return &Config{}
}

func (Definition) SteamShortcut(binDir string) *model.SteamShortcutInfo {
	return &model.SteamShortcutInfo{
		AppName: "EmulationStation DE",
		Tags:    []string{"Kyaraben"},
		GridAssets: &model.SteamGridAssets{
			Grid:    esdeGridAssets.Grid,
			Hero:    esdeGridAssets.Hero,
			Logo:    esdeGridAssets.Logo,
			Capsule: esdeGridAssets.Capsule,
		},
	}
}

func (d Definition) SteamShortcutExe(binDir string) string {
	return filepath.Join(binDir, "esde")
}
