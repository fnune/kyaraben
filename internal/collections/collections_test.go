package collections

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/store"
)

const testManifest = `{
  "schema": 1,
  "id": "shmups",
  "name": "Shmups",
  "games": [
    {
      "id": "dodonpachi",
      "title": "DoDonPachi",
      "system": "arcade",
      "orientation": "vertical",
      "sets": [
        {"emulator": "retroarch:fbneo", "name": "ddonpach", "core_options": {"fbneo-cpu-speed-adjust": "100%"}},
        {"emulator": "retroarch:mame", "name": "ddonpach"}
      ],
      "settings": {"run_ahead_frames": 1, "frame_delay": "auto", "integer_scale": true, "retroarch": {"video_smooth": "false"}}
    },
    {
      "id": "blazing-star",
      "title": "Blazing Star",
      "system": "neogeo",
      "sets": [
        {"emulator": "retroarch:mame", "name": "blazstar", "requires": ["neogeo"]},
        {"emulator": "retroarch:fbneo", "name": "blazstar", "requires": ["neogeo"]}
      ]
    },
    {
      "id": "g-darius",
      "title": "G-Darius",
      "system": "arcade",
      "sets": [{"emulator": "retroarch:mame", "name": "gdarius"}]
    }
  ]
}`

func mustManifest(t *testing.T) *Manifest {
	t.Helper()
	m, err := ParseManifest([]byte(testManifest))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestManifestValidation(t *testing.T) {
	cases := map[string]string{
		"schema":        `{"schema": 2, "id": "x", "games": []}`,
		"set name":      `{"schema": 1, "id": "x", "games": [{"id": "a", "system": "arcade", "sets": [{"emulator": "retroarch:fbneo", "name": "../evil"}]}]}`,
		"path key":      `{"schema": 1, "id": "x", "games": [{"id": "a", "system": "arcade", "settings": {"retroarch": {"video_shader_dir": "/tmp"}}}]}`,
		"unknown key":   `{"schema": 1, "id": "x", "games": [{"id": "a", "system": "arcade", "settings": {"retroarch": {"libretro_directory": "/tmp"}}}]}`,
		"quote":         `{"schema": 1, "id": "x", "games": [{"id": "a", "system": "arcade", "settings": {"retroarch": {"video_smooth": "a\"b"}}}]}`,
		"frame delay":   `{"schema": 1, "id": "x", "games": [{"id": "a", "system": "arcade", "settings": {"frame_delay": "99"}}]}`,
		"duplicate":     `{"schema": 1, "id": "x", "games": [{"id": "a", "system": "arcade"}, {"id": "a", "system": "arcade"}]}`,
		"collection id": `{"schema": 1, "id": "../x", "games": []}`,
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseManifest([]byte(data)); err == nil {
				t.Errorf("expected validation error")
			}
		})
	}
	mustManifest(t)
}

type testEnv struct {
	fs    vfs.FS
	store *store.Collection
}

func newTestEnv(t *testing.T, files map[string]any) *testEnv {
	t.Helper()
	fs, cleanup, err := vfst.NewTestFS(files)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	collection, err := store.NewCollection(fs, paths.NewPaths(""), "/emu")
	if err != nil {
		t.Fatal(err)
	}
	return &testEnv{fs: fs, store: collection}
}

