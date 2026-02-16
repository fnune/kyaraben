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
	paused      bool
	config      model.SyncConfig
}

func NewFakeClient(config model.SyncConfig) *FakeClient {
	return &FakeClient{
		running:     true,
		connections: make(map[string]ConnectionInfo),
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

func (c *FakeClient) IsPaused() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.paused
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

func (c *FakeClient) GetFolderStatus(_ context.Context, _ string) (*FolderStatus, error) {
	return &FolderStatus{State: "idle"}, nil
}

func (c *FakeClient) GetStatus(_ context.Context) (*Status, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var devices []DeviceStatus
	for _, dev := range c.config.Devices {
		conn, ok := c.connections[dev.ID]
		devices = append(devices, DeviceStatus{
			ID:        dev.ID,
			Name:      dev.Name,
			Connected: ok && conn.Connected,
		})
	}
	return &Status{
		Enabled:  c.config.Enabled,
		Mode:     c.config.Mode,
		DeviceID: c.deviceID,
		Devices:  devices,
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

func (c *FakeClient) PauseSync(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paused = true
	return nil
}

func (c *FakeClient) ResumeSync(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paused = false
	return nil
}

var _ SyncClient = (*FakeClient)(nil)
