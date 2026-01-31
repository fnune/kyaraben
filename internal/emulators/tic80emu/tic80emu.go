package tic80emu

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:         model.EmulatorTIC80,
		Name:       "TIC-80",
		Systems:    []model.SystemID{model.SystemTIC80},
		Package:    model.NixpkgsRef("tic-80"),
		Provisions: []model.Provision{},
		StateKinds: []model.StateKind{
			model.StateSaves,
		},
		Launcher: model.LauncherInfo{
			Binary:      "tic80",
			GenericName: "Fantasy Console",
			Categories:  []string{"Game", "Development"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return nil, nil
}
