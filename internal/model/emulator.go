package model

// PackageSource indicates where an emulator package comes from.
type PackageSource string

const (
	PackageSourceAppImage PackageSource = "appimage"
)

type PathUsage struct {
	UsesBiosDir        bool
	UsesSavesDir       bool
	UsesStatesDir      bool
	UsesScreenshotsDir bool
	OpaqueContents     string
}

func StandardPathUsage() PathUsage {
	return PathUsage{
		UsesBiosDir:        true,
		UsesSavesDir:       true,
		UsesStatesDir:      true,
		UsesScreenshotsDir: true,
	}
}

type Emulator struct {
	ID              EmulatorID
	Name            string
	Systems         []SystemID
	Package         PackageRef
	ProvisionGroups []ProvisionGroup
	StateKinds      []StateKind
	Launcher        LauncherInfo
	PathUsage       PathUsage
}

type LauncherInfo struct {
	// Binary must match the name passed to AppImageRef() for AppImage packages.
	Binary      string
	DisplayName string
	GenericName string
	Categories  []string

	// RomCommand builds the CLI command for launching a game file.
	// The returned string uses %ROM% as the placeholder for the game path.
	RomCommand func(opts RomLaunchOptions) string
}

type RomLaunchOptions struct {
	BinaryPath string
	Fullscreen bool
	SavesDir   string
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
