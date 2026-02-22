package xbox360

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDXbox360,
		Name:         "Microsoft Xbox 360",
		Description:  "Home console by Microsoft (2005)",
		Manufacturer: model.ManufacturerMicrosoft,
		Label:        "360",
		Extensions:   []string{".iso", ".xex", ".zar", ".god"},
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDXeniaEdge
}
