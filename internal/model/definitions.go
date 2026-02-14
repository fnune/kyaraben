package model

type StoreReader interface {
	RomsDir() string
	BiosDir() string
	SystemBiosDir(SystemID) string
	SystemSavesDir(SystemID) string
	EmulatorSavesDir(EmulatorID) string // Per-emulator saves for cores that need individual sync
	EmulatorStatesDir(EmulatorID) string
	EmulatorScreenshotsDir(EmulatorID) string
	SystemRomsDir(SystemID) string
}

type ConfigGenerator interface {
	Generate(store StoreReader) ([]ConfigPatch, error)
}

// LaunchArgsProvider is an optional interface that config generators can implement
// to provide command-line arguments for the emulator's .desktop file Exec line.
// This allows emulators to use CLI flags (like Dolphin's -u) to set the user directory
// instead of or in addition to config file manipulation.
type LaunchArgsProvider interface {
	LaunchArgs(store StoreReader) []string
}

type SystemDefinition interface {
	System() System
	DefaultEmulatorID() EmulatorID
}

type EmulatorDefinition interface {
	Emulator() Emulator
	ConfigGenerator() ConfigGenerator
}

type SymlinkSpec struct {
	Source string // Where the emulator expects the directory (e.g., ~/.local/share/dolphin-emu/GC)
	Target string // Where kyaraben stores data (e.g., ~/Emulation/saves/gamecube)
}

type SymlinkProvider interface {
	Symlinks(store StoreReader, resolver BaseDirResolver) ([]SymlinkSpec, error)
}

type SymlinkCreator interface {
	Create(spec SymlinkSpec) error
}
