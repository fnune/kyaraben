package daemon

import (
	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
)

// Request types

type SetConfigRequest struct {
	UserStore string                         `json:"userStore"`
	Systems   map[string][]string            `json:"systems"`
	Emulators map[string]EmulatorConfRequest `json:"emulators,omitempty"`
	Frontends map[string]FrontendConfRequest `json:"frontends,omitempty"`
}

type FrontendConfRequest struct {
	Enabled bool   `json:"enabled"`
	Version string `json:"version,omitempty"`
}

type EmulatorConfRequest struct {
	Version string `json:"version,omitempty"`
}

type SyncRemoveDeviceRequest struct {
	DeviceID string `json:"deviceId"`
}

type SyncJoinPrimaryRequest struct {
	Code        string `json:"code"`
	PairingAddr string `json:"pairingAddr,omitempty"`
	DeviceID    string `json:"deviceId,omitempty"`
}

type SyncEnableRequest struct{}

// Response types

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatusResponse struct {
	UserStore               string              `json:"userStore"`
	EnabledSystems          []model.SystemID    `json:"enabledSystems"`
	InstalledEmulators      []InstalledEmulator `json:"installedEmulators"`
	InstalledFrontends      []InstalledFrontend `json:"installedFrontends"`
	Symlinks                []SymlinkInfo       `json:"symlinks,omitempty"`
	LastApplied             string              `json:"lastApplied"`
	HealthWarning           string              `json:"healthWarning,omitempty"`
	KyarabenVersion         string              `json:"kyarabenVersion"`
	ManifestKyarabenVersion string              `json:"manifestKyarabenVersion,omitempty"`
}

type InstalledEmulator struct {
	ID             model.EmulatorID         `json:"id"`
	Version        string                   `json:"version"`
	ExecLine       string                   `json:"execLine,omitempty"`
	ManagedConfigs []ManagedConfigInfo      `json:"managedConfigs,omitempty"`
	IconPath       string                   `json:"iconPath,omitempty"`
	Paths          map[string]EmulatorPaths `json:"paths,omitempty"`
}

type ManagedRegionInfo struct {
	Type      string `json:"type"`
	Section   string `json:"section,omitempty"`
	KeyPrefix string `json:"keyPrefix,omitempty"`
}

type ManagedConfigInfo struct {
	Path           string              `json:"path"`
	ManagedRegions []ManagedRegionInfo `json:"managedRegions,omitempty"`
}

type SymlinkInfo struct {
	Source     string           `json:"source"`
	Target     string           `json:"target"`
	EmulatorID model.EmulatorID `json:"emulatorId"`
}

type EmulatorPaths struct {
	Roms        string `json:"roms"`
	Bios        string `json:"bios,omitempty"`
	Saves       string `json:"saves,omitempty"`
	Savestates  string `json:"states,omitempty"`
	Screenshots string `json:"screenshots,omitempty"`
}

type InstalledFrontend struct {
	ID      model.FrontendID `json:"id"`
	Version string           `json:"version"`
}

// DoctorResponse maps "systemId:emulatorId" to provisions for that system/emulator pair.
// Uses map[string] because tygo can't generate valid TypeScript for union index types.
type DoctorResponse map[string][]ProvisionResult

type ProvisionResult struct {
	Filename            string `json:"filename"`
	Kind                string `json:"kind"`
	Description         string `json:"description"`
	Status              string `json:"status"`
	ExpectedPath        string `json:"expectedPath,omitempty"`
	FoundPath           string `json:"foundPath,omitempty"`
	ImportViaUI         bool   `json:"importViaUI,omitempty"`
	GroupMessage        string `json:"groupMessage,omitempty"`
	GroupRequired       bool   `json:"groupRequired"`
	GroupSatisfied      bool   `json:"groupSatisfied"`
	GroupSize           int    `json:"groupSize"`
	DisplayName         string `json:"displayName"`
	VerifiedDisplayName string `json:"verifiedDisplayName,omitempty"`
	Instructions        string `json:"instructions,omitempty"`
}

type ProgressEvent struct {
	Step            string            `json:"step"`
	Message         string            `json:"message"`
	Output          string            `json:"output,omitempty"`
	BuildPhase      string            `json:"buildPhase,omitempty"`
	PackageName     string            `json:"packageName,omitempty"`
	ProgressPercent int               `json:"progressPercent,omitempty"`
	BytesDownloaded int64             `json:"bytesDownloaded,omitempty"`
	BytesTotal      int64             `json:"bytesTotal,omitempty"`
	BytesPerSecond  int64             `json:"bytesPerSecond,omitempty"`
	LogPosition     int64             `json:"logPosition"`
	LogEntry        *logging.LogEntry `json:"logEntry,omitempty"`
}

