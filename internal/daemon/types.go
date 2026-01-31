package daemon

// Protocol types for daemon-UI communication.
//
// These types are the source of truth. Run `just generate-types` to regenerate
// ui/src/types/generated.ts after modifying this file.

// CommandType identifies the type of command.
type CommandType string

const (
	CmdStatus           CommandType = "status"
	CmdDoctor           CommandType = "doctor"
	CmdApply            CommandType = "apply"
	CmdCancelApply      CommandType = "cancel_apply"
	CmdGetSystems       CommandType = "get_systems"
	CmdGetConfig        CommandType = "get_config"
	CmdSetConfig        CommandType = "set_config"
	CmdSyncStatus       CommandType = "sync_status"
	CmdSyncAddDevice    CommandType = "sync_add_device"
	CmdSyncRemoveDevice CommandType = "sync_remove_device"
	CmdUninstallPreview CommandType = "uninstall_preview"
)

// Command represents a command from the UI.
type Command struct {
	Type CommandType            `json:"type"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// EventType identifies the type of event.
type EventType string

const (
	EventReady     EventType = "ready"
	EventResult    EventType = "result"
	EventProgress  EventType = "progress"
	EventError     EventType = "error"
	EventCancelled EventType = "cancelled"
)

// Event represents an event sent to the UI.
type Event struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data,omitempty"`
}
