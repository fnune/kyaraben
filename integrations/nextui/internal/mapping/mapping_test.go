package mapping

import (
	"testing"

	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
)

func TestKyarabenFolderID(t *testing.T) {
	m := NewMapper("/mnt/SDCARD", config.DefaultConfig())

	tests := []struct {
		category Category
		system   string
		expected string
	}{
		{CategorySaves, "gb", "kyaraben-saves-gb"},
		{CategoryROMs, "snes", "kyaraben-roms-snes"},
		{CategoryBIOS, "psx", "kyaraben-bios-psx"},
		{CategoryScreenshots, "", "kyaraben-screenshots"},
	}

	for _, tt := range tests {
		got := m.KyarabenFolderID(tt.category, tt.system)
		if got != tt.expected {
			t.Errorf("KyarabenFolderID(%s, %s) = %q, want %q", tt.category, tt.system, got, tt.expected)
		}
	}
}

func TestDevicePath(t *testing.T) {
	m := NewMapper("/mnt/SDCARD", config.DefaultConfig())

	tests := []struct {
		category Category
		system   string
		expected string
	}{
		{CategorySaves, "gb", "/mnt/SDCARD/Saves/GB"},
		{CategorySaves, "snes", "/mnt/SDCARD/Saves/SFC"},
		{CategoryBIOS, "psx", "/mnt/SDCARD/Bios/PS"},
		{CategoryROMs, "gb", "/mnt/SDCARD/Roms/Game Boy (GB)"},
		{CategoryROMs, "unknown", ""},
		{CategoryScreenshots, "", "/mnt/SDCARD/Screenshots"},
	}

	for _, tt := range tests {
		got := m.DevicePath(tt.category, tt.system)
		if got != tt.expected {
			t.Errorf("DevicePath(%s, %s) = %q, want %q", tt.category, tt.system, got, tt.expected)
		}
	}
}

func TestDevicePathWithCustomConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Saves["gba"] = "Saves/MGBA"

	m := NewMapper("/mnt/SDCARD", cfg)

	got := m.DevicePath(CategorySaves, "gba")
	expected := "/mnt/SDCARD/Saves/MGBA"
	if got != expected {
		t.Errorf("DevicePath with custom config = %q, want %q", got, expected)
	}
}

func TestFolderMappings(t *testing.T) {
	cfg := config.Config{
		Saves: map[string]string{
			"gb": "Saves/GB",
		},
		ROMs: map[string]string{
			"gb": "Roms/Game Boy (GB)",
		},
		Screenshots: "Screenshots",
	}

	m := NewMapper("/mnt/SDCARD", cfg)
	mappings := m.FolderMappings()

	if len(mappings) != 3 {
		t.Errorf("expected 3 mappings, got %d", len(mappings))
	}

	foundSaves := false
	foundROMs := false
	foundScreenshots := false

	for _, mapping := range mappings {
		switch mapping.FolderID {
		case "kyaraben-saves-gb":
			foundSaves = true
			if mapping.DevicePath != "/mnt/SDCARD/Saves/GB" {
				t.Errorf("saves path = %q, want /mnt/SDCARD/Saves/GB", mapping.DevicePath)
			}
		case "kyaraben-roms-gb":
			foundROMs = true
		case "kyaraben-screenshots":
			foundScreenshots = true
		}
	}

	if !foundSaves {
		t.Error("saves mapping not found")
	}
	if !foundROMs {
		t.Error("roms mapping not found")
	}
	if !foundScreenshots {
		t.Error("screenshots mapping not found")
	}
}

func TestAllSystems(t *testing.T) {
	cfg := config.Config{
		Saves: map[string]string{
			"gb":  "Saves/GB",
			"gba": "Saves/GBA",
		},
		ROMs: map[string]string{
			"gb":   "Roms/Game Boy (GB)",
			"snes": "Roms/Super Nintendo (SFC)",
		},
	}

	m := NewMapper("/mnt/SDCARD", cfg)
	systems := m.AllSystems()

	if len(systems) != 3 {
		t.Errorf("expected 3 systems (gb, gba, snes), got %d", len(systems))
	}
}
