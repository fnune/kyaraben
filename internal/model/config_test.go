package model

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create and save config
	cfg := &KyarabenConfig{
		Global: GlobalConfig{
			UserStore: "~/Emulation",
		},
		Systems: map[SystemID][]EmulatorID{
			SystemIDSNES: {EmulatorIDRetroArchBsnes},
			SystemIDPSX:  {EmulatorIDDuckStation},
		},
	}

	if err := SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	// Load config back
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify contents
	if loaded.Global.UserStore != cfg.Global.UserStore {
		t.Errorf("UserStore mismatch: got %s, want %s", loaded.Global.UserStore, cfg.Global.UserStore)
	}

	if len(loaded.Systems) != len(cfg.Systems) {
		t.Errorf("Systems count mismatch: got %d, want %d", len(loaded.Systems), len(cfg.Systems))
	}

	snesEmulators := loaded.Systems[SystemIDSNES]
	if len(snesEmulators) != 1 || snesEmulators[0] != EmulatorIDRetroArchBsnes {
		t.Errorf("SNES emulator mismatch: got %v, want [%s]", snesEmulators, EmulatorIDRetroArchBsnes)
	}

	psxEmulators := loaded.Systems[SystemIDPSX]
	if len(psxEmulators) != 1 || psxEmulators[0] != EmulatorIDDuckStation {
		t.Errorf("PSX emulator mismatch: got %v, want [%s]", psxEmulators, EmulatorIDDuckStation)
	}
}

func TestExpandUserStore(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	tests := []struct {
		name      string
		userStore string
		want      string
	}{
		{
			name:      "expand tilde",
			userStore: "~/Emulation",
			want:      filepath.Join(home, "Emulation"),
		},
		{
			name:      "absolute path",
			userStore: "/tmp/Emulation",
			want:      "/tmp/Emulation",
		},
		{
			name:      "relative path",
			userStore: "Emulation",
			want:      "Emulation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &KyarabenConfig{
				Global: GlobalConfig{UserStore: tt.userStore},
			}

			got, err := cfg.ExpandUserStore()
			if err != nil {
				t.Fatalf("ExpandUserStore failed: %v", err)
			}

			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEnabledSystems(t *testing.T) {
	cfg := &KyarabenConfig{
		Systems: map[SystemID][]EmulatorID{
			SystemIDSNES: {EmulatorIDRetroArchBsnes},
			SystemIDPSX:  {EmulatorIDDuckStation},
			SystemIDGBA:  {EmulatorIDMGBA},
		},
	}

	systems := cfg.EnabledSystems()
	if len(systems) != 3 {
		t.Errorf("Expected 3 systems, got %d", len(systems))
	}

	systemMap := make(map[SystemID]bool)
	for _, s := range systems {
		systemMap[s] = true
	}

	for _, expected := range []SystemID{SystemIDSNES, SystemIDPSX, SystemIDGBA} {
		if !systemMap[expected] {
			t.Errorf("System %s not found in enabled systems", expected)
		}
	}
}

func TestEnabledEmulators(t *testing.T) {
	cfg := &KyarabenConfig{
		Systems: map[SystemID][]EmulatorID{
			SystemIDSNES:     {EmulatorIDRetroArchBsnes},
			SystemIDPSX:      {EmulatorIDDuckStation, EmulatorIDRetroArchBeetleSaturn},
			SystemIDGameCube: {EmulatorIDDolphin},
			SystemIDWii:      {EmulatorIDDolphin}, // Same emulator, should be deduplicated
		},
	}

	emulators := cfg.EnabledEmulators()

	emulatorMap := make(map[EmulatorID]bool)
	for _, e := range emulators {
		emulatorMap[e] = true
	}

	// Dolphin should appear only once despite being in two systems
	expectedCount := 4 // bsnes, duckstation, beetle_saturn, dolphin
	if len(emulators) != expectedCount {
		t.Errorf("Expected %d emulators, got %d: %v", expectedCount, len(emulators), emulators)
	}

	for _, expected := range []EmulatorID{EmulatorIDRetroArchBsnes, EmulatorIDDuckStation, EmulatorIDRetroArchBeetleSaturn, EmulatorIDDolphin} {
		if !emulatorMap[expected] {
			t.Errorf("Emulator %s not found in enabled emulators", expected)
		}
	}
}

func TestEmulatorVersion(t *testing.T) {
	cfg := &KyarabenConfig{
		Systems: map[SystemID][]EmulatorID{
			SystemIDPSX: {EmulatorIDDuckStation},
		},
		Emulators: map[EmulatorID]EmulatorConf{
			EmulatorIDDuckStation: {Version: "v0.1-10655"},
		},
	}

	if version := cfg.EmulatorVersion(EmulatorIDDuckStation); version != "v0.1-10655" {
		t.Errorf("Expected version v0.1-10655, got %s", version)
	}

	if version := cfg.EmulatorVersion(EmulatorIDMGBA); version != "" {
		t.Errorf("Expected empty version for unconfigured emulator, got %s", version)
	}
}
