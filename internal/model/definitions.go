package model

type StoreReader interface {
	BiosDir() string
	SystemBiosDir(SystemID) string
	SystemSavesDir(SystemID) string
	SystemStatesDir(SystemID) string
	SystemScreenshotsDir(SystemID) string
	SystemRomsDir(SystemID) string
}

type ConfigGenerator interface {
	Generate(store StoreReader, systems []SystemID) ([]ConfigPatch, error)
}

type SystemDefinition interface {
	System() System
	DefaultEmulatorID() EmulatorID
}

type EmulatorDefinition interface {
	Emulator() Emulator
	ConfigGenerator() ConfigGenerator
}
