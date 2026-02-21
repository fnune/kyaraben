package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultRelayURL = "https://kyaraben-relay.koyeb.app"
	relayTimeout    = 10 * time.Second
)

type RelayClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewRelayClient(baseURL string) *RelayClient {
	if baseURL == "" {
		baseURL = DefaultRelayURL
	}
	return &RelayClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: relayTimeout,
		},
	}
}

type CreateSessionResponse struct {
	Code      string `json:"code"`
	ExpiresIn int    `json:"expiresIn"`
}

type GetSessionResponse struct {
	DeviceID string `json:"deviceId"`
}

type GetResponseResponse struct {
	DeviceID string `json:"deviceId,omitempty"`
	Ready    bool   `json:"ready"`
}

type relayError struct {
	Error string `json:"error"`
}

func (c *RelayClient) CreateSession(ctx context.Context, deviceID string) (*CreateSessionResponse, error) {
	body := map[string]string{"deviceId": deviceID}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/pair", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result CreateSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func (c *RelayClient) GetSession(ctx context.Context, code string) (*GetSessionResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/pair/"+code, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result GetSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func (c *RelayClient) SubmitResponse(ctx context.Context, code, deviceID string) error {
	body := map[string]string{"deviceId": deviceID}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/pair/"+code+"/response", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}
	return nil
}

func (c *RelayClient) GetResponse(ctx context.Context, code string) (*GetResponseResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/pair/"+code+"/response", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result GetResponseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func (c *RelayClient) DeleteSession(ctx context.Context, code string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/pair/"+code, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return c.parseError(resp)
	}
	return nil
}

func (c *RelayClient) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp relayError
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return fmt.Errorf("relay error: %s", errResp.Error)
	}

	return fmt.Errorf("relay error: HTTP %d", resp.StatusCode)
}
