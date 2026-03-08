package config

import (
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
)

func TestLoadCreatesDefaultConfig(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "/data")
	cfg, err := store.Load(DefaultConfig())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !cfg.Service.Autostart {
		t.Error("expected Service.Autostart=true by default")
	}

	if _, err := fs.Stat("/data/config.toml"); err != nil {
		t.Errorf("config file should be created: %v", err)
	}
}

func TestLoadReadsExistingConfig(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data/config.toml": `[service]
autostart = false

[saves]
gba = "Saves/MGBA"
`,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "/data")
	cfg, err := store.Load(DefaultConfig())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Service.Autostart {
		t.Error("expected Service.Autostart=false from file")
	}
	if cfg.Saves["gba"] != "Saves/MGBA" {
		t.Errorf("expected gba saves path 'Saves/MGBA', got %q", cfg.Saves["gba"])
	}
}

func TestSaveConfig(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "/data")

	cfg := DefaultConfig()
	cfg.Service.Autostart = false
	cfg.Saves["test"] = "Test/Path"

	if err := store.Save(&cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load(DefaultConfig())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Service.Autostart {
		t.Error("expected Service.Autostart=false after save")
	}
	if loaded.Saves["test"] != "Test/Path" {
		t.Errorf("expected test saves path 'Test/Path', got %q", loaded.Saves["test"])
	}
}

func TestLoadWithEmptyDataDir(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "")
	cfg, err := store.Load(DefaultConfig())
	if err != nil {
		t.Fatalf("Load with empty dir: %v", err)
	}

	if !cfg.Service.Autostart {
		t.Error("expected defaults with empty dir")
	}
}

func TestDefaultConfigHasAllSystems(t *testing.T) {
	cfg := DefaultConfig()

	expectedSaves := []model.SystemID{
		model.SystemIDNES,
		model.SystemIDSNES,
		model.SystemIDGB,
		model.SystemIDGBC,
		model.SystemIDGBA,
		model.SystemIDPSX,
		model.SystemIDGenesis,
	}
	for _, sys := range expectedSaves {
		if _, ok := cfg.Saves[string(sys)]; !ok {
			t.Errorf("missing saves entry for %s", sys)
		}
	}

	expectedROMs := []model.SystemID{
		model.SystemIDNES,
		model.SystemIDSNES,
		model.SystemIDGB,
		model.SystemIDGBC,
		model.SystemIDGBA,
		model.SystemIDPSX,
		model.SystemIDGenesis,
	}
	for _, sys := range expectedROMs {
		if _, ok := cfg.ROMs[string(sys)]; !ok {
			t.Errorf("missing roms entry for %s", sys)
		}
	}
}
