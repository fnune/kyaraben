package guestapp

import (
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"
)

func testDefaults() Config {
	return Config{
		Service: ServiceConfig{
			Autostart: true,
		},
		PathMappings: PathMappings{
			Saves: map[string]string{
				"nes": "Saves/NES",
				"gba": "Saves/GBA",
			},
			ROMs: map[string]string{
				"nes": "Roms/NES",
			},
		},
	}
}

func TestConfigStoreLoadCreatesFile(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "/data")
	cfg, err := store.Load(testDefaults())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !cfg.Service.Autostart {
		t.Error("expected Autostart=true from defaults")
	}

	if _, err := fs.Stat("/data/config.toml"); err != nil {
		t.Errorf("config file should be created: %v", err)
	}
}

func TestConfigStoreLoadReadsExisting(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data/config.toml": `[service]
autostart = false
sync_states = true

[saves]
nes = "Custom/NES"
`,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "/data")
	cfg, err := store.Load(testDefaults())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Service.Autostart {
		t.Error("expected Autostart=false from file")
	}
	if !cfg.Service.SyncStates {
		t.Error("expected SyncStates=true from file")
	}
	if cfg.Saves["nes"] != "Custom/NES" {
		t.Errorf("expected saves nes='Custom/NES', got %q", cfg.Saves["nes"])
	}
}

func TestConfigStoreSave(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "/data")

	cfg := testDefaults()
	cfg.Service.Autostart = false
	cfg.Saves["test"] = "Test/Path"

	if err := store.Save(&cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load(testDefaults())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Service.Autostart {
		t.Error("expected Autostart=false after save")
	}
	if loaded.Saves["test"] != "Test/Path" {
		t.Errorf("expected test='Test/Path', got %q", loaded.Saves["test"])
	}
}

func TestConfigStoreEmptyDataDir(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "")
	cfg, err := store.Load(testDefaults())
	if err != nil {
		t.Fatalf("Load with empty dir: %v", err)
	}

	if !cfg.Service.Autostart {
		t.Error("expected defaults with empty dir")
	}
}

func TestConfigStoreSaveCreatesDir(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "/nested/deep/dir")
	cfg := testDefaults()

	if err := store.Save(&cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := fs.Stat("/nested/deep/dir/config.toml"); err != nil {
		t.Errorf("config file should exist: %v", err)
	}
}

func TestConfigStoreLoadSyncRelays(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data/config.toml": `[service]
autostart = true

[sync]
relays = ["https://relay1.example.com", "https://relay2.example.com"]
`,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "/data")
	cfg, err := store.Load(testDefaults())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(cfg.Sync.Relays) != 2 {
		t.Fatalf("expected 2 relays, got %d", len(cfg.Sync.Relays))
	}
	if cfg.Sync.Relays[0] != "https://relay1.example.com" {
		t.Errorf("expected first relay 'https://relay1.example.com', got %q", cfg.Sync.Relays[0])
	}
	if cfg.Sync.Relays[1] != "https://relay2.example.com" {
		t.Errorf("expected second relay 'https://relay2.example.com', got %q", cfg.Sync.Relays[1])
	}
}
