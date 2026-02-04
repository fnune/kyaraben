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
	if !strings.Contains(contentStr, "Icon=kyaraben-eden") {
		t.Errorf("eden.desktop should contain Icon=kyaraben-eden, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "Categories=Game;Emulator;") {
		t.Errorf("eden.desktop should contain Categories, got:\n%s", contentStr)
	}

	edenIconPath := filepath.Join(m.iconsDirForExt(".svg"), "kyaraben-eden.svg")
	if _, err := os.Stat(edenIconPath); err != nil {
		t.Errorf("kyaraben-eden.svg should exist in scalable/apps: %v", err)
	}

	duckstationIconPath := filepath.Join(m.iconsDirForExt(".png"), "kyaraben-duckstation.png")
	if _, err := os.Stat(duckstationIconPath); err != nil {
		t.Errorf("kyaraben-duckstation.png should exist in 256x256/apps: %v", err)
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

func TestInstallKyarabenWithAppImage(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	binDir := filepath.Join(homeDir, ".local", "bin")
	appsDir := filepath.Join(homeDir, ".local", "share", "applications")

	t.Setenv("HOME", homeDir)

	m := &Manager{profileDir: filepath.Join(tmpDir, "kyaraben")}

	appImagePath := filepath.Join(tmpDir, "Kyaraben.AppImage")
	if err := os.WriteFile(appImagePath, []byte("fake appimage content"), 0755); err != nil {
		t.Fatalf("creating fake AppImage: %v", err)
	}

	result, err := m.InstallKyaraben(appImagePath, "")
	if err != nil {
		t.Fatalf("InstallKyaraben() error = %v", err)
	}

	if result.AppPath != filepath.Join(binDir, "kyaraben-ui") {
		t.Errorf("AppPath = %s, want %s", result.AppPath, filepath.Join(binDir, "kyaraben-ui"))
	}
	if result.CLIPath != filepath.Join(binDir, "kyaraben") {
		t.Errorf("CLIPath = %s, want %s", result.CLIPath, filepath.Join(binDir, "kyaraben"))
	}
	if result.DesktopPath != filepath.Join(appsDir, "kyaraben.desktop") {
		t.Errorf("DesktopPath = %s, want %s", result.DesktopPath, filepath.Join(appsDir, "kyaraben.desktop"))
	}

	if _, err := os.Stat(result.AppPath); err != nil {
		t.Errorf("AppImage not copied: %v", err)
	}

	content, err := os.ReadFile(result.DesktopPath)
	if err != nil {
		t.Fatalf("reading desktop file: %v", err)
	}
	if !strings.Contains(string(content), "Exec="+result.AppPath) {
		t.Errorf("desktop file should exec AppImage, got:\n%s", content)
	}
}

func TestInstallKyarabenWithSidecar(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	binDir := filepath.Join(homeDir, ".local", "bin")

	t.Setenv("HOME", homeDir)

	m := &Manager{profileDir: filepath.Join(tmpDir, "kyaraben")}

	appImagePath := filepath.Join(tmpDir, "Kyaraben.AppImage")
	if err := os.WriteFile(appImagePath, []byte("fake appimage"), 0755); err != nil {
		t.Fatalf("creating fake AppImage: %v", err)
	}

	sidecarPath := filepath.Join(tmpDir, "kyaraben-sidecar")
	if err := os.WriteFile(sidecarPath, []byte("fake sidecar binary"), 0755); err != nil {
		t.Fatalf("creating fake sidecar: %v", err)
	}

	result, err := m.InstallKyaraben(appImagePath, sidecarPath)
	if err != nil {
		t.Fatalf("InstallKyaraben() error = %v", err)
	}

	cliContent, err := os.ReadFile(result.CLIPath)
	if err != nil {
		t.Fatalf("reading CLI: %v", err)
	}
	if string(cliContent) != "fake sidecar binary" {
		t.Errorf("CLI should be a copy of sidecar, got: %s", cliContent)
	}

	info, err := os.Lstat(result.CLIPath)
	if err != nil {
		t.Fatalf("stat CLI: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("CLI should be a file, not a symlink, when sidecarPath provided")
	}

	if result.AppPath != filepath.Join(binDir, "kyaraben-ui") {
		t.Errorf("AppPath = %s, want %s", result.AppPath, filepath.Join(binDir, "kyaraben-ui"))
	}
}

func TestInstallKyarabenCLIOnly(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	binDir := filepath.Join(homeDir, ".local", "bin")
	appsDir := filepath.Join(homeDir, ".local", "share", "applications")

	t.Setenv("HOME", homeDir)

	m := &Manager{profileDir: filepath.Join(tmpDir, "kyaraben")}

	result, err := m.InstallKyaraben("", "")
	if err != nil {
		t.Fatalf("InstallKyaraben() error = %v", err)
	}

	if result.AppPath != "" {
		t.Errorf("AppPath should be empty for CLI-only install, got %s", result.AppPath)
	}
	if result.CLIPath != filepath.Join(binDir, "kyaraben") {
		t.Errorf("CLIPath = %s, want %s", result.CLIPath, filepath.Join(binDir, "kyaraben"))
	}
	if result.DesktopPath != filepath.Join(appsDir, "kyaraben.desktop") {
		t.Errorf("DesktopPath = %s, want %s", result.DesktopPath, filepath.Join(appsDir, "kyaraben.desktop"))
	}

	linkTarget, err := os.Readlink(result.CLIPath)
	if err != nil {
		t.Fatalf("reading CLI symlink: %v", err)
	}
	if linkTarget == "" {
		t.Error("CLI symlink should point to current executable")
	}

	content, err := os.ReadFile(result.DesktopPath)
	if err != nil {
		t.Fatalf("reading desktop file: %v", err)
	}
	if !strings.Contains(string(content), "Exec="+result.CLIPath) {
		t.Errorf("desktop file should exec CLI, got:\n%s", content)
	}
}

func TestGetInstallStatus(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	binDir := filepath.Join(homeDir, ".local", "bin")
	appsDir := filepath.Join(homeDir, ".local", "share", "applications")

	t.Setenv("HOME", homeDir)

	m := &Manager{profileDir: filepath.Join(tmpDir, "kyaraben")}

	status := m.GetInstallStatus()
	if status.AppPath != "" || status.CLIPath != "" || status.DesktopPath != "" {
		t.Error("GetInstallStatus should return empty paths when nothing installed")
	}

	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("creating bin dir: %v", err)
	}
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("creating apps dir: %v", err)
	}

	appPath := filepath.Join(binDir, "kyaraben-ui")
	cliPath := filepath.Join(binDir, "kyaraben")
	desktopPath := filepath.Join(appsDir, "kyaraben.desktop")

	if err := os.WriteFile(appPath, []byte("fake"), 0755); err != nil {
		t.Fatalf("creating app: %v", err)
	}
	if err := os.WriteFile(cliPath, []byte("fake"), 0755); err != nil {
		t.Fatalf("creating cli: %v", err)
	}
	if err := os.WriteFile(desktopPath, []byte("fake"), 0644); err != nil {
		t.Fatalf("creating desktop: %v", err)
	}

	status = m.GetInstallStatus()
	if status.AppPath != appPath {
		t.Errorf("AppPath = %s, want %s", status.AppPath, appPath)
	}
	if status.CLIPath != cliPath {
		t.Errorf("CLIPath = %s, want %s", status.CLIPath, cliPath)
	}
	if status.DesktopPath != desktopPath {
		t.Errorf("DesktopPath = %s, want %s", status.DesktopPath, desktopPath)
	}
}
