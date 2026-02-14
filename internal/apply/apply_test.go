package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/emulators/symlink"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/packages"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
	"github.com/fnune/kyaraben/internal/versions"
)

type fakeBaseDirResolver struct {
	root string
}

func (f fakeBaseDirResolver) UserConfigDir() (string, error) {
	return filepath.Join(f.root, ".config"), nil
}

func (f fakeBaseDirResolver) UserHomeDir() (string, error) {
	return f.root, nil
}

func (f fakeBaseDirResolver) UserDataDir() (string, error) {
	return filepath.Join(f.root, ".local", "share"), nil
}

func TestMain(m *testing.M) {
	if err := versions.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestUnmanagedEntriesExcludedFromManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	packagesDir := filepath.Join(tmpDir, "packages")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
		},
	}

	reg := registry.NewDefault()
	userStore, err := store.NewUserStore(userStorePath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	installer := packages.NewFakeInstaller(packagesDir)
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	applier := &Applier{
		Installer:       installer,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		BaseDirResolver: fakeBaseDirResolver{root: tmpDir},
	}

	_, err = applier.Apply(context.Background(), cfg, userStore, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	manifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	for _, cfg := range manifest.ManagedConfigs {
		for _, key := range cfg.ManagedKeys {
			keyName := key.Path[len(key.Path)-1]
			if keyName == "menu_driver" {
				t.Errorf("menu_driver should not be in ManagedKeys (it's unmanaged)")
			}
		}
	}

	foundLibretroDir := false
	for _, cfg := range manifest.ManagedConfigs {
		for _, key := range cfg.ManagedKeys {
			keyName := key.Path[len(key.Path)-1]
			if keyName == "libretro_directory" {
				foundLibretroDir = true
			}
		}
	}
	if !foundLibretroDir {
		t.Error("libretro_directory should be in ManagedKeys (it's managed)")
	}
}

func TestApplyRemovesUnenabledEmulatorsFromManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	packagesDir := filepath.Join(tmpDir, "packages")

	oldManifest := &model.Manifest{
		Version:     1,
		LastApplied: time.Now().Add(-time.Hour),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDMGBA: {
				ID:          model.EmulatorIDMGBA,
				Version:     "0.10.0",
				PackagePath: "/old/packages",
				Installed:   time.Now().Add(-time.Hour),
			},
			model.EmulatorIDRetroArchBsnes: {
				ID:          model.EmulatorIDRetroArchBsnes,
				Version:     "1.22.0",
				PackagePath: "/old/packages",
				Installed:   time.Now().Add(-time.Hour),
			},
		},
	}
	if err := oldManifest.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save old manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
	}

	reg := registry.NewDefault()
	userStore, err := store.NewUserStore(userStorePath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	installer := packages.NewFakeInstaller(packagesDir)
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	applier := &Applier{
		Installer:       installer,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		BaseDirResolver: fakeBaseDirResolver{root: tmpDir},
	}

	_, err = applier.Apply(context.Background(), cfg, userStore, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	newManifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if len(newManifest.InstalledEmulators) != 1 {
		t.Errorf("Expected 1 installed emulator, got %d", len(newManifest.InstalledEmulators))
	}

	if _, ok := newManifest.InstalledEmulators[model.EmulatorIDMGBA]; !ok {
		t.Error("Expected mGBA to be in manifest")
	}

	if _, ok := newManifest.InstalledEmulators[model.EmulatorIDRetroArchBsnes]; ok {
		t.Error("RetroArch bsnes should have been removed from manifest")
	}
}

func TestApplyRemovesConfigDirsForDisabledEmulators(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	packagesDir := filepath.Join(tmpDir, "packages")
	configDir := filepath.Join(tmpDir, ".config")

	mgbaConfigDir := filepath.Join(configDir, "mgba")
	if err := os.MkdirAll(mgbaConfigDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mgbaConfigDir, "config.ini"), []byte("[test]"), 0644); err != nil {
		t.Fatal(err)
	}

	oldManifest := &model.Manifest{
		Version:     1,
		LastApplied: time.Now().Add(-time.Hour),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDMGBA: {
				ID:          model.EmulatorIDMGBA,
				Version:     "0.10.0",
				PackagePath: packagesDir,
				Installed:   time.Now().Add(-time.Hour),
			},
		},
		ManagedConfigs: []model.ManagedConfig{
			{
				EmulatorIDs: []model.EmulatorID{model.EmulatorIDMGBA},
				Target: model.ConfigTarget{
					RelPath: "mgba/config.ini",
					Format:  model.ConfigFormatINI,
					BaseDir: model.ConfigBaseDirUserConfig,
				},
				BaselineHash: "abc123",
				LastModified: time.Now().Add(-time.Hour),
			},
		},
	}
	if err := oldManifest.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save old manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{},
	}

	reg := registry.NewDefault()
	userStore, err := store.NewUserStore(userStorePath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	installer := packages.NewFakeInstaller(packagesDir)
	resolver := fakeBaseDirResolver{root: tmpDir}
	configWriter := emulators.NewConfigWriter(resolver)

	applier := &Applier{
		Installer:       installer,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		BaseDirResolver: resolver,
	}

	_, err = applier.Apply(context.Background(), cfg, userStore, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if _, err := os.Stat(mgbaConfigDir); !os.IsNotExist(err) {
		t.Error("mgba config directory should have been removed when emulator was disabled")
	}

	newManifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if len(newManifest.ManagedConfigs) != 0 {
		t.Errorf("Expected 0 managed configs after disabling emulator, got %d", len(newManifest.ManagedConfigs))
	}
}

func TestApplyCreatesEmulatorStatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	packagesDir := filepath.Join(tmpDir, "packages")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
			model.SystemIDGBA:  {model.EmulatorIDMGBA},
		},
	}

	reg := registry.NewDefault()
	userStore, err := store.NewUserStore(userStorePath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	installer := packages.NewFakeInstaller(packagesDir)
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	applier := &Applier{
		Installer:       installer,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		BaseDirResolver: fakeBaseDirResolver{root: tmpDir},
	}

	_, err = applier.Apply(context.Background(), cfg, userStore, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	for _, tc := range []struct {
		name string
		path string
	}{
		{"retroarch core", filepath.Join(userStorePath, "states", "retroarch:bsnes")},
		{"standalone emulator", filepath.Join(userStorePath, "states", "mgba")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if info, err := os.Stat(tc.path); err != nil || !info.IsDir() {
				t.Errorf("States directory not created: %s", tc.path)
			}
		})
	}
}

