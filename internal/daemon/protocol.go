package daemon

import "github.com/fnune/kyaraben/internal/model"

// Request types

type SetConfigRequest struct {
	UserStore string            `json:"userStore"`
	Systems   map[string]string `json:"systems"`
}

type SyncAddDeviceRequest struct {
	DeviceID string `json:"deviceId"`
	Name     string `json:"name,omitempty"`
}

type SyncRemoveDeviceRequest struct {
	DeviceID string `json:"deviceId"`
}

// Response types

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatusResponse struct {
	UserStore          string              `json:"userStore"`
	EnabledSystems     []model.SystemID    `json:"enabledSystems"`
	InstalledEmulators []InstalledEmulator `json:"installedEmulators"`
	LastApplied        string              `json:"lastApplied"`
}

type InstalledEmulator struct {
	ID      model.EmulatorID `json:"id"`
	Version string           `json:"version"`
}

// DoctorResponse uses map[string] because tygo can't generate valid TypeScript
// for map[SystemID] (index signatures don't support union types).
// The TypeScript type is manually defined as Record<SystemID, ProvisionResult[]>.
type DoctorResponse map[string][]ProvisionResult

type ProvisionResult struct {
	Filename    string `json:"filename"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Status      string `json:"status"`
	FoundPath   string `json:"foundPath,omitempty"`
}

type ProgressEvent struct {
	Step    string `json:"step"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
	Speed   string `json:"speed,omitempty"`
}

type ApplyResult struct {
	Success   bool   `json:"success"`
	StorePath string `json:"storePath"`
}

type CancelledResponse struct {
	Message string `json:"message"`
}

type GetSystemsResponse []SystemWithEmulators

type SystemWithEmulators struct {
	ID           model.SystemID     `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Manufacturer model.Manufacturer `json:"manufacturer"`
	Label        string             `json:"label"`
	Emulators    []EmulatorRef      `json:"emulators"`
}

type EmulatorRef struct {
	ID                model.EmulatorID `json:"id"`
	Name              string           `json:"name"`
	DefaultVersion    string           `json:"defaultVersion,omitempty"`
	AvailableVersions []string         `json:"availableVersions,omitempty"`
}

type ConfigResponse struct {
	UserStore string                `json:"userStore"`
	Systems   map[string]SystemConf `json:"systems"`
}

type SystemConf struct {
	Emulator      model.EmulatorID `json:"emulator"`
	PinnedVersion string           `json:"pinnedVersion,omitempty"`
}

type SetConfigResponse struct {
	Success bool `json:"success"`
}

// SyncState represents the current state of sync.
// Constants use the full type name as prefix (SyncState*) because tygo's
// enum_style: union requires the prefix to match the type name exactly.
type SyncState string

const (
	SyncStateDisabled     SyncState = "disabled"
	SyncStateSynced       SyncState = "synced"
	SyncStateSyncing      SyncState = "syncing"
	SyncStateDisconnected SyncState = "disconnected"
	SyncStateConflict     SyncState = "conflict"
	SyncStateError        SyncState = "error"
)

type SyncStatusResponse struct {
	Enabled  bool         `json:"enabled"`
	Mode     string       `json:"mode,omitempty"`
	Running  bool         `json:"running,omitempty"`
	DeviceID string       `json:"deviceId,omitempty"`
	GUIURL   string       `json:"guiURL,omitempty"`
	State    SyncState    `json:"state,omitempty"`
	Devices  []SyncDevice `json:"devices,omitempty"`
}

type SyncDevice struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
}

type SyncAddDeviceResponse struct {
	Success  bool   `json:"success"`
	DeviceID string `json:"deviceId"`
	Name     string `json:"name"`
}

type SyncRemoveDeviceResponse struct {
	Success  bool   `json:"success"`
	DeviceID string `json:"deviceId"`
	Name     string `json:"name"`
}

type UninstallPreviewResponse struct {
	StateDir       string         `json:"stateDir"`
	StateDirExists bool           `json:"stateDirExists"`
	DesktopFiles   []string       `json:"desktopFiles"`
	IconFiles      []string       `json:"iconFiles"`
	ConfigFiles    []string       `json:"configFiles"`
	Preserved      PreservedPaths `json:"preserved"`
}

type PreservedPaths struct {
	UserStore string `json:"userStore"`
	ConfigDir string `json:"configDir"`
}
