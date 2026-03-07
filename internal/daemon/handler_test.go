package daemon

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/versions"
)

func TestMain(m *testing.M) {
	if err := versions.Init(); err != nil {
		panic(err)
	}
	m.Run()
}

type testDaemonEnv struct {
	fs           vfs.FS
	cleanup      func()
	configPath   string
	stateDir     string
	manifestPath string
}

func newTestDaemonEnv(t *testing.T, cfg *model.KyarabenConfig) *testDaemonEnv {
	t.Helper()
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state/build": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}

	env := &testDaemonEnv{
		fs:           fs,
		cleanup:      cleanup,
		configPath:   "/config.toml",
		stateDir:     "/state",
		manifestPath: "/state/build/manifest.json",
	}

	if cfg != nil {
		if err := model.NewConfigStore(fs).Save(cfg, env.configPath); err != nil {
			cleanup()
			t.Fatalf("saving config: %v", err)
		}
	}

	return env
}

func (e *testDaemonEnv) newDaemon() *Daemon {
	return New(Deps{
		FS:           e.fs,
		Paths:        paths.DefaultPaths(),
		ConfigPath:   e.configPath,
		StateDir:     e.stateDir,
		ManifestPath: e.manifestPath,
		Registry:     registry.NewDefault(),
		Resolver:     fakeBaseDirResolver{root: "/"},
	})
}

func TestHandleUninstallPreview_EmptyManifest(t *testing.T) {
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: "~/Emulation",
		},
		Systems: make(map[model.SystemID][]model.EmulatorID),
	}
	env := newTestDaemonEnv(t, cfg)
	defer env.cleanup()
	d := env.newDaemon()

	events := d.Handle(Command{Type: CommandTypeUninstallPreview})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EventTypeResult {
		t.Fatalf("expected result event, got %s", event.Type)
	}

	resp, ok := event.Data.(UninstallPreviewResponse)
	if !ok {
		t.Fatalf("expected UninstallPreviewResponse, got %T", event.Data)
	}

	if resp.StateDir == "" {
		t.Error("expected stateDir to be set")
	}
	if len(resp.DesktopFiles) != 0 {
		t.Errorf("expected no desktop files, got %v", resp.DesktopFiles)
	}
	if len(resp.IconFiles) != 0 {
		t.Errorf("expected no icon files, got %v", resp.IconFiles)
	}
	if len(resp.ConfigFiles) != 0 {
		t.Errorf("expected no config files, got %v", resp.ConfigFiles)
	}
	if resp.Preserved.Collection != "~/Emulation" {
		t.Errorf("expected collection ~/Emulation, got %s", resp.Preserved.Collection)
	}
}

func TestHandleUninstallPreview_WithManifest(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state/build":  &vfst.Dir{Perm: 0755},
		"/test.desktop": "[Desktop Entry]",
		"/test.png":     "PNG",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	configPath := "/config.toml"
	stateDir := "/state"
	manifestPath := "/state/build/manifest.json"

	desktopFile := "/test.desktop"
	iconFile := "/test.png"

	manifest := &model.Manifest{
		Version:      1,
		LastApplied:  time.Now(),
		DesktopFiles: []string{desktopFile},
		IconFiles:    []string{iconFile},
	}
	if err := model.NewManifestStore(fs).Save(manifest, manifestPath); err != nil {
		t.Fatalf("saving manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: "~/Games",
		},
		Systems: make(map[model.SystemID][]model.EmulatorID),
	}
	if err := model.NewConfigStore(fs).Save(cfg, configPath); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	d := New(Deps{
		FS:           fs,
		Paths:        paths.DefaultPaths(),
		ConfigPath:   configPath,
		StateDir:     stateDir,
		ManifestPath: manifestPath,
		Registry:     registry.NewDefault(),
		Resolver:     fakeBaseDirResolver{root: "/"},
	})

	events := d.Handle(Command{Type: CommandTypeUninstallPreview})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EventTypeResult {
		t.Fatalf("expected result event, got %s", event.Type)
	}

	resp, ok := event.Data.(UninstallPreviewResponse)
	if !ok {
		t.Fatalf("expected UninstallPreviewResponse, got %T", event.Data)
	}

	if len(resp.DesktopFiles) != 1 || resp.DesktopFiles[0] != desktopFile {
		t.Errorf("expected desktop file %s, got %v", desktopFile, resp.DesktopFiles)
	}
	if len(resp.IconFiles) != 1 || resp.IconFiles[0] != iconFile {
		t.Errorf("expected icon file %s, got %v", iconFile, resp.IconFiles)
	}
	if resp.Preserved.Collection != "~/Games" {
		t.Errorf("expected collection ~/Games, got %s", resp.Preserved.Collection)
	}
}

