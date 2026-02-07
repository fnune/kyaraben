package esde

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Frontend() model.Frontend {
	return model.Frontend{
		ID:      model.FrontendIDESDE,
		Name:    "ES-DE",
		Package: model.AppImageRef("es-de"),
		Launcher: model.LauncherInfo{
			Binary:      "es-de",
			DisplayName: "ES-DE",
			GenericName: "Game Frontend",
			Categories:  []string{"Game"},
		},
	}
}

func (Definition) ConfigGenerator() model.FrontendConfigGenerator {
	return &Config{}
}
