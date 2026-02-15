package sync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestSyncBetweenInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	if _, err := exec.LookPath("syncthing"); err != nil {
		t.Skip("syncthing not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	inst1 := newTestInstance(t, "device1", 18384, 22000)
	inst2 := newTestInstance(t, "device2", 18385, 22001)

	t.Cleanup(func() {
		inst1.stop()
		inst2.stop()
	})

	if err := inst1.generate(); err != nil {
		t.Fatalf("inst1.generate: %v", err)
	}
	if err := inst2.generate(); err != nil {
		t.Fatalf("inst2.generate: %v", err)
	}

	if err := inst1.writeConfig(inst2.deviceID, inst2.listenPort); err != nil {
		t.Fatalf("inst1.writeConfig: %v", err)
	}
	if err := inst2.writeConfig(inst1.deviceID, inst1.listenPort); err != nil {
		t.Fatalf("inst2.writeConfig: %v", err)
	}

	if err := inst1.start(ctx); err != nil {
		t.Fatalf("inst1.start: %v", err)
	}
	if err := inst2.start(ctx); err != nil {
		t.Fatalf("inst2.start: %v", err)
	}

	if err := waitConnected(ctx, inst1, inst2.deviceID); err != nil {
		t.Fatalf("waitConnected: %v", err)
	}

	testContent := []byte("hello from device1")
	testFile := filepath.Join(inst1.syncDir, "test.txt")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	syncedFile := filepath.Join(inst2.syncDir, "test.txt")
	if err := waitForFile(ctx, syncedFile, testContent); err != nil {
		t.Fatalf("waitForFile: %v", err)
	}
}

func waitConnected(ctx context.Context, inst *testInstance, peerID string) error {
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

func waitForFile(ctx context.Context, path string, expectedContent []byte) error {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		content, err := os.ReadFile(path)
		if err == nil && string(content) == string(expectedContent) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("file did not sync")
}
