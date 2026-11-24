package config

import (
	"maps"
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

func TestDefaultROMFoldersMatchTheNextUISkeleton(t *testing.T) {
	nextUISkeleton := map[model.SystemID]string{
		model.SystemIDNES:          "Roms/Nintendo Entertainment System (FC)",
		model.SystemIDSNES:         "Roms/Super Nintendo Entertainment System (SFC)",
		model.SystemIDGB:           "Roms/Game Boy (GB)",
		model.SystemIDGBC:          "Roms/Game Boy Color (GBC)",
		model.SystemIDGBA:          "Roms/Game Boy Advance (GBA)",
		model.SystemIDPSX:          "Roms/Sony PlayStation (PS)",
		model.SystemIDGenesis:      "Roms/Sega Genesis (MD)",
		model.SystemIDGameGear:     "Roms/Sega Game Gear (GG)",
		model.SystemIDMasterSystem: "Roms/Sega Master System (SMS)",
		model.SystemIDPCEngine:     "Roms/TurboGrafx-16 (PCE)",
		model.SystemIDNGP:          "Roms/Neo Geo Pocket (NGP)",
		model.SystemIDAtari2600:    "Roms/Atari 2600 (A2600)",
		model.SystemIDC64:          "Roms/Commodore 64 (C64)",
		model.SystemIDArcade:       "Roms/Arcade (FBN)",
	}

	cfg := DefaultConfig()
	for system, want := range nextUISkeleton {
		if got := cfg.ROMs[string(system)]; got != want {
			t.Errorf("%s: got %q, want %q", system, got, want)
		}
	}
}

func TestDefaultConfigForCardAdoptsAnExistingFolderWithTheSameTag(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/sd/Roms/Super Famicom (SFC)":                        &vfst.Dir{Perm: 0755},
		"/sd/Roms/Super Nintendo Entertainment System (SUPA)": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cfg := DefaultConfigForCard(fs, "/sd")

	if got := cfg.ROMs[string(model.SystemIDSNES)]; got != "Roms/Super Famicom (SFC)" {
		t.Errorf("snes: got %q", got)
	}
	if got := cfg.ROMs[string(model.SystemIDNES)]; got != "Roms/Nintendo Entertainment System (FC)" {
		t.Errorf("nes without a folder on the card: got %q", got)
	}
}

func TestDefaultConfigForCardPrefersTheNextUIFolderOverOtherFoldersWithTheSameTag(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/sd/Roms/Super Nintendo (SFC)":                      &vfst.Dir{Perm: 0755},
		"/sd/Roms/Super Nintendo Entertainment System (SFC)": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cfg := DefaultConfigForCard(fs, "/sd")

	if got := cfg.ROMs[string(model.SystemIDSNES)]; got != "Roms/Super Nintendo Entertainment System (SFC)" {
		t.Errorf("snes: got %q", got)
	}
}

func TestDefaultConfigForCardKeepsTheSavedMappingOfAnExistingDevice(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/sd/Roms/Super Nintendo Entertainment System (SFC)": &vfst.Dir{Perm: 0755},
		"/data/config.toml": `[roms]
snes = "Roms/Super Nintendo (SFC)"
`,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cfg, err := NewConfigStore(fs, "/data").Load(DefaultConfigForCard(fs, "/sd"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := cfg.ROMs[string(model.SystemIDSNES)]; got != "Roms/Super Nintendo (SFC)" {
		t.Errorf("snes: got %q", got)
	}
}

func TestDefaultConfigForCardWithoutARomsFolderUsesTheDefaults(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/sd": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if got, want := DefaultConfigForCard(fs, "/sd").ROMs, DefaultConfig().ROMs; !maps.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
