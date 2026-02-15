package launcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/paths"
)

type fakeBaseDirResolver struct {
	homeDir   string
	configDir string
	dataDir   string
}

func (r fakeBaseDirResolver) UserHomeDir() (string, error) {
	return r.homeDir, nil
}

func (r fakeBaseDirResolver) UserConfigDir() (string, error) {
	if r.configDir != "" {
		return r.configDir, nil
	}
	return filepath.Join(r.homeDir, ".config"), nil
}

func (r fakeBaseDirResolver) UserDataDir() (string, error) {
	if r.dataDir != "" {
		return r.dataDir, nil
	}
	return filepath.Join(r.homeDir, ".local", "share"), nil
}

func TestGenerateDesktopFiles(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/kyaraben":              &vfst.Dir{Perm: 0755},
		"/home":                  &vfst.Dir{Perm: 0755},
		"/icons/eden.svg":        "<svg></svg>",
		"/icons/duckstation.png": "fake png",
		"/data":                  &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	profileDir := "/kyaraben"
	homeDir := "/home"
	dataDir := "/data"

	edenIconPath := "/icons/eden.svg"
	duckstationIconPath := "/icons/duckstation.png"

	resolver := fakeBaseDirResolver{homeDir: homeDir}
	m := &Manager{fs: fs, paths: paths.DefaultPaths(), profileDir: profileDir, dataDir: dataDir, resolver: resolver}

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

	icons := []InstalledIcon{
		{Name: "eden", Filename: "eden.svg", Path: edenIconPath},
		{Name: "duckstation", Filename: "duckstation.png", Path: duckstationIconPath},
	}

	result, err := m.GenerateDesktopFiles(entries, icons, nil)
	if err != nil {
		t.Fatalf("GenerateDesktopFiles() error = %v", err)
	}
	if len(result.DesktopFiles) != 2 {
		t.Errorf("should return 2 desktop files, got %d", len(result.DesktopFiles))
	}
	if len(result.IconFiles) != 2 {
		t.Errorf("should return 2 icon files, got %d", len(result.IconFiles))
	}

	edenPath := m.ApplicationsDir() + "/eden.desktop"
	content, err := fs.ReadFile(edenPath)
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
	expectedIconPath := "/data/icons/hicolor/scalable/apps/kyaraben-eden.svg"
	if !strings.Contains(contentStr, "Icon="+expectedIconPath) {
		t.Errorf("eden.desktop should contain Icon=%s, got:\n%s", expectedIconPath, contentStr)
	}
	if !strings.Contains(contentStr, "Categories=Game;Emulator;") {
		t.Errorf("eden.desktop should contain Categories, got:\n%s", contentStr)
	}

	edenIcon := m.iconsDirForExt(".svg") + "/kyaraben-eden.svg"
	if _, err := fs.Stat(edenIcon); err != nil {
		t.Errorf("kyaraben-eden.svg should exist in scalable/apps: %v", err)
	}

	duckstationIcon := m.iconsDirForExt(".png") + "/kyaraben-duckstation.png"
	if _, err := fs.Stat(duckstationIcon); err != nil {
		t.Errorf("kyaraben-duckstation.png should exist in 256x256/apps: %v", err)
	}
}

func TestInstallKyarabenWithAppImage(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/kyaraben":             &vfst.Dir{Perm: 0755},
		"/home":                 &vfst.Dir{Perm: 0755},
		"/Kyaraben.AppImage":    "fake appimage content",
		"/bin/kyaraben-current": "fake current executable",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	homeDir := "/home"
	binDir := filepath.Join(homeDir, ".local", "bin")
	appsDir := filepath.Join(homeDir, ".local", "share", "applications")

	resolver := fakeBaseDirResolver{homeDir: homeDir}
	m := &Manager{fs: fs, paths: paths.DefaultPaths(), profileDir: "/kyaraben", resolver: resolver, executablePath: "/bin/kyaraben-current"}

	appImagePath := "/Kyaraben.AppImage"

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

	if _, err := fs.Stat(result.AppPath); err != nil {
		t.Errorf("AppImage not copied: %v", err)
	}

	cliContent, err := fs.ReadFile(result.CLIPath)
	if err != nil {
		t.Fatalf("reading CLI: %v", err)
	}
	if string(cliContent) != "fake current executable" {
		t.Errorf("CLI should be copied from executablePath, got: %s", cliContent)
	}

	content, err := fs.ReadFile(result.DesktopPath)
	if err != nil {
		t.Fatalf("reading desktop file: %v", err)
	}
	if !strings.Contains(string(content), "Exec="+result.AppPath) {
		t.Errorf("desktop file should exec AppImage, got:\n%s", content)
	}
}

func TestInstallKyarabenWithSidecar(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/kyaraben":          &vfst.Dir{Perm: 0755},
		"/home":              &vfst.Dir{Perm: 0755},
		"/Kyaraben.AppImage": "fake appimage",
		"/kyaraben-sidecar":  "fake sidecar binary",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	homeDir := "/home"
	binDir := filepath.Join(homeDir, ".local", "bin")

	resolver := fakeBaseDirResolver{homeDir: homeDir}
	m := &Manager{fs: fs, paths: paths.DefaultPaths(), profileDir: "/kyaraben", resolver: resolver}

	appImagePath := "/Kyaraben.AppImage"
	sidecarPath := "/kyaraben-sidecar"

	result, err := m.InstallKyaraben(appImagePath, sidecarPath)
	if err != nil {
		t.Fatalf("InstallKyaraben() error = %v", err)
	}

	cliContent, err := fs.ReadFile(result.CLIPath)
	if err != nil {
		t.Fatalf("reading CLI: %v", err)
	}
	if string(cliContent) != "fake sidecar binary" {
		t.Errorf("CLI should be a copy of sidecar, got: %s", cliContent)
	}

	info, err := fs.Lstat(result.CLIPath)
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

func TestGetInstallStatus(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/kyaraben": &vfst.Dir{Perm: 0755},
		"/home":     &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	homeDir := "/home"
	binDir := filepath.Join(homeDir, ".local", "bin")
	appsDir := filepath.Join(homeDir, ".local", "share", "applications")

	resolver := fakeBaseDirResolver{homeDir: homeDir}
	m := &Manager{fs: fs, paths: paths.DefaultPaths(), profileDir: "/kyaraben", resolver: resolver}

	status := m.GetInstallStatus()
	if status.AppPath != "" || status.CLIPath != "" || status.DesktopPath != "" {
		t.Error("GetInstallStatus should return empty paths when nothing installed")
	}

	if err := vfs.MkdirAll(fs, binDir, 0755); err != nil {
		t.Fatalf("creating bin dir: %v", err)
	}
	if err := vfs.MkdirAll(fs, appsDir, 0755); err != nil {
		t.Fatalf("creating apps dir: %v", err)
	}

	appPath := filepath.Join(binDir, "kyaraben-ui")
	cliPath := filepath.Join(binDir, "kyaraben")
	desktopPath := filepath.Join(appsDir, "kyaraben.desktop")

	if err := fs.WriteFile(appPath, []byte("fake"), 0755); err != nil {
		t.Fatalf("creating app: %v", err)
	}
	if err := fs.WriteFile(cliPath, []byte("fake"), 0755); err != nil {
		t.Fatalf("creating cli: %v", err)
	}
	if err := fs.WriteFile(desktopPath, []byte("fake"), 0644); err != nil {
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
