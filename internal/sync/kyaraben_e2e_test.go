package sync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/folders"
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

	device1 := newKyarabenInstance(t, "device1", 18384, 22000)
	device2 := newKyarabenInstance(t, "device2", 18385, 22001)

	t.Cleanup(func() {
		device1.stop()
		device2.stop()
	})

	if err := device1.generate(); err != nil {
		t.Fatalf("device1.generate: %v", err)
	}
	if err := device2.generate(); err != nil {
		t.Fatalf("device2.generate: %v", err)
	}

	if err := device1.writeKyarabenConfig(device2); err != nil {
		t.Fatalf("device1.writeKyarabenConfig: %v", err)
	}
	if err := device2.writeKyarabenConfig(device1); err != nil {
		t.Fatalf("device2.writeKyarabenConfig: %v", err)
	}

	if err := device1.start(ctx); err != nil {
		t.Fatalf("device1.start: %v", err)
	}
	if err := device2.start(ctx); err != nil {
		t.Fatalf("device2.start: %v", err)
	}

	if err := waitConnected(ctx, device1.testInstance, device2.deviceID); err != nil {
		t.Fatalf("waitConnected: %v", err)
	}

	t.Run("roms sync bidirectionally", func(t *testing.T) {
		romData := []byte("fake rom data")
		romPath := filepath.Join(device1.collection, "roms", "snes", "game.sfc")

		if err := os.MkdirAll(filepath.Dir(romPath), 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(romPath, romData, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		syncedPath := filepath.Join(device2.collection, "roms", "snes", "game.sfc")
		if err := waitForFile(ctx, syncedPath, romData); err != nil {
			t.Fatalf("ROM did not sync device1 → device2: %v", err)
		}

		romData2 := []byte("rom from device2")
		romPath2 := filepath.Join(device2.collection, "roms", "psx", "game.bin")

		if err := os.MkdirAll(filepath.Dir(romPath2), 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(romPath2, romData2, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		syncedPath2 := filepath.Join(device1.collection, "roms", "psx", "game.bin")
		if err := waitForFile(ctx, syncedPath2, romData2); err != nil {
			t.Fatalf("ROM did not sync device2 → device1: %v", err)
		}
	})

	t.Run("rom deletions sync", func(t *testing.T) {
		romData := []byte("rom to delete")
		romPath := filepath.Join(device1.collection, "roms", "snes", "deleteme.sfc")

		if err := os.WriteFile(romPath, romData, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		syncedPath := filepath.Join(device2.collection, "roms", "snes", "deleteme.sfc")
		if err := waitForFile(ctx, syncedPath, romData); err != nil {
			t.Fatalf("ROM did not sync before deletion test: %v", err)
		}

		if err := os.Remove(romPath); err != nil {
			t.Fatalf("Remove: %v", err)
		}

		if err := waitForFileDeletion(ctx, syncedPath); err != nil {
			t.Fatalf("ROM deletion did not sync: %v", err)
		}
	})

	t.Run("saves sync bidirectionally", func(t *testing.T) {
		saveData := []byte("save from device1")
		savePath := filepath.Join(device1.collection, "saves", "snes", "game.srm")

		if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(savePath, saveData, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		syncedPath := filepath.Join(device2.collection, "saves", "snes", "game.srm")
		if err := waitForFile(ctx, syncedPath, saveData); err != nil {
			t.Fatalf("save did not sync device1 → device2: %v", err)
		}

		updatedSave := []byte("save from device2")
		if err := os.WriteFile(syncedPath, updatedSave, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if err := waitForFile(ctx, savePath, updatedSave); err != nil {
			t.Fatalf("save did not sync device2 → device1: %v", err)
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
	collection string
	systems    []model.SystemID
	emulators  []folders.EmulatorInfo
}

func newKyarabenInstance(t *testing.T, name string, guiPort, listenPort int) *kyarabenInstance {
	t.Helper()

	inst := newTestInstance(t, name, guiPort, listenPort)
	collection := filepath.Join(filepath.Dir(inst.configDir), "emulation")

	systems := []model.SystemID{"snes", "psx"}
	emulators := []folders.EmulatorInfo{
		{ID: "retroarch:bsnes", UsesStatesDir: true},
		{ID: "duckstation", UsesStatesDir: true},
	}
	for _, sys := range systems {
		for _, category := range []string{"roms", "saves", "bios"} {
			dir := filepath.Join(collection, category, string(sys))
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("creating %s: %v", dir, err)
			}
		}
	}
	for _, emu := range emulators {
		if emu.UsesStatesDir {
			dir := filepath.Join(collection, "states", string(emu.ID))
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("creating %s: %v", dir, err)
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(collection, "screenshots"), 0755); err != nil {
		t.Fatalf("creating screenshots: %v", err)
	}

	return &kyarabenInstance{
		testInstance: inst,
		collection:   collection,
		systems:      systems,
		emulators:    emulators,
	}
}

func waitForFileDeletion(ctx context.Context, path string) error {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("file still exists: %s", path)
}

func (k *kyarabenInstance) writeKyarabenConfig(peer *kyarabenInstance) error {
	cfg := model.SyncConfig{
		Enabled: true,
		Syncthing: model.SyncthingConfig{
			ListenPort:    k.listenPort,
			DiscoveryPort: 21027,
			GUIPort:       k.guiPort,
			RelayEnabled:  false,
		},
	}

	gen := NewDefaultConfigGenerator(cfg, k.collection, k.systems, k.emulators, nil)
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
