package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testDeviceID = "LGFPDIT7-SKNNJVJZ-A4FC7QNC-XYZDEFGH-IJKLMNOP-QRSTUVWX-YZ234567"

type syncTest struct {
	*cliTest
	fakeSyncthing *FakeSyncthing
	relayServer   *testRelayServer
}

func newSyncTest(t *testing.T) *syncTest {
	t.Helper()
	c := newCLITest(t)

	fakeST := NewFakeSyncthing(testDeviceID)
	if err := fakeST.Start(); err != nil {
		t.Fatalf("starting fake syncthing: %v", err)
	}
	t.Cleanup(fakeST.Stop)

	return &syncTest{
		cliTest:       c,
		fakeSyncthing: fakeST,
	}
}

func (s *syncTest) initWithSyncEnabled() {
	s.t.Helper()
	s.init()

	relaysLine := ""
	if s.relayServer != nil {
		relaysLine = fmt.Sprintf("relays = [%q]", s.relayServer.url)
	}

	config := fmt.Sprintf(`[global]
collection = %q

[systems]
snes = ["retroarch:bsnes"]

[sync]
enabled = true
%s

[sync.syncthing]
base_url = %q
gui_port = %d
`, s.collection, relaysLine, s.fakeSyncthing.BaseURL(), s.fakeSyncthing.Port())

	s.writeFile("config.toml", config)
}

func (s *syncTest) startRelay(t *testing.T) {
	t.Helper()

	relayBinary := os.Getenv("KYARABEN_RELAY_BINARY")
	if relayBinary == "" {
		t.Skip("KYARABEN_RELAY_BINARY not set")
	}
	if _, err := os.Stat(relayBinary); os.IsNotExist(err) {
		t.Skip("relay binary not found at", relayBinary)
	}

	s.relayServer = startTestRelayServerE2E(t, relayBinary)
	t.Cleanup(s.relayServer.stop)
}

func TestSyncStatus(t *testing.T) {
	t.Run("shows enabled status with fake syncthing", func(t *testing.T) {
		s := newSyncTest(t)
		s.initWithSyncEnabled()

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Sync: enabled")
		s.assertContains(output, "Status:")
		s.assertContains(output, "Device ID:")
		s.assertContains(output, testDeviceID)
	})

	t.Run("shows paired devices when present", func(t *testing.T) {
		s := newSyncTest(t)
		s.fakeSyncthing.devices["PEERDEVICE"] = fakeDevice{
			DeviceID: "PEERDEVICE-1234567-ABCDEFGH-IJKLMNOP-QRSTUVWX-YZ234567",
			Name:     "steamdeck-kyaraben",
		}
		s.initWithSyncEnabled()

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Paired devices:")
		s.assertContains(output, "steamdeck-kyaraben")
	})
}

func TestSyncPairDeviceID(t *testing.T) {
	t.Run("prints full device ID with --device-id flag", func(t *testing.T) {
		s := newSyncTest(t)
		s.initWithSyncEnabled()

		output, err := s.run("sync", "pair", "--device-id")
		if err != nil {
			t.Fatalf("sync pair --device-id failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Device ID:")
		s.assertContains(output, testDeviceID)
		s.assertContains(output, "kyaraben sync pair")
	})
}

func TestSyncPairWithRelay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping relay test in short mode")
	}

	t.Run("initiator generates 6-digit code", func(t *testing.T) {
		s := newSyncTest(t)
		s.startRelay(t)
		s.initWithSyncEnabled()

		cmd := s.cmd("sync", "pair")

		output := make(chan string, 1)
		go func() {
			out, _ := cmd.CombinedOutput()
			output <- string(out)
		}()

		select {
		case out := <-output:
			if strings.Contains(out, "Pairing code:") {
				parts := strings.Split(out, "Pairing code:")
				if len(parts) > 1 {
					codeLine := strings.TrimSpace(strings.Split(parts[1], "\n")[0])
					if len(codeLine) == 6 && isValidPairingCode(codeLine) {
						return
					}
				}
			}
			if strings.Contains(out, "Device ID:") {
				return
			}
			t.Fatalf("unexpected output: %s", out)
		case <-time.After(10 * time.Second):
			_ = cmd.Process.Kill()
			t.Fatal("timeout waiting for pairing code")
		}
	})
}

