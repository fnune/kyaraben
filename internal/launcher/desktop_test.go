package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDesktopFiles(t *testing.T) {
	tmpDir := t.TempDir()

	profileDir := filepath.Join(tmpDir, "kyaraben")
	npLocation := filepath.Join(tmpDir, "nix-portable")
	realStoreDir := filepath.Join(npLocation, ".nix-portable", "nix", "store", "xyz789-retroarch")
	realAppsDir := filepath.Join(realStoreDir, "share", "applications")

	currentDir := filepath.Join(tmpDir, "store", "nix", "store", "abc123-profile")
	currentAppsDir := filepath.Join(currentDir, "share", "applications")

	if err := os.MkdirAll(realAppsDir, 0755); err != nil {
		t.Fatalf("creating real apps dir: %v", err)
	}
	if err := os.MkdirAll(currentAppsDir, 0755); err != nil {
		t.Fatalf("creating current apps dir: %v", err)
	}

	realDesktop := filepath.Join(realAppsDir, "retroarch.desktop")
	desktopContent := "[Desktop Entry]\nName=RetroArch\nExec=/nix/store/xyz789-retroarch/bin/retroarch %f\n"
	if err := os.WriteFile(realDesktop, []byte(desktopContent), 0644); err != nil {
		t.Fatalf("creating real desktop file: %v", err)
	}

	virtualTarget := "/nix/store/xyz789-retroarch/share/applications/retroarch.desktop"
	symlinkPath := filepath.Join(currentAppsDir, "retroarch.desktop")
	if err := os.Symlink(virtualTarget, symlinkPath); err != nil {
		t.Fatalf("creating symlink: %v", err)
	}

	m := &Manager{profileDir: profileDir, nixPortableLocation: npLocation}

	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("creating profile dir: %v", err)
	}
	if err := os.Symlink(currentDir, m.CurrentLink()); err != nil {
		t.Fatalf("creating current symlink: %v", err)
	}

	entries := []DesktopEntry{
		NixStoreDesktop{BinaryName: "retroarch"},
		GeneratedDesktop{
			BinaryName:    "eden",
			Name:          "Eden",
			GenericName:   "Nintendo Switch Emulator",
			CategoriesStr: "Game;Emulator",
		},
	}

	result, err := m.GenerateDesktopFiles(entries, nil)
	if err != nil {
		t.Fatalf("GenerateDesktopFiles() error = %v", err)
	}
	if len(result.DesktopFiles) == 0 {
		t.Error("GenerateDesktopFiles() should return created desktop files")
	}

	retroarchPath := filepath.Join(m.ApplicationsDir(), "retroarch.desktop")
	info, err := os.Lstat(retroarchPath)
	if err != nil {
		t.Fatalf("checking retroarch.desktop: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("retroarch.desktop should be a regular file, not a symlink")
	}

	retroarchContent, err := os.ReadFile(retroarchPath)
	if err != nil {
		t.Fatalf("reading retroarch.desktop: %v", err)
	}
	if strings.Contains(string(retroarchContent), "/nix/store/") {
		t.Errorf("retroarch.desktop should not contain /nix/store/ paths, got:\n%s", retroarchContent)
	}
	expectedExec := fmt.Sprintf("Exec=%s/retroarch", m.BinDir())
	if !strings.Contains(string(retroarchContent), expectedExec) {
		t.Errorf("retroarch.desktop should contain %s, got:\n%s", expectedExec, retroarchContent)
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
	expectedEdenExec := fmt.Sprintf("Exec=%s/eden", m.BinDir())
	if !strings.Contains(contentStr, expectedEdenExec) {
		t.Errorf("eden.desktop should contain %s, got:\n%s", expectedEdenExec, contentStr)
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
}

func TestDesktopEntryInterface(t *testing.T) {
	nix := NixStoreDesktop{BinaryName: "retroarch"}
	gen := GeneratedDesktop{BinaryName: "eden", Name: "Eden"}

	var entries []DesktopEntry
	entries = append(entries, nix, gen)

	if entries[0].Binary() != "retroarch" {
		t.Errorf("NixStoreDesktop.Binary() = %s, want retroarch", entries[0].Binary())
	}
	if entries[1].Binary() != "eden" {
		t.Errorf("GeneratedDesktop.Binary() = %s, want eden", entries[1].Binary())
	}
}

func TestEmbeddedIcons(t *testing.T) {
	for _, name := range []string{"eden", "duckstation"} {
		data, err := embeddedIcons.ReadFile("icons/" + name + ".svg")
		if err != nil {
			t.Errorf("icon %s.svg not embedded: %v", name, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("icon %s.svg is empty", name)
		}
		if !strings.Contains(string(data), "<svg") {
			t.Errorf("icon %s.svg doesn't look like an SVG", name)
		}
	}
}

func TestRewriteDesktopExecLines(t *testing.T) {
	binDir := "/home/user/.local/state/kyaraben/bin"
	tests := []struct {
		name     string
		input    string
		binary   string
		expected string
	}{
		{
			name:     "simple exec line",
			input:    "Exec=/nix/store/abc123-foo/bin/retroarch %f",
			binary:   "retroarch",
			expected: "Exec=/home/user/.local/state/kyaraben/bin/retroarch %f",
		},
		{
			name:     "exec line with flags",
			input:    "Exec=/nix/store/xyz789-bar/bin/retroarch --verbose",
			binary:   "retroarch",
			expected: "Exec=/home/user/.local/state/kyaraben/bin/retroarch --verbose",
		},
		{
			name:     "multiple lines",
			input:    "[Desktop Entry]\nName=RetroArch\nExec=/nix/store/abc123/bin/retroarch %f\nIcon=retroarch",
			binary:   "retroarch",
			expected: "[Desktop Entry]\nName=RetroArch\nExec=/home/user/.local/state/kyaraben/bin/retroarch %f\nIcon=retroarch",
		},
		{
			name:     "no exec line",
			input:    "[Desktop Entry]\nName=Test\nIcon=test",
			binary:   "test",
			expected: "[Desktop Entry]\nName=Test\nIcon=test",
		},
		{
			name:     "non-nix exec line unchanged",
			input:    "Exec=/usr/bin/retroarch %f",
			binary:   "retroarch",
			expected: "Exec=/usr/bin/retroarch %f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriteDesktopExecLines([]byte(tt.input), binDir, tt.binary)
			if string(result) != tt.expected {
				t.Errorf("got:\n%s\nexpected:\n%s", result, tt.expected)
			}
		})
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