func TestResolve(t *testing.T) {
	m := mustManifest(t)
	both := map[model.SystemID][]model.EmulatorID{
		model.SystemIDArcade: {model.EmulatorIDRetroArchFBNeo, model.EmulatorIDRetroArchMAME},
		model.SystemIDNeoGeo: {model.EmulatorIDRetroArchFBNeo, model.EmulatorIDRetroArchMAME},
	}

	t.Run("first available set wins", func(t *testing.T) {
		env := newTestEnv(t, map[string]any{"/emu/roms/arcade/ddonpach.zip": ""})
		g, _ := m.Game("dodonpachi")
		res := NewResolver(env.fs, env.store, both).Resolve(g, model.CollectionGameConfig{})
		if !res.Playable || res.Emulator != model.EmulatorIDRetroArchFBNeo || res.RomPath != "/emu/roms/arcade/ddonpach.zip" {
			t.Errorf("unexpected resolution %+v", res)
		}
	})

	t.Run("local emulator choice", func(t *testing.T) {
		env := newTestEnv(t, map[string]any{"/emu/roms/arcade/ddonpach.zip": ""})
		g, _ := m.Game("dodonpachi")
		res := NewResolver(env.fs, env.store, both).Resolve(g, model.CollectionGameConfig{Emulator: model.EmulatorIDRetroArchMAME})
		if !res.Playable || res.Emulator != model.EmulatorIDRetroArchMAME {
			t.Errorf("unexpected resolution %+v", res)
		}
	})

	t.Run("MAME does not look in the plain BIOS folder", func(t *testing.T) {
		env := newTestEnv(t, map[string]any{
			"/emu/roms/neogeo/blazstar.zip": "",
			"/emu/bios/neogeo/neogeo.zip":   "",
		})
		g, _ := m.Game("blazing-star")
		res := NewResolver(env.fs, env.store, both).Resolve(g, model.CollectionGameConfig{})
		if !res.Playable || res.Emulator != model.EmulatorIDRetroArchFBNeo {
			t.Errorf("expected FBNeo fallback, got %+v", res)
		}
	})

	t.Run("MAME finds BIOS in mame/bios", func(t *testing.T) {
		env := newTestEnv(t, map[string]any{
			"/emu/roms/neogeo/blazstar.zip":         "",
			"/emu/bios/neogeo/mame/bios/neogeo.zip": "",
		})
		g, _ := m.Game("blazing-star")
		res := NewResolver(env.fs, env.store, both).Resolve(g, model.CollectionGameConfig{})
		if !res.Playable || res.Emulator != model.EmulatorIDRetroArchMAME {
			t.Errorf("expected MAME, got %+v", res)
		}
	})

	t.Run("missing files are reported", func(t *testing.T) {
		env := newTestEnv(t, map[string]any{"/emu/roms/neogeo/blazstar.zip": ""})
		g, _ := m.Game("blazing-star")
		res := NewResolver(env.fs, env.store, both).Resolve(g, model.CollectionGameConfig{})
		if res.Playable || len(res.Missing) != 1 || res.Missing[0] != "neogeo.zip" {
			t.Errorf("unexpected resolution %+v", res)
		}
	})

	t.Run("emulator not enabled", func(t *testing.T) {
		env := newTestEnv(t, map[string]any{"/emu/roms/arcade/gdarius.zip": ""})
		g, _ := m.Game("g-darius")
		fbneoOnly := map[model.SystemID][]model.EmulatorID{model.SystemIDArcade: {model.EmulatorIDRetroArchFBNeo}}
		res := NewResolver(env.fs, env.store, fbneoOnly).Resolve(g, model.CollectionGameConfig{})
		if res.Playable || !strings.Contains(res.Reason, "retroarch:mame") {
			t.Errorf("unexpected resolution %+v", res)
		}
	})
}

func TestGameOverrides(t *testing.T) {
	m := mustManifest(t)
	g, _ := m.Game("dodonpachi")
	env := newTestEnv(t, map[string]any{
		"/emu/roms/arcade/ddonpach.zip":                             "",
		"/cfg/retroarch/config/FinalBurn Neo/FinalBurn Neo.opt":     "fbneo-allow-depth-32 = \"enabled\"\n",
		"/cfg/retroarch/config/FinalBurn Neo/unrelated-content.opt": "",
	})
	enabled := map[model.SystemID][]model.EmulatorID{model.SystemIDArcade: {model.EmulatorIDRetroArchFBNeo}}
	res := NewResolver(env.fs, env.store, enabled).Resolve(g, model.CollectionGameConfig{})

	files, err := GameOverrides("/cfg", g, res, DeviceSettings{Vertical: model.VerticalNative, VRR: true}, env.fs)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected cfg and opt, got %d files", len(files))
	}
	if files[0].Path != "/cfg/retroarch/config/FinalBurn Neo/ddonpach.cfg" {
		t.Errorf("unexpected cfg path %s", files[0].Path)
	}
	cfg := parseCFG(files[0].Content)
	want := map[string]string{
		"video_allow_rotate":     "false",
		"run_ahead_enabled":      "true",
		"run_ahead_frames":       "1",
		"video_frame_delay_auto": "true",
		"video_scale_integer":    "true",
		"vrr_runloop_enable":     "true",
		"video_smooth":           "false",
	}
	for k, v := range want {
		if cfg[k] != v {
			t.Errorf("%s = %q, want %q", k, cfg[k], v)
		}
	}
	opt := parseCFG(files[1].Content)
	if opt["fbneo-allow-depth-32"] != "enabled" || opt["fbneo-cpu-speed-adjust"] != "100%" {
		t.Errorf("opt should inherit core options and add the game's: %v", opt)
	}
}

