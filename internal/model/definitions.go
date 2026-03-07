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
	FrontendGamelistDir(FrontendID, SystemID) string
	FrontendMediaDir(FrontendID, SystemID) string
	FrontendMediaBaseDir(FrontendID) string
}

// GenerateContext provides all dependencies a config generator needs.
// Emulators that do not need controller config ignore the nil ControllerConfig.
type GenerateContext struct {
	Store              StoreReader
	BaseDirResolver    BaseDirResolver
	ControllerConfig   *ControllerConfig
	Preset             string
	Bezels             bool
	TargetDevice       string
	Resume             string
	SystemDisplayTypes map[SystemID]DisplayType
}

// GenerateResult consolidates all outputs from a config generator:
// config patches, symlinks, and launch args.
type GenerateResult struct {
	Patches          []ConfigPatch
	Symlinks         []SymlinkSpec
	LaunchArgs       []string
	InitialDownloads []InitialDownload
	EmbeddedFiles    []EmbeddedFile
}

// InitialDownload describes a file that should be downloaded once if missing.
// Unlike provisions, these are user data files (e.g., HDD images) that get
// modified during use and should never be re-downloaded or validated.
// For archives, set ArchiveType and ExtractDir to extract contents.
type InitialDownload struct {
	URL         string
	SHA256      string
	DestPath    string
	ArchiveType string // Optional: "tar.gz", "zip", etc. If set, extracts to ExtractDir.
	ExtractDir  string // Directory to extract archive contents to. Required if ArchiveType is set.
	StripPrefix string // Optional: strip this prefix from paths when extracting (e.g., "koko-aio-slang-NG-1.9.85/").
}

// EmbeddedFile describes a file with content embedded in the binary.
// Used for assets like shaders that we bundle rather than download.
type EmbeddedFile struct {
	Content  []byte
	DestPath string
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
