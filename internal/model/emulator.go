package model

// EmulatorID uniquely identifies an emulator implementation.
// Format: "emulator" or "emulator:core" for RetroArch cores.
type EmulatorID string

const (
	EmulatorRetroArchBsnes   EmulatorID = "retroarch:bsnes"
	EmulatorRetroArchMGBA    EmulatorID = "retroarch:mgba"
	EmulatorRetroArchMelonDS EmulatorID = "retroarch:melonds"
	EmulatorRetroArchPPSSPP  EmulatorID = "retroarch:ppsspp"
	EmulatorDuckStation      EmulatorID = "duckstation"
	EmulatorTIC80            EmulatorID = "tic80"
	EmulatorEden             EmulatorID = "eden"
	EmulatorE2ETest          EmulatorID = "e2e-test"
)

// PackageSource indicates where an emulator package comes from.
type PackageSource string

const (
	PackageSourceNixpkgs   PackageSource = "nixpkgs"
	PackageSourceGitHub    PackageSource = "github"
	PackageSourceVersioned PackageSource = "versioned"
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
	Binary      string   // Binary name (e.g., "duckstation-qt", "retroarch")
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
