package vita3k

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorVita3K,
		Name:    "Vita3K",
		Systems: []model.SystemID{model.SystemPSVita},
		Package: model.AppImageRef("vita3k"),
		// PS Vita firmware is downloaded through emulator
		Provisions: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "vita3k",
			GenericName: "PlayStation Vita Emulator",
			Categories:  []string{"Game", "Emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "Vita3K/config.yml",
	Format:  model.ConfigFormatYAML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// Vita3K uses YAML config and stores data in its own directory structure
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"pref-path"}, Value: store.EmulatorOpaqueDir(model.EmulatorVita3K)},
		},
	}}, nil
}
