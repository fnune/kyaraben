package launcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDesktopFiles(t *testing.T) {
	tmpDir := t.TempDir()

	profileDir := filepath.Join(tmpDir, "kyaraben")
	npLocation := filepath.Join(tmpDir, "nix-portable")

	currentDir := filepath.Join(tmpDir, "store", "nix", "store", "abc123-profile")

	if err := os.MkdirAll(currentDir, 0755); err != nil {
		t.Fatalf("creating current dir: %v", err)
	}

	dataDir := filepath.Join(tmpDir, "data")
	m := &Manager{profileDir: profileDir, dataDir: dataDir, nixPortableLocation: npLocation}

	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("creating profile dir: %v", err)
	}
	if err := os.Symlink(currentDir, m.CurrentLink()); err != nil {
		t.Fatalf("creating current symlink: %v", err)
	}

	storeIconsDir := filepath.Join(currentDir, "share", "icons")
	if err := os.MkdirAll(storeIconsDir, 0755); err != nil {
		t.Fatalf("creating store icons dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storeIconsDir, "eden.svg"), []byte("<svg></svg>"), 0644); err != nil {
		t.Fatalf("writing eden icon: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storeIconsDir, "duckstation.png"), []byte("fake png"), 0644); err != nil {
		t.Fatalf("writing duckstation icon: %v", err)
	}

	entries := []GeneratedDesktop{
		{
			BinaryName:    "eden",
			Name:          "Eden",
			GenericName:   "Nintendo Switch Emulator",
			CategoriesStr: "Game;Emulator",
		},
		{
			BinaryName:    "duckstation",
			Name:          "DuckStation",
			GenericName:   "PlayStation Emulator",
			CategoriesStr: "Game;Emulator",
		},
	}

	result, err := m.GenerateDesktopFiles(entries, nil)
	if err != nil {
		t.Fatalf("GenerateDesktopFiles() error = %v", err)
	}
	if len(result.DesktopFiles) != 2 {
		t.Errorf("GenerateDesktopFiles() should return 2 desktop files, got %d", len(result.DesktopFiles))
	}
	if len(result.IconFiles) != 2 {
		t.Errorf("GenerateDesktopFiles() should return 2 icon files, got %d", len(result.IconFiles))
	}

	edenPath := filepath.Join(m.ApplicationsDir(), "eden.desktop")
	content, err := os.ReadFile(edenPath)
	if err != nil {
		t.Fatalf("reading eden.desktop: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Name=Eden") {
		t.Errorf("eden.desktop should contain Name=Eden, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "GenericName=Nintendo Switch Emulator") {
		t.Errorf("eden.desktop should contain GenericName, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "Icon=eden") {
		t.Errorf("eden.desktop should contain Icon=eden, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "Categories=Game;Emulator;") {
		t.Errorf("eden.desktop should contain Categories, got:\n%s", contentStr)
	}

	edenIconPath := filepath.Join(m.IconsDir(), "eden.svg")
	if _, err := os.Stat(edenIconPath); err != nil {
		t.Errorf("eden.svg should exist: %v", err)
	}

	duckstationIconPath := filepath.Join(m.IconsDir(), "duckstation.png")
	if _, err := os.Stat(duckstationIconPath); err != nil {
		t.Errorf("duckstation.png should exist: %v", err)
	}
}

func TestVirtualToRealStorePath(t *testing.T) {
	m := &Manager{nixPortableLocation: "/home/user/.local/state/kyaraben/build/nix"}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "typical virtual path",
			input:    "/nix/store/abc123-retroarch/share/applications/retroarch.desktop",
			expected: "/home/user/.local/state/kyaraben/build/nix/.nix-portable/nix/store/abc123-retroarch/share/applications/retroarch.desktop",
		},
		{
			name:     "non-nix path unchanged",
			input:    "/usr/share/applications/foo.desktop",
			expected: "/usr/share/applications/foo.desktop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.virtualToRealStorePath(tt.input)
			if result != tt.expected {
				t.Errorf("got %s, expected %s", result, tt.expected)
			}
		})
	}
}
