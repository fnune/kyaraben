package ppsspp

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDPPSSPP,
		Name:    "PPSSPP",
		Systems: []model.SystemID{model.SystemIDPSP},
		Package: model.AppImageRef("ppsspp"),
		// PPSSPP uses HLE - no BIOS required
		Provisions: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "ppsspp",
			GenericName: "PlayStation Portable Emulator",
			Categories:  []string{"Game", "Emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "ppsspp/PSP/SYSTEM/ppsspp.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// PPSSPP stores saves relative to the memstick directory
	// Use opaque dir since PPSSPP manages its own directory structure internally
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"General", "MemStickDirectory"}, Value: store.EmulatorOpaqueDir(model.EmulatorIDPPSSPP)},
			{Path: []string{"General", "ScreenshotsPath"}, Value: store.SystemScreenshotsDir(model.SystemIDPSP)},
		},
	}}, nil
}
