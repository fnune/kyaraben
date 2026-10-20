package collections

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/model"
)

const testPort = 47410

type fakeProcess struct{ done chan struct{} }

func (p *fakeProcess) Wait() error {
	<-p.done
	return nil
}

type fakeStarter struct {
	mu       sync.Mutex
	commands [][]string
	proc     *fakeProcess
}

func (s *fakeStarter) Start(command []string, _ string) (Process, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commands = append(s.commands, command)
	s.proc = &fakeProcess{done: make(chan struct{})}
	return s.proc, nil
}

type fakeCommander struct {
	mu   sync.Mutex
	sent []string
}

func (c *fakeCommander) Send(port int, command string, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = append(c.sent, fmt.Sprintf("%d %s", port, command))
	return nil
}

func (c *fakeCommander) Sent() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.sent...)
}

type testServer struct {
	handler   http.Handler
	starter   *fakeStarter
	commander *fakeCommander
	env       *testEnv
}

func newTestServer(t *testing.T) (http.Handler, *fakeStarter) {
	ts := newFullTestServer(t)
	return ts.handler, ts.starter
}

func newFullTestServer(t *testing.T) *testServer {
	t.Helper()
	env := newTestEnv(t, map[string]any{
		"/emu/roms/arcade/ddonpach.zip":  "",
		"/state/bin/retroarch":           "",
		"/state/cores/fbneo_libretro.so": "",
		"/state/collections/cache/.keep": "",
		"/cfg/retroarch/config/.keep":    "",
	})
	cfg := &model.KyarabenConfig{
		Systems: map[model.SystemID][]model.EmulatorID{model.SystemIDArcade: {model.EmulatorIDRetroArchFBNeo}},
		Collections: map[string]model.CollectionConfig{
			"shmups": {URL: "https://games.example.com/shmups/collection.json"},
		},
	}
	svc := NewService(ServiceDeps{
		FS:     env.fs,
		Config: cfg,
		Store:  env.store,
		GetEmulator: func(id model.EmulatorID) (model.Emulator, error) {
			return model.Emulator{ID: id, Launcher: model.LauncherInfo{Binary: "retroarch", CoreName: "fbneo_libretro"}}, nil
		},
		Source:        NewSource(env.fs, &fakeFetcher{data: []byte(testManifest)}, "/state/collections/cache"),
		Overrides:     NewOverrideStore(env.fs, "/state/collections/owned.json"),
		UserConfigDir: "/cfg",
		BinDir:        "/state/bin",
		CoresDir:      "/state/cores",
		CommandPort:   47420,
	})
	starter := &fakeStarter{}
	commander := &fakeCommander{}
	sessions := NewSessions(starter, commander, "/state/collections/logs")
	handler := NewServer(svc, sessions, cfg.Collections, testPort, "test").Handler()
	return &testServer{handler: handler, starter: starter, commander: commander, env: env}
}

func request(h http.Handler, method, path, origin string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	req.Host = "127.0.0.1:47410"
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestServerRejectsForeignOriginsAndHosts(t *testing.T) {
	h, _ := newTestServer(t)

	if rec := request(h, http.MethodGet, "/v1/collections/shmups", "https://evil.example.com", nil); rec.Code != http.StatusForbidden {
		t.Errorf("foreign origin: got %d", rec.Code)
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/health", nil)
	req.Host = "rebind.example.com:47410"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("foreign host: got %d", rec.Code)
	}
}

func TestServerPreflightAllowsPrivateNetwork(t *testing.T) {
	h, _ := newTestServer(t)
	rec := request(h, http.MethodOptions, "/v1/collections/shmups/games/dodonpachi/launch", "https://games.example.com",
		map[string]string{"Access-Control-Request-Private-Network": "true", "Access-Control-Request-Method": "POST"})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://games.example.com" || rec.Header().Get("Access-Control-Allow-Private-Network") != "true" {
		t.Errorf("missing CORS headers: %v", rec.Header())
	}
}

func TestServerStatusAndLaunch(t *testing.T) {
	h, starter := newTestServer(t)
	origin := "https://games.example.com"

	rec := request(h, http.MethodGet, "/v1/collections/shmups", origin, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body)
	}
	var st Status
	if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	playable := map[string]bool{}
	for _, g := range st.Games {
		playable[g.ID] = g.Playable
	}
	if !playable["dodonpachi"] || playable["g-darius"] {
		t.Errorf("unexpected playability %v", playable)
	}

	rec = request(h, http.MethodPost, "/v1/collections/shmups/games/g-darius/launch", origin, nil)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("unplayable launch: got %d", rec.Code)
	}

	rec = request(h, http.MethodPost, "/v1/collections/shmups/games/dodonpachi/launch", origin, nil)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("launch: %d %s", rec.Code, rec.Body)
	}
	var session Session
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	wantCmd := []string{"/state/bin/retroarch", "-f", "-L", "/state/cores/fbneo_libretro.so", "/emu/roms/arcade/ddonpach.zip"}
	if got := starter.commands[0]; len(got) != len(wantCmd) || got[4] != wantCmd[4] || got[3] != wantCmd[3] {
		t.Errorf("command = %v, want %v", got, wantCmd)
	}

	if rec := request(h, http.MethodPost, "/v1/collections/shmups/games/dodonpachi/launch", origin, nil); rec.Code != http.StatusConflict {
		t.Errorf("second launch while running: got %d", rec.Code)
	}
	if len(starter.commands) != 1 {
		t.Errorf("a refused launch must not start anything, got %d starts", len(starter.commands))
	}

	close(starter.proc.done)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		rec = request(h, http.MethodGet, "/v1/sessions/"+session.ID, origin, nil)
		_ = json.Unmarshal(rec.Body.Bytes(), &session)
		if !session.Running {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if session.Running || session.EndedAt == nil {
		t.Errorf("session should have ended: %+v", session)
	}
}

func TestServerMenuLaunchEnablesCommandsForThatLaunchOnly(t *testing.T) {
	ts := newFullTestServer(t)
	origin := "https://games.example.com"
	cfgPath := "/cfg/retroarch/config/FinalBurn Neo/ddonpach.cfg"

	rec := request(ts.handler, http.MethodPost, "/v1/collections/shmups/games/dodonpachi/launch?menu=true", origin, nil)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("menu launch: %d %s", rec.Code, rec.Body)
	}
	cfg, _ := ts.env.fs.ReadFile(cfgPath)
	if values := parseCFG(cfg); values["network_cmd_enable"] != "true" || values["network_cmd_port"] != "47420" {
		t.Errorf("menu launch should enable commands: %s", cfg)
	}
	deadline := time.Now().Add(2 * time.Second)
	for len(ts.commander.Sent()) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if sent := ts.commander.Sent(); len(sent) != 1 || sent[0] != "47420 MENU_TOGGLE" {
		t.Errorf("sent %v", sent)
	}

	close(ts.starter.proc.done)
	for time.Now().Before(deadline.Add(2 * time.Second)) {
		if rec := request(ts.handler, http.MethodPost, "/v1/collections/shmups/games/dodonpachi/launch", origin, nil); rec.Code == http.StatusAccepted {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cfg, _ = ts.env.fs.ReadFile(cfgPath)
	if _, ok := parseCFG(cfg)["network_cmd_enable"]; ok {
		t.Errorf("a normal launch should not enable commands: %s", cfg)
	}
}
