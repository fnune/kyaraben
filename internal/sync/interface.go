package sync

import (
	"context"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/syncthing"
)

type ConnectionInfo = syncthing.ConnectionInfo
type FolderStatus = syncthing.FolderStatus
type FolderConfig = syncthing.FolderConfig
type ConfiguredDevice = syncthing.ConfiguredDevice
type DiscoveredDevice = syncthing.DiscoveredDevice
type PendingDevice = syncthing.PendingDevice
type PendingFolder = syncthing.PendingFolder
type CompletionResponse = syncthing.CompletionResponse
type LocalChange = syncthing.LocalChange
type FolderSharingDrift = syncthing.FolderSharingDrift
type SystemStatus = syncthing.SystemStatus
type SyncProgressInfo = syncthing.SyncProgressInfo
type PendingStatus = syncthing.PendingStatus
type DeviceStats = syncthing.DeviceStats

type SyncClient interface {
	syncthing.SyncClient
	GetStatus(ctx context.Context, fs vfs.FS) (*Status, error)
}

var _ SyncClient = (*Client)(nil)
