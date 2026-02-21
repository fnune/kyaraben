package sync

import "context"

type SyncClient interface {
	IsRunning(ctx context.Context) bool
	GetDeviceID(ctx context.Context) (string, error)
	GetConnections(ctx context.Context) (map[string]ConnectionInfo, error)
	GetConfiguredDevices(ctx context.Context) ([]ConfiguredDevice, error)
	GetDiscoveredDevices(ctx context.Context) ([]DiscoveredDevice, error)
	GetPendingDevices(ctx context.Context) ([]PendingDevice, error)
	GetFolderStatus(ctx context.Context, folderID string) (*FolderStatus, error)
	GetFolderConfigs(ctx context.Context) ([]FolderConfig, error)
	GetStatus(ctx context.Context) (*Status, error)
	AddDevice(ctx context.Context, deviceID, name string) error
	AddDeviceAutoName(ctx context.Context, deviceID string) error
	RemoveDevice(ctx context.Context, deviceID string) error
	ShareFoldersWithDevice(ctx context.Context, deviceID string) error
	RevertFolder(ctx context.Context, folderID string) error
	GetLocalChanges(ctx context.Context, folderID string) ([]LocalChange, error)
	GetPendingFolders(ctx context.Context) ([]PendingFolder, error)
	DismissPendingFolder(ctx context.Context, folderID, deviceID string) error
}

var _ SyncClient = (*Client)(nil)
