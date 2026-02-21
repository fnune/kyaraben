package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/model"
)

const pairingTimeout = 5 * time.Minute
const pollInterval = 2 * time.Second

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
	ctx, cancel := context.WithTimeout(ctx, pairingTimeout)
	defer cancel()

	localID, err := f.cfg.Client.GetDeviceID(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting local device ID: %w", err)
	}

	f.emit("DEVICE_ID:%s", localID)
	f.emit("On your secondary device, open the Sync tab and select this device from the list.")
	f.emit("Waiting for secondary to connect...")

	seenPending := make(map[string]bool)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pending, err := f.cfg.Client.GetPendingDevices(ctx)
			if err != nil {
				log.Debug("error getting pending devices: %v", err)
				continue
			}

			for _, dev := range pending {
				if seenPending[dev.DeviceID] {
					continue
				}
				seenPending[dev.DeviceID] = true

				f.emit("Device found: %s", dev.Name)
				f.emit("Pairing with %s...", dev.Name)

				if err := f.cfg.Client.AddDevice(ctx, dev.DeviceID, dev.Name); err != nil {
					return nil, fmt.Errorf("adding device: %w", err)
				}
				if err := f.cfg.Client.ShareFoldersWithDevice(ctx, dev.DeviceID); err != nil {
					return nil, fmt.Errorf("sharing folders: %w", err)
				}

				f.emit("Paired with %s", dev.Name)
				return &PairResult{PeerDeviceID: dev.DeviceID, PeerName: dev.Name}, nil
			}

		case <-ctx.Done():
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
	ctx, cancel := context.WithTimeout(ctx, pairingTimeout)
	defer cancel()

	f.emit("Connecting to %s...", truncateDeviceID(primaryDeviceID))

	if err := f.cfg.Client.AddDeviceAutoName(ctx, primaryDeviceID); err != nil {
		return nil, fmt.Errorf("adding primary device: %w", err)
	}

	f.emit("Waiting for primary to accept connection...")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			devices, err := f.cfg.Client.GetConfiguredDevices(ctx)
			if err != nil {
				log.Debug("error getting configured devices: %v", err)
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
				f.emit("Device was removed from configuration")
				return nil, fmt.Errorf("pairing cancelled: device was removed")
			}

			connections, err := f.cfg.Client.GetConnections(ctx)
			if err != nil {
				log.Debug("error getting connections: %v", err)
				continue
			}

			conn, ok := connections[primaryDeviceID]
			if ok && conn.Connected {
				if err := f.cfg.Client.ShareFoldersWithDevice(ctx, primaryDeviceID); err != nil {
					return nil, fmt.Errorf("sharing folders: %w", err)
				}

				f.emit("Paired with %s", peerName)
				return &PairResult{PeerDeviceID: primaryDeviceID, PeerName: peerName}, nil
			}

			f.emit("Waiting for primary to accept... (device ID: %s)", truncateDeviceID(primaryDeviceID))

		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("pairing timed out: primary device did not accept connection")
			}
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
