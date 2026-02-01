package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/versions"
)

func TestMain(m *testing.M) {
	if err := versions.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func newTestDaemon(t *testing.T, cfg *model.KyarabenConfig) *Daemon {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	stateDir := filepath.Join(tmpDir, "state")
	manifestPath := filepath.Join(tmpDir, "state", "build", "manifest.json")

	if cfg != nil {
		if err := model.SaveConfig(cfg, configPath); err != nil {
			t.Fatalf("saving config: %v", err)
		}
	}

	return New(configPath, stateDir, manifestPath, registry.NewDefault(), nil, nil, nil, nil)
}

func TestHandleUninstallPreview_EmptyManifest(t *testing.T) {
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: "~/Emulation",
		},
		Systems: make(map[model.SystemID][]model.EmulatorID),
	}
	d := newTestDaemon(t, cfg)

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
	if resp.Preserved.UserStore != "~/Emulation" {
		t.Errorf("expected userStore ~/Emulation, got %s", resp.Preserved.UserStore)
	}
}

func TestHandleUninstallPreview_WithManifest(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	stateDir := filepath.Join(tmpDir, "state")
	manifestPath := filepath.Join(tmpDir, "state", "build", "manifest.json")

	desktopFile := filepath.Join(tmpDir, "test.desktop")
	iconFile := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(desktopFile, []byte("[Desktop Entry]"), 0644); err != nil {
		t.Fatalf("creating desktop file: %v", err)
	}
	if err := os.WriteFile(iconFile, []byte("PNG"), 0644); err != nil {
		t.Fatalf("creating icon file: %v", err)
	}

	manifest := &model.Manifest{
		Version:      1,
		LastApplied:  time.Now(),
		DesktopFiles: []string{desktopFile},
		IconFiles:    []string{iconFile},
	}
	if err := manifest.Save(manifestPath); err != nil {
		t.Fatalf("saving manifest: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: "~/Games",
		},
		Systems: make(map[model.SystemID][]model.EmulatorID),
	}
	if err := model.SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	d := New(configPath, stateDir, manifestPath, registry.NewDefault(), nil, nil, nil, nil)

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
	if resp.Preserved.UserStore != "~/Games" {
		t.Errorf("expected userStore ~/Games, got %s", resp.Preserved.UserStore)
	}
}

func TestHandleGetConfig_ReturnsConfig(t *testing.T) {
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: "~/TestEmulation",
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
		Emulators: map[model.EmulatorID]model.EmulatorConf{
			model.EmulatorIDMGBA: {Version: "0.10.0"},
		},
	}
	d := newTestDaemon(t, cfg)

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

	if resp.UserStore != "~/TestEmulation" {
		t.Errorf("expected userStore ~/TestEmulation, got %s", resp.UserStore)
	}
	if len(resp.Systems) != 1 {
		t.Errorf("expected 1 system, got %d", len(resp.Systems))
	}
	if emulators, ok := resp.Systems["gba"]; !ok || len(emulators) != 1 || emulators[0] != model.EmulatorIDMGBA {
		t.Errorf("expected gba system with mgba emulator, got %v", resp.Systems)
	}
	if emuConf, ok := resp.Emulators["mgba"]; !ok || emuConf.Version != "0.10.0" {
		t.Errorf("expected mgba version 0.10.0, got %v", resp.Emulators)
	}
}

func TestHandleGetConfig_DefaultConfig(t *testing.T) {
	d := newTestDaemon(t, nil)

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

	if resp.UserStore != "~/Emulation" {
		t.Errorf("expected default userStore ~/Emulation, got %s", resp.UserStore)
	}
}

func TestHandleGetSystems_ReturnsSystems(t *testing.T) {
	d := newTestDaemon(t, nil)

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

func TestHandleInstallStatus_EmptyManifest(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	stateDir := filepath.Join(tmpDir, "state")
	manifestPath := filepath.Join(tmpDir, "state", "build", "manifest.json")

	manifest := &model.Manifest{
		Version:     1,
		LastApplied: time.Now(),
		KyarabenInstall: &model.KyarabenInstall{
			CLIPath:     filepath.Join(tmpDir, "nonexistent", "kyaraben"),
			DesktopPath: filepath.Join(tmpDir, "nonexistent", "kyaraben.desktop"),
		},
	}
	if err := manifest.Save(manifestPath); err != nil {
		t.Fatalf("saving manifest: %v", err)
	}

	d := New(configPath, stateDir, manifestPath, registry.NewDefault(), nil, nil, nil, nil)

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
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	stateDir := filepath.Join(tmpDir, "state")
	manifestPath := filepath.Join(tmpDir, "state", "build", "manifest.json")

	cliPath := filepath.Join(tmpDir, "kyaraben")
	desktopPath := filepath.Join(tmpDir, "kyaraben.desktop")
	if err := os.WriteFile(cliPath, []byte("#!/bin/sh"), 0755); err != nil {
		t.Fatalf("creating cli file: %v", err)
	}
	if err := os.WriteFile(desktopPath, []byte("[Desktop Entry]"), 0644); err != nil {
		t.Fatalf("creating desktop file: %v", err)
	}

	manifest := &model.Manifest{
		Version:     1,
		LastApplied: time.Now(),
		KyarabenInstall: &model.KyarabenInstall{
			CLIPath:     cliPath,
			DesktopPath: desktopPath,
		},
	}
	if err := manifest.Save(manifestPath); err != nil {
		t.Fatalf("saving manifest: %v", err)
	}

	d := New(configPath, stateDir, manifestPath, registry.NewDefault(), nil, nil, nil, nil)

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
