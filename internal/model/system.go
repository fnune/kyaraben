package model

// SystemID uniquely identifies a gaming platform.
type SystemID string

const (
	SystemSNES      SystemID = "snes"
	SystemPSX       SystemID = "psx"
	SystemPS2       SystemID = "ps2"
	SystemPS3       SystemID = "ps3"
	SystemPSVita    SystemID = "psvita"
	SystemGBA       SystemID = "gba"
	SystemNDS       SystemID = "nds"
	SystemPSP       SystemID = "psp"
	SystemDreamcast SystemID = "dreamcast"
	SystemGameCube  SystemID = "gamecube"
	SystemWii       SystemID = "wii"
	SystemWiiU      SystemID = "wiiu"
	System3DS       SystemID = "3ds"
	SystemSwitch    SystemID = "switch"
	SystemE2ETest   SystemID = "e2e-test"
)

// System represents a gaming platform that can be emulated.
type System struct {
	ID          SystemID
	Name        string
	Description string
	Hidden      bool
}
