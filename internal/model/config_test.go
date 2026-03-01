package model

import (
	"path/filepath"
	"testing"

	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"
)

func newTestFS(t *testing.T, root map[string]any) vfs.FS {
	t.Helper()
	fs, cleanup, err := vfst.NewTestFS(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	return fs
}

func TestLoadSaveConfig(t *testing.T) {
	t.Parallel()

	fs := newTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	configPath := "/config/config.toml"

	cfg := &KyarabenConfig{
		Global: GlobalConfig{
			UserStore: "~/Emulation",
		},
		Systems: map[SystemID][]EmulatorID{
			SystemIDSNES: {EmulatorIDRetroArchBsnes},
			SystemIDPSX:  {EmulatorIDDuckStation},
		},
	}

	store := NewConfigStore(fs)
	if err := store.Save(cfg, configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := fs.Stat(configPath); err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	loaded, err := store.Load(configPath, nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

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
	t.Parallel()

	const homeDir = "/home/testuser"

	tests := []struct {
		name      string
		userStore string
		want      string
	}{
		{
			name:      "expand tilde",
			userStore: "~/Emulation",
			want:      filepath.Join(homeDir, "Emulation"),
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

			got := cfg.ExpandUserStoreWith(homeDir)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEnabledSystems(t *testing.T) {
	t.Parallel()

	cfg := &KyarabenConfig{
		Systems: map[SystemID][]EmulatorID{
			SystemIDSNES: {EmulatorIDRetroArchBsnes},
			SystemIDPSX:  {EmulatorIDDuckStation},
			SystemIDGBA:  {EmulatorIDRetroArchMGBA},
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
	t.Parallel()

	cfg := &KyarabenConfig{
		Systems: map[SystemID][]EmulatorID{
			SystemIDSNES:     {EmulatorIDRetroArchBsnes},
			SystemIDPSX:      {EmulatorIDDuckStation, EmulatorIDRetroArchBeetleSaturn},
			SystemIDGameCube: {EmulatorIDDolphin},
			SystemIDWii:      {EmulatorIDDolphin},
		},
	}

	emulators := cfg.EnabledEmulators()

	emulatorMap := make(map[EmulatorID]bool)
	for _, e := range emulators {
		emulatorMap[e] = true
	}

	expectedCount := 4
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
	t.Parallel()

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

	if version := cfg.EmulatorVersion(EmulatorIDRetroArchMGBA); version != "" {
		t.Errorf("Expected empty version for unconfigured emulator, got %s", version)
	}
}

func TestEmulatorShaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      *KyarabenConfig
		emulator EmulatorID
		want     string
	}{
		{
			name: "emulator override takes precedence",
			cfg: &KyarabenConfig{
				Graphics: GraphicsConfig{Shaders: ShadersOff},
				Emulators: map[EmulatorID]EmulatorConf{
					EmulatorIDDuckStation: {Shaders: ptrString(ShadersOn)},
				},
			},
			emulator: EmulatorIDDuckStation,
			want:     ShadersOn,
		},
		{
			name: "graphics default when no emulator override",
			cfg: &KyarabenConfig{
				Graphics:  GraphicsConfig{Shaders: ShadersOff},
				Emulators: map[EmulatorID]EmulatorConf{},
			},
			emulator: EmulatorIDDuckStation,
			want:     ShadersOff,
		},
		{
			name: "manual fallback when nothing configured",
			cfg: &KyarabenConfig{
				Graphics:  GraphicsConfig{},
				Emulators: map[EmulatorID]EmulatorConf{},
			},
			emulator: EmulatorIDDuckStation,
			want:     ShadersManual,
		},
		{
			name: "nil emulators map",
			cfg: &KyarabenConfig{
				Graphics:  GraphicsConfig{Shaders: ShadersOn},
				Emulators: nil,
			},
			emulator: EmulatorIDDuckStation,
			want:     ShadersOn,
		},
		{
			name: "emulator override to manual",
			cfg: &KyarabenConfig{
				Graphics: GraphicsConfig{Shaders: ShadersOn},
				Emulators: map[EmulatorID]EmulatorConf{
					EmulatorIDDuckStation: {Shaders: ptrString(ShadersManual)},
				},
			},
			emulator: EmulatorIDDuckStation,
			want:     ShadersManual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.EmulatorShaders(tt.emulator)
			if got != tt.want {
				t.Errorf("EmulatorShaders(%q) = %q, want %q", tt.emulator, got, tt.want)
			}
		})
	}
}

func ptrString(s string) *string {
	return &s
}
