package pcsx2

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorPCSX2,
		Name:    "PCSX2",
		Systems: []model.SystemID{model.SystemPS2},
		Package: model.VersionedAppImageRef("pcsx2"),
		Provisions: []model.Provision{
			{
				ID:          "ps2-bios-usa",
				Kind:        model.ProvisionBIOS,
				Filename:    "scph39001.bin",
				Description: "PlayStation 2 BIOS (USA)",
				Required:    true,
				MD5Hash:     "d5ce2c7d119f563ce04bc04dbc3a323e",
			},
			{
				ID:          "ps2-bios-europe",
				Kind:        model.ProvisionBIOS,
				Filename:    "scph70004.bin",
				Description: "PlayStation 2 BIOS (Europe)",
				Required:    false,
				MD5Hash:     "d333558cc14561c1fdc334c75d5f37b7",
			},
			{
				ID:          "ps2-bios-japan",
				Kind:        model.ProvisionBIOS,
				Filename:    "scph10000.bin",
				Description: "PlayStation 2 BIOS (Japan)",
				Required:    false,
				MD5Hash:     "2e6e6db3a66e65e86ad75389cd7fb4b6",
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "pcsx2",
			GenericName: "PlayStation 2 Emulator",
			Categories:  []string{"Game", "Emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "PCSX2/inis/PCSX2.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"Folders", "Bios"}, Value: store.SystemBiosDir(model.SystemPS2)},
			{Path: []string{"Folders", "MemoryCards"}, Value: store.SystemSavesDir(model.SystemPS2)},
			{Path: []string{"Folders", "Savestates"}, Value: store.EmulatorStatesDir(model.EmulatorPCSX2)},
			{Path: []string{"Folders", "Screenshots"}, Value: store.SystemScreenshotsDir(model.SystemPS2)},
			{Path: []string{"GameList", "RecursivePaths"}, Value: store.SystemRomsDir(model.SystemPS2)},
		},
	}}, nil
}
