package sync

import "context"

type SyncClient interface {
	IsRunning(ctx context.Context) bool
	GetDeviceID(ctx context.Context) (string, error)
	GetConnections(ctx context.Context) (map[string]ConnectionInfo, error)
	GetFolderStatus(ctx context.Context, folderID string) (*FolderStatus, error)
	GetStatus(ctx context.Context) (*Status, error)
	AddDevice(ctx context.Context, deviceID, name string) error
	RemoveDevice(ctx context.Context, deviceID string) error
	ShareFoldersWithDevice(ctx context.Context, deviceID string) error
	PauseSync(ctx context.Context) error
	ResumeSync(ctx context.Context) error
}

var _ SyncClient = (*Client)(nil)
