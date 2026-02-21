package model

import (
	"testing"
	"time"

	"github.com/twpayne/go-vfs/v5/vfst"
)

func TestNewManifest(t *testing.T) {
	t.Parallel()

	m := NewManifest()

	if m.Version != 1 {
		t.Errorf("Version = %d, want 1", m.Version)
	}
	if m.InstalledEmulators == nil {
		t.Error("InstalledEmulators is nil, want empty map")
	}
	if len(m.InstalledEmulators) != 0 {
		t.Errorf("InstalledEmulators has %d entries, want 0", len(m.InstalledEmulators))
	}
	if m.ManagedConfigs == nil {
		t.Error("ManagedConfigs is nil, want empty slice")
	}
	if len(m.ManagedConfigs) != 0 {
		t.Errorf("ManagedConfigs has %d entries, want 0", len(m.ManagedConfigs))
	}
}

func TestLoadManifest_NonExistent(t *testing.T) {
	t.Parallel()

	fs := newTestFS(t, map[string]any{
		"/data": &vfst.Dir{Perm: 0755},
	})

	m, err := NewManifestStore(fs).Load("/data/nonexistent.json")
	if err != nil {
		t.Fatalf("NewManifestStore().Load() error = %v", err)
	}

	if m.Version != 1 {
		t.Errorf("Version = %d, want 1", m.Version)
	}
}

func TestLoadManifest_Existing(t *testing.T) {
	t.Parallel()

	content := `{
  "version": 2,
  "last_applied": "2024-01-15T10:30:00Z",
  "installed_emulators": {
    "duckstation": {
      "id": "duckstation",
      "version": "1.0.0",
      "package_path": "/test/packages",
      "installed": "2024-01-15T10:00:00Z"
    }
  },
  "managed_configs": [
    {
      "emulator_id": "duckstation",
      "target": {
        "rel_path": "duckstation/settings.ini",
        "format": "ini",
        "base_dir": "config"
      },
      "baseline_hash": "abc123",
      "last_modified": "2024-01-15T10:00:00Z",
      "managed_keys": []
    }
  ]
}`

	fs := newTestFS(t, map[string]any{
		"/data/manifest.json": content,
	})

	m, err := NewManifestStore(fs).Load("/data/manifest.json")
	if err != nil {
		t.Fatalf("NewManifestStore().Load() error = %v", err)
	}

	if m.Version != 2 {
		t.Errorf("Version = %d, want 2", m.Version)
	}
	if len(m.InstalledEmulators) != 1 {
		t.Errorf("InstalledEmulators has %d entries, want 1", len(m.InstalledEmulators))
	}
	emu, ok := m.InstalledEmulators["duckstation"]
	if !ok {
		t.Fatal("missing duckstation emulator")
	}
	if emu.Version != "1.0.0" {
		t.Errorf("emulator version = %q, want %q", emu.Version, "1.0.0")
	}
	if len(m.ManagedConfigs) != 1 {
		t.Errorf("ManagedConfigs has %d entries, want 1", len(m.ManagedConfigs))
	}
}

func TestLoadManifest_InvalidJSON(t *testing.T) {
	t.Parallel()

	fs := newTestFS(t, map[string]any{
		"/data/manifest.json": "not valid json",
	})

	_, err := NewManifestStore(fs).Load("/data/manifest.json")
	if err == nil {
		t.Error("NewManifestStore().Load() expected error for invalid JSON")
	}
}

func TestManifest_Save_CreatesParentDirs(t *testing.T) {
	t.Parallel()

	fs := newTestFS(t, map[string]any{
		"/data": &vfst.Dir{Perm: 0755},
	})

	m := NewManifest()
	if err := NewManifestStore(fs).Save(m, "/data/nested/deep/manifest.json"); err != nil {
		t.Fatalf("NewManifestStore().Save() error = %v", err)
	}

	vfst.RunTests(t, fs, "",
		vfst.TestPath("/data/nested/deep/manifest.json",
			vfst.TestModeIsRegular(),
		),
	)
}

func TestManifest_Save_WritesValidJSON(t *testing.T) {
	t.Parallel()

	fs := newTestFS(t, map[string]any{
		"/data": &vfst.Dir{Perm: 0755},
	})

	m := NewManifest()
	m.AddEmulator(InstalledEmulator{
		ID:          "test-emu",
		Version:     "2.0.0",
		PackagePath: "/test/packages",
		Installed:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	})

	if err := NewManifestStore(fs).Save(m, "/data/manifest.json"); err != nil {
		t.Fatalf("NewManifestStore().Save() error = %v", err)
	}

	loaded, err := NewManifestStore(fs).Load("/data/manifest.json")
	if err != nil {
		t.Fatalf("NewManifestStore().Load() error = %v", err)
	}

	emu, ok := loaded.InstalledEmulators["test-emu"]
	if !ok {
		t.Fatal("missing test-emu after round-trip")
	}
	if emu.Version != "2.0.0" {
		t.Errorf("version = %q, want %q", emu.Version, "2.0.0")
	}
}

