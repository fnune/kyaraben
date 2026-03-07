package mapping

import (
	"testing"

	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/internal/folders"
)

func TestKyarabenFolderID(t *testing.T) {
	m := NewMapper("/mnt/SDCARD", config.DefaultConfig())

	tests := []struct {
		category folders.Category
		system   string
		expected string
	}{
		{folders.CategorySaves, "gb", "kyaraben-saves-gb"},
		{folders.CategoryROMs, "snes", "kyaraben-roms-snes"},
		{folders.CategoryBIOS, "psx", "kyaraben-bios-psx"},
		{folders.CategoryScreenshots, "retroarch", "kyaraben-screenshots-retroarch"},
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
		category folders.Category
		system   string
		expected string
	}{
		{folders.CategorySaves, "gb", "/mnt/SDCARD/Saves/GB"},
		{folders.CategorySaves, "snes", "/mnt/SDCARD/Saves/SFC"},
		{folders.CategoryBIOS, "psx", "/mnt/SDCARD/Bios/PS"},
		{folders.CategoryROMs, "gb", "/mnt/SDCARD/Roms/Game Boy (GB)"},
		{folders.CategoryROMs, "unknown", ""},
		{folders.CategoryScreenshots, "retroarch", "/mnt/SDCARD/Screenshots"},
		{folders.CategoryScreenshots, "duckstation", ""},
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

	got := m.DevicePath(folders.CategorySaves, "gba")
	expected := "/mnt/SDCARD/Saves/MGBA"
	if got != expected {
		t.Errorf("DevicePath with custom config = %q, want %q", got, expected)
	}
}

func TestFolderMappings(t *testing.T) {
	cfg := config.Config{
		PathMappings: config.PathMappings{
			Saves: map[string]string{
				"gb": "Saves/GB",
			},
			ROMs: map[string]string{
				"gb": "Roms/Game Boy (GB)",
			},
			Screenshots: map[string]string{
				"retroarch": "Screenshots",
			},
		},
	}

	m := NewMapper("/mnt/SDCARD", cfg)
	mappings := m.FolderMappings()

	if len(mappings) != 3 {
		t.Errorf("expected 3 mappings, got %d", len(mappings))
	}

	foundSaves := false
	foundROMs := false
	foundRetroarchScreenshots := false

	for _, mapping := range mappings {
		switch mapping.FolderID {
		case "kyaraben-saves-gb":
			foundSaves = true
			if mapping.DevicePath != "/mnt/SDCARD/Saves/GB" {
				t.Errorf("saves path = %q, want /mnt/SDCARD/Saves/GB", mapping.DevicePath)
			}
		case "kyaraben-roms-gb":
			foundROMs = true
		case "kyaraben-screenshots-retroarch":
			foundRetroarchScreenshots = true
			if mapping.DevicePath != "/mnt/SDCARD/Screenshots" {
				t.Errorf("retroarch screenshots path = %q, want /mnt/SDCARD/Screenshots", mapping.DevicePath)
			}
		}
	}

	if !foundSaves {
		t.Error("saves mapping not found")
	}
	if !foundROMs {
		t.Error("roms mapping not found")
	}
	if !foundRetroarchScreenshots {
		t.Error("retroarch screenshots mapping not found")
	}
}

func TestAllSystems(t *testing.T) {
	cfg := config.Config{
		PathMappings: config.PathMappings{
			Saves: map[string]string{
				"gb":  "Saves/GB",
				"gba": "Saves/GBA",
			},
			ROMs: map[string]string{
				"gb":   "Roms/Game Boy (GB)",
				"snes": "Roms/Super Nintendo (SFC)",
			},
		},
	}

	m := NewMapper("/mnt/SDCARD", cfg)
	systems := m.AllSystems()

	if len(systems) != 3 {
		t.Errorf("expected 3 systems (gb, gba, snes), got %d", len(systems))
	}
}
