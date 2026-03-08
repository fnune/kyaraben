package syncthing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const relayRequestTimeout = 10 * time.Second
const relayRetryDelay = 2 * time.Second
const relayMaxRetries = 3

func getRelayHealthRetries() int {
	if v := os.Getenv("KYARABEN_RELAY_HEALTH_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 30
}

var ProductionRelayURLs = []string{
	"https://kyaraben-kyaraben-245b3f77.koyeb.app/api",
}

type RelayClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewRelayClient(urls []string) (*RelayClient, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no relay URLs provided")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	maxRetries := getRelayHealthRetries()

	for _, url := range urls {
		for attempt := 0; attempt < maxRetries; attempt++ {
			resp, err := client.Get(url + "/health")
			if err != nil {
				defaultLogger.Debug("Relay health check attempt %d failed for %s: %v", attempt+1, url, err)
				time.Sleep(2 * time.Second)
				continue
			}
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return &RelayClient{
					baseURL:    url,
					httpClient: &http.Client{Timeout: relayRequestTimeout},
				}, nil
			}
			defaultLogger.Debug("Relay health check attempt %d got status %d for %s", attempt+1, resp.StatusCode, url)
			time.Sleep(2 * time.Second)
		}
	}

	return nil, fmt.Errorf("no healthy relay server found")
}

func NewDefaultRelayClient() (*RelayClient, error) {
	return NewRelayClient(ProductionRelayURLs)
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

	var lastErr error
	for attempt := 0; attempt < relayMaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(relayRetryDelay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/pair", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("sending request: %w", err)
			defaultLogger.Debug("CreateSession attempt %d failed: %v", attempt+1, lastErr)
			continue
		}

		if resp.StatusCode != http.StatusCreated {
			lastErr = c.parseError(resp)
			_ = resp.Body.Close()
			defaultLogger.Debug("CreateSession attempt %d got status %d", attempt+1, resp.StatusCode)
			continue
		}

		var result CreateSessionResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("decoding response: %w", err)
		}
		_ = resp.Body.Close()
		return &result, nil
	}

	return nil, lastErr
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