func TestOverrideStore(t *testing.T) {
	env := newTestEnv(t, map[string]any{"/cfg/user.cfg": "mine = \"1\"\n"})
	s := NewOverrideStore(env.fs, "/state/owned.json")
	a := OverrideFile{Path: "/cfg/a.cfg", Content: []byte("a = \"1\"\n")}
	b := OverrideFile{Path: "/cfg/b.cfg", Content: []byte("b = \"1\"\n")}
	user := OverrideFile{Path: "/cfg/user.cfg", Content: []byte("x")}

	results, err := s.Write("shmups", []OverrideFile{a, b, user}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	assertOutcomes(t, results, map[string]FileOutcome{"/cfg/a.cfg": FileWritten, "/cfg/b.cfg": FileWritten, "/cfg/user.cfg": FileForeign})

	if err := env.fs.WriteFile("/cfg/a.cfg", []byte("edited"), 0o644); err != nil {
		t.Fatal(err)
	}
	a2 := OverrideFile{Path: a.Path, Content: []byte("a = \"2\"\n")}
	results, _ = s.Write("shmups", []OverrideFile{a2, b}, true, false)
	assertOutcomes(t, results, map[string]FileOutcome{"/cfg/a.cfg": FileModified, "/cfg/b.cfg": FileUnchanged})

	results, _ = s.Write("shmups", []OverrideFile{a2}, true, true)
	assertOutcomes(t, results, map[string]FileOutcome{"/cfg/a.cfg": FileWritten, "/cfg/b.cfg": FileRemoved})
	if _, err := env.fs.Stat("/cfg/b.cfg"); err == nil {
		t.Error("pruned file should be deleted")
	}

	results, _ = s.Write("other", []OverrideFile{a2}, false, false)
	assertOutcomes(t, results, map[string]FileOutcome{"/cfg/a.cfg": FileForeign})

	results, _ = s.RemoveAll()
	assertOutcomes(t, results, map[string]FileOutcome{"/cfg/a.cfg": FileRemoved})
	if data, _ := env.fs.ReadFile("/cfg/user.cfg"); string(data) != "mine = \"1\"\n" {
		t.Errorf("user file was touched: %q", data)
	}
}

func assertOutcomes(t *testing.T, results []FileResult, want map[string]FileOutcome) {
	t.Helper()
	got := map[string]FileOutcome{}
	for _, r := range results {
		got[r.Path] = r.Outcome
	}
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
	}
	for p, o := range want {
		if got[p] != o {
			t.Errorf("%s: got %s, want %s", p, got[p], o)
		}
	}
}

type fakeFetcher struct {
	data []byte
	err  error
}

func (f *fakeFetcher) Fetch(context.Context, string) ([]byte, error) { return f.data, f.err }

func TestSourceFallsBackToCache(t *testing.T) {
	env := newTestEnv(t, map[string]any{})
	fetcher := &fakeFetcher{data: []byte(testManifest)}
	src := NewSource(env.fs, fetcher, "/state/cache")

	loaded, err := src.Load(context.Background(), "shmups", "https://example.com/c.json", 0)
	if err != nil || loaded.FromCache {
		t.Fatalf("first load: %v %+v", err, loaded)
	}

	fetcher.err = errors.New("offline")
	loaded, err = src.Load(context.Background(), "shmups", "https://example.com/c.json", 0)
	if err != nil || !loaded.FromCache || loaded.FetchErr == nil {
		t.Fatalf("expected cached copy with fetch error: %v %+v", err, loaded)
	}

	if _, err := src.Load(context.Background(), "other", "https://example.com/c.json", 0); err == nil {
		t.Error("expected error for mismatched collection id with no cache")
	}
}
