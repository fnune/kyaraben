package sync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

		romVersions := filepath.Join(device2.collection, "roms", "snes", ".stversions")
		if found, err := versionedFileExists(romVersions, "deleteme"); err != nil {
			t.Fatalf("checking rom versions: %v", err)
		} else if found {
			t.Errorf("ROM deletion was versioned, but ROM folders must not use versioning")
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

	t.Run("deleted saves are recoverable from peer versioning", func(t *testing.T) {
		saveData := []byte("precious save")
		savePath := filepath.Join(device1.collection, "saves", "psx", "recover.srm")

		if err := os.WriteFile(savePath, saveData, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		syncedPath := filepath.Join(device2.collection, "saves", "psx", "recover.srm")
		if err := waitForFile(ctx, syncedPath, saveData); err != nil {
			t.Fatalf("save did not sync before deletion: %v", err)
		}

		if err := os.Remove(savePath); err != nil {
			t.Fatalf("Remove: %v", err)
		}
		if err := waitForFileDeletion(ctx, syncedPath); err != nil {
			t.Fatalf("save deletion did not sync: %v", err)
		}

		versionsDir := filepath.Join(device2.collection, "saves", "psx", ".stversions")
		if err := waitForVersionedFile(ctx, versionsDir, "recover"); err != nil {
			t.Fatalf("deleted save was not preserved by versioning on peer: %v", err)
		}
	})
}

func TestKyarabenIgnoreDeleteAsymmetric(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	if _, err := exec.LookPath("syncthing"); err != nil {
		t.Skip("syncthing not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	deck := newKyarabenInstance(t, "deck", 18386, 22002)
	bilbo := newKyarabenInstance(t, "bilbo", 18387, 22003)
	deck.ignoreDeleteRoms = true
	bilbo.ignoreDeleteRoms = false

	t.Cleanup(func() {
		deck.stop()
		bilbo.stop()
	})

	for _, d := range []*kyarabenInstance{deck, bilbo} {
		if err := d.generate(); err != nil {
			t.Fatalf("%s.generate: %v", d.name, err)
		}
	}
	if err := deck.writeKyarabenConfig(bilbo); err != nil {
		t.Fatalf("deck.writeKyarabenConfig: %v", err)
	}
	if err := bilbo.writeKyarabenConfig(deck); err != nil {
		t.Fatalf("bilbo.writeKyarabenConfig: %v", err)
	}
	for _, d := range []*kyarabenInstance{deck, bilbo} {
		if err := d.start(ctx); err != nil {
			t.Fatalf("%s.start: %v", d.name, err)
		}
	}
	if err := waitConnected(ctx, deck.testInstance, bilbo.deviceID); err != nil {
		t.Fatalf("waitConnected: %v", err)
	}

	const observeWindow = 25 * time.Second

	t.Run("sanity: device with ignoreDelete keeps a ROM deleted elsewhere", func(t *testing.T) {
		data := []byte("sanity rom")
		bilboPath := filepath.Join(bilbo.collection, "roms", "snes", "sanity.sfc")
		deckPath := filepath.Join(deck.collection, "roms", "snes", "sanity.sfc")

		if err := os.WriteFile(bilboPath, data, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		if err := waitForFile(ctx, deckPath, data); err != nil {
			t.Fatalf("ROM did not sync bilbo -> deck: %v", err)
		}

		if err := os.Remove(bilboPath); err != nil {
			t.Fatalf("Remove: %v", err)
		}

		propagated := deletionPropagated(ctx, deckPath, observeWindow)
		t.Logf("SANITY: bilbo(ignoreDelete=off) deleted, deck(ignoreDelete=on) propagated=%v (want false)", propagated)
		if propagated {
			t.Errorf("ignoreDelete not effective: deck deleted a ROM it should have kept; harness/wiring is wrong, experiment result below is invalid")
		}
	})

	t.Run("question: delete on ignoreDelete device, observe the device without it", func(t *testing.T) {
		data := []byte("question rom")
		deckPath := filepath.Join(deck.collection, "roms", "snes", "question.sfc")
		bilboPath := filepath.Join(bilbo.collection, "roms", "snes", "question.sfc")

		if err := os.WriteFile(deckPath, data, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		if err := waitForFile(ctx, bilboPath, data); err != nil {
			t.Fatalf("ROM did not sync deck -> bilbo: %v", err)
		}

		if err := os.Remove(deckPath); err != nil {
			t.Fatalf("Remove: %v", err)
		}

		propagated := deletionPropagated(ctx, bilboPath, observeWindow)
		if !propagated {
			t.Errorf("expected bilbo (ignoreDelete=off) to lose a ROM deleted on the deck (ignoreDelete=on), but it was kept; syncthing deleter-side behavior may have changed")
		}
	})
}

func deletionPropagated(ctx context.Context, path string, window time.Duration) bool {
	deadline := time.Now().Add(window)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(500 * time.Millisecond):
		}
	}
	return false
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
	collection       string
	systems          []model.SystemID
	emulators        []folders.EmulatorInfo
	ignoreDeleteRoms bool
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

func versionedFileExists(dir, prefix string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			return true, nil
		}
	}
	return false, nil
}

func waitForVersionedFile(ctx context.Context, dir, prefix string) error {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		found, err := versionedFileExists(dir, prefix)
		if err != nil {
			return err
		}
		if found {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("no versioned file with prefix %q in %s", prefix, dir)
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
	folderRequests := gen.FolderCreateRequests()

	deviceRefs := []XMLFolderDevice{
		{ID: k.deviceID},
		{ID: peer.deviceID},
	}

	xmlFolders := make([]XMLFolder, len(folderRequests))
	for i, req := range folderRequests {
		folder := XMLFolder{
			ID:               req.ID,
			Label:            req.Label,
			Path:             req.Path,
			Type:             FolderType(req.Type),
			Devices:          deviceRefs,
			FSWatcherEnabled: true,
			IgnorePerms:      true,
		}
		if req.Versioning != nil {
			params := make([]XMLVersioningParam, 0, len(req.Versioning.Params))
			for key, val := range req.Versioning.Params {
				params = append(params, XMLVersioningParam{Key: key, Val: val})
			}
			folder.Versioning = XMLVersioning{Type: req.Versioning.Type, Params: params}
		}
		if k.ignoreDeleteRoms && strings.HasPrefix(req.ID, "kyaraben-roms-") {
			folder.IgnoreDelete = true
		}
		xmlFolders[i] = folder
	}

	xmlCfg := &SyncthingXMLConfig{
		Version: 37,
		Folders: xmlFolders,
		Devices: []XMLDevice{
			{
				ID:          k.deviceID,
				Name:        "this-device",
				Compression: "metadata",
			},
			{
				ID:          peer.deviceID,
				Name:        peer.name,
				Compression: "metadata",
				Addresses:   []string{fmt.Sprintf("tcp://127.0.0.1:%d", peer.listenPort)},
			},
		},
		GUI: XMLGUI{
			Enabled: true,
			Address: fmt.Sprintf("127.0.0.1:%d", k.guiPort),
			APIKey:  k.apiKey,
			Theme:   "default",
		},
		Options: XMLOptions{
			ListenAddresses: []string{
				fmt.Sprintf("tcp://0.0.0.0:%d", k.listenPort),
				fmt.Sprintf("quic://0.0.0.0:%d", k.listenPort),
			},
			GlobalAnnounceEnabled: false,
			LocalAnnounceEnabled:  false,
			LocalAnnouncePort:     21027,
			URAccepted:            -1,
			AutoUpgradeIntervalH:  0,
		},
		Defaults: XMLDefaults{
			Folder: XMLDefaultFolder{
				Path: k.collection,
			},
		},
	}

	return k.writeConfigXML(xmlCfg)
}
