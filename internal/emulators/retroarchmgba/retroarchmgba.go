package retroarchmgba

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorRetroArchMGBA,
		Name:    "RetroArch (mGBA)",
		Systems: []model.SystemID{model.SystemGBA},
		Package: model.NixpkgsOverlayRef(
			"retroarch-mgba",
			`pkgs.wrapRetroArch { cores = with pkgs.libretro; [ mgba ]; }`,
		),
		// BIOS is optional - mGBA has built-in high-level emulation.
		// See: https://docs.libretro.com/library/mgba/
		Provisions: []model.Provision{
			{
				ID:          "gba-bios",
				Kind:        model.ProvisionBIOS,
				Filename:    "gba_bios.bin",
				Description: "Game Boy Advance BIOS",
				Required:    false,
				MD5Hash:     "a860e8c0b6d573d191e4ec7db1b1e4f6",
			},
		},
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

const coreName = "mgba_libretro"

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(coreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"savefile_directory"}, Value: store.SystemSavesDir(model.SystemGBA)},
			{Path: []string{"savestate_directory"}, Value: store.EmulatorStatesDir(model.EmulatorRetroArchMGBA)},
			{Path: []string{"screenshot_directory"}, Value: store.SystemScreenshotsDir(model.SystemGBA)},
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemGBA)},
		},
	}
}