func TestHandleGetConfig_ReturnsConfig(t *testing.T) {
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: "~/TestEmulation",
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
		Emulators: map[model.EmulatorID]model.EmulatorConf{
			model.EmulatorIDRetroArchMGBA: {Version: "0.10.0"},
		},
	}
	env := newTestDaemonEnv(t, cfg)
	defer env.cleanup()
	d := env.newDaemon()

	events := d.Handle(Command{Type: CommandTypeGetConfig})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EventTypeResult {
		t.Fatalf("expected result event, got %s", event.Type)
	}

	resp, ok := event.Data.(ConfigResponse)
	if !ok {
		t.Fatalf("expected ConfigResponse, got %T", event.Data)
	}

	if resp.Collection != "~/TestEmulation" {
		t.Errorf("expected collection ~/TestEmulation, got %s", resp.Collection)
	}
	if len(resp.Systems) != 1 {
		t.Errorf("expected 1 system, got %d", len(resp.Systems))
	}
	if emulators, ok := resp.Systems["gba"]; !ok || len(emulators) != 1 || emulators[0] != model.EmulatorIDRetroArchMGBA {
		t.Errorf("expected gba system with mgba emulator, got %v", resp.Systems)
	}
	if emuConf, ok := resp.Emulators["retroarch:mgba"]; !ok || emuConf.Version != "0.10.0" {
		t.Errorf("expected retroarch:mgba version 0.10.0, got %v", resp.Emulators)
	}
}

func TestHandleGetConfig_DefaultConfig(t *testing.T) {
	env := newTestDaemonEnv(t, nil)
	defer env.cleanup()
	d := env.newDaemon()

	events := d.Handle(Command{Type: CommandTypeGetConfig})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EventTypeResult {
		t.Fatalf("expected result event, got %s", event.Type)
	}

	resp, ok := event.Data.(ConfigResponse)
	if !ok {
		t.Fatalf("expected ConfigResponse, got %T", event.Data)
	}

	if resp.Collection != "~/Emulation" {
		t.Errorf("expected default collection ~/Emulation, got %s", resp.Collection)
	}
}

func TestHandleGetSystems_ReturnsSystems(t *testing.T) {
	env := newTestDaemonEnv(t, nil)
	defer env.cleanup()
	d := env.newDaemon()

	events := d.Handle(Command{Type: CommandTypeGetSystems})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EventTypeResult {
		t.Fatalf("expected result event, got %s", event.Type)
	}

	resp, ok := event.Data.(GetSystemsResponse)
	if !ok {
		t.Fatalf("expected GetSystemsResponse, got %T", event.Data)
	}

	if len(resp) == 0 {
		t.Error("expected at least one system")
	}

	var foundGBA bool
	for _, sys := range resp {
		if sys.ID == model.SystemIDGBA {
			foundGBA = true
			if sys.Name != "Game Boy Advance" {
				t.Errorf("expected GBA name 'Game Boy Advance', got %s", sys.Name)
			}
			if len(sys.Emulators) == 0 {
				t.Error("expected GBA to have at least one emulator")
			}
		}
	}
	if !foundGBA {
		t.Error("expected to find GBA system")
	}
}

