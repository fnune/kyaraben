package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PairingClient struct {
	httpClient *http.Client
}

func NewPairingClient() *PairingClient {
	return &PairingClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *PairingClient) Pair(ctx context.Context, addr, code, localDeviceID, localName string) (*PairResult, error) {
	reqBody := PairingRequest{
		Code:     code,
		DeviceID: localDeviceID,
		Name:     localName,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := fmt.Sprintf("http://%s/pair", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending pairing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("invalid pairing code")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("too many pairing attempts")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pairing failed with status %d", resp.StatusCode)
	}

	var pairingResp PairingResponse
	if err := json.NewDecoder(resp.Body).Decode(&pairingResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &PairResult{
		PeerDeviceID: pairingResp.DeviceID,
		PeerName:     pairingResp.Name,
	}, nil
}