type ApplyResult struct {
	Success bool `json:"success"`
}

type CancelledResponse struct {
	Message string `json:"message"`
}

type GetSystemsResponse []SystemWithEmulators

type SystemWithEmulators struct {
	ID                model.SystemID     `json:"id"`
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	Manufacturer      model.Manufacturer `json:"manufacturer"`
	Label             string             `json:"label"`
	DefaultEmulatorID model.EmulatorID   `json:"defaultEmulatorId"`
	Emulators         []EmulatorRef      `json:"emulators"`
}

type EmulatorRef struct {
	ID                model.EmulatorID `json:"id"`
	Name              string           `json:"name"`
	DefaultVersion    string           `json:"defaultVersion,omitempty"`
	AvailableVersions []string         `json:"availableVersions,omitempty"`
	DownloadBytes     int64            `json:"downloadBytes,omitempty"`
	CoreBytes         int64            `json:"coreBytes,omitempty"`
	PackageName       string           `json:"packageName,omitempty"`
}

type ConfigResponse struct {
	UserStore string                          `json:"userStore"`
	Systems   map[string][]model.EmulatorID   `json:"systems"`
	Emulators map[string]EmulatorConfResponse `json:"emulators,omitempty"`
	Frontends map[string]FrontendConfResponse `json:"frontends,omitempty"`
}

type EmulatorConfResponse struct {
	Version string `json:"version,omitempty"`
}

type FrontendConfResponse struct {
	Enabled bool   `json:"enabled"`
	Version string `json:"version,omitempty"`
}

type GetFrontendsResponse []FrontendRef

type FrontendRef struct {
	ID                model.FrontendID `json:"id"`
	Name              string           `json:"name"`
	DefaultVersion    string           `json:"defaultVersion,omitempty"`
	AvailableVersions []string         `json:"availableVersions,omitempty"`
	DownloadBytes     int64            `json:"downloadBytes,omitempty"`
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
	Enabled          bool          `json:"enabled"`
	Running          bool          `json:"running,omitempty"`
	Installed        bool          `json:"installed,omitempty"`
	ServiceInstalled bool          `json:"serviceInstalled,omitempty"`
	DeviceID         string        `json:"deviceId,omitempty"`
	GUIURL           string        `json:"guiURL,omitempty"`
	State            SyncState     `json:"state,omitempty"`
	Devices          []SyncDevice  `json:"devices,omitempty"`
	Folders          []SyncFolder  `json:"folders,omitempty"`
	Pairing          bool          `json:"pairing,omitempty"`
	Progress         *SyncProgress `json:"progress,omitempty"`
	ServiceError     string        `json:"serviceError,omitempty"`
}

type SyncProgress struct {
	NeedFiles   int64 `json:"needFiles"`
	NeedBytes   int64 `json:"needBytes"`
	GlobalBytes int64 `json:"globalBytes"`
	Percent     int   `json:"percent"`
}

type SyncDevice struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Connected  bool   `json:"connected"`
	Paused     bool   `json:"paused,omitempty"`
	Completion *int   `json:"completion,omitempty"`
}

type SyncFolder struct {
	ID                 string `json:"id"`
	Path               string `json:"path"`
	Label              string `json:"label"`
	State              string `json:"state"`
	Error              string `json:"error,omitempty"`
	Type               string `json:"type"`
	GlobalSize         int64  `json:"globalSize"`
	LocalSize          int64  `json:"localSize"`
	NeedSize           int64  `json:"needSize"`
	ReceiveOnlyChanges int    `json:"receiveOnlyChanges"`
}

type SyncRevertFolderRequest struct {
	FolderID string `json:"folderId"`
}

type SyncRevertFolderResponse struct {
	Success bool `json:"success"`
}

type SyncLocalChangesRequest struct {
	FolderID string `json:"folderId"`
}

type SyncLocalChange struct {
	Action   string `json:"action"`
	Type     string `json:"type"`
	Path     string `json:"path"`
	Modified string `json:"modified"`
	Size     int64  `json:"size"`
}

type SyncLocalChangesResponse struct {
	Changes []SyncLocalChange `json:"changes"`
}

type SyncPendingResponse struct {
	Pending    bool  `json:"pending"`
	TotalFiles int64 `json:"totalFiles"`
	TotalBytes int64 `json:"totalBytes"`
}

type SyncRemoveDeviceResponse struct {
	Success  bool   `json:"success"`
	DeviceID string `json:"deviceId"`
	Name     string `json:"name"`
}

