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
	CommandTypeStatus                CommandType = "status"
	CommandTypeDoctor                CommandType = "doctor"
	CommandTypeApply                 CommandType = "apply"
	CommandTypeCancelApply           CommandType = "cancel_apply"
	CommandTypeGetSystems            CommandType = "get_systems"
	CommandTypeGetFrontends          CommandType = "get_frontends"
	CommandTypeGetConfig             CommandType = "get_config"
	CommandTypeSetConfig             CommandType = "set_config"
	CommandTypeSyncStatus            CommandType = "sync_status"
	CommandTypeSyncRemoveDevice      CommandType = "sync_remove_device"
	CommandTypeSyncStartPairing      CommandType = "sync_start_pairing"
	CommandTypeSyncJoinPeer          CommandType = "sync_join_peer"
	CommandTypeSyncCancelPairing     CommandType = "sync_cancel_pairing"
	CommandTypeSyncPending           CommandType = "sync_pending"
	CommandTypeUninstallPreview      CommandType = "uninstall_preview"
	CommandTypeUninstall             CommandType = "uninstall"
	CommandTypeInstallKyaraben       CommandType = "install_kyaraben"
	CommandTypeInstallStatus         CommandType = "install_status"
	CommandTypeRefreshIconCaches     CommandType = "refresh_icon_caches"
	CommandTypePreflight             CommandType = "preflight"
	CommandTypeSyncEnable            CommandType = "sync_enable"
	CommandTypeSyncRevertFolder      CommandType = "sync_revert_folder"
	CommandTypeSyncLocalChanges      CommandType = "sync_local_changes"
	CommandTypeSyncReset             CommandType = "sync_reset"
	CommandTypeSyncDiscoveredDevices CommandType = "sync_discovered_devices"
	CommandTypeSyncSetSettings       CommandType = "sync_set_settings"
	CommandTypeGetStorageDevices     CommandType = "get_storage_devices"
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

// SyncRemoveDeviceCommand includes the device to remove.
type SyncRemoveDeviceCommand struct {
	Type CommandType             `json:"type"`
	ID   string                  `json:"id,omitempty"`
	Data SyncRemoveDeviceRequest `json:"data"`
}

// SyncJoinPeerCommand includes the pairing code and peer device.
type SyncJoinPeerCommand struct {
	Type CommandType         `json:"type"`
	ID   string              `json:"id,omitempty"`
	Data SyncJoinPeerRequest `json:"data"`
}

// InstallKyarabenCommand includes the install options.
type InstallKyarabenCommand struct {
	Type CommandType            `json:"type"`
	ID   string                 `json:"id,omitempty"`
	Data InstallKyarabenRequest `json:"data"`
}

// SyncEnableCommand includes the sync enable options.
type SyncEnableCommand struct {
	Type CommandType       `json:"type"`
	ID   string            `json:"id,omitempty"`
	Data SyncEnableRequest `json:"data"`
}

// SyncRevertFolderCommand includes the folder to revert.
type SyncRevertFolderCommand struct {
	Type CommandType             `json:"type"`
	ID   string                  `json:"id,omitempty"`
	Data SyncRevertFolderRequest `json:"data"`
}

// SyncLocalChangesCommand includes the folder to get local changes for.
type SyncLocalChangesCommand struct {
	Type CommandType             `json:"type"`
	ID   string                  `json:"id,omitempty"`
	Data SyncLocalChangesRequest `json:"data"`
}

// SyncSetSettingsCommand includes the sync settings to update.
type SyncSetSettingsCommand struct {
	Type CommandType            `json:"type"`
	ID   string                 `json:"id,omitempty"`
	Data SyncSetSettingsRequest `json:"data"`
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
