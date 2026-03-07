package syncthing

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
)

func truncateID(id string) string {
	if len(id) > 7 {
		return id[:7] + "..."
	}
	return id
}

type Client struct {
	config     Config
	apiKey     string
	httpClient *http.Client
}

func NewClient(config Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) baseURL() string {
	if c.config.BaseURL != "" {
		return c.config.BaseURL
	}
	return fmt.Sprintf("http://127.0.0.1:%d", c.config.GUIPort)
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

func (c *Client) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/system/status", nil)
	if err != nil {
		return nil, fmt.Errorf("getting system status: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var status SystemStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &status, nil
}

func (c *Client) GetDeviceID(ctx context.Context) (string, error) {
	status, err := c.GetSystemStatus(ctx)
	if err != nil {
		return "", err
	}
	return status.MyID, nil
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

	var conns struct {
		Connections map[string]ConnectionInfo `json:"connections"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&conns); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return conns.Connections, nil
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

func (c *Client) GetDeviceCompletion(ctx context.Context, deviceID string) (*CompletionResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/db/completion?device="+deviceID, nil)
	if err != nil {
		return nil, fmt.Errorf("getting device completion: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var completion CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &completion, nil
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

	var response struct {
		Files []struct {
			Action   string `json:"action"`
			Type     string `json:"type"`
			Name     string `json:"name"`
			Modified string `json:"modified"`
			Size     int64  `json:"size"`
		} `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

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
		defaultLogger.Debug("Ping failed: %v", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		defaultLogger.Debug("Ping returned status %d", resp.StatusCode)
		return false
	}
	return true
}

func (c *Client) SetAPIKey(key string) {
	c.apiKey = key
}

func (c *Client) Config() Config {
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
	return c.AddDeviceWithAddresses(ctx, deviceID, name, []string{"dynamic"})
}

func (c *Client) AddDeviceWithAddresses(ctx context.Context, deviceID, name string, addresses []string) error {
	defaultLogger.Info("PUT /rest/config/devices/%s (name=%q, addresses=%v)", truncateID(deviceID), name, addresses)
	dev := syncthingDevice{
		DeviceID:          deviceID,
		Name:              name,
		Addresses:         addresses,
		Compression:       "metadata",
		AutoAcceptFolders: false,
	}

	resp, err := c.doRequest(ctx, http.MethodPut, "/rest/config/devices/"+deviceID, dev)
	if err != nil {
		defaultLogger.Error("PUT /rest/config/devices/%s failed: %v", truncateID(deviceID), err)
		return fmt.Errorf("adding device: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		defaultLogger.Error("PUT /rest/config/devices/%s returned status %d", truncateID(deviceID), resp.StatusCode)
		return fmt.Errorf("unexpected status adding device: %d", resp.StatusCode)
	}
	defaultLogger.Info("PUT /rest/config/devices/%s -> 200 OK", truncateID(deviceID))
	return nil
}

func (c *Client) AddDeviceAutoName(ctx context.Context, deviceID string) error {
	return c.AddDevice(ctx, deviceID, "")
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

func (c *Client) GetConfiguredDevices(ctx context.Context) ([]ConfiguredDevice, error) {
	return c.getConfiguredDevicesWithRetry(ctx, 0)
}

func (c *Client) getConfiguredDevicesWithRetry(ctx context.Context, attempt int) ([]ConfiguredDevice, error) {
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

	if len(result) == 0 && len(devices) > 0 {
		defaultLogger.Debug("GetConfiguredDevices: only self device in config (filtered %d)", len(devices))
	}

	if len(result) == 0 && attempt < 3 {
		status, _ := c.GetSystemStatus(ctx)
		uptime := 0
		if status != nil {
			uptime = status.Uptime
		}
		defaultLogger.Debug("GetConfiguredDevices: raw=%d, uptime=%ds, attempt=%d", len(devices), uptime, attempt)

		if uptime < 5 {
			defaultLogger.Info("GetConfiguredDevices: empty result with low uptime (%ds), retrying in 500ms (attempt %d)", uptime, attempt+1)
			select {
			case <-ctx.Done():
				return result, nil
			case <-time.After(500 * time.Millisecond):
			}
			return c.getConfiguredDevicesWithRetry(ctx, attempt+1)
		}
	}

	return result, nil
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
	defaultLogger.Info("Sharing folders with device %s", truncateID(deviceID))

	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/config/folders", nil)
	if err != nil {
		defaultLogger.Error("GET /rest/config/folders failed: %v", err)
		return fmt.Errorf("getting folders: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		defaultLogger.Error("GET /rest/config/folders returned status %d", resp.StatusCode)
		return fmt.Errorf("unexpected status getting folders: %d", resp.StatusCode)
	}

	var folders []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return fmt.Errorf("decoding folders: %w", err)
	}

	sharedCount := 0
	for i := range folders {
		devices, ok := folders[i]["devices"].([]any)
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
		folders[i]["devices"] = devices
		sharedCount++
	}

	if sharedCount == 0 {
		defaultLogger.Info("All folders already shared with device %s", truncateID(deviceID))
		return nil
	}

	putResp, err := c.doRequest(ctx, http.MethodPut, "/rest/config/folders", folders)
	if err != nil {
		defaultLogger.Error("PUT /rest/config/folders failed: %v", err)
		return fmt.Errorf("updating folders: %w", err)
	}
	_ = putResp.Body.Close()

	if putResp.StatusCode != http.StatusOK {
		defaultLogger.Error("PUT /rest/config/folders returned status %d", putResp.StatusCode)
		return fmt.Errorf("unexpected status updating folders: %d", putResp.StatusCode)
	}

	defaultLogger.Info("Shared %d folders with device %s", sharedCount, truncateID(deviceID))
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

func (c *Client) GetFoldersWithDevices(ctx context.Context) ([]FolderConfig, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/config/folders", nil)
	if err != nil {
		return nil, fmt.Errorf("getting folders: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var rawFolders []struct {
		ID      string `json:"id"`
		Path    string `json:"path"`
		Type    string `json:"type"`
		Devices []struct {
			DeviceID string `json:"deviceID"`
		} `json:"devices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawFolders); err != nil {
		return nil, fmt.Errorf("decoding folders: %w", err)
	}

	folders := make([]FolderConfig, len(rawFolders))
	for i, raw := range rawFolders {
		devices := make([]string, len(raw.Devices))
		for j, d := range raw.Devices {
			devices[j] = d.DeviceID
		}
		folders[i] = FolderConfig{
			ID:      raw.ID,
			Path:    raw.Path,
			Type:    raw.Type,
			Devices: devices,
		}
	}

	return folders, nil
}

func (c *Client) ReconcileFolderSharing(ctx context.Context, drift []FolderSharingDrift) error {
	if len(drift) == 0 {
		return nil
	}

	driftMap := make(map[string][]string)
	for _, d := range drift {
		driftMap[d.FolderID] = d.MissingDeviceIDs
	}

	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/config/folders", nil)
	if err != nil {
		return fmt.Errorf("getting folders: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status getting folders: %d", resp.StatusCode)
	}

	var folders []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return fmt.Errorf("decoding folders: %w", err)
	}

	modified := false
	for i := range folders {
		folderID, _ := folders[i]["id"].(string)
		missingDevices, needsFix := driftMap[folderID]
		if !needsFix {
			continue
		}

		devices, _ := folders[i]["devices"].([]any)
		for _, deviceID := range missingDevices {
			devices = append(devices, map[string]any{"deviceID": deviceID})
			defaultLogger.Info("Reconciling: adding device %s to folder %s", truncateID(deviceID), folderID)
		}
		folders[i]["devices"] = devices
		modified = true
	}

	if !modified {
		return nil
	}

	putResp, err := c.doRequest(ctx, http.MethodPut, "/rest/config/folders", folders)
	if err != nil {
		return fmt.Errorf("updating folders: %w", err)
	}
	_ = putResp.Body.Close()

	if putResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status updating folders: %d", putResp.StatusCode)
	}

	return nil
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
