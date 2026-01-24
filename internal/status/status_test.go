package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
)

func TestGet(t *testing.T) {
	tmpDir := t.TempDir()
	userStore := filepath.Join(tmpDir, "Emulation")
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStore,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemTIC80: {Emulator: model.EmulatorTIC80},
			model.SystemSNES:  {Emulator: model.EmulatorRetroArchBsnes},
		},
	}

	registry := emulators.NewRegistry()

	result, err := Get(cfg, configPath, registry)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result.ConfigPath != configPath {
		t.Errorf("ConfigPath: got %s, want %s", result.ConfigPath, configPath)
	}

	if result.UserStorePath != userStore {
		t.Errorf("UserStorePath: got %s, want %s", result.UserStorePath, userStore)
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
	userStore := filepath.Join(tmpDir, "Emulation")
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create all required directories for IsInitialized to return true
	for _, dir := range []string{"roms", "bios", "saves", "states", "screenshots"} {
		if err := os.MkdirAll(filepath.Join(userStore, dir), 0755); err != nil {
			t.Fatalf("Failed to create %s dir: %v", dir, err)
		}
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStore,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemTIC80: {Emulator: model.EmulatorTIC80},
		},
	}

	registry := emulators.NewRegistry()

	result, err := Get(cfg, configPath, registry)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.UserStoreInitialized {
		t.Error("UserStoreInitialized should be true when all directories exist")
	}
}

func TestGetSystemNames(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: tmpDir,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemTIC80: {Emulator: model.EmulatorTIC80},
		},
	}

	registry := emulators.NewRegistry()

	result, err := Get(cfg, tmpDir, registry)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(result.EnabledSystems) != 1 {
		t.Fatalf("Expected 1 system, got %d", len(result.EnabledSystems))
	}

	sys := result.EnabledSystems[0]
	if sys.ID != model.SystemTIC80 {
		t.Errorf("System ID: got %s, want %s", sys.ID, model.SystemTIC80)
	}
	if sys.Name != "TIC-80" {
		t.Errorf("System Name: got %s, want TIC-80", sys.Name)
	}
}

func TestGetMissingRequiredCount(t *testing.T) {
	tmpDir := t.TempDir()
	userStore := filepath.Join(tmpDir, "Emulation")

	if err := os.MkdirAll(filepath.Join(userStore, "bios", "psx"), 0755); err != nil {
		t.Fatalf("Failed to create bios dir: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStore,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemPSX:   {Emulator: model.EmulatorDuckStation},
			model.SystemTIC80: {Emulator: model.EmulatorTIC80},
		},
	}

	registry := emulators.NewRegistry()

	result, err := Get(cfg, tmpDir, registry)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// PSX has required BIOS, TIC-80 doesn't
	if result.MissingRequiredCount != 1 {
		t.Errorf("MissingRequiredCount: got %d, want 1 (PSX missing BIOS)", result.MissingRequiredCount)
	}
}

func TestGetWithManifest(t *testing.T) {
	tmpDir := t.TempDir()

	// Set XDG_STATE_HOME to our temp dir for manifest
	t.Setenv("XDG_STATE_HOME", tmpDir)

	// Create manifest
	manifestDir := filepath.Join(tmpDir, "kyaraben")
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		t.Fatalf("Failed to create manifest dir: %v", err)
	}

	manifest := &model.Manifest{
		LastApplied: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorTIC80: {
				ID:        model.EmulatorTIC80,
				Version:   "latest",
				StorePath: "/nix/store/abc123",
				Installed: time.Now(),
			},
		},
	}

	manifestPath := filepath.Join(manifestDir, "manifest.json")
	if err := manifest.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: tmpDir,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemTIC80: {Emulator: model.EmulatorTIC80},
		},
	}

	registry := emulators.NewRegistry()

	result, err := Get(cfg, tmpDir, registry)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(result.InstalledEmulators) != 1 {
		t.Fatalf("InstalledEmulators: got %d, want 1", len(result.InstalledEmulators))
	}

	emu := result.InstalledEmulators[0]
	if emu.ID != model.EmulatorTIC80 {
		t.Errorf("Emulator ID: got %s, want %s", emu.ID, model.EmulatorTIC80)
	}
	if emu.Name != "TIC-80" {
		t.Errorf("Emulator Name: got %s, want TIC-80", emu.Name)
	}
	if emu.Version != "latest" {
		t.Errorf("Emulator Version: got %s, want latest", emu.Version)
	}

	if result.LastApplied.IsZero() {
		t.Error("LastApplied should not be zero")
	}
}
