package syncthing

import "context"

type SyncClient interface {
	IsRunning(ctx context.Context) bool
	GetSystemStatus(ctx context.Context) (*SystemStatus, error)
	GetDeviceID(ctx context.Context) (string, error)
	GetConnections(ctx context.Context) (map[string]ConnectionInfo, error)
	GetConfiguredDevices(ctx context.Context) ([]ConfiguredDevice, error)
	GetDiscoveredDevices(ctx context.Context) ([]DiscoveredDevice, error)
	GetPendingDevices(ctx context.Context) ([]PendingDevice, error)
	GetFolderStatus(ctx context.Context, folderID string) (*FolderStatus, error)
	GetFolderConfigs(ctx context.Context) ([]FolderConfig, error)
	GetFoldersWithDevices(ctx context.Context) ([]FolderConfig, error)
	AddDevice(ctx context.Context, deviceID, name string) error
	AddDeviceWithAddresses(ctx context.Context, deviceID, name string, addresses []string) error
	AddDeviceAutoName(ctx context.Context, deviceID string) error
	RemoveDevice(ctx context.Context, deviceID string) error
	ShareFoldersWithDevice(ctx context.Context, deviceID string) error
	ReconcileFolderSharing(ctx context.Context, drift []FolderSharingDrift) error
	RevertFolder(ctx context.Context, folderID string) error
	GetLocalChanges(ctx context.Context, folderID string) ([]LocalChange, error)
	GetPendingFolders(ctx context.Context) ([]PendingFolder, error)
	DismissPendingFolder(ctx context.Context, folderID, deviceID string) error
	GetDeviceCompletion(ctx context.Context, deviceID string) (*CompletionResponse, error)
	GetFolderCompletionForDevice(ctx context.Context, folderID, deviceID string) (*CompletionResponse, error)
	GetSyncProgress(ctx context.Context) (*SyncProgressInfo, error)
	GetPendingStatus(ctx context.Context) (*PendingStatus, error)
	AddFolders(ctx context.Context, folders []FolderCreateRequest) error
	Restart(ctx context.Context) error
	SetAPIKey(key string)
	Config() Config
	DisableUsageReporting(ctx context.Context) error
	AllowInsecureAdmin(ctx context.Context) error
}

var _ SyncClient = (*Client)(nil)
