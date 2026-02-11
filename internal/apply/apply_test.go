package apply

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/emulators/symlink"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
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

type mockNixClient struct {
	storePath string
	flakePath string
	gcError   error
}

func (m *mockNixClient) IsAvailable() bool { return true }
func (m *mockNixClient) Build(ctx context.Context, flakeRef string) (string, error) {
	return m.storePath, nil
}
func (m *mockNixClient) BuildWithLink(ctx context.Context, flakeRef string, outLink string) error {
	return nil
}
func (m *mockNixClient) BuildMultiple(ctx context.Context, flakeRefs []string) (map[string]string, error) {
	return nil, nil
}
func (m *mockNixClient) Eval(ctx context.Context, expr string) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockNixClient) FlakeUpdate(ctx context.Context, flakePath string) error { return nil }
func (m *mockNixClient) GetVersion(ctx context.Context) (string, error)          { return "2.18.0", nil }
func (m *mockNixClient) EnsureFlakeDir() error                                   { return nil }
func (m *mockNixClient) GetFlakePath() string                                    { return m.flakePath }
func (m *mockNixClient) FlakeCheck(ctx context.Context, flakePath string) error  { return nil }
func (m *mockNixClient) SetOutputCallback(cb func(string))                       {}
func (m *mockNixClient) SetProgressCallback(cb func(nix.BuildProgress))          {}
func (m *mockNixClient) SetExpectedPackages(packages []nix.ExpectedPackage)      {}
func (m *mockNixClient) EnsurePersistentNixPortable() (string, error)            { return "", nil }
func (m *mockNixClient) GetPersistentNixPortablePath() string                    { return "" }
func (m *mockNixClient) GetNixPortableBinary() string                            { return "" }
func (m *mockNixClient) GetNixPortableLocation() string                          { return "" }
func (m *mockNixClient) RealStorePath(path string) string                        { return path }
func (m *mockNixClient) GarbageCollect(ctx context.Context) error                { return m.gcError }

func TestUnmanagedEntriesExcludedFromManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	flakePath := filepath.Join(tmpDir, "flake")

	if err := os.MkdirAll(flakePath, 0755); err != nil {
		t.Fatalf("Failed to create flake dir: %v", err)
	}

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

	flakeGen := nix.NewFlakeGenerator(reg, reg)
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	applier := &Applier{
		NixClient:       &mockNixClient{storePath: "/nix/store/test-path", flakePath: flakePath},
		FlakeGenerator:  flakeGen,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		LauncherManager: nil,
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
	flakePath := filepath.Join(tmpDir, "flake")

	if err := os.MkdirAll(flakePath, 0755); err != nil {
		t.Fatalf("Failed to create flake dir: %v", err)
	}

	oldManifest := &model.Manifest{
		Version:     1,
		LastApplied: time.Now().Add(-time.Hour),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDMGBA: {
				ID:        model.EmulatorIDMGBA,
				Version:   "0.10.0",
				StorePath: "/nix/store/old-path",
				Installed: time.Now().Add(-time.Hour),
			},
			model.EmulatorIDRetroArchBsnes: {
				ID:        model.EmulatorIDRetroArchBsnes,
				Version:   "1.22.0",
				StorePath: "/nix/store/old-path",
				Installed: time.Now().Add(-time.Hour),
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

	flakeGen := nix.NewFlakeGenerator(reg, reg)
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	applier := &Applier{
		NixClient:       &mockNixClient{storePath: "/nix/store/new-path", flakePath: flakePath},
		FlakeGenerator:  flakeGen,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		LauncherManager: nil,
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

func TestApplyCreatesEmulatorStatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	flakePath := filepath.Join(tmpDir, "flake")

	if err := os.MkdirAll(flakePath, 0755); err != nil {
		t.Fatalf("Failed to create flake dir: %v", err)
	}

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

	flakeGen := nix.NewFlakeGenerator(reg, reg)
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	applier := &Applier{
		NixClient:       &mockNixClient{storePath: "/nix/store/test-path", flakePath: flakePath},
		FlakeGenerator:  flakeGen,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		LauncherManager: nil,
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
	flakePath := filepath.Join(tmpDir, "flake")

	if err := os.MkdirAll(flakePath, 0755); err != nil {
		t.Fatalf("Failed to create flake dir: %v", err)
	}

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

	flakeGen := nix.NewFlakeGenerator(reg, reg)
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	fakeSymlinkCreator := &symlink.FakeCreator{}

	applier := &Applier{
		NixClient:       &mockNixClient{storePath: "/nix/store/test-path", flakePath: flakePath},
		FlakeGenerator:  flakeGen,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		LauncherManager: nil,
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
}

func TestApplySucceedsWhenGCFails(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	userStorePath := filepath.Join(tmpDir, "Emulation")
	flakePath := filepath.Join(tmpDir, "flake")

	if err := os.MkdirAll(flakePath, 0755); err != nil {
		t.Fatalf("Failed to create flake dir: %v", err)
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

	flakeGen := nix.NewFlakeGenerator(reg, reg)
	configWriter := emulators.NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	var progressSteps []Progress
	applier := &Applier{
		NixClient: &mockNixClient{
			storePath: "/nix/store/test-path",
			flakePath: flakePath,
			gcError:   fmt.Errorf("nix store gc failed: Operation not permitted"),
		},
		FlakeGenerator:  flakeGen,
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
		if p.Step == "gc" {
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
