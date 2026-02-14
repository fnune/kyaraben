package sync

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/fnune/kyaraben/internal/model"
)

const pairingTimeout = 5 * time.Minute

type PairingFlowConfig struct {
	SyncConfig model.SyncConfig
	Instance   string
	Advertiser Advertiser
	Browser    Browser
	Client     SyncClient
	OnProgress func(msg string)
}

type PrimaryPairingFlow struct {
	cfg PairingFlowConfig
}

func NewPrimaryPairingFlow(cfg PairingFlowConfig) *PrimaryPairingFlow {
	return &PrimaryPairingFlow{cfg: cfg}
}

func (f *PrimaryPairingFlow) Run(ctx context.Context) (*PairResult, string, error) {
	ctx, cancel := context.WithTimeout(ctx, pairingTimeout)
	defer cancel()

	code, err := GeneratePairingCode()
	if err != nil {
		return nil, "", fmt.Errorf("generating pairing code: %w", err)
	}

	localID, err := f.cfg.Client.GetDeviceID(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("getting local device ID: %w", err)
	}

	hostname, _ := os.Hostname()
	localName := hostname + "-kyaraben"
	if f.cfg.Instance != "" {
		localName = hostname + "-kyaraben-" + f.cfg.Instance
	}

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, "", fmt.Errorf("creating pairing listener: %w", err)
	}

	server := NewPairingServer(code, localID, localName)
	server.SetOnPairAccept(func(peerDeviceID, peerName string) error {
		if err := f.cfg.Client.AddDevice(ctx, peerDeviceID, peerName); err != nil {
			return fmt.Errorf("adding device to syncthing: %w", err)
		}
		if err := f.cfg.Client.ShareFoldersWithDevice(ctx, peerDeviceID); err != nil {
			return fmt.Errorf("sharing folders: %w", err)
		}
		return nil
	})

	if err := server.Start(listener); err != nil {
		return nil, "", fmt.Errorf("starting pairing server: %w", err)
	}
	defer server.Stop()

	port := server.Port()

	if err := f.cfg.Advertiser.Advertise(ctx, hostname, port); err != nil {
		return nil, "", fmt.Errorf("starting mDNS advertisement: %w", err)
	}
	defer f.cfg.Advertiser.Stop()

	f.emit("Pairing code: %s", code)
	f.emit("Waiting for devices... (expires in 5 minutes)")

	select {
	case result := <-server.Result():
		f.emit("Paired with %s", result.PeerName)
		return &result, code, nil
	case <-ctx.Done():
		return nil, code, fmt.Errorf("pairing timed out")
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

func (f *SecondaryPairingFlow) Run(ctx context.Context, code string) (*PairResult, error) {
	ctx, cancel := context.WithTimeout(ctx, pairingTimeout)
	defer cancel()

	localID, err := f.cfg.Client.GetDeviceID(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting local device ID: %w", err)
	}

	hostname, _ := os.Hostname()
	localName := hostname + "-kyaraben"
	if f.cfg.Instance != "" {
		localName = hostname + "-kyaraben-" + f.cfg.Instance
	}

	f.emit("Searching for primary...")

	offers, err := f.cfg.Browser.Browse(ctx)
	if err != nil {
		return nil, fmt.Errorf("browsing for devices: %w", err)
	}

	var offer PairingOffer
	for {
		select {
		case o, ok := <-offers:
			if !ok {
				return nil, fmt.Errorf("no primaries found on the network")
			}
			offer = o
			f.emit("Found: %s (%s)", offer.Hostname, offer.PairingAddr)
			goto pair
		case <-ctx.Done():
			return nil, fmt.Errorf("discovery timed out")
		}
	}

pair:
	f.emit("Pairing...")

	pairingClient := NewPairingClient()
	result, err := pairingClient.Pair(ctx, offer.PairingAddr, code, localID, localName, string(f.cfg.SyncConfig.Mode))
	if err != nil {
		return nil, fmt.Errorf("pairing with primary: %w", err)
	}

	if err := f.cfg.Client.AddDevice(ctx, result.PeerDeviceID, result.PeerName); err != nil {
		return nil, fmt.Errorf("adding device to syncthing: %w", err)
	}
	if err := f.cfg.Client.ShareFoldersWithDevice(ctx, result.PeerDeviceID); err != nil {
		return nil, fmt.Errorf("sharing folders: %w", err)
	}

	f.emit("Paired with %s", result.PeerName)
	return result, nil
}

func (f *SecondaryPairingFlow) emit(format string, args ...any) {
	if f.cfg.OnProgress != nil {
		f.cfg.OnProgress(fmt.Sprintf(format, args...))
	}
}
