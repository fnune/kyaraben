package e2etestemu

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:          model.EmulatorE2ETest,
		Name:        "E2E Test",
		Systems:     []model.SystemID{model.SystemE2ETest},
		Package:     model.NixpkgsRef("hello"),
		Provisions:  []model.Provision{},
		StateKinds:  []model.StateKind{},
		ConfigPaths: []string{},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

func (c *Config) ConfigPaths() []string {
	return []string{}
}

func (c *Config) Generate(store model.StoreReader, systems []model.SystemID) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{}, nil
}
