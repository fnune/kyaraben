package model

type StoreReader interface {
	RomsDir() string
	BiosDir() string
	SystemBiosDir(SystemID) string
	SystemSavesDir(SystemID) string
	EmulatorSavesDir(EmulatorID) string // Per-emulator saves for cores that need individual sync
	EmulatorStatesDir(EmulatorID) string
	SystemScreenshotsDir(SystemID) string
	SystemRomsDir(SystemID) string
	EmulatorOpaqueDir(EmulatorID) string
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
