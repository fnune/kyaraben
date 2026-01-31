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
		Systems: map[SystemID]SystemConf{
			SystemSNES: {Emulator: string(EmulatorRetroArchBsnes)},
			SystemPSX:  {Emulator: string(EmulatorDuckStation)},
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

	if loaded.Systems[SystemSNES].EmulatorID() != EmulatorRetroArchBsnes {
		t.Errorf("SNES emulator mismatch: got %s, want %s", loaded.Systems[SystemSNES].EmulatorID(), EmulatorRetroArchBsnes)
	}

	if loaded.Systems[SystemPSX].EmulatorID() != EmulatorDuckStation {
		t.Errorf("PSX emulator mismatch: got %s, want %s", loaded.Systems[SystemPSX].EmulatorID(), EmulatorDuckStation)
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
		Systems: map[SystemID]SystemConf{
			SystemSNES:  {Emulator: string(EmulatorRetroArchBsnes)},
			SystemPSX:   {Emulator: string(EmulatorDuckStation)},
			SystemTIC80: {Emulator: string(EmulatorTIC80)},
		},
	}

	systems := cfg.EnabledSystems()
	if len(systems) != 3 {
		t.Errorf("Expected 3 systems, got %d", len(systems))
	}

	// Check that all systems are present (order may vary)
	systemMap := make(map[SystemID]bool)
	for _, s := range systems {
		systemMap[s] = true
	}

	for _, expected := range []SystemID{SystemSNES, SystemPSX, SystemTIC80} {
		if !systemMap[expected] {
			t.Errorf("System %s not found in enabled systems", expected)
		}
	}
}
