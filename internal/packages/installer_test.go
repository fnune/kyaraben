package packages

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/versions"
)

func TestMain(m *testing.M) {
	if err := versions.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestPackageInstallerInstallEmulator(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	stateDir := "/state"
	dl := NewFakeDownloader(fs, []byte("fake-appimage-binary"))
	ext := NewFakeExtractor(fs, nil)

	installer := NewPackageInstaller(fs, stateDir, dl, ext)

	var progressPhases []string
	binary, err := installer.InstallEmulator(context.Background(), "mgba", func(p InstallProgress) {
		progressPhases = append(progressPhases, p.Phase)
	})
	if err != nil {
		t.Fatalf("InstallEmulator: %v", err)
	}

	if binary.Name != "mgba" {
		t.Errorf("binary name = %q, want %q", binary.Name, "mgba")
	}

	if _, err := fs.Stat(binary.Path); err != nil {
		t.Errorf("binary not found at %s", binary.Path)
	}

	info, err := fs.Stat(binary.Path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("binary should be executable")
	}

	if len(dl.Calls) != 1 {
		t.Fatalf("expected 1 download call, got %d", len(dl.Calls))
	}
	if dl.Calls[0].SHA256 == "" {
		t.Error("download should include SHA256")
	}
}

func TestPackageInstallerSkipsAlreadyInstalled(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	stateDir := "/state"
	dl := NewFakeDownloader(fs, []byte("fake-binary"))
	ext := NewFakeExtractor(fs, nil)

	installer := NewPackageInstaller(fs, stateDir, dl, ext)

	_, err = installer.InstallEmulator(context.Background(), "mgba", nil)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	dl.Calls = nil

	var skipped bool
	_, err = installer.InstallEmulator(context.Background(), "mgba", func(p InstallProgress) {
		if p.Phase == "skipped" {
			skipped = true
		}
	})
	if err != nil {
		t.Fatalf("second install: %v", err)
	}

	if len(dl.Calls) != 0 {
		t.Error("should not re-download already installed package")
	}
	if !skipped {
		t.Error("should report skipped phase")
	}
}

func TestPackageInstallerIsEmulatorInstalled(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	stateDir := "/state"
	dl := NewFakeDownloader(fs, []byte("fake-binary"))
	ext := NewFakeExtractor(fs, nil)

	installer := NewPackageInstaller(fs, stateDir, dl, ext)

	if installer.IsEmulatorInstalled("mgba") {
		t.Error("should not be installed yet")
	}

	_, err = installer.InstallEmulator(context.Background(), "mgba", nil)
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	if !installer.IsEmulatorInstalled("mgba") {
		t.Error("should be installed after InstallEmulator")
	}
}

func TestPackageInstallerInstallArchive(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	stateDir := "/state"
	dl := NewFakeDownloader(fs, []byte("fake-zip-content"))
	ext := NewFakeExtractor(fs, map[string]string{
		"melonDS-x86_64.AppImage": "fake-melonds-binary",
	})

	installer := NewPackageInstaller(fs, stateDir, dl, ext)

	binary, err := installer.InstallEmulator(context.Background(), "melonds", nil)
	if err != nil {
		t.Fatalf("InstallEmulator: %v", err)
	}

	if binary.Name != "melonds" {
		t.Errorf("binary name = %q, want melonds", binary.Name)
	}

	if _, err := fs.Stat(binary.Path); err != nil {
		t.Errorf("binary not found at %s: %v", binary.Path, err)
	}

	if len(ext.Calls) != 1 {
		t.Fatalf("expected 1 extract call, got %d", len(ext.Calls))
	}
	if ext.Calls[0].ArchiveType != "zip" {
		t.Errorf("archive type = %q, want zip", ext.Calls[0].ArchiveType)
	}
}

func TestPackageInstallerInstallCores(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	stateDir := "/state"
	dl := NewFakeDownloader(fs, []byte("fake-cores-bundle"))
	ext := NewFakeExtractor(fs, map[string]string{
		"cores/bsnes_libretro.so": "fake-bsnes-core",
		"cores/mesen_libretro.so": "fake-mesen-core",
	})

	installer := NewPackageInstaller(fs, stateDir, dl, ext)

	cores, err := installer.InstallCores(context.Background(), []string{"bsnes", "mesen"}, nil)
	if err != nil {
		t.Fatalf("InstallCores: %v", err)
	}

	if len(cores) != 2 {
		t.Fatalf("expected 2 cores, got %d", len(cores))
	}

	for _, core := range cores {
		if _, err := fs.Stat(core.Path); err != nil {
			t.Errorf("core not found at %s: %v", core.Path, err)
		}
	}
}

func TestPackageInstallerInstallIcon(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	stateDir := "/state"
	dl := NewFakeDownloader(fs, []byte("fake-icon-data"))
	ext := NewFakeExtractor(fs, nil)

	installer := NewPackageInstaller(fs, stateDir, dl, ext)

	icon, err := installer.InstallIcon(context.Background(), "eden", "https://example.com/eden.svg", "sha256-abc123")
	if err != nil {
		t.Fatalf("InstallIcon: %v", err)
	}

	if icon.Filename != "eden.svg" {
		t.Errorf("icon filename = %q, want eden.svg", icon.Filename)
	}

	if _, err := fs.Stat(icon.Path); err != nil {
		t.Errorf("icon not found at %s", icon.Path)
	}
}

func TestPackageInstallerGarbageCollect(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	stateDir := "/state"
	dl := NewFakeDownloader(fs, []byte("fake-binary"))
	ext := NewFakeExtractor(fs, nil)
	installer := NewPackageInstaller(fs, stateDir, dl, ext)

	_, _ = installer.InstallEmulator(context.Background(), "mgba", nil)
	_, _ = installer.InstallEmulator(context.Background(), "eden", nil)

	keep := map[string]string{
		"mgba": installer.ResolveVersion("mgba"),
	}

	if err := installer.GarbageCollect(keep); err != nil {
		t.Fatalf("GarbageCollect: %v", err)
	}

	if !installer.IsEmulatorInstalled("mgba") {
		t.Error("mgba should still be installed")
	}

	edenDir := filepath.Join(installer.PackagesDir(), "eden")
	if _, err := fs.Stat(edenDir); err == nil {
		t.Error("eden should have been garbage collected")
	}
}

func TestPackageInstallerResolveVersion(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	stateDir := "/state"
	dl := NewFakeDownloader(fs, nil)
	ext := NewFakeExtractor(fs, nil)
	installer := NewPackageInstaller(fs, stateDir, dl, ext)

	version := installer.ResolveVersion("mgba")
	if version == "" {
		t.Error("should resolve a version for mgba")
	}
}

func TestPackageInstallerVersionOverride(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	stateDir := "/state"
	dl := NewFakeDownloader(fs, []byte("fake-binary"))
	ext := NewFakeExtractor(fs, nil)
	installer := NewPackageInstaller(fs, stateDir, dl, ext)
	installer.SetVersionOverrides(map[string]string{"eden": "v0.1.0"})

	version := installer.ResolveVersion("eden")
	if version != "v0.1.0" {
		t.Errorf("version = %q, want v0.1.0", version)
	}
}

func TestConcurrentInstallerInstallAll(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	stateDir := "/state"
	dl := NewFakeDownloader(fs, []byte("fake-binary"))
	ext := NewFakeExtractor(fs, nil)
	installer := NewPackageInstaller(fs, stateDir, dl, ext)

	concurrent := NewConcurrentInstaller(installer, 3)

	binaries, err := concurrent.InstallAll(context.Background(), []string{"mgba", "eden"}, nil)
	if err != nil {
		t.Fatalf("InstallAll: %v", err)
	}

	if len(binaries) != 2 {
		t.Errorf("expected 2 binaries, got %d", len(binaries))
	}
}
