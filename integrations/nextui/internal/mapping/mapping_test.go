package mapping

import (
	"testing"
)

func TestKyarabenFolderID(t *testing.T) {
	m := NewMapper("/mnt/SDCARD", nil)

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

func TestNextUIPath(t *testing.T) {
	m := NewMapper("/mnt/SDCARD", nil)

	tests := []struct {
		category Category
		tag      string
		expected string
	}{
		{CategorySaves, "GB", "/mnt/SDCARD/Saves/GB"},
		{CategorySaves, "SFC", "/mnt/SDCARD/Saves/SFC"},
		{CategoryBIOS, "PS", "/mnt/SDCARD/Bios/PS"},
		{CategoryROMs, "GB", "/mnt/SDCARD/Roms/Game Boy (GB)"},
		{CategoryROMs, "UNKNOWN", "/mnt/SDCARD/Roms/UNKNOWN"},
		{CategoryScreenshots, "", "/mnt/SDCARD/Screenshots"},
	}

	for _, tt := range tests {
		got := m.NextUIPath(tt.category, tt.tag)
		if got != tt.expected {
			t.Errorf("NextUIPath(%s, %s) = %q, want %q", tt.category, tt.tag, got, tt.expected)
		}
	}
}

func TestTagToSystem(t *testing.T) {
	m := NewMapper("/mnt/SDCARD", nil)

	tests := []struct {
		tag      string
		expected string
		ok       bool
	}{
		{"GB", "gb", true},
		{"SFC", "snes", true},
		{"FC", "nes", true},
		{"PS", "psx", true},
		{"UNKNOWN", "", false},
	}

	for _, tt := range tests {
		got, ok := m.TagToSystem(tt.tag)
		if ok != tt.ok || got != tt.expected {
			t.Errorf("TagToSystem(%s) = (%q, %v), want (%q, %v)", tt.tag, got, ok, tt.expected, tt.ok)
		}
	}
}

func TestTagToSystemWithOverrides(t *testing.T) {
	overrides := map[string]string{
		"MGBA": "gba",
		"SUPA": "snes",
	}
	m := NewMapper("/mnt/SDCARD", overrides)

	tests := []struct {
		tag      string
		expected string
		ok       bool
	}{
		{"MGBA", "gba", true},
		{"SUPA", "snes", true},
		{"GBA", "gba", true},
	}

	for _, tt := range tests {
		got, ok := m.TagToSystem(tt.tag)
		if ok != tt.ok || got != tt.expected {
			t.Errorf("TagToSystem(%s) = (%q, %v), want (%q, %v)", tt.tag, got, ok, tt.expected, tt.ok)
		}
	}
}

func TestSystemToTag(t *testing.T) {
	m := NewMapper("/mnt/SDCARD", nil)

	tests := []struct {
		system   string
		expected string
		ok       bool
	}{
		{"gb", "GB", true},
		{"snes", "SFC", true},
		{"nes", "FC", true},
		{"psx", "PS", true},
		{"unknown", "", false},
	}

	for _, tt := range tests {
		got, ok := m.SystemToTag(tt.system)
		if ok != tt.ok || got != tt.expected {
			t.Errorf("SystemToTag(%s) = (%q, %v), want (%q, %v)", tt.system, got, ok, tt.expected, tt.ok)
		}
	}
}

func TestExtractTagFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/mnt/SDCARD/Roms/Game Boy (GB)", "GB"},
		{"/mnt/SDCARD/Roms/Super Nintendo (SFC)", "SFC"},
		{"/mnt/SDCARD/Saves/GB", "GB"},
		{"/mnt/SDCARD/Roms/NoTag", "NoTag"},
	}

	for _, tt := range tests {
		got := ExtractTagFromPath(tt.path)
		if got != tt.expected {
			t.Errorf("ExtractTagFromPath(%s) = %q, want %q", tt.path, got, tt.expected)
		}
	}
}
