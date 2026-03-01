package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/emulators/symlink"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/packages"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
	"github.com/fnune/kyaraben/internal/testutil"
	"github.com/fnune/kyaraben/internal/versions"
)

func TestMain(m *testing.M) {
	if err := versions.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type testEnv struct {
	fs           vfs.FS
	cleanup      func()
	rootDir      string
	manifestPath string
	collection   *store.Collection
	installer    *packages.FakeInstaller
	configWriter *emulators.ConfigWriter
	applier      *Applier
	reg          *registry.Registry
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	fs := testutil.NewTestFS(t, map[string]any{
		"/home":                     &vfst.Dir{Perm: 0755},
		"/home/Emulation":           &vfst.Dir{Perm: 0755},
		"/home/Emulation/bios":      &vfst.Dir{Perm: 0755},
		"/home/Emulation/roms":      &vfst.Dir{Perm: 0755},
		"/home/Emulation/saves":     &vfst.Dir{Perm: 0755},
		"/home/packages":            &vfst.Dir{Perm: 0755},
		"/home/.config":             &vfst.Dir{Perm: 0755},
		"/home/.config/retroarch":   &vfst.Dir{Perm: 0755},
		"/home/.config/ppsspp":      &vfst.Dir{Perm: 0755},
		"/home/.config/duckstation": &vfst.Dir{Perm: 0755},
		"/home/.local":              &vfst.Dir{Perm: 0755},
		"/home/.local/share":        &vfst.Dir{Perm: 0755},
	})

	rootDir := "/home"
	manifestPath := filepath.Join(rootDir, "manifest.json")
	collectionPath := filepath.Join(rootDir, "Emulation")
	packagesDir := filepath.Join(rootDir, "packages")

	collection, err := store.NewCollection(fs, paths.DefaultPaths(), collectionPath)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	reg := registry.NewDefault()
	resolver := testutil.FakeResolver{
		ConfigDir: filepath.Join(rootDir, ".config"),
		HomeDir:   rootDir,
		DataDir:   filepath.Join(rootDir, ".local", "share"),
	}
	installer := packages.NewFakeInstaller(fs, packagesDir)
	configWriter := emulators.NewConfigWriter(fs, resolver)
	symlinkCreator := symlink.NewCreator(fs)
	applier := NewApplier(fs, installer, configWriter, reg, manifestPath, nil, resolver, symlinkCreator)

	return &testEnv{
		fs:           fs,
		cleanup:      func() {},
		rootDir:      rootDir,
		manifestPath: manifestPath,
		collection:   collection,
		installer:    installer,
		configWriter: configWriter,
		applier:      applier,
		reg:          reg,
	}
}

func TestUnmanagedEntriesExcludedFromManifest(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
		},
	}

	_, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	manifestStore := model.NewManifestStore(env.fs)
	manifest, err := manifestStore.Load(env.manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if len(manifest.ManagedConfigs) == 0 {
		t.Error("expected at least one managed config in manifest")
	}
}