func TestHandleGetSystems_PopulatesPackageNameAndCoreBytes(t *testing.T) {
	env := newTestDaemonEnv(t, nil)
	defer env.cleanup()
	d := env.newDaemon()

	events := d.Handle(Command{Type: CommandTypeGetSystems})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	resp, ok := events[0].Data.(GetSystemsResponse)
	if !ok {
		t.Fatalf("expected GetSystemsResponse, got %T", events[0].Data)
	}

	var foundRetroArchEmulator bool
	for _, sys := range resp {
		for _, emu := range sys.Emulators {
			if emu.PackageName == "" {
				t.Errorf("emulator %s has empty PackageName", emu.ID)
			}

			if emu.PackageName == "retroarch" {
				foundRetroArchEmulator = true
				if emu.CoreBytes <= 0 {
					t.Errorf("RetroArch emulator %s should have CoreBytes > 0, got %d", emu.ID, emu.CoreBytes)
				}
				if emu.DownloadBytes <= 0 {
					t.Errorf("RetroArch emulator %s should have DownloadBytes > 0 (retroarch binary), got %d", emu.ID, emu.DownloadBytes)
				}
			}
		}
	}

	if !foundRetroArchEmulator {
		t.Error("expected to find at least one RetroArch emulator")
	}
}

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

func TestHandlePreflight_ReturnsPreflightResponse(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state/build": &vfst.Dir{Perm: 0755},
		"/Emulation":   &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	configPath := "/config.toml"
	stateDir := "/state"
	manifestPath := "/state/build/manifest.json"

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: "/Emulation",
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDRetroArchMGBA},
		},
	}
	if err := model.NewConfigStore(fs).Save(cfg, configPath); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	reg := registry.NewDefault()
	resolver := fakeBaseDirResolver{root: "/"}
	configWriter := emulators.NewConfigWriter(fs, resolver)
	d := New(Deps{
		FS:           fs,
		Paths:        paths.DefaultPaths(),
		ConfigPath:   configPath,
		StateDir:     stateDir,
		ManifestPath: manifestPath,
		Registry:     reg,
		ConfigWriter: configWriter,
		Resolver:     resolver,
	})

	events := d.Handle(Command{Type: CommandTypePreflight})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EventTypeResult {
		t.Fatalf("expected result event, got %s", event.Type)
	}

	resp, ok := event.Data.(PreflightResponse)
	if !ok {
		t.Fatalf("expected PreflightResponse, got %T", event.Data)
	}

	if resp.Diffs == nil {
		t.Error("expected diffs to be non-nil")
	}
}

func TestHandlePreflight_EmptyConfig(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state/build": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: "~/Emulation",
		},
		Systems: make(map[model.SystemID][]model.EmulatorID),
	}
	configPath := "/config.toml"
	if err := model.NewConfigStore(fs).Save(cfg, configPath); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	resolver := fakeBaseDirResolver{root: "/"}
	configWriter := emulators.NewConfigWriter(fs, resolver)
	d := New(Deps{
		FS:           fs,
		Paths:        paths.DefaultPaths(),
		ConfigPath:   configPath,
		StateDir:     "/state",
		ManifestPath: "/state/build/manifest.json",
		Registry:     registry.NewDefault(),
		ConfigWriter: configWriter,
		Resolver:     resolver,
	})

	events := d.Handle(Command{Type: CommandTypePreflight})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	resp, ok := events[0].Data.(PreflightResponse)
	if !ok {
		t.Fatalf("expected PreflightResponse, got %T", events[0].Data)
	}

	if len(resp.Diffs) != 0 {
		t.Errorf("expected 0 diffs for empty config, got %d", len(resp.Diffs))
	}
}

