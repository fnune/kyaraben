package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
)

const pairingTimeout = 5 * time.Minute
const pollInterval = 2 * time.Second

var pairingLog = logging.New("sync").WithPrefix("[pairing]")

type PairingFlowConfig struct {
	SyncConfig model.SyncConfig
	Instance   string
	Client     SyncClient
	OnProgress func(msg string)
}

type PrimaryPairingFlow struct {
	cfg PairingFlowConfig
}

func NewPrimaryPairingFlow(cfg PairingFlowConfig) *PrimaryPairingFlow {
	return &PrimaryPairingFlow{cfg: cfg}
}

func (f *PrimaryPairingFlow) Run(ctx context.Context) (*PairResult, error) {
	pairingLog.Info("Starting primary pairing flow")

	ctx, cancel := context.WithTimeout(ctx, pairingTimeout)
	defer cancel()

	localID, err := f.cfg.Client.GetDeviceID(ctx)
	if err != nil {
		pairingLog.Error("Failed to get local device ID: %v", err)
		return nil, fmt.Errorf("getting local device ID: %w", err)
	}
	pairingLog.Info("Local device ID: %s", truncateDeviceID(localID))

	f.emit("DEVICE_ID:%s", localID)
	f.emit("On your secondary device, open the Sync tab and select this device from the list.")
	f.emit("Waiting for secondary to connect...")

	seenPending := make(map[string]bool)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	pollCount := 0
	for {
		select {
		case <-ticker.C:
			pollCount++
			pending, err := f.cfg.Client.GetPendingDevices(ctx)
			if err != nil {
				pairingLog.Info("Poll %d: error getting pending devices: %v", pollCount, err)
				continue
			}

			for _, dev := range pending {
				if seenPending[dev.DeviceID] {
					continue
				}
				seenPending[dev.DeviceID] = true

				pairingLog.Info("Poll %d: found pending device %s (%s)", pollCount, dev.Name, truncateDeviceID(dev.DeviceID))
				f.emit("Device found: %s", dev.Name)
				f.emit("Pairing with %s...", dev.Name)

				pairingLog.Info("Adding device %s to syncthing config", truncateDeviceID(dev.DeviceID))
				if err := f.cfg.Client.AddDevice(ctx, dev.DeviceID, dev.Name); err != nil {
					pairingLog.Error("Failed to add device: %v", err)
					return nil, fmt.Errorf("adding device: %w", err)
				}

				pairingLog.Info("Sharing folders with device %s", truncateDeviceID(dev.DeviceID))
				if err := f.cfg.Client.ShareFoldersWithDevice(ctx, dev.DeviceID); err != nil {
					pairingLog.Error("Failed to share folders: %v", err)
					return nil, fmt.Errorf("sharing folders: %w", err)
				}

				pairingLog.Info("Primary pairing completed successfully with %s", dev.Name)
				f.emit("Paired with %s", dev.Name)
				return &PairResult{PeerDeviceID: dev.DeviceID, PeerName: dev.Name}, nil
			}

			pairingLog.Debug("Poll %d: no pending devices", pollCount)

		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				pairingLog.Error("Primary pairing timed out after %d polls", pollCount)
			} else {
				pairingLog.Info("Primary pairing cancelled after %d polls", pollCount)
			}
			return nil, ctx.Err()
		}
	}
}

func (f *PrimaryPairingFlow) emit(format string, args ...any) {
	if f.cfg.OnProgress != nil {
		f.cfg.OnProgress(fmt.Sprintf(format, args...))
	}
}

type SecondaryPairingFlow struct {
	cfg PairingFlowConfig
}

func NewSecondaryPairingFlow(cfg PairingFlowConfig) *SecondaryPairingFlow {
	return &SecondaryPairingFlow{cfg: cfg}
}