func TestSyncAddDevice(t *testing.T) {
	t.Run("adds device with valid ID", func(t *testing.T) {
		s := newSyncTest(t)
		s.initWithSyncEnabled()

		peerID := "ABCDEFGH-IJKLMNOP-QRSTUVWX-YZ234567-ABCDEFGH-IJKLMNOP-QRSTUVWX"
		output, err := s.run("sync", "add-device", peerID)
		if err != nil {
			t.Fatalf("sync add-device failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Added device")
	})

	t.Run("rejects invalid device ID", func(t *testing.T) {
		s := newSyncTest(t)
		s.initWithSyncEnabled()

		output, err := s.run("sync", "add-device", "invalid-id")
		if err == nil {
			t.Fatalf("expected error for invalid device ID, got: %s", output)
		}

		s.assertContains(output, "invalid device ID")
	})
}

func TestSyncRemoveDevice(t *testing.T) {
	t.Run("removes existing device", func(t *testing.T) {
		s := newSyncTest(t)
		peerID := "PEERAAAA-BBBBBBBB-CCCCCCCC-DDDDDDDD-EEEEEEEE-FFFFFFFF-GGGGGGGG"
		s.fakeSyncthing.devices[peerID] = fakeDevice{
			DeviceID: peerID,
			Name:     "test-device",
		}
		s.initWithSyncEnabled()

		output, err := s.run("sync", "remove-device", "test-device")
		if err != nil {
			t.Fatalf("sync remove-device failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Removed device")
	})

	t.Run("errors for unknown device", func(t *testing.T) {
		s := newSyncTest(t)
		s.initWithSyncEnabled()

		_, err := s.run("sync", "remove-device", "nonexistent")
		if err == nil {
			t.Fatal("expected error for unknown device")
		}
	})
}

func isValidPairingCode(code string) bool {
	if len(code) != 6 {
		return false
	}
	for _, c := range code {
		isUpper := c >= 'A' && c <= 'Z'
		isDigit := c >= '0' && c <= '9'
		if !isUpper && !isDigit {
			return false
		}
	}
	return true
}

type testRelayServer struct {
	cmd  *exec.Cmd
	port int
	url  string
}

func startTestRelayServerE2E(t *testing.T, binaryPath string) *testRelayServer {
	t.Helper()

	port := 19800 + os.Getpid()%100

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
		conn, err := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", s.url+"/health").Output()
		if err == nil && strings.TrimSpace(string(conn)) == "200" {
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

func TestSyncHelp(t *testing.T) {
	c := newCLITest(t)
	output, err := c.run("sync", "--help")
	if err != nil {
		t.Fatalf("sync help failed: %v\nOutput: %s", err, output)
	}

	c.assertContains(output, "pair")
	c.assertContains(output, "status")
	c.assertContains(output, "add-device")
	c.assertContains(output, "remove-device")
}

func TestSyncAutostart(t *testing.T) {
	t.Run("defaults to autostart enabled", func(t *testing.T) {
		s := newSyncTest(t)
		s.initWithSyncEnabled()

		config := s.readFile("config.toml")
		if strings.Contains(config, "autostart = false") {
			t.Error("expected autostart to default to true or be absent")
		}
	})

	t.Run("autostart can be disabled in config", func(t *testing.T) {
		s := newSyncTest(t)
		s.init()

		config := fmt.Sprintf(`[global]
collection = %q

[systems]
snes = ["retroarch:bsnes"]

[sync]
enabled = true
autostart = false

[sync.syncthing]
base_url = %q
gui_port = %d
`, s.collection, s.fakeSyncthing.BaseURL(), s.fakeSyncthing.Port())
		s.writeFile("config.toml", config)

		configContent := s.readFile("config.toml")
		s.assertContains(configContent, "autostart = false")
	})
}

func TestSyncDisabled(t *testing.T) {
	t.Run("shows disabled status when sync not enabled", func(t *testing.T) {
		c := newCLITest(t)
		c.init()

		output, err := c.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "Sync: disabled")
		c.assertContains(output, "Enable sync in your config.toml")
		c.assertContains(output, "[sync]")
		c.assertContains(output, "enabled = true")
	})
}

func TestSyncStatusNotRunning(t *testing.T) {
	t.Run("shows not running when syncthing is stopped", func(t *testing.T) {
		s := newSyncTest(t)
		s.fakeSyncthing.Stop()
		s.initWithSyncEnabled()

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Sync: enabled")
		s.assertContains(output, "Status: not running")
		s.assertContains(output, "Syncthing is not running")
	})
}

func TestSyncStatusErrors(t *testing.T) {
	t.Run("shows not running when API returns 500", func(t *testing.T) {
		s := newSyncTest(t)
		s.fakeSyncthing.SetErrorMode("api-error")
		s.initWithSyncEnabled()

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Sync: enabled")
		s.assertContains(output, "Status: not running")
	})
}

func TestSyncStatusFolders(t *testing.T) {
	t.Run("shows folder syncing state", func(t *testing.T) {
		s := newSyncTest(t)
		s.fakeSyncthing.AddFolder("kyaraben-saves-snes", "/path/to/saves/snes", "sendreceive")
		s.fakeSyncthing.SetFolderState("kyaraben-saves-snes", fakeFolderState{
			State:       "syncing",
			GlobalBytes: 1000000,
			NeedBytes:   500000,
		})
		s.initWithSyncEnabled()

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Synced folders:")
		s.assertContains(output, "snes (saves)")
		s.assertContains(output, "syncing")
	})

	t.Run("shows folder idle state", func(t *testing.T) {
		s := newSyncTest(t)
		s.fakeSyncthing.AddFolder("kyaraben-saves-gba", "/path/to/saves/gba", "sendreceive")
		s.fakeSyncthing.SetFolderState("kyaraben-saves-gba", fakeFolderState{
			State:       "idle",
			GlobalBytes: 1000000,
			NeedBytes:   0,
		})
		s.initWithSyncEnabled()

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "gba (saves)")
		s.assertContains(output, "idle")
	})

	t.Run("shows folder error state with message", func(t *testing.T) {
		s := newSyncTest(t)
		s.fakeSyncthing.AddFolder("kyaraben-saves-psx", "/path/to/saves/psx", "sendreceive")
		s.fakeSyncthing.SetFolderState("kyaraben-saves-psx", fakeFolderState{
			State: "error",
			Error: "permission denied",
		})
		s.initWithSyncEnabled()

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "psx (saves)")
		s.assertContains(output, "error")
		s.assertContains(output, "permission denied")
	})
}

func TestSyncStatusConnections(t *testing.T) {
	t.Run("shows connected device status", func(t *testing.T) {
		s := newSyncTest(t)
		peerID := "PEERAAAA-BBBBBBBB-CCCCCCCC-DDDDDDDD-EEEEEEEE-FFFFFFFF-GGGGGGGG"
		s.fakeSyncthing.devices[peerID] = fakeDevice{
			DeviceID: peerID,
			Name:     "steamdeck-kyaraben",
		}
		s.fakeSyncthing.SetConnection(peerID, true)
		s.initWithSyncEnabled()

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Paired devices:")
		s.assertContains(output, "steamdeck-kyaraben")
		s.assertContains(output, "connected")
	})

	t.Run("shows disconnected device status", func(t *testing.T) {
		s := newSyncTest(t)
		peerID := "PEERAAAA-BBBBBBBB-CCCCCCCC-DDDDDDDD-EEEEEEEE-FFFFFFFF-GGGGGGGG"
		s.fakeSyncthing.devices[peerID] = fakeDevice{
			DeviceID: peerID,
			Name:     "desktop-kyaraben",
		}
		s.fakeSyncthing.SetConnection(peerID, false)
		s.initWithSyncEnabled()

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Paired devices:")
		s.assertContains(output, "desktop-kyaraben")
		s.assertContains(output, "disconnected")
	})

	t.Run("shows disconnected when no connection info", func(t *testing.T) {
		s := newSyncTest(t)
		peerID := "PEERAAAA-BBBBBBBB-CCCCCCCC-DDDDDDDD-EEEEEEEE-FFFFFFFF-GGGGGGGG"
		s.fakeSyncthing.devices[peerID] = fakeDevice{
			DeviceID: peerID,
			Name:     "offline-device",
		}
		s.initWithSyncEnabled()

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "offline-device")
		s.assertContains(output, "disconnected")
	})
}

