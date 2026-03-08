package nintendo3ds

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDN3DS,
		Name:         "Nintendo 3DS",
		Description:  "Handheld console by Nintendo (2011)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "3DS",
		Extensions:   []string{".3ds", ".3dsx", ".cia", ".cxi", ".cci", ".app", ".7z", ".zip"},
		DisplayType:  model.DisplayTypeLCD,
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	// Azahar crashes on Vulkan: https://github.com/azahar-emu/azahar/pull/1825
	return model.EmulatorIDRetroArchCitra
}
