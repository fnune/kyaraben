package e2etestemu

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:         model.EmulatorE2ETest,
		Name:       "E2E Test",
		Systems:    []model.SystemID{model.SystemE2ETest},
		Package:    model.NixpkgsRef("hello"),
		Provisions: []model.Provision{},
		StateKinds: []model.StateKind{},
		Launcher: model.LauncherInfo{
			Binary:      "hello",
			GenericName: "Test Application",
			Categories:  []string{"Utility"},
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
