package esde

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

type fakeStoreReader struct {
	root string
}

func (f *fakeStoreReader) RomsDir() string {
	return filepath.Join(f.root, "roms")
}

func (f *fakeStoreReader) BiosDir() string {
	return filepath.Join(f.root, "bios")
}

func (f *fakeStoreReader) SystemBiosDir(sys model.SystemID) string {
	return filepath.Join(f.root, "bios", string(sys))
}

func (f *fakeStoreReader) SystemSavesDir(sys model.SystemID) string {
	return filepath.Join(f.root, "saves", string(sys))
}

func (f *fakeStoreReader) EmulatorSavesDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "saves", string(emu))
}

func (f *fakeStoreReader) EmulatorStatesDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "states", string(emu))
}

func (f *fakeStoreReader) EmulatorScreenshotsDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "screenshots", string(emu))
}

func (f *fakeStoreReader) SystemRomsDir(sys model.SystemID) string {
	return filepath.Join(f.root, "roms", string(sys))
}

func (f *fakeStoreReader) CoresDir() string {
	return "/state/cores"
}

func TestBuildCommandPassesSavesDir(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}

	emulators := map[model.EmulatorID]model.Emulator{
		model.EmulatorIDRetroArchMGBA: {
			ID: model.EmulatorIDRetroArchMGBA,
			Launcher: model.LauncherInfo{
				Binary: "mgba",
				RomCommand: func(opts model.RomLaunchOptions) string {
					cmd := opts.BinaryPath
					if opts.Fullscreen {
						cmd += " -f"
					}
					if opts.SavesDir != "" {
						cmd += " -C savegamePath=" + opts.SavesDir
					}
					cmd += " %ROM%"
					return cmd
				},
			},
		},
	}

	ctx := model.FrontendContext{
		BinDir: "/opt/bin",
		Store:  store,
		GetEmulator: func(id model.EmulatorID) (model.Emulator, error) {
			return emulators[id], nil
		},
		GetLaunchArgs: func(id model.EmulatorID) []string {
			return nil
		},
	}

	c := &Config{}

	tests := []struct {
		sysID       model.SystemID
		wantSavesIn string
	}{
		{model.SystemIDGB, "/emulation/saves/gb"},
		{model.SystemIDGBC, "/emulation/saves/gbc"},
		{model.SystemIDGBA, "/emulation/saves/gba"},
	}

	for _, tt := range tests {
		t.Run(string(tt.sysID), func(t *testing.T) {
			cmd := c.buildCommand(ctx, model.EmulatorIDRetroArchMGBA, tt.sysID)

			if !strings.Contains(cmd, "-C savegamePath="+tt.wantSavesIn) {
				t.Errorf("buildCommand() = %q, want savegamePath=%s", cmd, tt.wantSavesIn)
			}
		})
	}
}

func TestGenerateSettingsIncludesDefaults(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}

	ctx := model.FrontendContext{
		Store: store,
	}

	c := &Config{}
	patch, err := c.generateSettings(ctx)
	if err != nil {
		t.Fatalf("generateSettings() error = %v", err)
	}

	expectedSettings := map[string]struct {
		value       string
		defaultOnly bool
	}{
		"ROMDirectory":     {"/emulation/roms", false},
		"Theme":            {"linear-es-de", true},
		"ThemeVariant":     {"simpleCarousel", true},
		"SystemsSorting":   {"manufacturer_year", true},
		"DefaultSortOrder": {"last played, descending", true},
	}

	if len(patch.Entries) != len(expectedSettings) {
		t.Fatalf("expected %d entries, got %d", len(expectedSettings), len(patch.Entries))
	}

	for _, entry := range patch.Entries {
		if len(entry.Path) != 1 {
			t.Errorf("expected path length 1, got %d for %v", len(entry.Path), entry.Path)
			continue
		}
		name := entry.Path[0]
		expected, ok := expectedSettings[name]
		if !ok {
			t.Errorf("unexpected setting: %s", name)
			continue
		}
		if entry.Value != expected.value {
			t.Errorf("setting %s: got %q, want %q", name, entry.Value, expected.value)
		}
		if entry.DefaultOnly != expected.defaultOnly {
			t.Errorf("setting %s: DefaultOnly got %v, want %v", name, entry.DefaultOnly, expected.defaultOnly)
		}
	}
}

func TestBuildCommandIncludesLaunchArgs(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}

	emulators := map[model.EmulatorID]model.Emulator{
		model.EmulatorIDCemu: {
			ID: model.EmulatorIDCemu,
			Launcher: model.LauncherInfo{
				Binary: "cemu",
				RomCommand: func(opts model.RomLaunchOptions) string {
					cmd := opts.BinaryPath
					if len(opts.LaunchArgs) > 0 {
						cmd += " " + strings.Join(opts.LaunchArgs, " ")
					}
					if opts.Fullscreen {
						cmd += " -f"
					}
					cmd += " -g %ROM%"
					return cmd
				},
			},
		},
	}

	launchArgs := map[model.EmulatorID][]string{
		model.EmulatorIDCemu: {"-c", "/some/config/path"},
	}

	ctx := model.FrontendContext{
		BinDir: "/opt/bin",
		Store:  store,
		GetEmulator: func(id model.EmulatorID) (model.Emulator, error) {
			return emulators[id], nil
		},
		GetLaunchArgs: func(id model.EmulatorID) []string {
			return launchArgs[id]
		},
	}

	c := &Config{}
	cmd := c.buildCommand(ctx, model.EmulatorIDCemu, model.SystemIDWiiU)

	if !strings.Contains(cmd, "-c /some/config/path") {
		t.Errorf("buildCommand() = %q, want LaunchArgs included", cmd)
	}

	expectedOrder := "/opt/bin/cemu -c /some/config/path -f -g %ROM%"
	if cmd != expectedOrder {
		t.Errorf("buildCommand() = %q, want %q", cmd, expectedOrder)
	}
}
