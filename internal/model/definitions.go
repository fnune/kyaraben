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
	CoresDir() string
}

// GenerateContext provides all dependencies a config generator needs.
// Emulators that do not need controller config ignore the nil ControllerConfig.
type GenerateContext struct {
	Store            StoreReader
	BaseDirResolver  BaseDirResolver
	ControllerConfig *ControllerConfig
}

// GenerateResult consolidates all outputs from a config generator:
// config patches, symlinks, and launch args.
type GenerateResult struct {
	Patches    []ConfigPatch
	Symlinks   []SymlinkSpec
	LaunchArgs []string
}

type ConfigGenerator interface {
	Generate(ctx GenerateContext) (GenerateResult, error)
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

type SymlinkCreator interface {
	Create(spec SymlinkSpec) error
}
