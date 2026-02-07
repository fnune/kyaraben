package status

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
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
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA:  {model.EmulatorIDMGBA},
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
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
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
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
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
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
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDPSX: {model.EmulatorIDDuckStation},
			model.SystemIDGBA: {model.EmulatorIDMGBA},
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
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
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

func TestGetWithVersionPinning(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &model.Manifest{
		LastApplied: time.Now(),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDDuckStation: {
				ID:        model.EmulatorIDDuckStation,
				Version:   "v0.1-10655",
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
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDPSX: {model.EmulatorIDDuckStation},
		},
		Emulators: map[model.EmulatorID]model.EmulatorConf{
			model.EmulatorIDDuckStation: {Version: "v0.1-10655"},
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
	if emu.PinnedVersion != "v0.1-10655" {
		t.Errorf("PinnedVersion: got %s, want v0.1-10655", emu.PinnedVersion)
	}
}

func TestGetManagedConfigsIncludesKeys(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &model.Manifest{
		LastApplied: time.Now(),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDMGBA: {
				ID:        model.EmulatorIDMGBA,
				Version:   "latest",
				StorePath: "/nix/store/abc123",
				Installed: time.Now(),
			},
		},
		ManagedConfigs: []model.ManagedConfig{
			{
				EmulatorIDs: []model.EmulatorID{model.EmulatorIDMGBA},
				Target: model.ConfigTarget{
					RelPath: "mgba/config.ini",
					BaseDir: model.ConfigBaseDirUserConfig,
				},
				ManagedKeys: []model.ManagedKey{
					{Path: []string{"ports.qt", "savegamePath"}, Value: "/test/saves"},
					{Path: []string{"ports.qt", "bios"}, Value: "/test/bios"},
				},
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, "manifest.json")
	if err := manifest.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global:  model.GlobalConfig{UserStore: tmpDir},
		Systems: map[model.SystemID][]model.EmulatorID{model.SystemIDGBA: {model.EmulatorIDMGBA}},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, tmpDir)

	result, err := Get(context.Background(), cfg, tmpDir, reg, userStore, manifestPath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(result.InstalledEmulators) != 1 {
		t.Fatalf("InstalledEmulators: got %d, want 1", len(result.InstalledEmulators))
	}

	emu := result.InstalledEmulators[0]
	if len(emu.ManagedConfigs) != 1 {
		t.Fatalf("ManagedConfigs: got %d, want 1", len(emu.ManagedConfigs))
	}

	cfg0 := emu.ManagedConfigs[0]
	if len(cfg0.Keys) != 2 {
		t.Fatalf("Keys: got %d, want 2", len(cfg0.Keys))
	}

	if cfg0.Keys[0].Key != "savegamePath" {
		t.Errorf("Key[0]: got %s, want savegamePath", cfg0.Keys[0].Key)
	}
	if cfg0.Keys[0].Value != "/test/saves" {
		t.Errorf("Value[0]: got %s, want /test/saves", cfg0.Keys[0].Value)
	}
}

func TestGetRetroArchCoreIncludesSharedConfig(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &model.Manifest{
		LastApplied: time.Now(),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDRetroArchBsnes: {
				ID:        model.EmulatorIDRetroArchBsnes,
				Version:   "latest",
				StorePath: "/nix/store/abc123",
				Installed: time.Now(),
			},
			model.EmulatorIDRetroArchMesen: {
				ID:        model.EmulatorIDRetroArchMesen,
				Version:   "latest",
				StorePath: "/nix/store/def456",
				Installed: time.Now(),
			},
		},
		ManagedConfigs: []model.ManagedConfig{
			{
				EmulatorIDs: []model.EmulatorID{model.EmulatorIDRetroArchBsnes, model.EmulatorIDRetroArchMesen},
				Target:      retroarch.MainConfigTarget,
				ManagedKeys: []model.ManagedKey{
					{Path: []string{"system_directory"}, Value: "/test/bios"},
					{Path: []string{"sort_savefiles_enable"}, Value: "false"},
				},
			},
			{
				EmulatorIDs: []model.EmulatorID{model.EmulatorIDRetroArchBsnes},
				Target: model.ConfigTarget{
					RelPath: "retroarch/config/bsnes_libretro/bsnes_libretro.cfg",
					BaseDir: model.ConfigBaseDirUserConfig,
				},
				ManagedKeys: []model.ManagedKey{
					{Path: []string{"savefile_directory"}, Value: "/test/saves/snes"},
				},
			},
			{
				EmulatorIDs: []model.EmulatorID{model.EmulatorIDRetroArchMesen},
				Target: model.ConfigTarget{
					RelPath: "retroarch/config/mesen_libretro/mesen_libretro.cfg",
					BaseDir: model.ConfigBaseDirUserConfig,
				},
				ManagedKeys: []model.ManagedKey{
					{Path: []string{"savefile_directory"}, Value: "/test/saves/nes"},
				},
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, "manifest.json")
	if err := manifest.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{UserStore: tmpDir},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
			model.SystemIDNES:  {model.EmulatorIDRetroArchMesen},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, tmpDir)

	result, err := Get(context.Background(), cfg, tmpDir, reg, userStore, manifestPath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(result.InstalledEmulators) != 2 {
		t.Fatalf("InstalledEmulators: got %d, want 2", len(result.InstalledEmulators))
	}

	var bsnes, mesen *EmulatorInfo
	for i := range result.InstalledEmulators {
		switch result.InstalledEmulators[i].ID {
		case model.EmulatorIDRetroArchBsnes:
			bsnes = &result.InstalledEmulators[i]
		case model.EmulatorIDRetroArchMesen:
			mesen = &result.InstalledEmulators[i]
		}
	}

	if bsnes == nil || mesen == nil {
		t.Fatal("Missing expected emulators")
	}

	if len(bsnes.ManagedConfigs) != 2 {
		t.Errorf("bsnes ManagedConfigs: got %d, want 2 (core override + shared)", len(bsnes.ManagedConfigs))
	}

	if len(mesen.ManagedConfigs) != 2 {
		t.Errorf("mesen ManagedConfigs: got %d, want 2 (core override + shared)", len(mesen.ManagedConfigs))
	}

	hasSharedConfig := false
	for _, cfg := range bsnes.ManagedConfigs {
		for _, key := range cfg.Keys {
			if key.Key == "sort_savefiles_enable" {
				hasSharedConfig = true
				break
			}
		}
	}
	if !hasSharedConfig {
		t.Error("bsnes should include shared RetroArch config with sort_savefiles_enable")
	}
}
