package daemon

// CommandType identifies the type of command.
type CommandType string

const (
	CmdStatus     CommandType = "status"
	CmdDoctor     CommandType = "doctor"
	CmdApply      CommandType = "apply"
	CmdGetSystems CommandType = "get_systems"
	CmdGetConfig  CommandType = "get_config"
	CmdSetConfig  CommandType = "set_config"
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
