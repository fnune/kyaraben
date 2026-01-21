package model

// EmulatorID uniquely identifies an emulator implementation.
// Format: "emulator" or "emulator:core" for RetroArch cores.
type EmulatorID string

const (
	EmulatorRetroArchBsnes EmulatorID = "retroarch:bsnes"
	EmulatorDuckStation    EmulatorID = "duckstation"
	EmulatorTIC80          EmulatorID = "tic80"
)

// PackageSource indicates where an emulator package comes from.
type PackageSource string

const (
	PackageSourceNixpkgs PackageSource = "nixpkgs"
	PackageSourceGitHub  PackageSource = "github"
)

// Emulator represents an implementation that runs a system's games.
type Emulator struct {
	ID          EmulatorID
	Name        string
	Systems     []SystemID
	Source      PackageSource
	NixAttr     string // Attribute path in nixpkgs or flake
	Provisions  []Provision
	StateKinds  []StateKind
	ConfigPaths []string // Paths to config files this emulator uses
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
