package daemon

// Protocol types for daemon-UI communication.
//
// Run `just generate-types` to regenerate ui/src/types/daemon.gen.ts after
// modifying this file or protocol.go.

// CommandType identifies the type of command.
//
// Constants use the full type name as prefix (CommandType*) because tygo's
// enum_style: union requires the prefix to match the type name exactly.
type CommandType string

const (
	CommandTypeStatus           CommandType = "status"
	CommandTypeDoctor           CommandType = "doctor"
	CommandTypeApply            CommandType = "apply"
	CommandTypeCancelApply      CommandType = "cancel_apply"
	CommandTypeGetSystems       CommandType = "get_systems"
	CommandTypeGetConfig        CommandType = "get_config"
	CommandTypeSetConfig        CommandType = "set_config"
	CommandTypeSyncStatus       CommandType = "sync_status"
	CommandTypeSyncAddDevice    CommandType = "sync_add_device"
	CommandTypeSyncRemoveDevice CommandType = "sync_remove_device"
	CommandTypeUninstallPreview CommandType = "uninstall_preview"
	CommandTypeInstallKyaraben  CommandType = "install_kyaraben"
	CommandTypeInstallStatus    CommandType = "install_status"
)

// Command represents a command from the UI.
type Command struct {
	Type CommandType `json:"type"`
	ID   string      `json:"id,omitempty"`
}

// SetConfigCommand includes the config data to set.
type SetConfigCommand struct {
	Type CommandType      `json:"type"`
	ID   string           `json:"id,omitempty"`
	Data SetConfigRequest `json:"data"`
}

// SyncAddDeviceCommand includes the device to add.
type SyncAddDeviceCommand struct {
	Type CommandType          `json:"type"`
	ID   string               `json:"id,omitempty"`
	Data SyncAddDeviceRequest `json:"data"`
}

// SyncRemoveDeviceCommand includes the device to remove.
type SyncRemoveDeviceCommand struct {
	Type CommandType             `json:"type"`
	ID   string                  `json:"id,omitempty"`
	Data SyncRemoveDeviceRequest `json:"data"`
}

// InstallKyarabenCommand includes the install options.
type InstallKyarabenCommand struct {
	Type CommandType            `json:"type"`
	ID   string                 `json:"id,omitempty"`
	Data InstallKyarabenRequest `json:"data"`
}

// EventType identifies the type of event.
//
// Constants use the full type name as prefix (EventType*) because tygo's
// enum_style: union requires the prefix to match the type name exactly.
type EventType string

const (
	EventTypeReady     EventType = "ready"
	EventTypeResult    EventType = "result"
	EventTypeProgress  EventType = "progress"
	EventTypeError     EventType = "error"
	EventTypeCancelled EventType = "cancelled"
)

// Event represents an event sent to the UI.
// The Data field contains a typed response struct depending on the command.
type Event struct {
	Type EventType   `json:"type"`
	ID   string      `json:"id,omitempty"`
	Data interface{} `json:"data,omitempty" tstype:"unknown"`
}
