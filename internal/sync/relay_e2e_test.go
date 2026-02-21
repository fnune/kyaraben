package sync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestRelayPairing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	if _, err := exec.LookPath("syncthing"); err != nil {
		t.Skip("syncthing not installed")
	}

	relayBinary := os.Getenv("KYARABEN_RELAY_BINARY")
	if relayBinary == "" {
		t.Skip("KYARABEN_RELAY_BINARY not set")
	}
	if _, err := os.Stat(relayBinary); os.IsNotExist(err) {
		t.Skip("relay binary not found at", relayBinary)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	relay := startTestRelayServer(t, relayBinary)
	defer relay.stop()

	primary := newTestInstance(t, "primary", 18390, 22010)
	secondary := newTestInstance(t, "secondary", 18391, 22011)

	t.Cleanup(func() {
		primary.stop()
		secondary.stop()
	})

	if err := primary.generate(); err != nil {
		t.Fatalf("primary.generate: %v", err)
	}
	if err := secondary.generate(); err != nil {
		t.Fatalf("secondary.generate: %v", err)
	}

	if err := writeTestConfig(primary); err != nil {
		t.Fatalf("writeTestConfig(primary): %v", err)
	}
	if err := writeTestConfig(secondary); err != nil {
		t.Fatalf("writeTestConfig(secondary): %v", err)
	}

	if err := primary.start(ctx); err != nil {
		t.Fatalf("primary.start: %v", err)
	}
	if err := secondary.start(ctx); err != nil {
		t.Fatalf("secondary.start: %v", err)
	}

	t.Run("relay pairing flow", func(t *testing.T) {
		relayClient, err := NewRelayClient([]string{relay.url})
		if err != nil {
			t.Fatalf("NewRelayClient: %v", err)
		}

		session, err := relayClient.CreateSession(ctx, primary.deviceID)
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		if len(session.Code) != 6 {
			t.Errorf("expected 6-char code, got %q", session.Code)
		}

		sessionInfo, err := relayClient.GetSession(ctx, session.Code)
		if err != nil {
			t.Fatalf("GetSession: %v", err)
		}
		if sessionInfo.DeviceID != primary.deviceID {
			t.Errorf("expected device ID %q, got %q", primary.deviceID, sessionInfo.DeviceID)
		}

		if err := relayClient.SubmitResponse(ctx, session.Code, secondary.deviceID); err != nil {
			t.Fatalf("SubmitResponse: %v", err)
		}

		resp, err := relayClient.GetResponse(ctx, session.Code)
		if err != nil {
			t.Fatalf("GetResponse: %v", err)
		}
		if !resp.Ready {
			t.Error("expected response to be ready")
		}
		if resp.DeviceID != secondary.deviceID {
			t.Errorf("expected device ID %q, got %q", secondary.deviceID, resp.DeviceID)
		}

		primaryAddr := fmt.Sprintf("tcp://127.0.0.1:%d", primary.listenPort)
		secondaryAddr := fmt.Sprintf("tcp://127.0.0.1:%d", secondary.listenPort)

		if err := primary.client.AddDeviceWithAddresses(ctx, secondary.deviceID, "secondary", []string{secondaryAddr}); err != nil {
			t.Fatalf("primary.AddDeviceWithAddresses: %v", err)
		}
		if err := secondary.client.AddDeviceWithAddresses(ctx, primary.deviceID, "primary", []string{primaryAddr}); err != nil {
			t.Fatalf("secondary.AddDeviceWithAddresses: %v", err)
		}

		if err := waitForConnection(ctx, primary, secondary.deviceID); err != nil {
			t.Fatalf("devices did not connect: %v", err)
		}

		if err := relayClient.DeleteSession(ctx, session.Code); err != nil {
			t.Errorf("DeleteSession: %v", err)
		}
	})
}

type testRelayServer struct {
	cmd  *exec.Cmd
	port int
	url  string
}

func startTestRelayServer(t *testing.T, binaryPath string) *testRelayServer {
	t.Helper()

	port := 19700

	cmd := exec.Command(binaryPath, "-addr", fmt.Sprintf(":%d", port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("starting relay server: %v", err)
	}

	server := &testRelayServer{
		cmd:  cmd,
		port: port,
		url:  fmt.Sprintf("http://localhost:%d", port),
	}

	if err := server.waitReady(); err != nil {
		_ = cmd.Process.Kill()
		t.Fatalf("relay server did not become ready: %v", err)
	}

	return server
}

func (s *testRelayServer) waitReady() error {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		client, err := NewRelayClient([]string{s.url})
		if err == nil && client != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("relay server did not become ready")
}

func (s *testRelayServer) stop() {
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- s.cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = s.cmd.Process.Kill()
			<-done
		}
	}
}

func writeTestConfig(inst *testInstance) error {
	cfg := &SyncthingXMLConfig{
		Version: 51,
		Devices: []XMLDevice{
			{
				ID:          inst.deviceID,
				Name:        inst.name,
				Compression: "metadata",
				Addresses:   []string{"dynamic"},
			},
		},
		Options: XMLOptions{
			ListenAddresses:       []string{fmt.Sprintf("tcp://127.0.0.1:%d", inst.listenPort)},
			GlobalAnnounceEnabled: false,
			LocalAnnounceEnabled:  false,
			RelaysEnabled:         false,
			AutoUpgradeIntervalH:  0,
		},
		GUI: XMLGUI{
			Enabled: true,
			Address: fmt.Sprintf("127.0.0.1:%d", inst.guiPort),
			APIKey:  inst.apiKey,
		},
	}
	return inst.writeConfigXML(cfg)
}

func waitForConnection(ctx context.Context, inst *testInstance, peerID string) error {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		conns, err := inst.client.GetConnections(ctx)
		if err == nil {
			if conn, ok := conns[peerID]; ok && conn.Connected {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("peer did not connect")
}
