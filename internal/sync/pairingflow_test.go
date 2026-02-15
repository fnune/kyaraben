package sync

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/model"
)

func TestPairingServerRejectsInvalidCode(t *testing.T) {
	server := NewPairingServer("ABC123", "LOCAL-DEVICE-ID", "test-primary")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	if err := server.Start(listener); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer server.Stop()

	ctx := context.Background()
	client := NewPairingClient()
	addr := fmt.Sprintf("127.0.0.1:%d", server.Port())

	_, err = client.Pair(ctx, addr, "WRONG1", "peer-id", "peer-name")
	if err == nil {
		t.Fatal("expected error for wrong code")
	}
}

func TestPairingServerAcceptsCorrectCode(t *testing.T) {
	server := NewPairingServer("ABC123", "LOCAL-DEVICE-ID", "test-primary")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	if err := server.Start(listener); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer server.Stop()

	ctx := context.Background()
	client := NewPairingClient()
	addr := fmt.Sprintf("127.0.0.1:%d", server.Port())

	result, err := client.Pair(ctx, addr, "ABC123", "PEER-DEVICE-ID", "peer-name")
	if err != nil {
		t.Fatalf("pair: %v", err)
	}

	if result.PeerDeviceID != "LOCAL-DEVICE-ID" {
		t.Errorf("expected device ID LOCAL-DEVICE-ID, got %s", result.PeerDeviceID)
	}
	if result.PeerName != "test-primary" {
		t.Errorf("expected name test-primary, got %s", result.PeerName)
	}
}

func TestPairingServerRateLimits(t *testing.T) {
	server := NewPairingServer("ABC123", "LOCAL-DEVICE-ID", "test-primary")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	if err := server.Start(listener); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer server.Stop()

	ctx := context.Background()
	client := NewPairingClient()
	addr := fmt.Sprintf("127.0.0.1:%d", server.Port())

	for i := 0; i < maxPairingAttempts; i++ {
		_, _ = client.Pair(ctx, addr, "WRONG1", "peer-id", "peer-name")
	}

	_, err = client.Pair(ctx, addr, "ABC123", "peer-id", "peer-name")
	if err == nil {
		t.Fatal("expected rate limit error")
	}
}

