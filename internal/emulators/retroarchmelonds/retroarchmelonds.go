package retroarchmelonds

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorRetroArchMelonDS,
		Name:    "RetroArch (melonDS)",
		Systems: []model.SystemID{model.SystemNDS},
		Package: model.NixpkgsOverlayRef(
			"retroarch-melonds",
			`pkgs.wrapRetroArch { cores = with pkgs.libretro; [ melonds ]; }`,
		),
		// BIOS files are optional - melonDS has built-in replacements.
		// See: https://docs.libretro.com/library/melonds/
		Provisions: []model.Provision{
			{
				ID:          "nds-bios-arm7",
				Kind:        model.ProvisionBIOS,
				Filename:    "bios7.bin",
				Description: "Nintendo DS ARM7 BIOS",
				Required:    false,
				MD5Hash:     "df692a80a5b1bc90728bc3dfc76cd948",
			},
			{
				ID:          "nds-bios-arm9",
				Kind:        model.ProvisionBIOS,
				Filename:    "bios9.bin",
				Description: "Nintendo DS ARM9 BIOS",
				Required:    false,
				MD5Hash:     "a392174eb3e572fed6447e956bde4b25",
			},
			{
				ID:          "nds-firmware",
				Kind:        model.ProvisionFirmware,
				Filename:    "firmware.bin",
				Description: "Nintendo DS Firmware",
				Required:    false,
				MD5Hash:     "e45033d9b0fa6b0de071292bba7c9d13",
			},
		},
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

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{
		retroarch.SharedConfig(store),
		coreOverrideConfig(store),
	}, nil
}

const coreName = "melonds_libretro"

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(coreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"savefile_directory"}, Value: store.SystemSavesDir(model.SystemNDS)},
			{Path: []string{"savestate_directory"}, Value: store.EmulatorStatesDir(model.EmulatorRetroArchMelonDS)},
			{Path: []string{"screenshot_directory"}, Value: store.SystemScreenshotsDir(model.SystemNDS)},
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemNDS)},
		},
	}
}
