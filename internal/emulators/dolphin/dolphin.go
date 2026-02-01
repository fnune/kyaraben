package dolphin

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorDolphin,
		Name:    "Dolphin",
		Systems: []model.SystemID{model.SystemGameCube, model.SystemWii},
		Package: model.VersionedAppImageRef("dolphin"),
		// Wii system menu optional, GameCube IPL optional
		Provisions: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "dolphin-emu",
			GenericName: "GameCube/Wii Emulator",
			Categories:  []string{"Game", "Emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "dolphin-emu/Dolphin.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// Dolphin uses INI config
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"General", "ISOPath0"}, Value: store.SystemRomsDir(model.SystemGameCube)},
			{Path: []string{"General", "ISOPath1"}, Value: store.SystemRomsDir(model.SystemWii)},
			{Path: []string{"General", "ISOPaths"}, Value: "2"},
		},
	}}, nil
}