func TestApplyRemovesUnenabledEmulatorsFromManifest(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	manifestStore := model.NewManifestStore(env.fs)
	oldManifest := &model.Manifest{
		Version:     1,
		LastApplied: time.Now().Add(-time.Hour),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDRetroArchMGBA: {
				ID:          model.EmulatorIDRetroArchMGBA,
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
	if err := manifestStore.Save(oldManifest, env.manifestPath); err != nil {
		t.Fatalf("Failed to save old manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
	}

	_, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	newManifest, err := manifestStore.Load(env.manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if len(newManifest.InstalledEmulators) != 1 {
		t.Errorf("Expected 1 installed emulator, got %d", len(newManifest.InstalledEmulators))
	}

	if _, ok := newManifest.InstalledEmulators[model.EmulatorIDRetroArchMGBA]; !ok {
		t.Error("Expected mGBA to be in manifest")
	}

	if _, ok := newManifest.InstalledEmulators[model.EmulatorIDRetroArchBsnes]; ok {
		t.Error("RetroArch bsnes should have been removed from manifest")
	}
}

func TestApplyRemovesConfigDirsForDisabledEmulators(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	ppssppConfigDir := filepath.Join(env.rootDir, ".config", "ppsspp")
	if err := env.fs.WriteFile(filepath.Join(ppssppConfigDir, "ppsspp.ini"), []byte("[test]"), 0644); err != nil {
		t.Fatal(err)
	}

	manifestStore := model.NewManifestStore(env.fs)
	packagesDir := filepath.Join(env.rootDir, "packages")
	oldManifest := &model.Manifest{
		Version:     1,
		LastApplied: time.Now().Add(-time.Hour),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDPPSSPP: {
				ID:          model.EmulatorIDPPSSPP,
				Version:     "v1.19.0",
				PackagePath: packagesDir,
				Installed:   time.Now().Add(-time.Hour),
			},
		},
		ManagedConfigs: []model.ManagedConfig{
			{
				EmulatorIDs: []model.EmulatorID{model.EmulatorIDPPSSPP},
				Target: model.ConfigTarget{
					RelPath: "ppsspp/ppsspp.ini",
					Format:  model.ConfigFormatINI,
					BaseDir: model.ConfigBaseDirUserConfig,
				},
				WrittenEntries: map[string]string{"section.key": "value"},
				LastModified:   time.Now().Add(-time.Hour),
			},
		},
	}
	if err := manifestStore.Save(oldManifest, env.manifestPath); err != nil {
		t.Fatalf("Failed to save old manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{},
	}

	_, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if _, err := env.fs.Stat(ppssppConfigDir); err == nil {
		t.Error("ppsspp config directory should have been removed when emulator was disabled")
	}

	newManifest, err := manifestStore.Load(env.manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if len(newManifest.ManagedConfigs) != 0 {
		t.Errorf("Expected 0 managed configs after disabling emulator, got %d", len(newManifest.ManagedConfigs))
	}
}

func TestApplyCreatesEmulatorStatesDirectories(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
			model.SystemIDGBA:  {model.EmulatorIDRetroArchMGBA},
		},
	}

	_, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	collectionPath := filepath.Join(env.rootDir, "Emulation")
	for _, tc := range []struct {
		name string
		path string
	}{
		{"retroarch bsnes core", filepath.Join(collectionPath, "states", "retroarch:bsnes")},
		{"retroarch mgba core", filepath.Join(collectionPath, "states", "retroarch:mgba")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if info, err := env.fs.Stat(tc.path); err != nil || !info.IsDir() {
				t.Errorf("States directory not created: %s", tc.path)
			}
		})
	}
}

func TestApplyCreatesSymlinks(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	fakeSymlinkCreator := &symlink.FakeCreator{}
	env.applier.SymlinkCreator = fakeSymlinkCreator

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGameCube: {model.EmulatorIDDolphin},
		},
	}

	_, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(fakeSymlinkCreator.Created) != 4 {
		t.Errorf("Expected 4 symlinks created for Dolphin, got %d", len(fakeSymlinkCreator.Created))
	}

	expectedSources := map[string]bool{
		filepath.Join(env.rootDir, ".local", "share", "dolphin-emu", "GC"):          false,
		filepath.Join(env.rootDir, ".local", "share", "dolphin-emu", "Wii"):         false,
		filepath.Join(env.rootDir, ".local", "share", "dolphin-emu", "StateSaves"):  false,
		filepath.Join(env.rootDir, ".local", "share", "dolphin-emu", "ScreenShots"): false,
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

	manifestStore := model.NewManifestStore(env.fs)
	manifest, err := manifestStore.Load(env.manifestPath)
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
	t.Parallel()

	env := newTestEnv(t)

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGameCube: {model.EmulatorIDDolphin},
		},
	}

	if _, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	collectionPath := filepath.Join(env.rootDir, "Emulation")
	paths := []string{
		filepath.Join(collectionPath, "saves", "gamecube", "USA"),
		filepath.Join(collectionPath, "saves", "gamecube", "EUR"),
		filepath.Join(collectionPath, "saves", "gamecube", "JAP"),
		filepath.Join(collectionPath, "bios", "gba"),
	}

	for _, path := range paths {
		assertDirExistsVFS(t, env.fs, path)
	}
}

func assertDirExistsVFS(t *testing.T, fs vfs.FS, path string) {
	t.Helper()
	info, err := fs.Stat(path)
	if err != nil {
		t.Fatalf("directory %s missing: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s exists but is not a directory", path)
	}
}