func TestApplyCreatesSymlinksForSymlinkProviders(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	packagesDir := filepath.Join(tmpDir, "packages")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGameCube: {model.EmulatorIDDolphin},
		},
	}

	reg := registry.NewDefault()
	userStore, err := store.NewUserStore(userStorePath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	installer := packages.NewFakeInstaller(packagesDir)
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})
	fakeSymlinkCreator := &symlink.FakeCreator{}

	applier := &Applier{
		Installer:       installer,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		BaseDirResolver: fakeBaseDirResolver{root: tmpDir},
		SymlinkCreator:  fakeSymlinkCreator,
	}

	_, err = applier.Apply(context.Background(), cfg, userStore, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(fakeSymlinkCreator.Created) != 4 {
		t.Errorf("Expected 4 symlinks created for Dolphin, got %d", len(fakeSymlinkCreator.Created))
	}

	expectedSources := map[string]bool{
		filepath.Join(tmpDir, ".local", "share", "dolphin-emu", "GC"):          false,
		filepath.Join(tmpDir, ".local", "share", "dolphin-emu", "Wii"):         false,
		filepath.Join(tmpDir, ".local", "share", "dolphin-emu", "StateSaves"):  false,
		filepath.Join(tmpDir, ".local", "share", "dolphin-emu", "ScreenShots"): false,
	}

	for _, spec := range fakeSymlinkCreator.Created {
		if _, ok := expectedSources[spec.Source]; ok {
			expectedSources[spec.Source] = true
		}
	}

	for source, found := range expectedSources {
		if !found {
			t.Errorf("Expected symlink source %q not found", source)
		}
	}

	manifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if len(manifest.Symlinks) != 4 {
		t.Errorf("Expected 4 symlinks in manifest, got %d", len(manifest.Symlinks))
	}

	for _, s := range manifest.Symlinks {
		if s.EmulatorID != model.EmulatorIDDolphin {
			t.Errorf("Expected emulator ID %s, got %s", model.EmulatorIDDolphin, s.EmulatorID)
		}
	}
}

func TestApplyCreatesProvisionDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	packagesDir := filepath.Join(tmpDir, "packages")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGameCube: {model.EmulatorIDDolphin},
		},
	}

	reg := registry.NewDefault()
	userStore, err := store.NewUserStore(userStorePath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	installer := packages.NewFakeInstaller(packagesDir)
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	applier := &Applier{
		Installer:       installer,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		BaseDirResolver: fakeBaseDirResolver{root: tmpDir},
		SymlinkCreator:  &symlink.FakeCreator{},
	}

	if _, err := applier.Apply(context.Background(), cfg, userStore, Options{}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	paths := []string{
		filepath.Join(userStorePath, "saves", "gamecube", "USA"),
		filepath.Join(userStorePath, "saves", "gamecube", "EUR"),
		filepath.Join(userStorePath, "saves", "gamecube", "JAP"),
		filepath.Join(userStorePath, "bios", "gba"),
	}

	for _, path := range paths {
		assertDirExists(t, path)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory %s missing: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s exists but is not a directory", path)
	}
}

func TestApplySucceedsWhenGCFails(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	packagesDir := filepath.Join(tmpDir, "packages")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
	}

	reg := registry.NewDefault()
	userStore, err := store.NewUserStore(userStorePath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	installer := packages.NewFakeInstaller(packagesDir)
	origGC := installer.GarbageCollect
	_ = origGC
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	var progressSteps []Progress
	applier := &Applier{
		Installer:       &failingGCInstaller{FakeInstaller: installer},
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		BaseDirResolver: fakeBaseDirResolver{root: tmpDir},
	}

	_, err = applier.Apply(context.Background(), cfg, userStore, Options{
		OnProgress: func(p Progress) {
			progressSteps = append(progressSteps, p)
		},
	})
	if err != nil {
		t.Fatalf("Apply should succeed even when GC fails, got: %v", err)
	}

	var gcSteps []Progress
	for _, p := range progressSteps {
		if p.Step == "cleanup" {
			gcSteps = append(gcSteps, p)
		}
	}

	if len(gcSteps) != 2 {
		t.Fatalf("Expected 2 gc progress events, got %d", len(gcSteps))
	}
	if gcSteps[1].Message == "" {
		t.Error("GC failure should report a message")
	}
}

type failingGCInstaller struct {
	*packages.FakeInstaller
}

func (f *failingGCInstaller) GarbageCollect(keep map[string]string) error {
	return fmt.Errorf("gc failed: Operation not permitted")
}

func TestApplyWithFakeInstallerE2E(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	packagesDir := filepath.Join(tmpDir, "packages")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA:  {model.EmulatorIDMGBA},
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
			model.SystemIDPSX:  {model.EmulatorIDDuckStation},
		},
	}

	reg := registry.NewDefault()
	userStore, err := store.NewUserStore(userStorePath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	installer := packages.NewFakeInstaller(packagesDir)
	installer.Versions["mgba"] = "0.10.3"
	installer.Versions["duckstation"] = "v0.1-10655"
	installer.Versions["retroarch"] = "1.22.0"

	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	var progressEvents []Progress
	applier := &Applier{
		Installer:       installer,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		BaseDirResolver: fakeBaseDirResolver{root: tmpDir},
	}

	result, err := applier.Apply(context.Background(), cfg, userStore, Options{
		OnProgress: func(p Progress) {
			progressEvents = append(progressEvents, p)
		},
	})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(result.Patches) == 0 {
		t.Error("Expected config patches to be generated")
	}

	manifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if len(manifest.InstalledEmulators) < 3 {
		t.Errorf("Expected at least 3 installed emulators, got %d", len(manifest.InstalledEmulators))
	}

	if mgba, ok := manifest.InstalledEmulators[model.EmulatorIDMGBA]; ok {
		if mgba.Version != "0.10.3" {
			t.Errorf("mGBA version = %q, want 0.10.3", mgba.Version)
		}
	} else {
		t.Error("mGBA should be in manifest")
	}

	for id, emu := range manifest.InstalledEmulators {
		if emu.PackagePath != packagesDir {
			t.Errorf("InstalledEmulators[%s].PackagePath = %q, want %q", id, emu.PackagePath, packagesDir)
		}
	}

	if len(installer.GCCalls) != 1 {
		t.Errorf("Expected 1 GC call, got %d", len(installer.GCCalls))
	}

	hasBuild := false
	hasFinalize := false
	for _, p := range progressEvents {
		if p.Step == "build" {
			hasBuild = true
		}
		if p.Step == "finalize" {
			hasFinalize = true
		}
	}
	if !hasBuild {
		t.Error("Expected build progress events")
	}
	if !hasFinalize {
		t.Error("Expected finalize progress events")
	}
}

