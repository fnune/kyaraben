package rpcs3

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorRPCS3,
		Name:    "RPCS3",
		Systems: []model.SystemID{model.SystemPS3},
		Package: model.VersionedAppImageRef("rpcs3"),
		// PS3 firmware is installed through the emulator
		Provisions: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "rpcs3",
			GenericName: "PlayStation 3 Emulator",
			Categories:  []string{"Game", "Emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "rpcs3/config.yml",
	Format:  model.ConfigFormatYAML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// RPCS3 uses YAML config with complex nested structure
	// Most paths are relative to RPCS3's data directory
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"VFS", "$(EmulatorDir)"}, Value: store.EmulatorOpaqueDir(model.EmulatorRPCS3)},
		},
	}}, nil
}
