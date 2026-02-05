package model

type FrontendDefinition interface {
	Frontend() Frontend
	ConfigGenerator() FrontendConfigGenerator
}

type FrontendConfigGenerator interface {
	Generate(ctx FrontendContext) ([]ConfigPatch, error)
}

type FrontendContext struct {
	EnabledSystems  []SystemID
	SystemEmulators map[SystemID][]EmulatorID
	GetSystem       func(SystemID) (System, error)
	GetEmulator     func(EmulatorID) (Emulator, error)
	Store           StoreReader
	BinDir          string
}
