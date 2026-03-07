package model

import (
	"fmt"
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
			Collection: "~/Emulation",
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

	if loaded.Global.Collection != cfg.Global.Collection {
		t.Errorf("Collection mismatch: got %s, want %s", loaded.Global.Collection, cfg.Global.Collection)
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

func TestExpandCollection(t *testing.T) {
	t.Parallel()

	const homeDir = "/home/testuser"

	tests := []struct {
		name       string
		collection string
		want       string
	}{
		{
			name:       "expand tilde",
			collection: "~/Emulation",
			want:       filepath.Join(homeDir, "Emulation"),
		},
		{
			name:       "absolute path",
			collection: "/tmp/Emulation",
			want:       "/tmp/Emulation",
		},
		{
			name:       "relative path",
			collection: "Emulation",
			want:       "Emulation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &KyarabenConfig{
				Global: GlobalConfig{Collection: tt.collection},
			}

			got := cfg.ExpandCollectionWith(homeDir)
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

func TestEmulatorPreset(t *testing.T) {
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
				Graphics: GraphicsConfig{Preset: PresetRetro},
				Emulators: map[EmulatorID]EmulatorConf{
					EmulatorIDDuckStation: {Preset: ptrString(PresetClean)},
				},
			},
			emulator: EmulatorIDDuckStation,
			want:     PresetClean,
		},
		{
			name: "global preset used when no override",
			cfg: &KyarabenConfig{
				Graphics:  GraphicsConfig{Preset: PresetRetro},
				Emulators: map[EmulatorID]EmulatorConf{},
			},
			emulator: EmulatorIDDuckStation,
			want:     PresetRetro,
		},
		{
			name: "default to clean when nothing configured",
			cfg: &KyarabenConfig{
				Graphics:  GraphicsConfig{},
				Emulators: map[EmulatorID]EmulatorConf{},
			},
			emulator: EmulatorIDDuckStation,
			want:     PresetClean,
		},
		{
			name: "nil emulators map uses global",
			cfg: &KyarabenConfig{
				Graphics:  GraphicsConfig{Preset: PresetRetro},
				Emulators: nil,
			},
			emulator: EmulatorIDDuckStation,
			want:     PresetRetro,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.EmulatorPreset(tt.emulator)
			if got != tt.want {
				t.Errorf("EmulatorPreset(%q) = %q, want %q", tt.emulator, got, tt.want)
			}
		})
	}
}

func ptrString(s string) *string {
	return &s
}

func TestEmulatorResume(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		cfg               *KyarabenConfig
		emulator          EmulatorID
		resumeRecommended bool
		want              string
	}{
		{
			name: "emulator override takes precedence over recommended",
			cfg: &KyarabenConfig{
				Savestate: SavestateConfig{Resume: ResumeRecommended},
				Emulators: map[EmulatorID]EmulatorConf{
					EmulatorIDDuckStation: {Resume: ptrString(EmulatorResumeOff)},
				},
			},
			emulator:          EmulatorIDDuckStation,
			resumeRecommended: true,
			want:              EmulatorResumeOff,
		},
		{
			name: "recommended global with recommended emulator enables resume",
			cfg: &KyarabenConfig{
				Savestate: SavestateConfig{Resume: ResumeRecommended},
				Emulators: map[EmulatorID]EmulatorConf{},
			},
			emulator:          EmulatorIDDuckStation,
			resumeRecommended: true,
			want:              EmulatorResumeOn,
		},
		{
			name: "recommended global with non-recommended emulator returns manual",
			cfg: &KyarabenConfig{
				Savestate: SavestateConfig{Resume: ResumeRecommended},
				Emulators: map[EmulatorID]EmulatorConf{},
			},
			emulator:          EmulatorIDCemu,
			resumeRecommended: false,
			want:              EmulatorResumeManual,
		},
		{
			name: "off global returns off regardless of recommendation",
			cfg: &KyarabenConfig{
				Savestate: SavestateConfig{Resume: ResumeOff},
				Emulators: map[EmulatorID]EmulatorConf{},
			},
			emulator:          EmulatorIDDuckStation,
			resumeRecommended: true,
			want:              EmulatorResumeOff,
		},
		{
			name: "manual fallback when nothing configured",
			cfg: &KyarabenConfig{
				Savestate: SavestateConfig{},
				Emulators: map[EmulatorID]EmulatorConf{},
			},
			emulator:          EmulatorIDDuckStation,
			resumeRecommended: true,
			want:              EmulatorResumeManual,
		},
		{
			name: "nil emulators map with recommended",
			cfg: &KyarabenConfig{
				Savestate: SavestateConfig{Resume: ResumeRecommended},
				Emulators: nil,
			},
			emulator:          EmulatorIDDuckStation,
			resumeRecommended: true,
			want:              EmulatorResumeOn,
		},
		{
			name: "emulator override to on for non-recommended emulator",
			cfg: &KyarabenConfig{
				Savestate: SavestateConfig{Resume: ResumeManual},
				Emulators: map[EmulatorID]EmulatorConf{
					EmulatorIDCemu: {Resume: ptrString(EmulatorResumeOn)},
				},
			},
			emulator:          EmulatorIDCemu,
			resumeRecommended: false,
			want:              EmulatorResumeOn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.EmulatorResume(tt.emulator, tt.resumeRecommended)
			if got != tt.want {
				t.Errorf("EmulatorResume(%q, %v) = %q, want %q", tt.emulator, tt.resumeRecommended, got, tt.want)
			}
		})
	}
}

