package sync

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"strings"
)

const (
	PairingCodeLength  = 6
	PairingCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

type PairingOffer struct {
	Hostname    string
	PairingAddr string
}

type Advertiser interface {
	Advertise(ctx context.Context, hostname string, port int) error
	Stop()
}

type Browser interface {
	Browse(ctx context.Context) (<-chan PairingOffer, error)
}

type PairResult struct {
	PeerDeviceID string
	PeerName     string
}

func GeneratePairingCode() (string, error) {
	var b strings.Builder
	for i := 0; i < PairingCodeLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(PairingCodeCharset))))
		if err != nil {
			return "", fmt.Errorf("generating random index: %w", err)
		}
		b.WriteByte(PairingCodeCharset[n.Int64()])
	}
	return b.String(), nil
}

func FindAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, fmt.Errorf("finding available port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	return port, nil
}