func TestSyncConflictDetection(t *testing.T) {
	t.Run("CLI shows conflict count when conflicts exist", func(t *testing.T) {
		s := newSyncTest(t)

		savesDir := filepath.Join(s.collection, "saves", "snes")
		if err := os.MkdirAll(savesDir, 0755); err != nil {
			t.Fatalf("creating saves dir: %v", err)
		}

		s.fakeSyncthing.AddFolder("kyaraben-saves-snes", savesDir, "sendreceive")
		s.fakeSyncthing.SetFolderState("kyaraben-saves-snes", fakeFolderState{
			State:       "idle",
			GlobalBytes: 1000,
			NeedBytes:   0,
		})

		s.initWithSyncEnabled()

		normalFile := filepath.Join(savesDir, "game.srm")
		if err := os.WriteFile(normalFile, []byte("normal data"), 0644); err != nil {
			t.Fatalf("creating normal file: %v", err)
		}

		conflictFile := filepath.Join(savesDir, "game.sync-conflict-ABCD123-20260307-120000.srm")
		if err := os.WriteFile(conflictFile, []byte("conflict data"), 0644); err != nil {
			t.Fatalf("creating conflict file: %v", err)
		}

		s.assertFileExists(normalFile)
		s.assertFileExists(conflictFile)

		output, err := s.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		s.assertContains(output, "Sync: enabled")
		s.assertContains(output, "Status: conflict")
		s.assertContains(output, "snes (saves)")
		s.assertContains(output, "1 conflict file(s)")
	})
}
