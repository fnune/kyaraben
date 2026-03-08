package config

import (
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"
)

func TestDefaultConfigReturnsExpectedStructure(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Service.Autostart {
		t.Error("expected Service.Autostart=false by default")
	}

	expectedSystems := []string{
		"nes", "snes", "n64", "gb", "gbc", "gba", "nds",
		"gamecube", "wii",
		"psx", "ps2", "psp",
		"genesis", "mastersystem", "gamegear", "saturn", "dreamcast",
		"pcengine", "ngp", "neogeo",
		"atari2600", "c64",
	}

	for _, sys := range expectedSystems {
		if _, ok := cfg.ROMs[sys]; !ok {
			t.Errorf("missing ROMs entry for %s", sys)
		}
		if _, ok := cfg.Saves[sys]; !ok {
			t.Errorf("missing Saves entry for %s", sys)
		}
	}

	if cfg.ROMs["genesis"] != "roms/megadrive" {
		t.Errorf("genesis ROMs path should map to megadrive, got %q", cfg.ROMs["genesis"])
	}
	if cfg.Saves["genesis"] != "saves/megadrive" {
		t.Errorf("genesis Saves path should map to megadrive, got %q", cfg.Saves["genesis"])
	}

	if cfg.Screenshots["retroarch"] != "screenshots" {
		t.Errorf("expected retroarch screenshots path 'screenshots', got %q", cfg.Screenshots["retroarch"])
	}
}

func TestLoadReturnsDefaultsWhenFileDoesNotExist(t *testing.T) {
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

	if cfg.Service.Autostart {
		t.Error("expected Service.Autostart=false from defaults")
	}

	if cfg.ROMs["nes"] != "roms/nes" {
		t.Errorf("expected default nes ROMs path 'roms/nes', got %q", cfg.ROMs["nes"])
	}
}

func TestLoadParsesTomlCorrectly(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data/config.toml": `[service]
autostart = true

[roms]
nes = "custom/roms/nes"
snes = "custom/roms/snes"

[saves]
nes = "custom/saves/nes"

[screenshots]
retroarch = "custom/screenshots"
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

	if !cfg.Service.Autostart {
		t.Error("expected Service.Autostart=true from file")
	}

	if cfg.ROMs["nes"] != "custom/roms/nes" {
		t.Errorf("expected nes ROMs 'custom/roms/nes', got %q", cfg.ROMs["nes"])
	}
	if cfg.ROMs["snes"] != "custom/roms/snes" {
		t.Errorf("expected snes ROMs 'custom/roms/snes', got %q", cfg.ROMs["snes"])
	}

	if cfg.Saves["nes"] != "custom/saves/nes" {
		t.Errorf("expected nes Saves 'custom/saves/nes', got %q", cfg.Saves["nes"])
	}

	if cfg.Screenshots["retroarch"] != "custom/screenshots" {
		t.Errorf("expected retroarch screenshots 'custom/screenshots', got %q", cfg.Screenshots["retroarch"])
	}
}

func TestLoadMergesWithDefaults(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data/config.toml": `[service]
autostart = true
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

	if !cfg.Service.Autostart {
		t.Error("expected Service.Autostart=true from file")
	}

	if cfg.ROMs["nes"] != "roms/nes" {
		t.Errorf("expected default nes ROMs path to be preserved, got %q", cfg.ROMs["nes"])
	}
}

func TestSaveWritesValidToml(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "/data")

	cfg := DefaultConfig()
	cfg.Service.Autostart = true
	cfg.ROMs["test"] = "test/roms"
	cfg.Saves["test"] = "test/saves"

	if err := store.Save(&cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load(DefaultConfig())
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}

	if !loaded.Service.Autostart {
		t.Error("expected Service.Autostart=true after save")
	}
	if loaded.ROMs["test"] != "test/roms" {
		t.Errorf("expected test ROMs 'test/roms', got %q", loaded.ROMs["test"])
	}
	if loaded.Saves["test"] != "test/saves" {
		t.Errorf("expected test Saves 'test/saves', got %q", loaded.Saves["test"])
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	store := NewConfigStore(fs, "/nested/deep/dir")
	cfg := DefaultConfig()

	if err := store.Save(&cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := fs.Stat("/nested/deep/dir/config.toml"); err != nil {
		t.Errorf("config file should exist: %v", err)
	}
}