func TestManifest_AddEmulator(t *testing.T) {
	t.Parallel()

	m := NewManifest()

	emu1 := InstalledEmulator{
		ID:          "emu1",
		Version:     "1.0.0",
		PackagePath: "/test/packages",
	}
	m.AddEmulator(emu1)

	if len(m.InstalledEmulators) != 1 {
		t.Fatalf("InstalledEmulators has %d entries, want 1", len(m.InstalledEmulators))
	}

	got, ok := m.InstalledEmulators["emu1"]
	if !ok {
		t.Fatal("emu1 not found")
	}
	if got.Version != "1.0.0" {
		t.Errorf("version = %q, want %q", got.Version, "1.0.0")
	}
}

func TestManifest_AddEmulator_Overwrites(t *testing.T) {
	t.Parallel()

	m := NewManifest()

	m.AddEmulator(InstalledEmulator{ID: "emu", Version: "1.0.0"})
	m.AddEmulator(InstalledEmulator{ID: "emu", Version: "2.0.0"})

	if len(m.InstalledEmulators) != 1 {
		t.Fatalf("InstalledEmulators has %d entries, want 1", len(m.InstalledEmulators))
	}

	got := m.InstalledEmulators["emu"]
	if got.Version != "2.0.0" {
		t.Errorf("version = %q, want %q", got.Version, "2.0.0")
	}
}

func TestManifest_GetEmulator(t *testing.T) {
	t.Parallel()

	m := NewManifest()
	m.AddEmulator(InstalledEmulator{ID: "existing", Version: "1.0.0"})

	emu, ok := m.GetEmulator("existing")
	if !ok {
		t.Error("GetEmulator returned false for existing emulator")
	}
	if emu.Version != "1.0.0" {
		t.Errorf("version = %q, want %q", emu.Version, "1.0.0")
	}

	_, ok = m.GetEmulator("missing")
	if ok {
		t.Error("GetEmulator returned true for missing emulator")
	}
}

func TestManifest_AddManagedConfig(t *testing.T) {
	t.Parallel()

	m := NewManifest()

	cfg := ManagedConfig{
		EmulatorIDs:  []EmulatorID{"emu1"},
		Target:       ConfigTarget{RelPath: "config.ini"},
		BaselineHash: "hash1",
	}
	if err := m.AddManagedConfig(cfg); err != nil {
		t.Fatalf("AddManagedConfig failed: %v", err)
	}

	if len(m.ManagedConfigs) != 1 {
		t.Fatalf("ManagedConfigs has %d entries, want 1", len(m.ManagedConfigs))
	}

	if m.ManagedConfigs[0].BaselineHash != "hash1" {
		t.Errorf("BaselineHash = %q, want %q", m.ManagedConfigs[0].BaselineHash, "hash1")
	}
}

func TestManifest_AddManagedConfig_MergesEmulatorIDs(t *testing.T) {
	t.Parallel()

	m := NewManifest()
	target := ConfigTarget{RelPath: "config.ini", BaseDir: ConfigBaseDirUserConfig}

	if err := m.AddManagedConfig(ManagedConfig{
		EmulatorIDs:  []EmulatorID{"emu1"},
		Target:       target,
		BaselineHash: "hash1",
	}); err != nil {
		t.Fatalf("AddManagedConfig failed: %v", err)
	}
	if err := m.AddManagedConfig(ManagedConfig{
		EmulatorIDs:  []EmulatorID{"emu2"},
		Target:       target,
		BaselineHash: "hash2",
	}); err != nil {
		t.Fatalf("AddManagedConfig failed: %v", err)
	}

	if len(m.ManagedConfigs) != 1 {
		t.Fatalf("ManagedConfigs has %d entries, want 1", len(m.ManagedConfigs))
	}

	if len(m.ManagedConfigs[0].EmulatorIDs) != 2 {
		t.Errorf("EmulatorIDs has %d entries, want 2", len(m.ManagedConfigs[0].EmulatorIDs))
	}

	if m.ManagedConfigs[0].BaselineHash != "hash2" {
		t.Errorf("BaselineHash = %q, want %q", m.ManagedConfigs[0].BaselineHash, "hash2")
	}
}