func TestApplyInstallsFrontend(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	packagesDir := filepath.Join(tmpDir, "packages")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
		Frontends: map[model.FrontendID]model.FrontendConfig{
			model.FrontendIDESDE: {Enabled: true},
		},
	}

	reg := registry.NewDefault()
	userStore, err := store.NewUserStore(userStorePath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	installer := packages.NewFakeInstaller(packagesDir)
	installer.Versions["mgba"] = "0.10.3"
	installer.Versions["es-de"] = "3.0.0"

	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	applier := &Applier{
		Installer:       installer,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		BaseDirResolver: fakeBaseDirResolver{root: tmpDir},
	}

	_, err = applier.Apply(context.Background(), cfg, userStore, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	manifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if _, ok := manifest.InstalledFrontends[model.FrontendIDESDE]; !ok {
		t.Error("ES-DE frontend should be in manifest")
	}

	if esde, ok := manifest.InstalledFrontends[model.FrontendIDESDE]; ok {
		if esde.Version != "3.0.0" {
			t.Errorf("ES-DE version = %q, want 3.0.0", esde.Version)
		}
	}

	binaryPath := filepath.Join(packagesDir, "es-de", "bin", "es-de")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Errorf("ES-DE binary should be installed at %s", binaryPath)
	}
}

type iconTrackingInstaller struct {
	*packages.FakeInstaller
	iconCalls []string
}

func (i *iconTrackingInstaller) InstallIcon(ctx context.Context, binaryName, url, sha256 string) (*packages.InstalledIcon, error) {
	i.iconCalls = append(i.iconCalls, binaryName)
	return i.FakeInstaller.InstallIcon(ctx, binaryName, url, sha256)
}

func TestApplyInstallsFrontendIcon(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	packagesDir := filepath.Join(tmpDir, "packages")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
		Frontends: map[model.FrontendID]model.FrontendConfig{
			model.FrontendIDESDE: {Enabled: true},
		},
	}

	reg := registry.NewDefault()
	userStore, err := store.NewUserStore(userStorePath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	fakeInstaller := packages.NewFakeInstaller(packagesDir)
	fakeInstaller.Versions["mgba"] = "0.10.3"
	fakeInstaller.Versions["es-de"] = "3.0.0"
	installer := &iconTrackingInstaller{FakeInstaller: fakeInstaller}

	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	applier := &Applier{
		Installer:       installer,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		BaseDirResolver: fakeBaseDirResolver{root: tmpDir},
	}

	_, err = applier.Apply(context.Background(), cfg, userStore, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	hasESDE := false
	for _, name := range installer.iconCalls {
		if name == "es-de" {
			hasESDE = true
			break
		}
	}

	if !hasESDE {
		t.Errorf("InstallIcon should be called for es-de, got calls for: %v", installer.iconCalls)
	}
}
