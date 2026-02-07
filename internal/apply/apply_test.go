package apply

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
	"github.com/fnune/kyaraben/internal/versions"
)

func TestMain(m *testing.M) {
	if err := versions.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type mockNixClient struct {
	storePath string
	flakePath string
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
func (m *mockNixClient) EnsurePersistentNixPortable() (string, error)            { return "", nil }
func (m *mockNixClient) GetPersistentNixPortablePath() string                    { return "" }
func (m *mockNixClient) GetNixPortableBinary() string                            { return "" }
func (m *mockNixClient) GetNixPortableLocation() string                          { return "" }
func (m *mockNixClient) RealStorePath(path string) string                        { return path }

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

	flakeGen := nix.NewFlakeGenerator(reg)
	configWriter := emulators.NewConfigWriter()

	applier := &Applier{
		NixClient:       &mockNixClient{storePath: "/nix/store/test-path", flakePath: flakePath},
		FlakeGenerator:  flakeGen,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		LauncherManager: nil,
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

	foundSystemDir := false
	for _, cfg := range manifest.ManagedConfigs {
		for _, key := range cfg.ManagedKeys {
			keyName := key.Path[len(key.Path)-1]
			if keyName == "system_directory" {
				foundSystemDir = true
			}
		}
	}
	if !foundSystemDir {
		t.Error("system_directory should be in ManagedKeys (it's managed)")
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

	flakeGen := nix.NewFlakeGenerator(reg)
	configWriter := emulators.NewConfigWriter()

	applier := &Applier{
		NixClient:       &mockNixClient{storePath: "/nix/store/new-path", flakePath: flakePath},
		FlakeGenerator:  flakeGen,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		LauncherManager: nil,
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