func (f *SecondaryPairingFlow) Run(ctx context.Context, primaryDeviceID string) (*PairResult, error) {
	pairingLog.Info("Starting secondary pairing flow for primary device %s", truncateDeviceID(primaryDeviceID))

	ctx, cancel := context.WithTimeout(ctx, pairingTimeout)
	defer cancel()

	f.emit("Connecting to %s...", truncateDeviceID(primaryDeviceID))

	pairingLog.Info("Adding primary device to syncthing config")
	if err := f.cfg.Client.AddDeviceAutoName(ctx, primaryDeviceID); err != nil {
		pairingLog.Error("Failed to add primary device: %v", err)
		return nil, fmt.Errorf("adding primary device: %w", err)
	}
	pairingLog.Info("Primary device added successfully")

	f.emit("Waiting for primary to accept connection...")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	pollCount := 0
	for {
		select {
		case <-ticker.C:
			pollCount++
			devices, err := f.cfg.Client.GetConfiguredDevices(ctx)
			if err != nil {
				pairingLog.Info("Poll %d: error getting configured devices: %v", pollCount, err)
				continue
			}

			deviceStillConfigured := false
			peerName := ""
			for _, dev := range devices {
				if dev.ID == primaryDeviceID {
					deviceStillConfigured = true
					peerName = dev.Name
					break
				}
			}

			if !deviceStillConfigured {
				pairingLog.Error("Poll %d: primary device was removed from configuration", pollCount)
				f.emit("Device was removed from configuration")
				return nil, fmt.Errorf("pairing cancelled: device was removed")
			}

			connections, err := f.cfg.Client.GetConnections(ctx)
			if err != nil {
				pairingLog.Info("Poll %d: error getting connections: %v", pollCount, err)
				continue
			}

			conn, ok := connections[primaryDeviceID]
			if ok && conn.Connected {
				pairingLog.Info("Poll %d: connection established with primary, sharing folders", pollCount)
				if err := f.cfg.Client.ShareFoldersWithDevice(ctx, primaryDeviceID); err != nil {
					pairingLog.Error("Failed to share folders: %v", err)
					return nil, fmt.Errorf("sharing folders: %w", err)
				}

				pairingLog.Info("Secondary pairing completed successfully with %s", peerName)
				f.emit("Paired with %s", peerName)
				return &PairResult{PeerDeviceID: primaryDeviceID, PeerName: peerName}, nil
			}

			pairingLog.Debug("Poll %d: waiting for connection (configured=%t, connected=%t)", pollCount, deviceStillConfigured, ok && conn.Connected)
			f.emit("Waiting for primary to accept... (device ID: %s)", truncateDeviceID(primaryDeviceID))

		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				pairingLog.Error("Secondary pairing timed out after %d polls", pollCount)
				return nil, fmt.Errorf("pairing timed out: primary device did not accept connection")
			}
			pairingLog.Info("Secondary pairing cancelled after %d polls", pollCount)
			return nil, fmt.Errorf("pairing cancelled")
		}
	}
}

func (f *SecondaryPairingFlow) emit(format string, args ...any) {
	if f.cfg.OnProgress != nil {
		f.cfg.OnProgress(fmt.Sprintf(format, args...))
	}
}

func truncateDeviceID(id string) string {
	if len(id) > 7 {
		return id[:7] + "..."
	}
	return id
}

type FormattedDeviceID struct {
	Full   string
	Groups []string
}

func FormatDeviceID(id string) FormattedDeviceID {
	clean := strings.ReplaceAll(id, "-", "")
	var groups []string
	for i := 0; i < len(clean); i += 7 {
		end := i + 7
		if end > len(clean) {
			end = len(clean)
		}
		groups = append(groups, clean[i:end])
	}
	return FormattedDeviceID{
		Full:   id,
		Groups: groups,
	}
}

type RelayPairingFlowConfig struct {
	SyncConfig model.SyncConfig
	Instance   string
	Client     SyncClient
	RelayURLs  []string
	OnProgress func(msg string)
	OnCode     func(code string, expiresIn int)
}

type RelayPrimaryResult struct {
	Code         string
	DeviceID     string
	PeerDeviceID string
	PeerName     string
}

type RelayPrimaryPairingFlow struct {
	cfg RelayPairingFlowConfig
}

func NewRelayPrimaryPairingFlow(cfg RelayPairingFlowConfig) *RelayPrimaryPairingFlow {
	return &RelayPrimaryPairingFlow{cfg: cfg}
}

