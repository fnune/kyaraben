package model

// StateKind categorizes what type of state this is.
type StateKind string

const (
	StateSaves       StateKind = "saves"       // Game progress (memory cards, battery saves)
	StateSavestates  StateKind = "savestates"  // Emulator snapshots
	StateScreenshots StateKind = "screenshots" // Captured images
	StateCache       StateKind = "cache"       // Shader cache, regenerable data
	StatePersistent  StateKind = "persistent"  // Emulator-managed storage
)

// SyncStrategy determines how state is synchronized across devices.
type SyncStrategy string

const (
	SyncBidirectional SyncStrategy = "bidirectional" // Sync both ways
	SyncSendOnly      SyncStrategy = "send-only"     // Push only
	SyncIgnore        SyncStrategy = "ignore"        // Don't sync
)

// State represents data an emulator produces during operation.
type State struct {
	Kind     StateKind
	Path     string       // Relative path within Collection
	Sync     SyncStrategy // How this state should be synced
	SystemID SystemID     // Which system this state belongs to
}
