package model

// SystemID uniquely identifies a gaming platform.
//
// Constants use the full type name as prefix (SystemID*) because tygo's
// enum_style: union requires the prefix to match the type name exactly.
type SystemID string

const (
	SystemIDSNES      SystemID = "snes"
	SystemIDPSX       SystemID = "psx"
	SystemIDPS2       SystemID = "ps2"
	SystemIDPS3       SystemID = "ps3"
	SystemIDPSVita    SystemID = "psvita"
	SystemIDGBA       SystemID = "gba"
	SystemIDNDS       SystemID = "nds"
	SystemIDPSP       SystemID = "psp"
	SystemIDDreamcast SystemID = "dreamcast"
	SystemIDGameCube  SystemID = "gamecube"
	SystemIDWii       SystemID = "wii"
	SystemIDWiiU      SystemID = "wiiu"
	SystemID3DS       SystemID = "3ds"
	SystemIDSwitch    SystemID = "switch"
)

// Manufacturer represents the company that made a gaming system.
type Manufacturer string

const (
	ManufacturerNintendo Manufacturer = "Nintendo"
	ManufacturerSony     Manufacturer = "Sony"
	ManufacturerSega     Manufacturer = "Sega"
	ManufacturerOther    Manufacturer = "Other"
)

// System represents a gaming platform that can be emulated.
type System struct {
	ID           SystemID     `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Manufacturer Manufacturer `json:"manufacturer"`
	Label        string       `json:"label"`
}
