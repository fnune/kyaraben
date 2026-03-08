package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/twpayne/go-vfs/v5"
)

type Status struct {
	Enabled  bool
	DeviceID string
	GUIURL   string
	Devices  []DeviceStatus
	Folders  []FolderStatusSummary
}

type DeviceStatus struct {
	ID        string
	Name      string
	Connected bool
	Paused    bool
}

type FolderStatusSummary struct {
	ID                 string
	Label              string
	Path               string
	Type               string
	State              string
	Error              string
	Completion         float64
	GlobalSize         int64
	LocalSize          int64
	NeedSize           int64
	ReceiveOnlyChanges int
	ConflictCount      int
}

func FolderLabel(id string) string {
	id = strings.TrimPrefix(id, "kyaraben-")

	if strings.HasPrefix(id, "frontends-esde-") {
		rest := strings.TrimPrefix(id, "frontends-esde-")
		parts := strings.SplitN(rest, "-", 2)
		if len(parts) == 2 {
			return fmt.Sprintf("%s (ES-DE %s)", parts[1], parts[0])
		}
	}

	parts := strings.SplitN(id, "-", 2)
	if len(parts) == 2 {
		return fmt.Sprintf("%s (%s)", parts[1], parts[0])
	}
	return id
}

func (c *Client) GetStatus(ctx context.Context, fsys vfs.FS) (*Status, error) {
	if !c.config.Enabled {
		return &Status{Enabled: false}, nil
	}

	deviceID, err := c.GetDeviceID(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting device ID: %w", err)
	}

	connections, err := c.GetConnections(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting connections: %w", err)
	}

	configuredDevices, err := c.GetConfiguredDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting configured devices: %w", err)
	}

	var devices []DeviceStatus
	for _, dev := range configuredDevices {
		conn, ok := connections[dev.ID]
		devices = append(devices, DeviceStatus{
			ID:        dev.ID,
			Name:      dev.Name,
			Connected: ok && conn.Connected,
			Paused:    dev.Paused || (ok && conn.Paused),
		})
	}

	var folders []FolderStatusSummary
	if folderConfigs, err := c.GetFolderConfigs(ctx); err == nil {
		for _, fc := range folderConfigs {
			folderStatus, err := c.GetFolderStatus(ctx, fc.ID)
			if err != nil {
				continue
			}
			var conflictCount int
			if fsys != nil {
				if conflicts, err := ScanForConflicts(fsys, fc.Path); err == nil {
					conflictCount = len(conflicts)
				}
			}
			folders = append(folders, FolderStatusSummary{
				ID:                 fc.ID,
				Label:              FolderLabel(fc.ID),
				Path:               fc.Path,
				Type:               fc.Type,
				State:              folderStatus.State,
				Error:              folderStatus.Error,
				GlobalSize:         folderStatus.GlobalBytes,
				LocalSize:          folderStatus.LocalBytes,
				NeedSize:           folderStatus.NeedBytes,
				ReceiveOnlyChanges: folderStatus.ReceiveOnlyTotalItems,
				ConflictCount:      conflictCount,
			})
		}
	}

	status := &Status{
		Enabled:  true,
		DeviceID: deviceID,
		GUIURL:   fmt.Sprintf("http://127.0.0.1:%d", c.config.Syncthing.GUIPort),
		Devices:  devices,
		Folders:  folders,
	}

	return status, nil
}

type OverallSyncState string

const (
	SyncStateDisabled     OverallSyncState = "disabled"
	SyncStateSynced       OverallSyncState = "synced"
	SyncStateSyncing      OverallSyncState = "syncing"
	SyncStateDisconnected OverallSyncState = "disconnected"
	SyncStateConflict     OverallSyncState = "conflict"
	SyncStateError        OverallSyncState = "error"
)

func (s *Status) OverallState() OverallSyncState {
	if !s.Enabled {
		return SyncStateDisabled
	}

	for _, f := range s.Folders {
		if f.ConflictCount > 0 {
			return SyncStateConflict
		}
	}

	if len(s.Devices) == 0 {
		return SyncStateDisconnected
	}

	connectedCount := 0
	for _, d := range s.Devices {
		if d.Connected {
			connectedCount++
		}
	}

	if connectedCount == 0 {
		return SyncStateDisconnected
	}

	for _, f := range s.Folders {
		if f.State == "syncing" || f.NeedSize > 0 {
			return SyncStateSyncing
		}
		if f.State == "error" {
			return SyncStateError
		}
	}

	return SyncStateSynced
}
