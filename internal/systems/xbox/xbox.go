package xbox

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDXbox,
		Name:         "Microsoft Xbox",
		Description:  "Home console by Microsoft (2001)",
		Manufacturer: model.ManufacturerMicrosoft,
		Label:        "XBOX",
		Extensions:   []string{".iso", ".xiso"},
		DisplayType:  model.DisplayTypeCRT,
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDXemu
}