func TestManifest_GetManagedConfig(t *testing.T) {
	t.Parallel()

	m := NewManifest()
	target := ConfigTarget{RelPath: "config.ini", BaseDir: ConfigBaseDirUserConfig}

	if err := m.AddManagedConfig(ManagedConfig{
		EmulatorIDs:  []EmulatorID{"emu1"},
		Target:       target,
		BaselineHash: "hash1",
	}); err != nil {
		t.Fatalf("AddManagedConfig failed: %v", err)
	}

	cfg, ok := m.GetManagedConfig(target)
	if !ok {
		t.Error("GetManagedConfig returned false for existing config")
	}
	if cfg.BaselineHash != "hash1" {
		t.Errorf("BaselineHash = %q, want %q", cfg.BaselineHash, "hash1")
	}

	_, ok = m.GetManagedConfig(ConfigTarget{RelPath: "other.ini"})
	if ok {
		t.Error("GetManagedConfig returned true for missing config")
	}
}

func TestManifest_GetManagedConfigsForEmulator(t *testing.T) {
	t.Parallel()

	m := NewManifest()

	configs := m.GetManagedConfigsForEmulator("emu1")
	if len(configs) != 0 {
		t.Errorf("expected empty slice, got %d configs", len(configs))
	}

	_ = m.AddManagedConfig(ManagedConfig{
		EmulatorIDs: []EmulatorID{"emu1"},
		Target:      ConfigTarget{RelPath: "config1.ini"},
	})
	_ = m.AddManagedConfig(ManagedConfig{
		EmulatorIDs: []EmulatorID{"emu1"},
		Target:      ConfigTarget{RelPath: "config2.ini"},
	})
	_ = m.AddManagedConfig(ManagedConfig{
		EmulatorIDs: []EmulatorID{"emu2"},
		Target:      ConfigTarget{RelPath: "other.ini"},
	})

	configs = m.GetManagedConfigsForEmulator("emu1")
	if len(configs) != 2 {
		t.Fatalf("expected 2 configs for emu1, got %d", len(configs))
	}

	configs = m.GetManagedConfigsForEmulator("emu2")
	if len(configs) != 1 {
		t.Fatalf("expected 1 config for emu2, got %d", len(configs))
	}

	configs = m.GetManagedConfigsForEmulator("emu3")
	if len(configs) != 0 {
		t.Errorf("expected 0 configs for emu3, got %d", len(configs))
	}
}

func TestManifest_AddManagedConfig_UpdatesKeysOnChange(t *testing.T) {
	t.Parallel()

	m := NewManifest()
	target := ConfigTarget{RelPath: "shared.cfg", BaseDir: ConfigBaseDirUserConfig}

	regions1 := ManagedRegions{SectionRegion{Section: "sec", KeyPrefix: "pfx1"}}
	if err := m.AddManagedConfig(ManagedConfig{
		EmulatorIDs:    []EmulatorID{"emu1"},
		Target:         target,
		ManagedRegions: regions1,
	}); err != nil {
		t.Fatalf("first AddManagedConfig failed: %v", err)
	}

	regions2 := ManagedRegions{SectionRegion{Section: "sec", KeyPrefix: "pfx2"}}
	err := m.AddManagedConfig(ManagedConfig{
		EmulatorIDs:    []EmulatorID{"emu2"},
		Target:         target,
		ManagedRegions: regions2,
	})

	if err != nil {
		t.Fatalf("second AddManagedConfig failed: %v", err)
	}

	if len(m.ManagedConfigs) != 1 {
		t.Fatalf("expected 1 managed config, got %d", len(m.ManagedConfigs))
	}

	cfg := m.ManagedConfigs[0]
	if len(cfg.EmulatorIDs) != 2 {
		t.Errorf("expected 2 emulator IDs, got %d", len(cfg.EmulatorIDs))
	}
	if len(cfg.ManagedRegions) != 1 {
		t.Errorf("expected 1 managed region, got %d", len(cfg.ManagedRegions))
	}
	sr, ok := cfg.ManagedRegions[0].(SectionRegion)
	if !ok {
		t.Fatalf("expected SectionRegion, got %T", cfg.ManagedRegions[0])
	}
	if sr.KeyPrefix != "pfx2" {
		t.Errorf("expected managed region to be updated to new value, got %v", sr.KeyPrefix)
	}
}

func TestManifest_GetManagedConfigsForEmulator_SharedConfig(t *testing.T) {
	t.Parallel()

	m := NewManifest()

	_ = m.AddManagedConfig(ManagedConfig{
		EmulatorIDs: []EmulatorID{"emu1", "emu2"},
		Target:      ConfigTarget{RelPath: "shared.cfg"},
	})
	_ = m.AddManagedConfig(ManagedConfig{
		EmulatorIDs: []EmulatorID{"emu1"},
		Target:      ConfigTarget{RelPath: "emu1-only.cfg"},
	})

	configs := m.GetManagedConfigsForEmulator("emu1")
	if len(configs) != 2 {
		t.Errorf("expected 2 configs for emu1 (shared + own), got %d", len(configs))
	}

	configs = m.GetManagedConfigsForEmulator("emu2")
	if len(configs) != 1 {
		t.Errorf("expected 1 config for emu2 (shared only), got %d", len(configs))
	}
}
