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

	// RomArgs is the CLI argument pattern for launching a ROM/game file.
	// Use %ROM% as the placeholder for the game path.
	// Empty means the ROM path is a positional argument: <binary> <rom>.
	// Examples:
	//   ""                          → positional: <binary> <rom>
	//   "-e %ROM%"                  → Dolphin: dolphin -e <rom>
	//   "-g %ROM%"                  → Cemu/Eden: cemu -g <rom>
	//   "-r %ROM%"                  → Vita3K: vita3k -r <rom>
	//   "-L bsnes_libretro %ROM%"   → RetroArch: retroarch -L bsnes_libretro <rom>
	RomArgs string
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