func TestBuildVersionOverrides(t *testing.T) {
	t.Parallel()

	fs := newTestFS(t, map[string]any{
		"/config/config.toml": `[global]
collection = "~/Emulation"

[systems]
snes = ["retroarch:bsnes"]

[emulators."retroarch:bsnes"]
version = "1.19.1"
`,
	})

	store := NewConfigStore(fs)
	cfg, err := store.Load("/config/config.toml", nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Emulators) == 0 {
		t.Fatal("expected Emulators map to be populated")
	}

	if cfg.Emulators[EmulatorIDRetroArchBsnes].Version != "1.19.1" {
		t.Errorf("expected bsnes version 1.19.1, got %q", cfg.Emulators[EmulatorIDRetroArchBsnes].Version)
	}

	getEmulator := func(id EmulatorID) (Emulator, error) {
		return Emulator{ID: id, Package: AppImage{Name: "retroarch"}}, nil
	}
	getFrontend := func(id FrontendID) (Frontend, error) {
		return Frontend{}, fmt.Errorf("not found")
	}

	overrides, err := cfg.BuildVersionOverrides(getEmulator, getFrontend)
	if err != nil {
		t.Fatalf("BuildVersionOverrides failed: %v", err)
	}

	if overrides["bsnes"] != "1.19.1" {
		t.Errorf("expected override bsnes=1.19.1, got %v", overrides)
	}
}

func TestConfigWarningsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		warnings ConfigWarnings
		want     string
	}{
		{
			name:     "empty warnings",
			warnings: nil,
			want:     "",
		},
		{
			name: "single warning",
			warnings: ConfigWarnings{
				{Field: "systems.foo", Message: "unknown system"},
			},
			want: "config has 1 warning(s):\n  - systems.foo: unknown system",
		},
		{
			name: "multiple warnings",
			warnings: ConfigWarnings{
				{Field: "systems.foo", Message: "unknown system"},
				{Field: "frontends.bar", Message: "unknown frontend"},
			},
			want: "config has 2 warning(s):\n  - systems.foo: unknown system\n  - frontends.bar: unknown frontend",
		},
		{
			name: "warning without field",
			warnings: ConfigWarnings{
				{Message: "something went wrong"},
			},
			want: "config has 1 warning(s):\n  - something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.warnings.Error()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadWithWarnings_UnknownSystem(t *testing.T) {
	t.Parallel()

	fs := newTestFS(t, map[string]any{
		"/config/config.toml": `[global]
collection = "~/Emulation"

[systems]
snes = ["retroarch:bsnes"]
unknown_system = ["duckstation"]
`,
	})

	validators := &ConfigValidators{
		GetSystem: func(id SystemID) (System, error) {
			if id == SystemIDSNES {
				return System{ID: id}, nil
			}
			return System{}, fmt.Errorf("unknown")
		},
		GetEmulator: func(id EmulatorID) (Emulator, error) {
			if id == EmulatorIDRetroArchBsnes {
				return Emulator{ID: id}, nil
			}
			return Emulator{}, fmt.Errorf("unknown")
		},
		GetFrontend: func(id FrontendID) (Frontend, error) {
			return Frontend{}, fmt.Errorf("unknown")
		},
	}

	store := NewConfigStore(fs)
	result, err := store.LoadWithWarnings("/config/config.toml", validators)
	if err != nil {
		t.Fatalf("LoadWithWarnings failed: %v", err)
	}

	if !result.Warnings.HasWarnings() {
		t.Error("expected warnings for unknown system, got none")
	}

	if len(result.Config.Systems) != 1 {
		t.Errorf("expected 1 system after filtering, got %d", len(result.Config.Systems))
	}
	if _, ok := result.Config.Systems[SystemIDSNES]; !ok {
		t.Error("expected SNES to be preserved")
	}
}

func TestLoadWithWarnings_UnknownEmulator(t *testing.T) {
	t.Parallel()

	fs := newTestFS(t, map[string]any{
		"/config/config.toml": `[global]
collection = "~/Emulation"

[systems]
snes = ["retroarch:bsnes", "unknown_emulator"]
`,
	})

	validators := &ConfigValidators{
		GetSystem: func(id SystemID) (System, error) {
			return System{ID: id}, nil
		},
		GetEmulator: func(id EmulatorID) (Emulator, error) {
			if id == EmulatorIDRetroArchBsnes {
				return Emulator{ID: id}, nil
			}
			return Emulator{}, fmt.Errorf("unknown")
		},
		GetFrontend: func(id FrontendID) (Frontend, error) {
			return Frontend{}, fmt.Errorf("unknown")
		},
	}

	store := NewConfigStore(fs)
	result, err := store.LoadWithWarnings("/config/config.toml", validators)
	if err != nil {
		t.Fatalf("LoadWithWarnings failed: %v", err)
	}

	if !result.Warnings.HasWarnings() {
		t.Error("expected warnings for unknown emulator, got none")
	}

	emulators := result.Config.Systems[SystemIDSNES]
	if len(emulators) != 1 {
		t.Errorf("expected 1 emulator after filtering, got %d", len(emulators))
	}
	if emulators[0] != EmulatorIDRetroArchBsnes {
		t.Errorf("expected bsnes to be preserved, got %s", emulators[0])
	}
}

func TestLoadWithWarnings_UnknownFrontend(t *testing.T) {
	t.Parallel()

	fs := newTestFS(t, map[string]any{
		"/config/config.toml": `[global]
collection = "~/Emulation"

[frontends.esde]
enabled = true

[frontends.unknown_frontend]
enabled = true
`,
	})

	validators := &ConfigValidators{
		GetSystem: func(id SystemID) (System, error) {
			return System{ID: id}, nil
		},
		GetEmulator: func(id EmulatorID) (Emulator, error) {
			return Emulator{ID: id}, nil
		},
		GetFrontend: func(id FrontendID) (Frontend, error) {
			if id == FrontendIDESDE {
				return Frontend{ID: id}, nil
			}
			return Frontend{}, fmt.Errorf("unknown")
		},
	}

	store := NewConfigStore(fs)
	result, err := store.LoadWithWarnings("/config/config.toml", validators)
	if err != nil {
		t.Fatalf("LoadWithWarnings failed: %v", err)
	}

	if !result.Warnings.HasWarnings() {
		t.Error("expected warnings for unknown frontend, got none")
	}

	if len(result.Config.Frontends) != 1 {
		t.Errorf("expected 1 frontend after filtering, got %d", len(result.Config.Frontends))
	}
	if _, ok := result.Config.Frontends[FrontendIDESDE]; !ok {
		t.Error("expected ESDE to be preserved")
	}
}

func TestLoadWithWarnings_CollectsAllWarnings(t *testing.T) {
	t.Parallel()

	fs := newTestFS(t, map[string]any{
		"/config/config.toml": `[global]
collection = "~/Emulation"

[systems]
unknown1 = ["emu1"]
unknown2 = ["emu2"]
snes = ["retroarch:bsnes"]

[frontends.unknown_fe]
enabled = true
`,
	})

	validators := &ConfigValidators{
		GetSystem: func(id SystemID) (System, error) {
			if id == SystemIDSNES {
				return System{ID: id}, nil
			}
			return System{}, fmt.Errorf("unknown")
		},
		GetEmulator: func(id EmulatorID) (Emulator, error) {
			if id == EmulatorIDRetroArchBsnes {
				return Emulator{ID: id}, nil
			}
			return Emulator{}, fmt.Errorf("unknown")
		},
		GetFrontend: func(id FrontendID) (Frontend, error) {
			return Frontend{}, fmt.Errorf("unknown")
		},
	}

	store := NewConfigStore(fs)
	result, err := store.LoadWithWarnings("/config/config.toml", validators)
	if err != nil {
		t.Fatalf("LoadWithWarnings failed: %v", err)
	}

	if len(result.Warnings) < 3 {
		t.Errorf("expected at least 3 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
}
