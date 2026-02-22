package sync

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/model"
)

func TestInitiatorPairingFlowAcceptsDevice(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{
		Enabled: true,
	})
	client.SetDeviceID("INITIATOR-ID")

	var messages []string
	flow := NewInitiatorPairingFlow(PairingFlowConfig{
		SyncConfig: client.config,
		Client:     client,
		OnProgress: func(msg string) {
			messages = append(messages, msg)
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct {
		result *PairResult
		err    error
	}, 1)

	go func() {
		result, err := flow.Run(ctx)
		done <- struct {
			result *PairResult
			err    error
		}{result, err}
	}()

	time.Sleep(100 * time.Millisecond)

	client.SetPendingDevices([]PendingDevice{
		{DeviceID: "JOINER-ID", Name: "steamdeck", Address: "192.168.1.100:22000"},
	})

	select {
	case out := <-done:
		if out.err != nil {
			t.Fatalf("unexpected error: %v", out.err)
		}
		if out.result.PeerDeviceID != "JOINER-ID" {
			t.Errorf("expected PeerDeviceID JOINER-ID, got %s", out.result.PeerDeviceID)
		}
		if out.result.PeerName != "steamdeck" {
			t.Errorf("expected PeerName steamdeck, got %s", out.result.PeerName)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for pairing")
	}

	addedPeers := client.AddedPeers()
	if len(addedPeers) != 1 {
		t.Fatalf("expected 1 added peer, got %d", len(addedPeers))
	}
	if addedPeers[0].ID != "JOINER-ID" {
		t.Errorf("expected added peer JOINER-ID, got %s", addedPeers[0].ID)
	}

	sharedWith := client.SharedWith()
	if len(sharedWith) != 1 || sharedWith[0] != "JOINER-ID" {
		t.Errorf("expected folders shared with JOINER-ID, got %v", sharedWith)
	}
}

func TestJoinerPairingFlowConnectsToPeer(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{
		Enabled: true,
	})
	client.SetDeviceID("JOINER-ID")

	var messages []string
	flow := NewJoinerPairingFlow(PairingFlowConfig{
		SyncConfig: client.config,
		Client:     client,
		OnProgress: func(msg string) {
			messages = append(messages, msg)
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct {
		result *PairResult
		err    error
	}, 1)

	go func() {
		result, err := flow.Run(ctx, "PEER-ID")
		done <- struct {
			result *PairResult
			err    error
		}{result, err}
	}()

	time.Sleep(100 * time.Millisecond)

	client.SetConfiguredDevice("PEER-ID", "feanor")
	client.SetConnection("PEER-ID", ConnectionInfo{Connected: true})

	select {
	case out := <-done:
		if out.err != nil {
			t.Fatalf("unexpected error: %v", out.err)
		}
		if out.result.PeerDeviceID != "PEER-ID" {
			t.Errorf("expected PeerDeviceID PEER-ID, got %s", out.result.PeerDeviceID)
		}
		if out.result.PeerName != "feanor" {
			t.Errorf("expected PeerName feanor, got %s", out.result.PeerName)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for pairing")
	}

	sharedWith := client.SharedWith()
	if len(sharedWith) != 1 || sharedWith[0] != "PEER-ID" {
		t.Errorf("expected folders shared with PEER-ID, got %v", sharedWith)
	}
}

func TestJoinerPairingFlowTimesOut(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{
		Enabled: true,
	})
	client.SetDeviceID("JOINER-ID")

	flow := NewJoinerPairingFlow(PairingFlowConfig{
		SyncConfig: client.config,
		Client:     client,
		OnProgress: func(msg string) {},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := flow.Run(ctx, "PEER-ID")
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestJoinerPairingFlowDetectsDeviceRemoval(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{
		Enabled: true,
	})
	client.SetDeviceID("JOINER-ID")

	flow := NewJoinerPairingFlow(PairingFlowConfig{
		SyncConfig: client.config,
		Client:     client,
		OnProgress: func(msg string) {},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		_, err := flow.Run(ctx, "PEER-ID")
		done <- err
	}()

	time.Sleep(100 * time.Millisecond)

	client.SetConfiguredDevice("PEER-ID", "feanor")

	time.Sleep(100 * time.Millisecond)

	client.RemoveConfiguredDevice("PEER-ID")

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error when device was removed")
		}
		if !strings.Contains(err.Error(), "device was removed") {
			t.Errorf("expected 'device was removed' error, got: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out")
	}
}

func TestTruncateDeviceID(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"ABC", "ABC"},
		{"ABCDEFG", "ABCDEFG"},
		{"ABCDEFGH", "ABCDEFG..."},
		{"ABCDEFGHIJKLMNOP", "ABCDEFG..."},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			result := truncateDeviceID(tt.id)
			if result != tt.expected {
				t.Errorf("truncateDeviceID(%q) = %q, want %q", tt.id, result, tt.expected)
			}
		})
	}
}

func TestIsRelayCode(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"ABC123", true},
		{"ABCDEF", true},
		{"123456", true},
		{"A1B2C3", true},
		{"abc123", false},
		{"ABCDE", false},
		{"ABCDEFG", false},
		{"ABC-12", false},
		{"", false},
		{"LGFPDIT7-SKNNJVJZ", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsRelayCode(tt.input)
			if result != tt.expected {
				t.Errorf("IsRelayCode(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
