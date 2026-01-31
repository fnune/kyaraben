package model

// EmulatorID uniquely identifies an emulator implementation.
// Format: "emulator" or "emulator:core" for RetroArch cores.
type EmulatorID string

const (
	// EmulatorRetroArch is the base RetroArch installation shared by all cores.
	// Used for shared data like the cores directory.
	EmulatorRetroArch      EmulatorID = "retroarch"
	EmulatorRetroArchBsnes EmulatorID = "retroarch:bsnes"
	EmulatorDuckStation    EmulatorID = "duckstation"
	EmulatorPCSX2          EmulatorID = "pcsx2"
	EmulatorRPCS3          EmulatorID = "rpcs3"
	EmulatorVita3K         EmulatorID = "vita3k"
	EmulatorPPSSPP         EmulatorID = "ppsspp"
	EmulatorMGBA           EmulatorID = "mgba"
	EmulatorMelonDS        EmulatorID = "melonds"
	EmulatorFlycast        EmulatorID = "flycast"
	EmulatorCemu           EmulatorID = "cemu"
	EmulatorAzahar         EmulatorID = "azahar"
	EmulatorDolphin        EmulatorID = "dolphin"
	EmulatorEden           EmulatorID = "eden"
)

// PackageSource indicates where an emulator package comes from.
type PackageSource string

const (
	PackageSourceAppImage PackageSource = "appimage"
)

type Emulator struct {
	ID         EmulatorID
	Name       string
	Systems    []SystemID
	Package    PackageRef
	Provisions []Provision
	StateKinds []StateKind
	Launcher   LauncherInfo
}

type LauncherInfo struct {
	// Binary is the executable name installed to $out/bin/. For AppImage packages,
	// this must match the name passed to AppImageRef() (the versions.toml key).
	Binary      string
	DisplayName string   // Name for .desktop file (uses Emulator.Name if empty)
	GenericName string   // For .desktop generation (e.g., "PlayStation Emulator")
	Categories  []string // XDG categories (e.g., ["Game", "Emulator"])
}

// SupportsSystem checks if this emulator can run games for the given system.
func (e *Emulator) SupportsSystem(sys SystemID) bool {
	for _, s := range e.Systems {
		if s == sys {
			return true
		}
	}
	return false
}
