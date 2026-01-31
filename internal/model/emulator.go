package model

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
