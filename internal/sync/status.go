package sync

import (
	"context"
	"fmt"

	"github.com/fnune/kyaraben/internal/model"
)

type Status struct {
	Enabled   bool
	Mode      model.SyncMode
	DeviceID  string
	GUIURL    string
	Devices   []DeviceStatus
	Folders   []FolderStatusSummary
	Conflicts []Conflict
}

type DeviceStatus struct {
	ID        string
	Name      string
	Connected bool
	Paused    bool
}

type FolderStatusSummary struct {
	ID                 string
	Path               string
	Type               string
	State              string
	Completion         float64
	GlobalSize         int64
	LocalSize          int64
	NeedSize           int64
	ReceiveOnlyChanges int
}

type Conflict struct {
	Path         string
	LocalModTime string
	Size         int64
}

func (c *Client) GetStatus(ctx context.Context) (*Status, error) {
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
		name := dev.Name
		if ok && conn.DeviceName != "" {
			name = conn.DeviceName + " (primary)"
		} else if c.config.Mode == model.SyncModeSecondary {
			name = "primary"
		}
		devices = append(devices, DeviceStatus{
			ID:        dev.ID,
			Name:      name,
			Connected: ok && conn.Connected,
			Paused:    dev.Paused || (ok && conn.Paused),
		})
	}

	var folders []FolderStatusSummary
	if folderConfigs, err := c.GetFolderConfigs(ctx); err == nil {
		for _, fc := range folderConfigs {
			fs, err := c.GetFolderStatus(ctx, fc.ID)
			if err != nil {
				continue
			}
			folders = append(folders, FolderStatusSummary{
				ID:                 fc.ID,
				Path:               fc.Path,
				Type:               fc.Type,
				State:              fs.State,
				GlobalSize:         fs.GlobalBytes,
				LocalSize:          fs.LocalBytes,
				NeedSize:           fs.NeedBytes,
				ReceiveOnlyChanges: fs.ReceiveOnlyTotalItems,
			})
		}
	}

	status := &Status{
		Enabled:  true,
		Mode:     c.config.Mode,
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

	if len(s.Conflicts) > 0 {
		return SyncStateConflict
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
