package model

// SystemID uniquely identifies a gaming platform.
//
// Constants use the full type name as prefix (SystemID*) because tygo's
// enum_style: union requires the prefix to match the type name exactly.
type SystemID string

const (
	SystemIDNES       SystemID = "nes"
	SystemIDSNES      SystemID = "snes"
	SystemIDN64       SystemID = "n64"
	SystemIDGB        SystemID = "gb"
	SystemIDGBC       SystemID = "gbc"
	SystemIDGBA       SystemID = "gba"
	SystemIDNDS       SystemID = "nds"
	SystemID3DS       SystemID = "3ds"
	SystemIDGameCube  SystemID = "gamecube"
	SystemIDWii       SystemID = "wii"
	SystemIDWiiU      SystemID = "wiiu"
	SystemIDSwitch    SystemID = "switch"
	SystemIDPSX       SystemID = "psx"
	SystemIDPS2       SystemID = "ps2"
	SystemIDPS3       SystemID = "ps3"
	SystemIDPSP       SystemID = "psp"
	SystemIDPSVita    SystemID = "psvita"
	SystemIDGenesis   SystemID = "genesis"
	SystemIDSaturn    SystemID = "saturn"
	SystemIDDreamcast SystemID = "dreamcast"
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