func TestHandlePreflight_DetectsVersionUpgrade(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state/build":                      &vfst.Dir{Perm: 0755},
		"/Emulation":                        &vfst.Dir{Perm: 0755},
		"/Emulation/roms":                   &vfst.Dir{Perm: 0755},
		"/Emulation/bios":                   &vfst.Dir{Perm: 0755},
		"/Emulation/saves":                  &vfst.Dir{Perm: 0755},
		"/Emulation/states":                 &vfst.Dir{Perm: 0755},
		"/.config/duckstation/settings.ini": "[Main]\nSettingsVersion = 3\n[AutoUpdater]\nCheckAtStartup = true\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	configPath := "/config.toml"
	stateDir := "/state"
	manifestPath := "/state/build/manifest.json"

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: "/Emulation",
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDPSX: {model.EmulatorIDDuckStation},
		},
	}
	if err := model.NewConfigStore(fs).Save(cfg, configPath); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	manifest := model.NewManifest()
	manifest.InstalledEmulators[model.EmulatorIDDuckStation] = model.InstalledEmulator{
		ID:          model.EmulatorIDDuckStation,
		Version:     "0.1.0",
		PackagePath: "/packages/duckstation",
		Installed:   time.Now(),
	}
	if err := manifest.AddManagedConfig(model.ManagedConfig{
		EmulatorIDs: []model.EmulatorID{model.EmulatorIDDuckStation},
		Target: model.ConfigTarget{
			RelPath: "duckstation/settings.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		WrittenEntries: map[string]string{
			"AutoUpdater.CheckAtStartup": "true",
		},
		LastModified:   time.Now(),
		ManagedRegions: model.ManagedRegions{model.FileRegion{}},
	}); err != nil {
		t.Fatalf("adding managed config: %v", err)
	}
	if err := model.NewManifestStore(fs).Save(manifest, manifestPath); err != nil {
		t.Fatalf("saving manifest: %v", err)
	}

	reg := registry.NewDefault()
	resolver := fakeBaseDirResolver{root: "/"}
	configWriter := emulators.NewConfigWriter(fs, resolver)
	d := New(Deps{
		FS:           fs,
		Paths:        paths.DefaultPaths(),
		ConfigPath:   configPath,
		StateDir:     stateDir,
		ManifestPath: manifestPath,
		Registry:     reg,
		ConfigWriter: configWriter,
		Resolver:     resolver,
	})

	events := d.Handle(Command{Type: CommandTypePreflight})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	resp, ok := events[0].Data.(PreflightResponse)
	if !ok {
		t.Fatalf("expected PreflightResponse, got %T", events[0].Data)
	}

	var duckstationDiff *ConfigFileDiff
	for i := range resp.Diffs {
		if resp.Diffs[i].Path == "/.config/duckstation/settings.ini" {
			duckstationDiff = &resp.Diffs[i]
			break
		}
	}

	if duckstationDiff == nil {
		t.Fatalf("expected diff for duckstation settings.ini, got diffs: %+v", resp.Diffs)
	}

	if !duckstationDiff.KyarabenChanged {
		t.Error("expected KyarabenChanged to be true")
	}

	if len(duckstationDiff.KyarabenUpdates) == 0 {
		t.Fatal("expected KyarabenUpdates to be non-empty")
	}

	found := false
	for _, update := range duckstationDiff.KyarabenUpdates {
		if update.Key == "CheckAtStartup" {
			found = true
			if update.OldValue != "true" {
				t.Errorf("expected old value 'true', got %q", update.OldValue)
			}
			if update.NewValue != "false" {
				t.Errorf("expected new value 'false', got %q", update.NewValue)
			}
		}
	}
	if !found {
		t.Errorf("expected KyarabenUpdate for CheckAtStartup, got: %+v", duckstationDiff.KyarabenUpdates)
	}
}

func TestRetroArchCoreName(t *testing.T) {
	tests := []struct {
		id       model.EmulatorID
		expected string
	}{
		{model.EmulatorIDRetroArchBsnes, "bsnes"},
		{model.EmulatorIDRetroArchMesen, "mesen"},
		{model.EmulatorIDRetroArchGenesisPlusGX, "genesis_plus_gx"},
		{model.EmulatorIDRetroArchMupen64Plus, "mupen64plus_next"},
		{model.EmulatorIDRetroArchBeetleSaturn, "mednafen_saturn"},
		{model.EmulatorIDRetroArchMGBA, "mgba"},
		{model.EmulatorIDRetroArchMelonDS, "melondsds"},
		{model.EmulatorIDDolphin, ""},
	}

	for _, tt := range tests {
		got := retroArchCoreName(tt.id)
		if got != tt.expected {
			t.Errorf("retroArchCoreName(%s) = %q, want %q", tt.id, got, tt.expected)
		}
	}
}

