package neogeo

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDNeoGeo,
		Name:         "SNK Neo Geo",
		Description:  "24-bit arcade/home system by SNK (1990)",
		Manufacturer: model.ManufacturerSNK,
		Label:        "NEOGEO",
		Extensions:   []string{".zip", ".7z"},
		DisplayType:  model.DisplayTypeCRT,
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDRetroArchFBNeo
}