type SyncStartPairingResponse struct {
	DeviceID string `json:"deviceId"`
	Code     string `json:"code,omitempty"`
}

type SyncPairingCompleteResponse struct {
	Success      bool   `json:"success"`
	PeerDeviceID string `json:"peerDeviceId"`
	PeerName     string `json:"peerName"`
}

type SyncJoinPrimaryResponse struct {
	Success      bool   `json:"success"`
	PeerDeviceID string `json:"peerDeviceId"`
	PeerName     string `json:"peerName"`
}

type SyncPairingProgressEvent struct {
	Message  string `json:"message"`
	DeviceID string `json:"deviceId,omitempty"`
	Code     string `json:"code,omitempty"`
}

type SyncEnableProgressEvent struct {
	Phase   string `json:"phase"`
	Message string `json:"message"`
	Percent int    `json:"percent"`
}

type SyncEnableResponse struct {
	Success bool `json:"success"`
}

type SyncResetResponse struct {
	Success      bool     `json:"success"`
	RemovedFiles []string `json:"removedFiles,omitempty"`
}

type SyncDiscoveredDevice struct {
	DeviceID  string   `json:"deviceId"`
	Addresses []string `json:"addresses"`
}

type SyncDiscoveredDevicesResponse struct {
	Devices []SyncDiscoveredDevice `json:"devices"`
}

type UninstallPreviewResponse struct {
	StateDir           string         `json:"stateDir"`
	StateDirExists     bool           `json:"stateDirExists"`
	RetroArchCoresDir  string         `json:"retroArchCoresDir,omitempty"`
	RetroArchCoreFiles []string       `json:"retroArchCoreFiles,omitempty"`
	DesktopFiles       []string       `json:"desktopFiles"`
	IconFiles          []string       `json:"iconFiles"`
	ConfigFiles        []string       `json:"configFiles"`
	KyarabenFiles      []string       `json:"kyarabenFiles"`
	SyncthingFiles     []string       `json:"syncthingFiles,omitempty"`
	Preserved          PreservedPaths `json:"preserved"`
}

type PreservedPaths struct {
	UserStore string `json:"userStore"`
	ConfigDir string `json:"configDir"`
}

type InstallKyarabenRequest struct {
	AppImagePath string `json:"appImagePath,omitempty"`
	SidecarPath  string `json:"sidecarPath,omitempty"`
}

type InstallKyarabenResponse struct {
	Success bool `json:"success"`
}

type InstallStatusResponse struct {
	Installed   bool   `json:"installed"`
	AppPath     string `json:"appPath,omitempty"`
	DesktopPath string `json:"desktopPath,omitempty"`
	CLIPath     string `json:"cliPath,omitempty"`
}

type RefreshIconCachesResponse struct {
	Refreshed []string `json:"refreshed"`
}

type UninstallResponse struct {
	Success      bool     `json:"success"`
	RemovedFiles []string `json:"removedFiles"`
	Errors       []string `json:"errors,omitempty"`
}

type PreflightResponse struct {
	Diffs         []ConfigFileDiff `json:"diffs"`
	FilesToBackup []string         `json:"filesToBackup"`
}

type ConfigFileDiff struct {
	Path            string                 `json:"path"`
	IsNewFile       bool                   `json:"isNewFile"`
	HasChanges      bool                   `json:"hasChanges"`
	UserModified    bool                   `json:"userModified"`
	KyarabenChanged bool                   `json:"kyarabenChanged"`
	UserChanges     []UserChangeDetail     `json:"userChanges,omitempty"`
	KyarabenUpdates []KyarabenUpdateDetail `json:"kyarabenUpdates,omitempty"`
	Changes         []ConfigChangeDetail   `json:"changes,omitempty"`
	ManagedRegions  []ManagedRegionInfo    `json:"managedRegions,omitempty"`
}

type KyarabenUpdateDetail struct {
	Key      string `json:"key"`
	OldValue string `json:"oldValue"`
	NewValue string `json:"newValue"`
}

type UserChangeDetail struct {
	Key           string `json:"key"`
	BaselineValue string `json:"baselineValue"`
	CurrentValue  string `json:"currentValue"`
}

type ConfigChangeDetail struct {
	Type     string `json:"type"`
	Key      string `json:"key"`
	Section  string `json:"section,omitempty"`
	OldValue string `json:"oldValue,omitempty"`
	NewValue string `json:"newValue,omitempty"`
}

type StorageDevice struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Path       string `json:"path"`
	FreeBytes  int64  `json:"freeBytes"`
	TotalBytes int64  `json:"totalBytes"`
}

type StorageDevicesResponse struct {
	Devices []StorageDevice `json:"devices"`
}
