package mapping

import (
	"testing"

	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
)

func TestSyncguestFolderMappings(t *testing.T) {
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
	mappings := m.SyncguestFolderMappings()

	if len(mappings) != 3 {
		t.Errorf("expected 3 mappings, got %d", len(mappings))
	}

	found := map[string]bool{
		"kyaraben-saves-gb":              false,
		"kyaraben-roms-gb":               false,
		"kyaraben-screenshots-retroarch": false,
	}

	for _, mapping := range mappings {
		found[mapping.ID] = true
		switch mapping.ID {
		case "kyaraben-saves-gb":
			if mapping.Path != "/mnt/SDCARD/Saves/GB" {
				t.Errorf("saves path = %q, want /mnt/SDCARD/Saves/GB", mapping.Path)
			}
		case "kyaraben-screenshots-retroarch":
			if mapping.Path != "/mnt/SDCARD/Screenshots" {
				t.Errorf("screenshots path = %q, want /mnt/SDCARD/Screenshots", mapping.Path)
			}
		}
	}

	for id, ok := range found {
		if !ok {
			t.Errorf("mapping %s not found", id)
		}
	}
}

func TestSyncguestFolderMappingsWithStates(t *testing.T) {
	cfg := config.Config{
		Service: config.ServiceConfig{
			SyncStates: true,
		},
		PathMappings: config.PathMappings{
			Saves: map[string]string{
				"gb": "Saves/GB",
			},
			States: map[string]string{
				"retroarch:snes9x": ".userdata/shared/SFC-snes9x",
			},
		},
	}

	m := NewMapper("/mnt/SDCARD", cfg)
	mappings := m.SyncguestFolderMappings()

	if len(mappings) != 2 {
		t.Errorf("expected 2 mappings, got %d", len(mappings))
	}

	foundStates := false
	for _, mapping := range mappings {
		if mapping.ID == "kyaraben-states-retroarch:snes9x" {
			foundStates = true
			if mapping.Path != "/mnt/SDCARD/.userdata/shared/SFC-snes9x" {
				t.Errorf("states path = %q, want /mnt/SDCARD/.userdata/shared/SFC-snes9x", mapping.Path)
			}
		}
	}

	if !foundStates {
		t.Error("states mapping not found")
	}
}

func TestSyncguestFolderMappingsStatesDisabled(t *testing.T) {
	cfg := config.Config{
		Service: config.ServiceConfig{
			SyncStates: false,
		},
		PathMappings: config.PathMappings{
			Saves: map[string]string{
				"gb": "Saves/GB",
			},
			States: map[string]string{
				"retroarch:snes9x": ".userdata/shared/SFC-snes9x",
			},
		},
	}

	m := NewMapper("/mnt/SDCARD", cfg)
	mappings := m.SyncguestFolderMappings()

	if len(mappings) != 1 {
		t.Errorf("expected 1 mapping (states disabled), got %d", len(mappings))
	}

	for _, mapping := range mappings {
		if mapping.ID == "kyaraben-states-retroarch:snes9x" {
			t.Error("states should not be included when SyncStates is false")
		}
	}
}