func TestHandleInstallStatus_EmptyManifest(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state/build": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	configPath := "/config.toml"
	stateDir := "/state"
	manifestPath := "/state/build/manifest.json"

	manifest := &model.Manifest{
		Version:     1,
		LastApplied: time.Now(),
		KyarabenInstall: &model.KyarabenInstall{
			CLIPath:     "/nonexistent/kyaraben",
			DesktopPath: "/nonexistent/kyaraben.desktop",
		},
	}
	if err := model.NewManifestStore(fs).Save(manifest, manifestPath); err != nil {
		t.Fatalf("saving manifest: %v", err)
	}

	d := New(Deps{
		FS:           fs,
		Paths:        paths.DefaultPaths(),
		ConfigPath:   configPath,
		StateDir:     stateDir,
		ManifestPath: manifestPath,
		Registry:     registry.NewDefault(),
		Resolver:     fakeBaseDirResolver{root: "/"},
	})

	events := d.Handle(Command{Type: CommandTypeInstallStatus})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EventTypeResult {
		t.Fatalf("expected result event, got %s", event.Type)
	}

	resp, ok := event.Data.(InstallStatusResponse)
	if !ok {
		t.Fatalf("expected InstallStatusResponse, got %T", event.Data)
	}

	if resp.Installed {
		t.Error("expected installed to be false when files don't exist")
	}
	if resp.CLIPath != "" {
		t.Errorf("expected empty CLI path for nonexistent file, got %s", resp.CLIPath)
	}
}

func TestHandleInstallStatus_WithManifest(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state/build":      &vfst.Dir{Perm: 0755},
		"/kyaraben":         "#!/bin/sh",
		"/kyaraben.desktop": "[Desktop Entry]",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	configPath := "/config.toml"
	stateDir := "/state"
	manifestPath := "/state/build/manifest.json"

	cliPath := "/kyaraben"
	desktopPath := "/kyaraben.desktop"

	manifest := &model.Manifest{
		Version:     1,
		LastApplied: time.Now(),
		KyarabenInstall: &model.KyarabenInstall{
			CLIPath:     cliPath,
			DesktopPath: desktopPath,
		},
	}
	if err := model.NewManifestStore(fs).Save(manifest, manifestPath); err != nil {
		t.Fatalf("saving manifest: %v", err)
	}

	d := New(Deps{
		FS:           fs,
		Paths:        paths.DefaultPaths(),
		ConfigPath:   configPath,
		StateDir:     stateDir,
		ManifestPath: manifestPath,
		Registry:     registry.NewDefault(),
		Resolver:     fakeBaseDirResolver{root: "/"},
	})

	events := d.Handle(Command{Type: CommandTypeInstallStatus})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EventTypeResult {
		t.Fatalf("expected result event, got %s", event.Type)
	}

	resp, ok := event.Data.(InstallStatusResponse)
	if !ok {
		t.Fatalf("expected InstallStatusResponse, got %T", event.Data)
	}

	if !resp.Installed {
		t.Error("expected installed to be true")
	}
	if resp.CLIPath != cliPath {
		t.Errorf("expected CLI path %s, got %s", cliPath, resp.CLIPath)
	}
	if resp.DesktopPath != desktopPath {
		t.Errorf("expected desktop path %s, got %s", desktopPath, resp.DesktopPath)
	}
}

func TestHandleSyncStatus_Disabled(t *testing.T) {
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: "~/Emulation",
		},
		Systems: make(map[model.SystemID][]model.EmulatorID),
		Sync:    model.DefaultSyncConfig(),
	}
	env := newTestDaemonEnv(t, cfg)
	defer env.cleanup()
	d := env.newDaemon()

	events := d.Handle(Command{Type: CommandTypeSyncStatus})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EventTypeResult {
		t.Fatalf("expected result event, got %s", event.Type)
	}

	resp, ok := event.Data.(SyncStatusResponse)
	if !ok {
		t.Fatalf("expected SyncStatusResponse, got %T", event.Data)
	}

	if resp.Enabled {
		t.Error("expected enabled to be false")
	}
	if resp.Installed {
		t.Error("expected installed to be false when no manifest")
	}
}

