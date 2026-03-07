package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesDefaultConfig(t *testing.T) {
	dir := t.TempDir()

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !cfg.Service.Autostart {
		t.Error("expected Service.Autostart=true by default")
	}

	if _, err := os.Stat(filepath.Join(dir, "config.toml")); err != nil {
		t.Errorf("config file should be created: %v", err)
	}
}

func TestLoadReadsExistingConfig(t *testing.T) {
	dir := t.TempDir()

	content := `[service]
autostart = false

[saves]
gba = "Saves/MGBA"
`
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
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
	dir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Service.Autostart = false
	cfg.Saves["test"] = "Test/Path"

	if err := cfg.Save(dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(dir)
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
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load with empty dir: %v", err)
	}

	if !cfg.Service.Autostart {
		t.Error("expected defaults with empty dir")
	}
}

func TestDefaultConfigHasAllSystems(t *testing.T) {
	cfg := DefaultConfig()

	expectedSaves := []string{"nes", "snes", "gb", "gbc", "gba", "psx", "genesis"}
	for _, sys := range expectedSaves {
		if _, ok := cfg.Saves[sys]; !ok {
			t.Errorf("missing saves entry for %s", sys)
		}
	}

	expectedROMs := []string{"nes", "snes", "gb", "gbc", "gba", "psx", "genesis"}
	for _, sys := range expectedROMs {
		if _, ok := cfg.ROMs[sys]; !ok {
			t.Errorf("missing roms entry for %s", sys)
		}
	}
}
