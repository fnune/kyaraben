package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
)

var log = logging.New("sync")

type Client struct {
	config     model.SyncConfig
	apiKey     string
	httpClient *http.Client
}

func NewClient(config model.SyncConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) baseURL() string {
	if c.config.Syncthing.BaseURL != "" {
		return c.config.Syncthing.BaseURL
	}
	return fmt.Sprintf("http://127.0.0.1:%d", c.config.Syncthing.GUIPort)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	url := c.baseURL() + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

type SystemStatus struct {
	MyID string `json:"myID"`
}

func (c *Client) GetDeviceID(ctx context.Context) (string, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/system/status", nil)
	if err != nil {
		return "", fmt.Errorf("getting system status: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var status SystemStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return status.MyID, nil
}

type ConnectionInfo struct {
	Connected  bool   `json:"connected"`
	Address    string `json:"address"`
	Paused     bool   `json:"paused"`
	DeviceName string `json:"deviceName"`
}

type ConnectionsResponse struct {
	Connections map[string]ConnectionInfo `json:"connections"`
}

func (c *Client) GetConnections(ctx context.Context) (map[string]ConnectionInfo, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/system/connections", nil)
	if err != nil {
		return nil, fmt.Errorf("getting connections: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var conns ConnectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&conns); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return conns.Connections, nil
}

type FolderStatus struct {
	State                  string `json:"state"`
	GlobalFiles            int    `json:"globalFiles"`
	GlobalBytes            int64  `json:"globalBytes"`
	LocalFiles             int    `json:"localFiles"`
	LocalBytes             int64  `json:"localBytes"`
	NeedFiles              int    `json:"needFiles"`
	NeedBytes              int64  `json:"needBytes"`
	PullErrors             int    `json:"pullErrors"`
	InSyncFiles            int    `json:"inSyncFiles"`
	InSyncBytes            int64  `json:"inSyncBytes"`
	ReceiveOnlyTotalItems  int    `json:"receiveOnlyTotalItems"`
	ReceiveOnlyChangedSize int64  `json:"receiveOnlyChangedBytes"`
}

func (c *Client) GetFolderStatus(ctx context.Context, folderID string) (*FolderStatus, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/db/status?folder="+folderID, nil)
	if err != nil {
		return nil, fmt.Errorf("getting folder status: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var status FolderStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &status, nil
}

func (c *Client) RevertFolder(ctx context.Context, folderID string) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/rest/db/revert?folder="+folderID, nil)
	if err != nil {
		return fmt.Errorf("reverting folder: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

type LocalChange struct {
	Action   string `json:"action"`
	Type     string `json:"type"`
	Path     string `json:"path"`
	Modified string `json:"modified"`
	Size     int64  `json:"size"`
}

type localChangedResponse struct {
	Files []struct {
		Action   string `json:"action"`
		Type     string `json:"type"`
		Name     string `json:"name"`
		Modified string `json:"modified"`
		Size     int64  `json:"size"`
	} `json:"files"`
}

func (c *Client) GetLocalChanges(ctx context.Context, folderID string) ([]LocalChange, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/db/localchanged?folder="+folderID, nil)
	if err != nil {
		return nil, fmt.Errorf("getting local changes: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var response localChangedResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Get folder path to check local filesystem
	folderPath := ""
	folders, err := c.GetFolderConfigs(ctx)
	if err == nil {
		for _, f := range folders {
			if f.ID == folderID {
				folderPath = f.Path
				break
			}
		}
	}

	changes := make([]LocalChange, len(response.Files))
	for i, f := range response.Files {
		action := ""
		if folderPath != "" {
			fullPath := filepath.Join(folderPath, f.Name)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				action = "deleted"
			} else if err == nil {
				action = "changed"
			}
		}
		changes[i] = LocalChange{
			Action:   action,
			Type:     f.Type,
			Path:     f.Name,
			Modified: f.Modified,
			Size:     f.Size,
		}
	}
	return changes, nil
}

func (c *Client) IsRunning(ctx context.Context) bool {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/system/ping", nil)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}

func (c *Client) SetAPIKey(key string) {
	c.apiKey = key
}

func (c *Client) Config() model.SyncConfig {
	return c.config
}

type syncthingDevice struct {
	DeviceID          string   `json:"deviceID"`
	Name              string   `json:"name,omitempty"`
	Addresses         []string `json:"addresses"`
	Compression       string   `json:"compression"`
	AutoAcceptFolders bool     `json:"autoAcceptFolders"`
	Paused            bool     `json:"paused,omitempty"`
}

func (c *Client) AddDevice(ctx context.Context, deviceID, name string) error {
	dev := syncthingDevice{
		DeviceID:          deviceID,
		Name:              name,
		Addresses:         []string{"dynamic"},
		Compression:       "metadata",
		AutoAcceptFolders: false,
	}

	resp, err := c.doRequest(ctx, http.MethodPut, "/rest/config/devices/"+deviceID, dev)
	if err != nil {
		return fmt.Errorf("adding device: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status adding device: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) RemoveDevice(ctx context.Context, deviceID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/rest/config/devices/"+deviceID, nil)
	if err != nil {
		return fmt.Errorf("removing device: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status removing device: %d", resp.StatusCode)
	}
	return nil
}

type ConfiguredDevice struct {
	ID     string
	Name   string
	Paused bool
}

func (c *Client) GetConfiguredDevices(ctx context.Context) ([]ConfiguredDevice, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/config/devices", nil)
	if err != nil {
		return nil, fmt.Errorf("getting devices config: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var devices []syncthingDevice
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, fmt.Errorf("decoding devices: %w", err)
	}

	myID, _ := c.GetDeviceID(ctx)
	var result []ConfiguredDevice
	for _, dev := range devices {
		if dev.DeviceID == myID {
			continue
		}
		result = append(result, ConfiguredDevice{
			ID:     dev.DeviceID,
			Name:   dev.Name,
			Paused: dev.Paused,
		})
	}
	return result, nil
}

type DiscoveredDevice struct {
	DeviceID  string
	Addresses []string
}

func (c *Client) GetDiscoveredDevices(ctx context.Context) ([]DiscoveredDevice, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/system/discovery", nil)
	if err != nil {
		return nil, fmt.Errorf("getting discovery: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var discovery map[string]struct {
		Addresses []string `json:"addresses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("decoding discovery: %w", err)
	}

	myID, _ := c.GetDeviceID(ctx)
	var result []DiscoveredDevice
	for deviceID, info := range discovery {
		if deviceID == myID {
			continue
		}
		result = append(result, DiscoveredDevice{
			DeviceID:  deviceID,
			Addresses: info.Addresses,
		})
	}
	return result, nil
}

type PendingDevice struct {
	DeviceID string
	Name     string
	Address  string
	Time     time.Time
}

func (c *Client) GetPendingDevices(ctx context.Context) ([]PendingDevice, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/cluster/pending/devices", nil)
	if err != nil {
		return nil, fmt.Errorf("getting pending devices: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var pending map[string]struct {
		Name    string    `json:"name"`
		Address string    `json:"address"`
		Time    time.Time `json:"time"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pending); err != nil {
		return nil, fmt.Errorf("decoding pending devices: %w", err)
	}

	var result []PendingDevice
	for deviceID, info := range pending {
		result = append(result, PendingDevice{
			DeviceID: deviceID,
			Name:     info.Name,
			Address:  info.Address,
			Time:     info.Time,
		})
	}
	return result, nil
}

func (c *Client) ShareFoldersWithDevice(ctx context.Context, deviceID string) error {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/config/folders", nil)
	if err != nil {
		return fmt.Errorf("getting folders: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status getting folders: %d", resp.StatusCode)
	}

	var folders []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return fmt.Errorf("decoding folders: %w", err)
	}

	for _, raw := range folders {
		var folder map[string]any
		if err := json.Unmarshal(raw, &folder); err != nil {
			continue
		}

		devices, ok := folder["devices"].([]any)
		if !ok {
			continue
		}

		alreadyShared := false
		for _, d := range devices {
			if devMap, ok := d.(map[string]any); ok {
				if devMap["deviceID"] == deviceID {
					alreadyShared = true
					break
				}
			}
		}

		if alreadyShared {
			continue
		}

		devices = append(devices, map[string]any{"deviceID": deviceID})
		folder["devices"] = devices

		folderID, _ := folder["id"].(string)
		patchResp, err := c.doRequest(ctx, http.MethodPut, "/rest/config/folders/"+folderID, folder)
		if err != nil {
			return fmt.Errorf("updating folder %s: %w", folderID, err)
		}
		_ = patchResp.Body.Close()
	}

	return nil
}

func (c *Client) Restart(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/rest/system/restart", nil)
	if err != nil {
		return fmt.Errorf("restarting syncthing: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

type PendingFolder struct {
	ID         string
	Label      string
	OfferedBy  string
	DeviceName string
}

func (c *Client) GetPendingFolders(ctx context.Context) ([]PendingFolder, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/cluster/pending/folders", nil)
	if err != nil {
		return nil, fmt.Errorf("getting pending folders: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var pending map[string]map[string]struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pending); err != nil {
		return nil, fmt.Errorf("decoding pending folders: %w", err)
	}

	var result []PendingFolder
	for folderID, devices := range pending {
		for deviceID, info := range devices {
			result = append(result, PendingFolder{
				ID:        folderID,
				Label:     info.Label,
				OfferedBy: deviceID,
			})
		}
	}
	return result, nil
}

func (c *Client) DismissPendingFolder(ctx context.Context, folderID, deviceID string) error {
	url := fmt.Sprintf("/rest/cluster/pending/folders?folder=%s&device=%s", folderID, deviceID)
	resp, err := c.doRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("dismissing pending folder: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

type FolderConfig struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Type string `json:"type"`
}

func (c *Client) GetFolderConfigs(ctx context.Context) ([]FolderConfig, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/config/folders", nil)
	if err != nil {
		return nil, fmt.Errorf("getting folders: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var folders []FolderConfig
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return nil, fmt.Errorf("decoding folders: %w", err)
	}

	return folders, nil
}

type PendingStatus struct {
	TotalFiles int64
	TotalBytes int64
}

func (c *Client) GetPendingStatus(ctx context.Context) (*PendingStatus, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/config/folders", nil)
	if err != nil {
		return nil, fmt.Errorf("getting folders: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var folders []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return nil, fmt.Errorf("decoding folders: %w", err)
	}

	var pending PendingStatus
	for _, folder := range folders {
		status, err := c.GetFolderStatus(ctx, folder.ID)
		if err != nil {
			continue
		}
		pending.TotalFiles += int64(status.NeedFiles)
		pending.TotalBytes += status.NeedBytes
	}

	return &pending, nil
}

type SyncProgressInfo struct {
	NeedFiles   int64
	NeedBytes   int64
	GlobalBytes int64
	Percent     int
}

func (c *Client) GetSyncProgress(ctx context.Context) (*SyncProgressInfo, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/config/folders", nil)
	if err != nil {
		return nil, fmt.Errorf("getting folders: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var folders []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return nil, fmt.Errorf("decoding folders: %w", err)
	}

	var progress SyncProgressInfo
	for _, folder := range folders {
		status, err := c.GetFolderStatus(ctx, folder.ID)
		if err != nil {
			continue
		}
		progress.NeedFiles += int64(status.NeedFiles)
		progress.NeedBytes += status.NeedBytes
		progress.GlobalBytes += status.GlobalBytes
	}

	if progress.GlobalBytes > 0 {
		syncedBytes := progress.GlobalBytes - progress.NeedBytes
		progress.Percent = int(syncedBytes * 100 / progress.GlobalBytes)
	} else {
		progress.Percent = 100
	}

	return &progress, nil
}
