package e2e

import (
	"context"
	"testing"

	"github.com/fnune/kyaraben/integrations/nextui/internal/app"
	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui/fake"
	"github.com/fnune/kyaraben/internal/syncguest"
)

func TestAppShowsMainMenu(t *testing.T) {
	fakeUI := fake.New()
	fakeUI.MenuUI.SelectAction = ui.ActionBack

	env := app.Env{
		SDCardPath: "/mnt/SDCARD",
		Platform:   "tg5040",
	}
	cfg := config.DefaultConfig()
	mgr := syncguest.New(syncguest.DefaultConfig(t.TempDir()))

	application := app.New(env, cfg, mgr, fakeUI)

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
	fakeUI := fake.New()
	fakeUI.MenuUI.SelectAction = ui.ActionBack

	env := app.Env{
		SDCardPath: "/mnt/SDCARD",
		Platform:   "tg5040",
	}
	cfg := config.DefaultConfig()
	mgr := syncguest.New(syncguest.DefaultConfig(t.TempDir()))

	application := app.New(env, cfg, mgr, fakeUI)
	_ = application.Run(context.Background())

	call := fakeUI.MenuUI.ShowCalls[0]

	expectedValues := map[string]bool{
		"status":  false,
		"pair":    false,
		"devices": false,
		"url":     false,
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
	env := app.Env{
		SDCardPath: "/mnt/SDCARD",
		Platform:   "tg5040",
	}
	cfg := config.DefaultConfig()
	cfg.Saves["gba"] = "Saves/MGBA"

	fakeUI := fake.New()
	fakeUI.MenuUI.SelectAction = ui.ActionBack

	mgr := syncguest.New(syncguest.DefaultConfig(t.TempDir()))
	application := app.New(env, cfg, mgr, fakeUI)

	err := application.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
