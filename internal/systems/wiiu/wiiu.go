package wiiu

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
	return model.System{
		ID:           model.SystemIDWiiU,
		Name:         "Wii U",
		Description:  "Home console by Nintendo (2012)",
		Manufacturer: model.ManufacturerNintendo,
		Label:        "WiiU",
		Extensions:   []string{".wua", ".wux", ".rpx", ".wud", ".elf"},
		DisplayType:  model.DisplayTypeCRT,
	}
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
	return model.EmulatorIDCemu
}
