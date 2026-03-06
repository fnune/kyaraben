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
	ID                 EmulatorID
	Name               string
	Systems            []SystemID
	Package            PackageRef
	ProvisionGroups    []ProvisionGroup
	StateKinds         []StateKind
	Launcher           LauncherInfo
	PathUsage          PathUsage
	SupportedSettings  []string
	SupportedHotkeys   []HotkeyID
	ShadersRecommended bool
	ResumeRecommended  bool
}

const SettingPreset = "preset"
const SettingResumeAutosave = "resume:autosave"
const SettingResumeAutoload = "resume:autoload"

type LauncherInfo struct {
	// Binary must match the name passed to AppImageRef() for AppImage packages.
	Binary      string
	DisplayName string
	GenericName string
	Categories  []string
	Keywords    []string

	// Env specifies environment variables to set when launching the emulator.
	Env map[string]string

	// CoreName is the libretro core filename (without .so extension) for RetroArch cores.
	// When set, the daemon builds the -L argument with the full path to the core.
	CoreName string

	// RomCommand builds the CLI command for launching a game file.
	// The returned string uses %ROM% as the placeholder for the game path.
	RomCommand func(opts RomLaunchOptions) string
}

type RomLaunchOptions struct {
	BinaryPath string
	Fullscreen bool
	SavesDir   string
	LaunchArgs []string
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
