package sync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/model"
)

func TestKyarabenSync(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	if _, err := exec.LookPath("syncthing"); err != nil {
		t.Skip("syncthing not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	primary := newKyarabenInstance(t, "primary", model.SyncModePrimary, 18384, 22000)
	secondary := newKyarabenInstance(t, "secondary", model.SyncModeSecondary, 18385, 22001)

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

	if err := primary.writeKyarabenConfig(secondary); err != nil {
		t.Fatalf("primary.writeKyarabenConfig: %v", err)
	}
	if err := secondary.writeKyarabenConfig(primary); err != nil {
		t.Fatalf("secondary.writeKyarabenConfig: %v", err)
	}

	if err := primary.start(ctx); err != nil {
		t.Fatalf("primary.start: %v", err)
	}
	if err := secondary.start(ctx); err != nil {
		t.Fatalf("secondary.start: %v", err)
	}

	if err := waitConnected(ctx, primary.testInstance, secondary.deviceID); err != nil {
		t.Fatalf("waitConnected: %v", err)
	}

	t.Run("roms sync primary to secondary", func(t *testing.T) {
		romData := []byte("fake rom data")
		romPath := filepath.Join(primary.userStore, "roms", "snes", "game.sfc")

		if err := os.MkdirAll(filepath.Dir(romPath), 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(romPath, romData, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		syncedPath := filepath.Join(secondary.userStore, "roms", "snes", "game.sfc")
		if err := waitForFile(ctx, syncedPath, romData); err != nil {
			t.Fatalf("ROM did not sync to secondary: %v", err)
		}
	})

	t.Run("roms do not sync secondary to primary", func(t *testing.T) {
		romData := []byte("rom from secondary")
		romPath := filepath.Join(secondary.userStore, "roms", "psx", "game.bin")

		if err := os.MkdirAll(filepath.Dir(romPath), 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(romPath, romData, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		primaryPath := filepath.Join(primary.userStore, "roms", "psx", "game.bin")
		if err := assertFileNeverAppears(ctx, primaryPath, 5*time.Second); err != nil {
			t.Fatalf("ROM incorrectly synced to primary: %v", err)
		}
	})

	t.Run("saves sync bidirectionally", func(t *testing.T) {
		saveData := []byte("save from primary")
		savePath := filepath.Join(primary.userStore, "saves", "snes", "game.srm")

		if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(savePath, saveData, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		syncedPath := filepath.Join(secondary.userStore, "saves", "snes", "game.srm")
		if err := waitForFile(ctx, syncedPath, saveData); err != nil {
			t.Fatalf("save did not sync primary → secondary: %v", err)
		}

		updatedSave := []byte("save from secondary")
		if err := os.WriteFile(syncedPath, updatedSave, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if err := waitForFile(ctx, savePath, updatedSave); err != nil {
			t.Fatalf("save did not sync secondary → primary: %v", err)
		}
	})
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

type kyarabenInstance struct {
	*testInstance
	mode      model.SyncMode
	userStore string
	systems   []model.SystemID
}

func newKyarabenInstance(t *testing.T, name string, mode model.SyncMode, guiPort, listenPort int) *kyarabenInstance {
	t.Helper()

	inst := newTestInstance(t, name, guiPort, listenPort)
	userStore := filepath.Join(filepath.Dir(inst.configDir), "emulation")

	systems := []model.SystemID{"snes", "psx"}
	for _, sys := range systems {
		for _, category := range []string{"roms", "saves", "states", "bios"} {
			dir := filepath.Join(userStore, category, string(sys))
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("creating %s: %v", dir, err)
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(userStore, "screenshots"), 0755); err != nil {
		t.Fatalf("creating screenshots: %v", err)
	}

	return &kyarabenInstance{
		testInstance: inst,
		mode:         mode,
		userStore:    userStore,
		systems:      systems,
	}
}

func assertFileNeverAppears(ctx context.Context, path string, duration time.Duration) error {
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("file appeared: %s", path)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return nil
}

func (k *kyarabenInstance) writeKyarabenConfig(peer *kyarabenInstance) error {
	cfg := model.SyncConfig{
		Enabled: true,
		Mode:    k.mode,
		Syncthing: model.SyncthingConfig{
			ListenPort:    k.listenPort,
			DiscoveryPort: 21027,
			GUIPort:       k.guiPort,
			RelayEnabled:  false,
		},
	}

	gen := NewDefaultConfigGenerator(cfg, k.userStore, k.systems)
	gen.SetDeviceID(k.deviceID)
	gen.SetAPIKey(k.apiKey)

	xmlCfg, err := gen.Generate()
	if err != nil {
		return err
	}

	xmlCfg.Devices = append(xmlCfg.Devices, XMLDevice{
		ID:          peer.deviceID,
		Name:        peer.name,
		Compression: "metadata",
		Addresses:   []string{fmt.Sprintf("tcp://127.0.0.1:%d", peer.listenPort)},
	})

	for i := range xmlCfg.Folders {
		xmlCfg.Folders[i].Devices = append(xmlCfg.Folders[i].Devices, XMLFolderDevice{ID: peer.deviceID})
	}

	xmlCfg.Options.GlobalAnnounceEnabled = false
	xmlCfg.Options.LocalAnnounceEnabled = false

	return k.writeConfigXML(xmlCfg)
}
