package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
)

var log = logging.New("sync")

type Client struct {
	config     model.SyncConfig
	apiKey     string
	httpClient *http.Client
	process    *Process
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
	return fmt.Sprintf("http://127.0.0.1:%d", c.config.Syncthing.GUIPort)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	url := c.baseURL() + path

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)

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
	defer resp.Body.Close()

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
	Connected bool   `json:"connected"`
	Address   string `json:"address"`
}

type ConnectionsResponse struct {
	Connections map[string]ConnectionInfo `json:"connections"`
}

func (c *Client) GetConnections(ctx context.Context) (map[string]ConnectionInfo, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/system/connections", nil)
	if err != nil {
		return nil, fmt.Errorf("getting connections: %w", err)
	}
	defer resp.Body.Close()

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
	State       string `json:"state"`
	GlobalFiles int    `json:"globalFiles"`
	GlobalBytes int64  `json:"globalBytes"`
	LocalFiles  int    `json:"localFiles"`
	LocalBytes  int64  `json:"localBytes"`
	NeedFiles   int    `json:"needFiles"`
	NeedBytes   int64  `json:"needBytes"`
	PullErrors  int    `json:"pullErrors"`
	InSyncFiles int    `json:"inSyncFiles"`
	InSyncBytes int64  `json:"inSyncBytes"`
}

func (c *Client) GetFolderStatus(ctx context.Context, folderID string) (*FolderStatus, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/db/status?folder="+folderID, nil)
	if err != nil {
		return nil, fmt.Errorf("getting folder status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var status FolderStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &status, nil
}

func (c *Client) IsRunning(ctx context.Context) bool {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/system/ping", nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (c *Client) SetAPIKey(key string) {
	c.apiKey = key
}

func (c *Client) Config() model.SyncConfig {
	return c.config
}