func TestPairingServerOnPairAcceptCallback(t *testing.T) {
	var acceptedID, acceptedName string
	server := NewPairingServer("ABC123", "LOCAL-DEVICE-ID", "test-primary")
	server.SetOnPairAccept(func(peerDeviceID, peerName string) error {
		acceptedID = peerDeviceID
		acceptedName = peerName
		return nil
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	if err := server.Start(listener); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer server.Stop()

	ctx := context.Background()
	client := NewPairingClient()
	addr := fmt.Sprintf("127.0.0.1:%d", server.Port())

	_, err = client.Pair(ctx, addr, "ABC123", "PEER-ID-123", "steam-deck")
	if err != nil {
		t.Fatalf("pair: %v", err)
	}

	if acceptedID != "PEER-ID-123" {
		t.Errorf("expected accepted ID PEER-ID-123, got %s", acceptedID)
	}
	if acceptedName != "steam-deck" {
		t.Errorf("expected accepted name steam-deck, got %s", acceptedName)
	}
}

func TestPrimaryPairingFlowWithFakes(t *testing.T) {
	fakeClient := NewFakeClient(model.SyncConfig{
		Enabled: true,
		Mode:    model.SyncModePrimary,
	})
	fakeClient.SetDeviceID("PRIMARY-DEVICE-ID-AAAA-BBBB-CCCC-DDDD")
	fakeAdvertiser := NewFakeAdvertiser()

	var messages []string

	flow := NewPrimaryPairingFlow(PairingFlowConfig{
		SyncConfig: fakeClient.config,
		Advertiser: fakeAdvertiser,
		Client:     fakeClient,
		OnProgress: func(msg string) { messages = append(messages, msg) },
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		time.Sleep(100 * time.Millisecond)

		if !fakeAdvertiser.IsAdvertising() {
			t.Error("expected advertiser to be advertising")
			return
		}

		port := fakeAdvertiser.Port()
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		client := NewPairingClient()

		// The code is in the messages; parse it out. In a real flow we'd
		// read it from the screen.
		var code string
		for _, msg := range messages {
			if len(msg) > len("Pairing code: ") {
				var c string
				if _, err := fmt.Sscanf(msg, "Pairing code: %s", &c); err == nil {
					code = c
					break
				}
			}
		}

		if code == "" {
			t.Error("no pairing code found in messages")
			return
		}

		_, err := client.Pair(ctx, addr, code, "SECONDARY-DEVICE-ID", "steamdeck")
		if err != nil {
			t.Errorf("pair: %v", err)
		}
	}()

	result, _, err := flow.Run(ctx)
	if err != nil {
		t.Fatalf("primary flow: %v", err)
	}

	if result.PeerDeviceID != "SECONDARY-DEVICE-ID" {
		t.Errorf("expected peer ID SECONDARY-DEVICE-ID, got %s", result.PeerDeviceID)
	}

	added := fakeClient.AddedPeers()
	if len(added) != 1 {
		t.Fatalf("expected 1 added peer, got %d", len(added))
	}
	if added[0].ID != "SECONDARY-DEVICE-ID" {
		t.Errorf("expected added peer SECONDARY-DEVICE-ID, got %s", added[0].ID)
	}

	shared := fakeClient.SharedWith()
	if len(shared) != 1 || shared[0] != "SECONDARY-DEVICE-ID" {
		t.Errorf("expected shared with SECONDARY-DEVICE-ID, got %v", shared)
	}
}

func TestSecondaryPairingFlowWithFakes(t *testing.T) {
	primaryClient := NewFakeClient(model.SyncConfig{
		Enabled: true,
		Mode:    model.SyncModePrimary,
	})
	primaryClient.SetDeviceID("PRIMARY-DEVICE-ID")

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	code := "XY7K9M"
	server := NewPairingServer(code, "PRIMARY-DEVICE-ID", "desktop-kyaraben")
	if err := server.Start(listener); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer server.Stop()

	secondaryClient := NewFakeClient(model.SyncConfig{
		Enabled: true,
		Mode:    model.SyncModeSecondary,
	})
	secondaryClient.SetDeviceID("SECONDARY-DEVICE-ID")

	fakeBrowser := NewFakeBrowser()
	fakeBrowser.SetOffers([]PairingOffer{{
		Hostname:    "desktop",
		PairingAddr: fmt.Sprintf("127.0.0.1:%d", port),
	}})

	var messages []string
	flow := NewSecondaryPairingFlow(PairingFlowConfig{
		SyncConfig: secondaryClient.config,
		Browser:    fakeBrowser,
		Client:     secondaryClient,
		OnProgress: func(msg string) { messages = append(messages, msg) },
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := flow.Run(ctx, code)
	if err != nil {
		t.Fatalf("secondary flow: %v", err)
	}

	if result.PeerDeviceID != "PRIMARY-DEVICE-ID" {
		t.Errorf("expected peer ID PRIMARY-DEVICE-ID, got %s", result.PeerDeviceID)
	}
	if result.PeerName != "desktop-kyaraben" {
		t.Errorf("expected peer name desktop-kyaraben, got %s", result.PeerName)
	}

	added := secondaryClient.AddedPeers()
	if len(added) != 1 || added[0].ID != "PRIMARY-DEVICE-ID" {
		t.Errorf("expected added peer PRIMARY-DEVICE-ID, got %v", added)
	}

	shared := secondaryClient.SharedWith()
	if len(shared) != 1 || shared[0] != "PRIMARY-DEVICE-ID" {
		t.Errorf("expected shared with PRIMARY-DEVICE-ID, got %v", shared)
	}
}

func TestFakeAdvertiserStopsOnFlowComplete(t *testing.T) {
	fakeClient := NewFakeClient(model.SyncConfig{
		Enabled: true,
		Mode:    model.SyncModePrimary,
	})
	fakeClient.SetDeviceID("PRIMARY-ID")
	fakeAdvertiser := NewFakeAdvertiser()

	flow := NewPrimaryPairingFlow(PairingFlowConfig{
		SyncConfig: fakeClient.config,
		Advertiser: fakeAdvertiser,
		Client:     fakeClient,
		OnProgress: func(msg string) {},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		time.Sleep(100 * time.Millisecond)
		port := fakeAdvertiser.Port()
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		client := NewPairingClient()

		var code string
		for code == "" {
			time.Sleep(10 * time.Millisecond)
		}
		_ = code
		_ = client
		_ = addr
		cancel()
	}()

	_, _, _ = flow.Run(ctx)

	if fakeAdvertiser.IsAdvertising() {
		t.Error("expected advertiser to have stopped")
	}
}
