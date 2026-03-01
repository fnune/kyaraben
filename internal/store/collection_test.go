package store

import (
	"testing"

	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/testutil"
)

func mustNewCollection(t *testing.T, fs vfs.FS, path string) *Collection {
	t.Helper()
	s, err := NewCollection(fs, paths.DefaultPaths(), path)
	if err != nil {
		t.Fatalf("NewCollection(%q) failed: %v", path, err)
	}
	return s
}

func TestCollectionInitialize(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/emulation": &vfst.Dir{Perm: 0755},
	})

	store := mustNewCollection(t, fs, "/emulation")

	if err := store.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	vfst.RunTests(t, fs, "",
		vfst.TestPath("/emulation/roms", vfst.TestIsDir()),
		vfst.TestPath("/emulation/bios", vfst.TestIsDir()),
		vfst.TestPath("/emulation/saves", vfst.TestIsDir()),
		vfst.TestPath("/emulation/states", vfst.TestIsDir()),
		vfst.TestPath("/emulation/screenshots", vfst.TestIsDir()),
	)

	if !store.IsInitialized() {
		t.Error("IsInitialized returned false after Initialize")
	}
}

func TestCollectionInitializeForEmulator(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/emulation": &vfst.Dir{Perm: 0755},
	})

	store := mustNewCollection(t, fs, "/emulation")

	if err := store.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if err := store.InitializeForEmulator(model.SystemIDSNES, model.EmulatorIDRetroArchBsnes, model.StandardPathUsage()); err != nil {
		t.Fatalf("InitializeForEmulator failed: %v", err)
	}

	vfst.RunTests(t, fs, "",
		vfst.TestPath("/emulation/roms/snes", vfst.TestIsDir()),
		vfst.TestPath("/emulation/bios/snes", vfst.TestIsDir()),
		vfst.TestPath("/emulation/saves/snes", vfst.TestIsDir()),
		vfst.TestPath("/emulation/states/retroarch:bsnes", vfst.TestIsDir()),
		vfst.TestPath("/emulation/screenshots/retroarch", vfst.TestIsDir()),
	)
}

func TestCollectionPaths(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{})

	store := mustNewCollection(t, fs, "/home/user/Emulation")

	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{"RomsDir", store.RomsDir, "/home/user/Emulation/roms"},
		{"BiosDir", store.BiosDir, "/home/user/Emulation/bios"},
		{"SavesDir", store.SavesDir, "/home/user/Emulation/saves"},
		{"StatesDir", store.StatesDir, "/home/user/Emulation/states"},
		{"ScreenshotsDir", store.ScreenshotsDir, "/home/user/Emulation/screenshots"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestCollectionSystemPaths(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{})

	store := mustNewCollection(t, fs, "/home/user/Emulation")

	tests := []struct {
		name   string
		system model.SystemID
		fn     func(model.SystemID) string
		want   string
	}{
		{"SystemRomsDir", model.SystemIDSNES, store.SystemRomsDir, "/home/user/Emulation/roms/snes"},
		{"SystemBiosDir", model.SystemIDPSX, store.SystemBiosDir, "/home/user/Emulation/bios/psx"},
		{"SystemSavesDir", model.SystemIDGBA, store.SystemSavesDir, "/home/user/Emulation/saves/gba"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(tt.system)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestCollectionEmulatorPaths(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{})

	store := mustNewCollection(t, fs, "/home/user/Emulation")

	got := store.EmulatorStatesDir(model.EmulatorIDRetroArchBsnes)
	want := "/home/user/Emulation/states/retroarch:bsnes"
	if got != want {
		t.Errorf("EmulatorStatesDir got %s, want %s", got, want)
	}

	gotScreenshots := store.EmulatorScreenshotsDir(model.EmulatorIDDuckStation)
	wantScreenshots := "/home/user/Emulation/screenshots/duckstation"
	if gotScreenshots != wantScreenshots {
		t.Errorf("EmulatorScreenshotsDir got %s, want %s", gotScreenshots, wantScreenshots)
	}

	gotRetroArchScreenshots := store.EmulatorScreenshotsDir(model.EmulatorIDRetroArchBsnes)
	wantRetroArchScreenshots := "/home/user/Emulation/screenshots/retroarch"
	if gotRetroArchScreenshots != wantRetroArchScreenshots {
		t.Errorf("EmulatorScreenshotsDir (RetroArch) got %s, want %s", gotRetroArchScreenshots, wantRetroArchScreenshots)
	}
}

func TestCollectionExists(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/existing": &vfst.Dir{Perm: 0755},
		"/file":     "test content",
	})

	store := mustNewCollection(t, fs, "/nonexistent")
	if store.Exists() {
		t.Error("Exists returned true for non-existent directory")
	}

	store = mustNewCollection(t, fs, "/existing")
	if !store.Exists() {
		t.Error("Exists returned false for existing directory")
	}

	store = mustNewCollection(t, fs, "/file")
	if store.Exists() {
		t.Error("Exists returned true for file (not directory)")
	}
}
