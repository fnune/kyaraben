package status

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
)

func mustNewUserStore(t *testing.T, path string) *store.UserStore {
	t.Helper()
	s, err := store.NewUserStore(path)
	if err != nil {
		t.Fatalf("NewUserStore(%q) failed: %v", path, err)
	}
	return s
}

func TestGet(t *testing.T) {
	tmpDir := t.TempDir()
	userStorePath := filepath.Join(tmpDir, "Emulation")
	configPath := filepath.Join(tmpDir, "config.toml")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemIDGBA:  {Emulator: string(model.EmulatorIDMGBA)},
			model.SystemIDSNES: {Emulator: string(model.EmulatorIDRetroArchBsnes)},
		},
	}

	registry := registry.NewDefault()
	userStore := mustNewUserStore(t, userStorePath)

	result, err := Get(context.Background(), cfg, configPath, registry, userStore, manifestPath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result.ConfigPath != configPath {
		t.Errorf("ConfigPath: got %s, want %s", result.ConfigPath, configPath)
	}

	if result.UserStorePath != userStorePath {
		t.Errorf("UserStorePath: got %s, want %s", result.UserStorePath, userStorePath)
	}

	if result.UserStoreInitialized {
		t.Error("UserStoreInitialized should be false for non-existent directory")
	}

	if len(result.EnabledSystems) != 2 {
		t.Errorf("EnabledSystems: got %d, want 2", len(result.EnabledSystems))
	}

	if len(result.InstalledEmulators) != 0 {
		t.Errorf("InstalledEmulators: got %d, want 0 (no manifest)", len(result.InstalledEmulators))
	}
}

func TestGetWithInitializedStore(t *testing.T) {
	tmpDir := t.TempDir()
	userStorePath := filepath.Join(tmpDir, "Emulation")
	configPath := filepath.Join(tmpDir, "config.toml")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create all required directories for IsInitialized to return true
	for _, dir := range []string{"roms", "bios", "saves", "states", "screenshots", "opaque"} {
		if err := os.MkdirAll(filepath.Join(userStorePath, dir), 0755); err != nil {
			t.Fatalf("Failed to create %s dir: %v", dir, err)
		}
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemIDGBA: {Emulator: string(model.EmulatorIDMGBA)},
		},
	}

	registry := registry.NewDefault()
	userStore := mustNewUserStore(t, userStorePath)

	result, err := Get(context.Background(), cfg, configPath, registry, userStore, manifestPath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.UserStoreInitialized {
		t.Error("UserStoreInitialized should be true when all directories exist")
	}
}

func TestGetSystemNames(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: tmpDir,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemIDGBA: {Emulator: string(model.EmulatorIDMGBA)},
		},
	}

	registry := registry.NewDefault()
	userStore := mustNewUserStore(t, tmpDir)

	result, err := Get(context.Background(), cfg, tmpDir, registry, userStore, manifestPath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(result.EnabledSystems) != 1 {
		t.Fatalf("Expected 1 system, got %d", len(result.EnabledSystems))
	}

	sys := result.EnabledSystems[0]
	if sys.ID != model.SystemIDGBA {
		t.Errorf("System ID: got %s, want %s", sys.ID, model.SystemIDGBA)
	}
	if sys.Name != "Game Boy Advance" {
		t.Errorf("System Name: got %s, want Game Boy Advance", sys.Name)
	}
}

func TestGetMissingRequiredCount(t *testing.T) {
	tmpDir := t.TempDir()
	userStorePath := filepath.Join(tmpDir, "Emulation")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	if err := os.MkdirAll(filepath.Join(userStorePath, "bios", "psx"), 0755); err != nil {
		t.Fatalf("Failed to create bios dir: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemIDPSX: {Emulator: string(model.EmulatorIDDuckStation)},
			model.SystemIDGBA: {Emulator: string(model.EmulatorIDMGBA)},
		},
	}

	registry := registry.NewDefault()
	userStore := mustNewUserStore(t, userStorePath)

	result, err := Get(context.Background(), cfg, tmpDir, registry, userStore, manifestPath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result.MissingRequiredCount != 1 {
		t.Errorf("MissingRequiredCount: got %d, want 1 (PSX missing BIOS)", result.MissingRequiredCount)
	}
}

func TestGetWithManifest(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &model.Manifest{
		LastApplied: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDMGBA: {
				ID:        model.EmulatorIDMGBA,
				Version:   "latest",
				StorePath: "/nix/store/abc123",
				Installed: time.Now(),
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, "manifest.json")
	if err := manifest.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: tmpDir,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemIDGBA: {Emulator: string(model.EmulatorIDMGBA)},
		},
	}

	registry := registry.NewDefault()
	userStore := mustNewUserStore(t, tmpDir)

	result, err := Get(context.Background(), cfg, tmpDir, registry, userStore, manifestPath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(result.InstalledEmulators) != 1 {
		t.Fatalf("InstalledEmulators: got %d, want 1", len(result.InstalledEmulators))
	}

	emu := result.InstalledEmulators[0]
	if emu.ID != model.EmulatorIDMGBA {
		t.Errorf("Emulator ID: got %s, want %s", emu.ID, model.EmulatorIDMGBA)
	}
	if emu.Name != "mGBA" {
		t.Errorf("Emulator Name: got %s, want mGBA", emu.Name)
	}
	if emu.Version != "latest" {
		t.Errorf("Emulator Version: got %s, want latest", emu.Version)
	}

	if result.LastApplied.IsZero() {
		t.Error("LastApplied should not be zero")
	}
}
