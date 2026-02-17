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
	userStore    *store.UserStore
	installer    *packages.FakeInstaller
	configWriter *emulators.ConfigWriter
	applier      *Applier
	reg          *registry.Registry
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	fs := testutil.NewTestFS(t, map[string]any{
		"/home":                      &vfst.Dir{Perm: 0755},
		"/home/Emulation":            &vfst.Dir{Perm: 0755},
		"/home/Emulation/bios":       &vfst.Dir{Perm: 0755},
		"/home/Emulation/roms":       &vfst.Dir{Perm: 0755},
		"/home/Emulation/saves":      &vfst.Dir{Perm: 0755},
		"/home/packages":             &vfst.Dir{Perm: 0755},
		"/home/.config":              &vfst.Dir{Perm: 0755},
		"/home/.config/retroarch":    &vfst.Dir{Perm: 0755},
		"/home/.config/mgba":         &vfst.Dir{Perm: 0755},
		"/home/.config/duckstation":  &vfst.Dir{Perm: 0755},
		"/home/.local":               &vfst.Dir{Perm: 0755},
		"/home/.local/share":         &vfst.Dir{Perm: 0755},
		"/home/.local/share/melonDS": &vfst.Dir{Perm: 0755},
	})

	rootDir := "/home"
	manifestPath := filepath.Join(rootDir, "manifest.json")
	userStorePath := filepath.Join(rootDir, "Emulation")
	packagesDir := filepath.Join(rootDir, "packages")

	userStore, err := store.NewUserStore(fs, paths.DefaultPaths(), userStorePath)
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
		userStore:    userStore,
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
			UserStore: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
		},
	}

	_, err := env.applier.Apply(context.Background(), cfg, env.userStore, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	manifestStore := model.NewManifestStore(env.fs)
	manifest, err := manifestStore.Load(env.manifestPath)
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
	t.Parallel()

	env := newTestEnv(t)

	manifestStore := model.NewManifestStore(env.fs)
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
	if err := manifestStore.Save(oldManifest, env.manifestPath); err != nil {
		t.Fatalf("Failed to save old manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
	}

	_, err := env.applier.Apply(context.Background(), cfg, env.userStore, Options{})
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

	if _, ok := newManifest.InstalledEmulators[model.EmulatorIDMGBA]; !ok {
		t.Error("Expected mGBA to be in manifest")
	}

	if _, ok := newManifest.InstalledEmulators[model.EmulatorIDRetroArchBsnes]; ok {
		t.Error("RetroArch bsnes should have been removed from manifest")
	}
}

func TestApplyRemovesConfigDirsForDisabledEmulators(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	mgbaConfigDir := filepath.Join(env.rootDir, ".config", "mgba")
	if err := env.fs.WriteFile(filepath.Join(mgbaConfigDir, "config.ini"), []byte("[test]"), 0644); err != nil {
		t.Fatal(err)
	}

	manifestStore := model.NewManifestStore(env.fs)
	packagesDir := filepath.Join(env.rootDir, "packages")
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
	if err := manifestStore.Save(oldManifest, env.manifestPath); err != nil {
		t.Fatalf("Failed to save old manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{},
	}

	_, err := env.applier.Apply(context.Background(), cfg, env.userStore, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if _, err := env.fs.Stat(mgbaConfigDir); err == nil {
		t.Error("mgba config directory should have been removed when emulator was disabled")
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
			UserStore: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
			model.SystemIDGBA:  {model.EmulatorIDMGBA},
		},
	}

	_, err := env.applier.Apply(context.Background(), cfg, env.userStore, Options{})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	userStorePath := filepath.Join(env.rootDir, "Emulation")
	for _, tc := range []struct {
		name string
		path string
	}{
		{"retroarch core", filepath.Join(userStorePath, "states", "retroarch:bsnes")},
		{"standalone emulator", filepath.Join(userStorePath, "states", "mgba")},
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
			UserStore: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGameCube: {model.EmulatorIDDolphin},
		},
	}

	_, err := env.applier.Apply(context.Background(), cfg, env.userStore, Options{})
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
			UserStore: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGameCube: {model.EmulatorIDDolphin},
		},
	}

	if _, err := env.applier.Apply(context.Background(), cfg, env.userStore, Options{}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	userStorePath := filepath.Join(env.rootDir, "Emulation")
	paths := []string{
		filepath.Join(userStorePath, "saves", "gamecube", "USA"),
		filepath.Join(userStorePath, "saves", "gamecube", "EUR"),
		filepath.Join(userStorePath, "saves", "gamecube", "JAP"),
		filepath.Join(userStorePath, "bios", "gba"),
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
			UserStore: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
	}

	env.applier.Installer = &failingGCInstaller{FakeInstaller: env.installer}

	var progressSteps []Progress
	_, err := env.applier.Apply(context.Background(), cfg, env.userStore, Options{
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
			UserStore: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA:  {model.EmulatorIDMGBA},
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
			model.SystemIDPSX:  {model.EmulatorIDDuckStation},
		},
	}

	env.installer.Versions["mgba"] = "0.10.3"
	env.installer.Versions["duckstation"] = "v0.1-10655"
	env.installer.Versions["retroarch"] = "1.22.0"

	var progressEvents []Progress
	result, err := env.applier.Apply(context.Background(), cfg, env.userStore, Options{
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

	if mgba, ok := manifest.InstalledEmulators[model.EmulatorIDMGBA]; ok {
		if mgba.Version != "0.10.3" {
			t.Errorf("mGBA version = %q, want 0.10.3", mgba.Version)
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
			UserStore: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
		Frontends: map[model.FrontendID]model.FrontendConfig{
			model.FrontendIDESDE: {Enabled: true},
		},
	}

	env.installer.Versions["mgba"] = "0.10.3"
	env.installer.Versions["es-de"] = "3.0.0"

	_, err := env.applier.Apply(context.Background(), cfg, env.userStore, Options{})
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
	binaryPath := filepath.Join(packagesDir, "es-de", "bin", "es-de")
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
			UserStore: filepath.Join(env.rootDir, "Emulation"),
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
		Frontends: map[model.FrontendID]model.FrontendConfig{
			model.FrontendIDESDE: {Enabled: true},
		},
	}

	env.installer.Versions["mgba"] = "0.10.3"
	env.installer.Versions["es-de"] = "3.0.0"
	installer := &iconTrackingInstaller{FakeInstaller: env.installer}
	env.applier.Installer = installer

	_, err := env.applier.Apply(context.Background(), cfg, env.userStore, Options{})
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
