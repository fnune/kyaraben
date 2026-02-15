package sync

import (
	"context"
	gosync "sync"

	"github.com/fnune/kyaraben/internal/model"
)

type FakeClient struct {
	mu          gosync.Mutex
	running     bool
	deviceID    string
	connections map[string]ConnectionInfo
	addedPeers  []model.SyncDevice
	removedIDs  []string
	sharedWith  []string
	config      model.SyncConfig
	folders     map[string]FolderStatusSummary
}

func NewFakeClient(config model.SyncConfig) *FakeClient {
	return &FakeClient{
		running:     true,
		connections: make(map[string]ConnectionInfo),
		folders:     make(map[string]FolderStatusSummary),
		config:      config,
	}
}

func (c *FakeClient) SetDeviceID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deviceID = id
}

func (c *FakeClient) SetRunning(running bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.running = running
}

func (c *FakeClient) SetConnection(deviceID string, info ConnectionInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connections[deviceID] = info
}

func (c *FakeClient) AddedPeers() []model.SyncDevice {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]model.SyncDevice, len(c.addedPeers))
	copy(result, c.addedPeers)
	return result
}

func (c *FakeClient) RemovedIDs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]string, len(c.removedIDs))
	copy(result, c.removedIDs)
	return result
}

func (c *FakeClient) SharedWith() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]string, len(c.sharedWith))
	copy(result, c.sharedWith)
	return result
}

func (c *FakeClient) SetFolderStatus(id string, status FolderStatusSummary) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.folders[id] = status
}

func (c *FakeClient) SetFolders(folders []FolderStatusSummary) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.folders = make(map[string]FolderStatusSummary)
	for _, f := range folders {
		c.folders[f.ID] = f
	}
}

func (c *FakeClient) IsRunning(_ context.Context) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

func (c *FakeClient) GetDeviceID(_ context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.deviceID, nil
}

func (c *FakeClient) GetConnections(_ context.Context) (map[string]ConnectionInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make(map[string]ConnectionInfo)
	for k, v := range c.connections {
		result[k] = v
	}
	return result, nil
}

func (c *FakeClient) GetConfiguredDevices(_ context.Context) ([]ConfiguredDevice, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var result []ConfiguredDevice
	for _, dev := range c.config.Devices {
		result = append(result, ConfiguredDevice{
			ID:   dev.ID,
			Name: dev.Name,
		})
	}
	for _, dev := range c.addedPeers {
		result = append(result, ConfiguredDevice{
			ID:   dev.ID,
			Name: dev.Name,
		})
	}
	return result, nil
}

func (c *FakeClient) GetFolderStatus(_ context.Context, folderID string) (*FolderStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if summary, ok := c.folders[folderID]; ok {
		return &FolderStatus{
			State:                 summary.State,
			GlobalBytes:           summary.GlobalSize,
			LocalBytes:            summary.LocalSize,
			NeedBytes:             summary.NeedSize,
			ReceiveOnlyTotalItems: summary.ReceiveOnlyChanges,
		}, nil
	}
	return &FolderStatus{State: "idle"}, nil
}

func (c *FakeClient) RevertFolder(_ context.Context, _ string) error {
	return nil
}

func (c *FakeClient) GetLocalChanges(_ context.Context, _ string) ([]LocalChange, error) {
	return nil, nil
}

func (c *FakeClient) GetFolderConfigs(_ context.Context) ([]FolderConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var configs []FolderConfig
	for _, f := range c.folders {
		configs = append(configs, FolderConfig{
			ID:   f.ID,
			Path: f.Path,
			Type: f.Type,
		})
	}
	return configs, nil
}

func (c *FakeClient) GetPendingFolders(_ context.Context) ([]PendingFolder, error) {
	return nil, nil
}

func (c *FakeClient) DismissPendingFolder(_ context.Context, _, _ string) error {
	return nil
}

func (c *FakeClient) GetStatus(_ context.Context) (*Status, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	allDevices := make([]model.SyncDevice, 0, len(c.config.Devices)+len(c.addedPeers))
	allDevices = append(allDevices, c.config.Devices...)
	allDevices = append(allDevices, c.addedPeers...)

	var devices []DeviceStatus
	for _, dev := range allDevices {
		conn, ok := c.connections[dev.ID]
		devices = append(devices, DeviceStatus{
			ID:        dev.ID,
			Name:      dev.Name,
			Connected: ok && conn.Connected,
			Paused:    ok && conn.Paused,
		})
	}

	var folders []FolderStatusSummary
	for _, f := range c.folders {
		folders = append(folders, f)
	}

	return &Status{
		Enabled:  c.config.Enabled,
		Mode:     c.config.Mode,
		DeviceID: c.deviceID,
		Devices:  devices,
		Folders:  folders,
	}, nil
}

func (c *FakeClient) AddDevice(_ context.Context, deviceID, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.addedPeers = append(c.addedPeers, model.SyncDevice{ID: deviceID, Name: name})
	return nil
}

func (c *FakeClient) RemoveDevice(_ context.Context, deviceID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.removedIDs = append(c.removedIDs, deviceID)
	return nil
}

func (c *FakeClient) ShareFoldersWithDevice(_ context.Context, deviceID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sharedWith = append(c.sharedWith, deviceID)
	return nil
}

var _ SyncClient = (*FakeClient)(nil)
