package e2e

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/fnune/kyaraben/integrations/nextui/internal/app"
	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/integrations/nextui/internal/service"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui/fake"
	"github.com/fnune/kyaraben/internal/syncguest"
)

func setupTest(t *testing.T) (app.Env, *config.Config, string, *syncguest.Manager, *service.Manager, *fake.UI) {
	t.Helper()

	dataDir := t.TempDir()
	userdataPath := t.TempDir()
	logsPath := filepath.Join(userdataPath, "logs")
	pakPath := t.TempDir()

	env := app.Env{
		SDCardPath:   "/mnt/SDCARD",
		UserdataPath: userdataPath,
		LogsPath:     logsPath,
		Platform:     "tg5040",
		PakPath:      pakPath,
	}

	cfg := config.DefaultConfig()
	cfg.Service.Autostart = false

	syncMgr := syncguest.New(syncguest.DefaultConfig(dataDir))

	svcMgr := service.NewManager(service.Config{
		DataDir:       dataDir,
		PakPath:       pakPath,
		UserdataPath:  userdataPath,
		Platform:      "tg5040",
		LogsPath:      logsPath,
		SyncthingPath: filepath.Join(pakPath, "syncthing"),
		GUIPort:       8484,
	})

	fakeUI := fake.New()

	return env, &cfg, dataDir, syncMgr, svcMgr, fakeUI
}

func TestAppShowsMainMenu(t *testing.T) {
	env, cfg, dataDir, syncMgr, svcMgr, fakeUI := setupTest(t)
	fakeUI.MenuUI.SelectAction = ui.ActionBack

	application := app.New(env, cfg, dataDir, syncMgr, svcMgr, fakeUI)

	err := application.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fakeUI.MenuUI.ShowCalls) == 0 {
		t.Fatal("expected menu to be shown")
	}

	call := fakeUI.MenuUI.ShowCalls[0]
	if call.Options.Title != "Kyaraben" {
		t.Errorf("expected title 'Kyaraben', got %q", call.Options.Title)
	}

	if len(call.Items) < 3 {
		t.Errorf("expected at least 3 menu items, got %d", len(call.Items))
	}
}

func TestAppMenuHasExpectedItems(t *testing.T) {
	env, cfg, dataDir, syncMgr, svcMgr, fakeUI := setupTest(t)
	fakeUI.MenuUI.SelectAction = ui.ActionBack

	application := app.New(env, cfg, dataDir, syncMgr, svcMgr, fakeUI)
	_ = application.Run(context.Background())

	call := fakeUI.MenuUI.ShowCalls[0]

	expectedValues := map[string]bool{
		"status":             false,
		"toggle_sync":        false,
		"toggle_autostart":   false,
		"toggle_sync_states": false,
		"pair":               false,
		"devices":            false,
		"url":                false,
	}

	for _, item := range call.Items {
		if _, ok := expectedValues[item.Value]; ok {
			expectedValues[item.Value] = true
		}
	}

	for value, found := range expectedValues {
		if !found {
			t.Errorf("expected menu item with value %q not found", value)
		}
	}
}

func TestMappingWithCustomConfig(t *testing.T) {
	env, cfg, dataDir, syncMgr, svcMgr, fakeUI := setupTest(t)
	cfg.Saves["gba"] = "Saves/MGBA"
	fakeUI.MenuUI.SelectAction = ui.ActionBack

	application := app.New(env, cfg, dataDir, syncMgr, svcMgr, fakeUI)

	err := application.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToggleAutostartSavesConfig(t *testing.T) {
	env, cfg, dataDir, syncMgr, svcMgr, fakeUI := setupTest(t)

	callCount := 0
	fakeUI.MenuUI.SelectFunc = func(items []ui.MenuItem) (int, ui.Action) {
		callCount++
		if callCount == 1 {
			for i, item := range items {
				if item.Value == "toggle_autostart" {
					return i, ui.ActionSelect
				}
			}
		}
		return 0, ui.ActionBack
	}

	application := app.New(env, cfg, dataDir, syncMgr, svcMgr, fakeUI)
	_ = application.Run(context.Background())

	if !cfg.Service.Autostart {
		t.Error("expected Service.Autostart to be true after toggle")
	}

	if !svcMgr.IsAutostartEnabled() {
		t.Error("expected autostart to be enabled")
	}
}
