package ppsspp

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorPPSSPP,
		Name:    "PPSSPP",
		Systems: []model.SystemID{model.SystemPSP},
		Package: model.NixpkgsRef("ppsspp"),
		// PPSSPP is an HLE emulator - no BIOS required.
		// See: https://www.ppsspp.org/docs/settings/
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

// PPSSPP uses a "memstick" directory structure mimicking the PSP.
// Config is at ~/.config/ppsspp/PSP/SYSTEM/ppsspp.ini
// See: https://www.ppsspp.org/docs/settings/
var configTarget = model.ConfigTarget{
	RelPath: "ppsspp/PSP/SYSTEM/ppsspp.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader, systems []model.SystemID) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			// General section - set the current directory for ROM browsing
			{Path: []string{"General", "CurrentDirectory"}, Value: store.SystemRomsDir(model.SystemPSP)},
		},
	}}, nil
}
