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

	// Create the fake nix-portable store structure
	realIconsDir := filepath.Join(npLocation, ".nix-portable", "nix", "store", "abc123-icons", "share", "icons")
	if err := os.MkdirAll(realIconsDir, 0755); err != nil {
		t.Fatalf("creating real icons dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(realIconsDir, "eden.svg"), []byte("<svg></svg>"), 0644); err != nil {
		t.Fatalf("writing eden icon: %v", err)
	}
	if err := os.WriteFile(filepath.Join(realIconsDir, "duckstation.png"), []byte("fake png"), 0644); err != nil {
		t.Fatalf("writing duckstation icon: %v", err)
	}

	// Create the profile with symlinks pointing to virtual /nix/store paths
	currentDir := filepath.Join(npLocation, ".nix-portable", "nix", "store", "xyz789-profile")
	profileIconsDir := filepath.Join(currentDir, "share", "icons")
	if err := os.MkdirAll(profileIconsDir, 0755); err != nil {
		t.Fatalf("creating profile icons dir: %v", err)
	}
	if err := os.Symlink("/nix/store/abc123-icons/share/icons/eden.svg", filepath.Join(profileIconsDir, "eden.svg")); err != nil {
		t.Fatalf("creating eden symlink: %v", err)
	}
	if err := os.Symlink("/nix/store/abc123-icons/share/icons/duckstation.png", filepath.Join(profileIconsDir, "duckstation.png")); err != nil {
		t.Fatalf("creating duckstation symlink: %v", err)
	}

	dataDir := filepath.Join(tmpDir, "data")
	m := &Manager{profileDir: profileDir, dataDir: dataDir, nixPortableLocation: npLocation}

	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("creating profile dir: %v", err)
	}
	if err := os.Symlink(currentDir, m.CurrentLink()); err != nil {
		t.Fatalf("creating current symlink: %v", err)
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

	edenIconPath := filepath.Join(m.iconsDirForExt(".svg"), "eden.svg")
	if _, err := os.Stat(edenIconPath); err != nil {
		t.Errorf("eden.svg should exist in scalable/apps: %v", err)
	}

	duckstationIconPath := filepath.Join(m.iconsDirForExt(".png"), "duckstation.png")
	if _, err := os.Stat(duckstationIconPath); err != nil {
		t.Errorf("duckstation.png should exist in 256x256/apps: %v", err)
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
