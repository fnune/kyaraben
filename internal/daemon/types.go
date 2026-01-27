package daemon

// Protocol types for daemon-UI communication.
//
// These types are the source of truth. When modifying this file, also update
// the TypeScript definitions in ui/electron/main.ts (DaemonCommand, DaemonEvent
// interfaces). Type generation from Go is planned but not yet implemented.

// CommandType identifies the type of command.
type CommandType string

const (
	CmdStatus           CommandType = "status"
	CmdDoctor           CommandType = "doctor"
	CmdApply            CommandType = "apply"
	CmdGetSystems       CommandType = "get_systems"
	CmdGetConfig        CommandType = "get_config"
	CmdSetConfig        CommandType = "set_config"
	CmdSyncStatus       CommandType = "sync_status"
	CmdSyncAddDevice    CommandType = "sync_add_device"
	CmdSyncRemoveDevice CommandType = "sync_remove_device"
)

// Command represents a command from the UI.
type Command struct {
	Type CommandType            `json:"type"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// EventType identifies the type of event.
type EventType string

const (
	EventReady    EventType = "ready"
	EventResult   EventType = "result"
	EventProgress EventType = "progress"
	EventError    EventType = "error"
)

// Event represents an event sent to the UI.
type Event struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data,omitempty"`
}
