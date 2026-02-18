package status

import (
	"context"
	"testing"
	"time"

	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
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
	m.Run()
}

func mustNewUserStore(t *testing.T, fs vfs.FS, path string) *store.UserStore {
	t.Helper()
	s, err := store.NewUserStore(fs, paths.DefaultPaths(), path)
	if err != nil {
		t.Fatalf("NewUserStore(%q) failed: %v", path, err)
	}
	return s
}

func TestGet(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation": &vfst.Dir{Perm: 0755},
		"/state":     &vfst.Dir{Perm: 0755},
	})

	userStorePath := "/Emulation"
	configPath := "/config.toml"
	manifestPath := "/state/manifest.json"

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA:  {model.EmulatorIDRetroArchMGBA},
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, userStorePath)

	getter := NewGetter(fs, paths.DefaultPaths())
	result, err := getter.Get(context.Background(), cfg, configPath, reg, userStore, manifestPath)
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
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation/roms":        &vfst.Dir{Perm: 0755},
		"/Emulation/bios":        &vfst.Dir{Perm: 0755},
		"/Emulation/saves":       &vfst.Dir{Perm: 0755},
		"/Emulation/states":      &vfst.Dir{Perm: 0755},
		"/Emulation/screenshots": &vfst.Dir{Perm: 0755},
		"/state":                 &vfst.Dir{Perm: 0755},
	})

	userStorePath := "/Emulation"
	configPath := "/config.toml"
	manifestPath := "/state/manifest.json"

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, userStorePath)

	getter := NewGetter(fs, paths.DefaultPaths())
	result, err := getter.Get(context.Background(), cfg, configPath, reg, userStore, manifestPath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.UserStoreInitialized {
		t.Error("UserStoreInitialized should be true when all directories exist")
	}
}

func TestGetSystemNames(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation": &vfst.Dir{Perm: 0755},
		"/state":     &vfst.Dir{Perm: 0755},
	})

	manifestPath := "/state/manifest.json"

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: "/Emulation",
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, "/Emulation")

	getter := NewGetter(fs, paths.DefaultPaths())
	result, err := getter.Get(context.Background(), cfg, "/Emulation", reg, userStore, manifestPath)
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
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation/bios/psx": &vfst.Dir{Perm: 0755},
		"/state":              &vfst.Dir{Perm: 0755},
	})

	userStorePath := "/Emulation"
	manifestPath := "/state/manifest.json"

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDPSX: {model.EmulatorIDDuckStation},
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, userStorePath)

	getter := NewGetter(fs, paths.DefaultPaths())
	result, err := getter.Get(context.Background(), cfg, "/config", reg, userStore, manifestPath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result.MissingRequiredCount != 1 {
		t.Errorf("MissingRequiredCount: got %d, want 1 (PSX missing BIOS)", result.MissingRequiredCount)
	}
}

func TestGetWithManifest(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation": &vfst.Dir{Perm: 0755},
		"/state":     &vfst.Dir{Perm: 0755},
	})

	manifest := &model.Manifest{
		LastApplied: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDRetroArchMGBA: {
				ID:          model.EmulatorIDRetroArchMGBA,
				Version:     "latest",
				PackagePath: "/test/packages",
				Installed:   time.Now(),
			},
		},
	}

	manifestPath := "/state/manifest.json"
	if err := model.NewManifestStore(fs).Save(manifest, manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: "/Emulation",
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, "/Emulation")

	getter := NewGetter(fs, paths.DefaultPaths())
	result, err := getter.Get(context.Background(), cfg, "/Emulation", reg, userStore, manifestPath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(result.InstalledEmulators) != 1 {
		t.Fatalf("InstalledEmulators: got %d, want 1", len(result.InstalledEmulators))
	}

	emu := result.InstalledEmulators[0]
	if emu.ID != model.EmulatorIDRetroArchMGBA {
		t.Errorf("Emulator ID: got %s, want %s", emu.ID, model.EmulatorIDRetroArchMGBA)
	}
	if emu.Name != "mGBA (RetroArch)" {
		t.Errorf("Emulator Name: got %s, want mGBA (RetroArch)", emu.Name)
	}
	if emu.Version != "latest" {
		t.Errorf("Emulator Version: got %s, want latest", emu.Version)
	}

	if result.LastApplied.IsZero() {
		t.Error("LastApplied should not be zero")
	}
}