func (f *RelayPrimaryPairingFlow) Run(ctx context.Context) (*RelayPrimaryResult, error) {
	pairingLog.Info("Starting relay primary pairing flow")

	localID, err := f.cfg.Client.GetDeviceID(ctx)
	if err != nil {
		pairingLog.Error("Failed to get local device ID: %v", err)
		return nil, fmt.Errorf("getting local device ID: %w", err)
	}
	pairingLog.Info("Local device ID: %s", truncateDeviceID(localID))

	relayURLs := f.cfg.RelayURLs
	if len(relayURLs) == 0 {
		relayURLs = ProductionRelayURLs
	}

	relayClient, err := NewRelayClient(relayURLs)
	if err != nil {
		pairingLog.Info("Relay unavailable: %v", err)
		return &RelayPrimaryResult{DeviceID: localID}, nil
	}

	session, err := relayClient.CreateSession(ctx, localID)
	if err != nil {
		pairingLog.Info("Failed to create relay session: %v", err)
		return &RelayPrimaryResult{DeviceID: localID}, nil
	}

	pairingLog.Info("Created relay session with code %s", session.Code)

	if f.cfg.OnCode != nil {
		f.cfg.OnCode(session.Code, session.ExpiresIn)
	}

	pairingCtx, cancel := context.WithTimeout(ctx, time.Duration(session.ExpiresIn)*time.Second)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	pollCount := 0
	for {
		select {
		case <-pairingCtx.Done():
			_ = relayClient.DeleteSession(ctx, session.Code)
			if pairingCtx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("pairing timed out")
			}
			return nil, pairingCtx.Err()
		case <-ticker.C:
			pollCount++
			resp, err := relayClient.GetResponse(ctx, session.Code)
			if err != nil {
				pairingLog.Debug("Poll %d: error getting response: %v", pollCount, err)
				continue
			}
			if !resp.Ready {
				pairingLog.Debug("Poll %d: not ready", pollCount)
				continue
			}

			pairingLog.Info("Poll %d: secondary device ID received: %s", pollCount, truncateDeviceID(resp.DeviceID))

			if err := f.cfg.Client.AddDeviceAutoName(ctx, resp.DeviceID); err != nil {
				_ = relayClient.DeleteSession(ctx, session.Code)
				return nil, fmt.Errorf("adding device: %w", err)
			}
			if err := f.cfg.Client.ShareFoldersWithDevice(ctx, resp.DeviceID); err != nil {
				_ = relayClient.DeleteSession(ctx, session.Code)
				return nil, fmt.Errorf("sharing folders: %w", err)
			}

			_ = relayClient.DeleteSession(ctx, session.Code)

			devices, _ := f.cfg.Client.GetConfiguredDevices(ctx)
			peerName := truncateDeviceID(resp.DeviceID)
			for _, d := range devices {
				if d.ID == resp.DeviceID {
					peerName = d.Name
					break
				}
			}

			pairingLog.Info("Primary pairing via relay completed with %s", peerName)
			f.emit("Paired with %s", peerName)
			return &RelayPrimaryResult{
				Code:         session.Code,
				DeviceID:     localID,
				PeerDeviceID: resp.DeviceID,
				PeerName:     peerName,
			}, nil
		}
	}
}

func (f *RelayPrimaryPairingFlow) emit(format string, args ...any) {
	if f.cfg.OnProgress != nil {
		f.cfg.OnProgress(fmt.Sprintf(format, args...))
	}
}

type RelaySecondaryPairingFlow struct {
	cfg RelayPairingFlowConfig
}

func NewRelaySecondaryPairingFlow(cfg RelayPairingFlowConfig) *RelaySecondaryPairingFlow {
	return &RelaySecondaryPairingFlow{cfg: cfg}
}

func (f *RelaySecondaryPairingFlow) Run(ctx context.Context, code string) (*PairResult, error) {
	pairingLog.Info("Starting relay secondary pairing flow with code %s", code)

	relayURLs := f.cfg.RelayURLs
	if len(relayURLs) == 0 {
		relayURLs = ProductionRelayURLs
	}

	relayClient, err := NewRelayClient(relayURLs)
	if err != nil {
		return nil, fmt.Errorf("relay unavailable: %w", err)
	}

	session, err := relayClient.GetSession(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("invalid pairing code: %w", err)
	}

	primaryDeviceID := session.DeviceID
	pairingLog.Info("Resolved code %s to device ID %s", code, truncateDeviceID(primaryDeviceID))

	f.emit("Connecting to %s...", truncateDeviceID(primaryDeviceID))

	localID, err := f.cfg.Client.GetDeviceID(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting local device ID: %w", err)
	}

	if err := relayClient.SubmitResponse(ctx, code, localID); err != nil {
		pairingLog.Error("Failed to submit response to relay: %v", err)
	} else {
		pairingLog.Info("Submitted device ID to relay")
	}

	flow := NewSecondaryPairingFlow(PairingFlowConfig{
		SyncConfig: f.cfg.SyncConfig,
		Instance:   f.cfg.Instance,
		Client:     f.cfg.Client,
		OnProgress: f.cfg.OnProgress,
	})

	return flow.Run(ctx, primaryDeviceID)
}

func (f *RelaySecondaryPairingFlow) emit(format string, args ...any) {
	if f.cfg.OnProgress != nil {
		f.cfg.OnProgress(fmt.Sprintf(format, args...))
	}
}

func IsRelayCode(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, c := range s {
		isUpper := c >= 'A' && c <= 'Z'
		isDigit := c >= '0' && c <= '9'
		if !isUpper && !isDigit {
			return false
		}
	}
	return true
}
