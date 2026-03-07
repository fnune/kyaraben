package sync

import (
	"context"
	gosync "sync"

	"github.com/fnune/kyaraben/internal/model"
)

type FakeClient struct {
	mu                gosync.Mutex
	running           bool
	deviceID          string
	connections       map[string]ConnectionInfo
	addedPeers        []ConfiguredDevice
	removedIDs        []string
	sharedWith        []string
	config            model.SyncConfig
	folders           map[string]FolderStatusSummary
	folderDevices     map[string][]string
	discoveredDevs    []DiscoveredDevice
	pendingDevs       []PendingDevice
	deviceCompletions map[string]CompletionResponse
	reconciledDrift   []FolderSharingDrift
}

func NewFakeClient(config model.SyncConfig) *FakeClient {
	return &FakeClient{
		running:           true,
		connections:       make(map[string]ConnectionInfo),
		folders:           make(map[string]FolderStatusSummary),
		folderDevices:     make(map[string][]string),
		deviceCompletions: make(map[string]CompletionResponse),
		config:            config,
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

func (c *FakeClient) AddedPeers() []ConfiguredDevice {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]ConfiguredDevice, len(c.addedPeers))
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

func (c *FakeClient) GetSystemStatus(_ context.Context) (*SystemStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &SystemStatus{MyID: c.deviceID, Uptime: 60}, nil
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
	result := make([]ConfiguredDevice, len(c.addedPeers))
	copy(result, c.addedPeers)
	return result, nil
}

func (c *FakeClient) SetConfiguredDevice(id, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, dev := range c.addedPeers {
		if dev.ID == id {
			c.addedPeers[i].Name = name
			return
		}
	}
	c.addedPeers = append(c.addedPeers, ConfiguredDevice{ID: id, Name: name})
}

func (c *FakeClient) RemoveConfiguredDevice(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, dev := range c.addedPeers {
		if dev.ID == id {
			c.addedPeers = append(c.addedPeers[:i], c.addedPeers[i+1:]...)
			return
		}
	}
}

func (c *FakeClient) SetDiscoveredDevices(devs []DiscoveredDevice) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.discoveredDevs = devs
}

func (c *FakeClient) GetDiscoveredDevices(_ context.Context) ([]DiscoveredDevice, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]DiscoveredDevice, len(c.discoveredDevs))
	copy(result, c.discoveredDevs)
	return result, nil
}

func (c *FakeClient) SetPendingDevices(devs []PendingDevice) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pendingDevs = devs
}

func (c *FakeClient) GetPendingDevices(_ context.Context) ([]PendingDevice, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]PendingDevice, len(c.pendingDevs))
	copy(result, c.pendingDevs)
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

func (c *FakeClient) SetFolderDevices(folderID string, devices []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.folderDevices[folderID] = devices
}

func (c *FakeClient) GetFoldersWithDevices(_ context.Context) ([]FolderConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var configs []FolderConfig
	for _, f := range c.folders {
		configs = append(configs, FolderConfig{
			ID:      f.ID,
			Path:    f.Path,
			Type:    f.Type,
			Devices: c.folderDevices[f.ID],
		})
	}
	return configs, nil
}

func (c *FakeClient) ReconcileFolderSharing(_ context.Context, drift []FolderSharingDrift) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reconciledDrift = append(c.reconciledDrift, drift...)
	return nil
}

func (c *FakeClient) ReconciledDrift() []FolderSharingDrift {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]FolderSharingDrift, len(c.reconciledDrift))
	copy(result, c.reconciledDrift)
	return result
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

	var devices []DeviceStatus
	for _, dev := range c.addedPeers {
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
		DeviceID: c.deviceID,
		Devices:  devices,
		Folders:  folders,
	}, nil
}

func (c *FakeClient) AddDevice(ctx context.Context, deviceID, name string) error {
	return c.AddDeviceWithAddresses(ctx, deviceID, name, nil)
}

func (c *FakeClient) AddDeviceWithAddresses(_ context.Context, deviceID, name string, _ []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.addedPeers = append(c.addedPeers, ConfiguredDevice{ID: deviceID, Name: name})
	return nil
}

func (c *FakeClient) AddDeviceAutoName(ctx context.Context, deviceID string) error {
	return c.AddDevice(ctx, deviceID, "")
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

func (c *FakeClient) SetDeviceCompletion(deviceID string, completion CompletionResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deviceCompletions[deviceID] = completion
}

func (c *FakeClient) GetDeviceCompletion(_ context.Context, deviceID string) (*CompletionResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if comp, ok := c.deviceCompletions[deviceID]; ok {
		return &comp, nil
	}
	return &CompletionResponse{Completion: 100}, nil
}

var _ SyncClient = (*FakeClient)(nil)