func TestGetWithVersionPinning(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation": &vfst.Dir{Perm: 0755},
		"/state":     &vfst.Dir{Perm: 0755},
	})

	manifest := &model.Manifest{
		LastApplied: time.Now(),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDDuckStation: {
				ID:          model.EmulatorIDDuckStation,
				Version:     "v0.1-10655",
				PackagePath: "/test/packages",
				Installed:   time.Now(),
			},
		},
	}

	manifestPath := "/state/manifest.json"
	if err := model.NewManifestStore(fs).Save(manifest, manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: "/Emulation",
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDPSX: {model.EmulatorIDDuckStation},
		},
		Emulators: map[model.EmulatorID]model.EmulatorConf{
			model.EmulatorIDDuckStation: {Version: "v0.1-10655"},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, "/Emulation")

	getter := NewGetter(fs, paths.DefaultPaths())
	result, err := getter.Get(context.Background(), cfg, "/Emulation", reg, userStore, manifestPath)
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
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation": &vfst.Dir{Perm: 0755},
		"/state":     &vfst.Dir{Perm: 0755},
	})

	manifest := &model.Manifest{
		LastApplied: time.Now(),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDRetroArchMGBA: {
				ID:          model.EmulatorIDRetroArchMGBA,
				Version:     "latest",
				PackagePath: "/test/packages",
				Installed:   time.Now(),
			},
		},
		ManagedConfigs: []model.ManagedConfig{
			{
				EmulatorIDs: []model.EmulatorID{model.EmulatorIDRetroArchMGBA},
				Target: model.ConfigTarget{
					RelPath: "mgba/config.ini",
					BaseDir: model.ConfigBaseDirUserConfig,
				},
			},
		},
	}

	manifestPath := "/state/manifest.json"
	if err := model.NewManifestStore(fs).Save(manifest, manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global:  model.GlobalConfig{UserStore: "/Emulation"},
		Systems: map[model.SystemID][]model.EmulatorID{model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA}},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, "/Emulation")

	getter := NewGetter(fs, paths.DefaultPaths())
	result, err := getter.Get(context.Background(), cfg, "/Emulation", reg, userStore, manifestPath)
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
	if cfg0.Path == "" {
		t.Error("expected non-empty path for managed config")
	}
}

func TestGetRetroArchCoreIncludesSharedConfig(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation": &vfst.Dir{Perm: 0755},
		"/state":     &vfst.Dir{Perm: 0755},
	})

	manifest := &model.Manifest{
		LastApplied: time.Now(),
		InstalledEmulators: map[model.EmulatorID]model.InstalledEmulator{
			model.EmulatorIDRetroArchBsnes: {
				ID:          model.EmulatorIDRetroArchBsnes,
				Version:     "latest",
				PackagePath: "/test/packages",
				Installed:   time.Now(),
			},
			model.EmulatorIDRetroArchMesen: {
				ID:          model.EmulatorIDRetroArchMesen,
				Version:     "latest",
				PackagePath: "/test/packages",
				Installed:   time.Now(),
			},
		},
		ManagedConfigs: []model.ManagedConfig{
			{
				EmulatorIDs: []model.EmulatorID{model.EmulatorIDRetroArchBsnes, model.EmulatorIDRetroArchMesen},
				Target:      retroarch.MainConfigTarget,
			},
			{
				EmulatorIDs: []model.EmulatorID{model.EmulatorIDRetroArchBsnes},
				Target: model.ConfigTarget{
					RelPath: "retroarch/config/bsnes_libretro/bsnes_libretro.cfg",
					BaseDir: model.ConfigBaseDirUserConfig,
				},
			},
			{
				EmulatorIDs: []model.EmulatorID{model.EmulatorIDRetroArchMesen},
				Target: model.ConfigTarget{
					RelPath: "retroarch/config/mesen_libretro/mesen_libretro.cfg",
					BaseDir: model.ConfigBaseDirUserConfig,
				},
			},
		},
	}

	manifestPath := "/state/manifest.json"
	if err := model.NewManifestStore(fs).Save(manifest, manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{UserStore: "/Emulation"},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
			model.SystemIDNES:  {model.EmulatorIDRetroArchMesen},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, "/Emulation")

	getter := NewGetter(fs, paths.DefaultPaths())
	result, err := getter.Get(context.Background(), cfg, "/Emulation", reg, userStore, manifestPath)
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
		if cfg.Path != "" {
			hasSharedConfig = true
			break
		}
	}
	if !hasSharedConfig {
		t.Error("bsnes should include shared RetroArch config")
	}
}
