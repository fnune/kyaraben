package mapping

import (
	"testing"

	"github.com/fnune/kyaraben/integrations/batocera/internal/config"
)

func TestNewMapper(t *testing.T) {
	cfg := config.Config{}
	m := NewMapper("/userdata", cfg)

	if m.basePath != "/userdata" {
		t.Errorf("basePath = %q, want /userdata", m.basePath)
	}
}

func TestSyncguestFolderMappings(t *testing.T) {
	cfg := config.Config{
		ROMs: map[string]string{
			"gb":  "roms/gb",
			"nes": "roms/nes",
		},
		Saves: map[string]string{
			"gb": "saves/gb",
		},
		Screenshots: map[string]string{
			"retroarch": "screenshots",
		},
	}

	m := NewMapper("/userdata", cfg)
	mappings := m.SyncguestFolderMappings()

	if len(mappings) != 4 {
		t.Errorf("expected 4 mappings, got %d", len(mappings))
	}

	found := map[string]bool{
		"kyaraben-roms-gb":               false,
		"kyaraben-roms-nes":              false,
		"kyaraben-saves-gb":              false,
		"kyaraben-screenshots-retroarch": false,
	}

	for _, mapping := range mappings {
		found[mapping.ID] = true
		switch mapping.ID {
		case "kyaraben-roms-gb":
			if mapping.Path != "/userdata/roms/gb" {
				t.Errorf("roms/gb path = %q, want /userdata/roms/gb", mapping.Path)
			}
		case "kyaraben-roms-nes":
			if mapping.Path != "/userdata/roms/nes" {
				t.Errorf("roms/nes path = %q, want /userdata/roms/nes", mapping.Path)
			}
		case "kyaraben-saves-gb":
			if mapping.Path != "/userdata/saves/gb" {
				t.Errorf("saves/gb path = %q, want /userdata/saves/gb", mapping.Path)
			}
		case "kyaraben-screenshots-retroarch":
			if mapping.Path != "/userdata/screenshots" {
				t.Errorf("screenshots path = %q, want /userdata/screenshots", mapping.Path)
			}
		}
	}

	for id, ok := range found {
		if !ok {
			t.Errorf("mapping %s not found", id)
		}
	}
}

func TestSyncguestFolderMappingsEmpty(t *testing.T) {
	cfg := config.Config{}

	m := NewMapper("/userdata", cfg)
	mappings := m.SyncguestFolderMappings()

	if len(mappings) != 0 {
		t.Errorf("expected 0 mappings for empty config, got %d", len(mappings))
	}
}