func TestApplySucceedsWhenGCFails(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
	}

	env.applier.Installer = &failingGCInstaller{FakeInstaller: env.installer}

	var progressSteps []Progress
	_, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{
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
	t.Parallel()

	env := newTestEnv(t)

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA:  {model.EmulatorIDRetroArchMGBA},
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
			model.SystemIDPSX:  {model.EmulatorIDDuckStation},
		},
	}

	env.installer.Versions["duckstation"] = "v0.1-10655"
	env.installer.Versions["retroarch"] = "1.22.0"

	var progressEvents []Progress
	result, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{
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

	manifestStore := model.NewManifestStore(env.fs)
	manifest, err := manifestStore.Load(env.manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if len(manifest.InstalledEmulators) < 3 {
		t.Errorf("Expected at least 3 installed emulators, got %d", len(manifest.InstalledEmulators))
	}

	if mgba, ok := manifest.InstalledEmulators[model.EmulatorIDRetroArchMGBA]; ok {
		if mgba.Version != "1.22.2" {
			t.Errorf("mGBA (RetroArch) version = %q, want 1.22.2", mgba.Version)
		}
	} else {
		t.Error("mGBA should be in manifest")
	}

	packagesDir := filepath.Join(env.rootDir, "packages")
	for id, emu := range manifest.InstalledEmulators {
		if emu.PackagePath != packagesDir {
			t.Errorf("InstalledEmulators[%s].PackagePath = %q, want %q", id, emu.PackagePath, packagesDir)
		}
	}

	if len(env.installer.GCCalls) != 1 {
		t.Errorf("Expected 1 GC call, got %d", len(env.installer.GCCalls))
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
	t.Parallel()

	env := newTestEnv(t)

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
		Frontends: map[model.FrontendID]model.FrontendConfig{
			model.FrontendIDESDE: {Enabled: true},
		},
	}

	env.installer.Versions["retroarch"] = "1.22.0"
	env.installer.Versions["esde"] = "3.0.0"

	_, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	manifestStore := model.NewManifestStore(env.fs)
	manifest, err := manifestStore.Load(env.manifestPath)
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

	packagesDir := filepath.Join(env.rootDir, "packages")
	binaryPath := filepath.Join(packagesDir, "esde", "bin", "esde")
	if _, err := env.fs.Stat(binaryPath); err != nil {
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
	t.Parallel()

	env := newTestEnv(t)

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
		Frontends: map[model.FrontendID]model.FrontendConfig{
			model.FrontendIDESDE: {Enabled: true},
		},
	}

	env.installer.Versions["retroarch"] = "1.22.0"
	env.installer.Versions["esde"] = "3.0.0"
	installer := &iconTrackingInstaller{FakeInstaller: env.installer}
	env.applier.Installer = installer

	_, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	hasESDE := false
	for _, name := range installer.iconCalls {
		if name == "esde" {
			hasESDE = true
			break
		}
	}

	if !hasESDE {
		t.Errorf("InstallIcon should be called for esde, got calls for: %v", installer.iconCalls)
	}
}

func TestGarbageCollectKeepsSyncthing(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
	}

	env.installer.Versions["retroarch"] = "1.22.0"
	env.installer.Versions["syncthing"] = "v2.0.14"

	_, err := env.applier.Apply(context.Background(), cfg, env.collection, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(env.installer.GCCalls) != 1 {
		t.Fatalf("Expected 1 GC call, got %d", len(env.installer.GCCalls))
	}

	keepMap := env.installer.GCCalls[0]
	if _, ok := keepMap["syncthing"]; !ok {
		t.Errorf("syncthing should always be in the GC keep map, got: %v", keepMap)
	}
}

func TestCoreDownloadSizeUsesBundleSize(t *testing.T) {
	t.Parallel()

	v := versions.MustGet()
	installer := packages.NewFakeInstaller(nil, "/packages")

	size := coreDownloadSize([]string{"bsnes", "snes9x", "mesen"}, "x64", v, installer)

	bsnesSpec, _ := v.GetPackage("bsnes")
	if bsnesSpec.BundleSize == 0 {
		t.Fatal("bsnes should have BundleSize configured")
	}

	if size != bsnesSpec.BundleSize {
		t.Errorf("coreDownloadSize = %d, want %d (bundle size counted once despite multiple cores)", size, bsnesSpec.BundleSize)
	}
}

func TestCoreDownloadSizeFallsBackToIndividualSize(t *testing.T) {
	t.Parallel()

	v := versions.MustGet()
	installer := packages.NewFakeInstaller(nil, "/packages")

	size := coreDownloadSize([]string{"melondsds"}, "x64", v, installer)

	spec, _ := v.GetPackage("melondsds")
	entry := spec.GetDefault()
	build := entry.Target("x64")

	if size != build.Size {
		t.Errorf("coreDownloadSize for non-bundled core = %d, want %d", size, build.Size)
	}
}

func TestCoreDownloadSizeCountsMultipleBundles(t *testing.T) {
	t.Parallel()

	v := versions.MustGet()
	installer := packages.NewFakeInstaller(nil, "/packages")
	installer.Versions["bsnes"] = "1.22.2"
	installer.Versions["snes9x"] = "1.19.1"

	size := coreDownloadSize([]string{"bsnes", "snes9x"}, "x64", v, installer)

	bsnesSpec, _ := v.GetPackage("bsnes")
	snes9xSpec, _ := v.GetPackage("snes9x")

	expectedSize := bsnesSpec.BundleSize + snes9xSpec.BundleSize
	if size != expectedSize {
		t.Errorf("coreDownloadSize with different versions = %d, want %d (two bundles)", size, expectedSize)
	}
}

func TestCoreDownloadSizeUsesResolvedVersion(t *testing.T) {
	t.Parallel()

	v := versions.MustGet()
	installer := packages.NewFakeInstaller(nil, "/packages")
	installer.Versions["bsnes"] = "1.19.1"

	size := coreDownloadSize([]string{"bsnes"}, "x64", v, installer)

	spec, _ := v.GetPackage("bsnes")
	if size != spec.BundleSize {
		t.Errorf("coreDownloadSize should use resolved version, got %d want %d", size, spec.BundleSize)
	}
}

func TestCoreArchiveTypeUsesResolvedVersion(t *testing.T) {
	t.Parallel()

	v := versions.MustGet()
	installer := packages.NewFakeInstaller(nil, "/packages")
	installer.Versions["bsnes"] = "1.22.2"

	archiveType := coreArchiveType([]string{"bsnes"}, "x64", v, installer)

	if archiveType != "7z" {
		t.Errorf("coreArchiveType = %q, want 7z", archiveType)
	}
}