func TestHandleSyncStatus_DisabledWithSyncthingInstalled(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/state/build":                  &vfst.Dir{Perm: 0755},
		"/state/packages/syncthing/bin": "fake-syncthing",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	configPath := "/config.toml"
	stateDir := "/state"
	manifestPath := "/state/build/manifest.json"

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: "~/Emulation",
		},
		Systems: make(map[model.SystemID][]model.EmulatorID),
		Sync:    model.DefaultSyncConfig(),
	}
	if err := model.NewConfigStore(fs).Save(cfg, configPath); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	manifest := &model.Manifest{
		Version:     1,
		LastApplied: time.Now(),
		SyncthingInstall: &model.SyncthingInstall{
			BinaryPath: "/state/packages/syncthing/bin",
		},
	}
	if err := model.NewManifestStore(fs).Save(manifest, manifestPath); err != nil {
		t.Fatalf("saving manifest: %v", err)
	}

	d := New(Deps{
		FS:           fs,
		Paths:        paths.DefaultPaths(),
		ConfigPath:   configPath,
		StateDir:     stateDir,
		ManifestPath: manifestPath,
		Registry:     registry.NewDefault(),
		Resolver:     fakeBaseDirResolver{root: "/"},
	})

	events := d.Handle(Command{Type: CommandTypeSyncStatus})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	resp, ok := events[0].Data.(SyncStatusResponse)
	if !ok {
		t.Fatalf("expected SyncStatusResponse, got %T", events[0].Data)
	}

	if resp.Enabled {
		t.Error("expected enabled to be false")
	}
	if !resp.Installed {
		t.Error("expected installed to be true when syncthing binary exists")
	}
}

func TestHandleSetConfig_MergesEmulatorSettings(t *testing.T) {
	presetDefault := model.PresetClean
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			Collection: "~/Emulation",
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDSNES: {model.EmulatorIDRetroArchBsnes},
			model.SystemIDGBA:  {model.EmulatorIDRetroArchMGBA},
		},
		Emulators: map[model.EmulatorID]model.EmulatorConf{
			model.EmulatorIDRetroArchBsnes: {Preset: &presetDefault},
			model.EmulatorIDRetroArchMGBA:  {Preset: &presetDefault, Version: "0.10.0"},
		},
		Frontends: map[model.FrontendID]model.FrontendConfig{},
	}
	env := newTestDaemonEnv(t, cfg)
	defer env.cleanup()
	d := env.newDaemon()

	presetRetro := model.PresetRetro
	setCmd := SetConfigCommand{
		Type: CommandTypeSetConfig,
		Data: SetConfigRequest{
			Collection: "~/Emulation",
			Systems: map[string][]string{
				"snes": {"retroarch:bsnes"},
				"gba":  {"retroarch:mgba"},
			},
			Emulators: map[string]EmulatorConfRequest{
				"retroarch:mgba": {Preset: &presetRetro},
			},
		},
	}

	events := d.HandleSetConfig(setCmd, nil)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	getEvents := d.Handle(Command{Type: CommandTypeGetConfig})
	resp, ok := getEvents[0].Data.(ConfigResponse)
	if !ok {
		t.Fatalf("expected ConfigResponse, got %T", getEvents[0].Data)
	}

	bsnesConf := resp.Emulators["retroarch:bsnes"]
	if bsnesConf.Preset == nil || *bsnesConf.Preset != model.PresetClean {
		t.Errorf("expected bsnes preset to be preserved as clean, got %v", bsnesConf.Preset)
	}

	mgbaConf := resp.Emulators["retroarch:mgba"]
	if mgbaConf.Preset == nil || *mgbaConf.Preset != model.PresetRetro {
		t.Errorf("expected mgba preset to be updated to retro, got %v", mgbaConf.Preset)
	}
	if mgbaConf.Version != "0.10.0" {
		t.Errorf("expected mgba version to be preserved as 0.10.0, got %s", mgbaConf.Version)
	}
}
