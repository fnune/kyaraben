package e2e

import (
	"context"
	"testing"

	"github.com/fnune/kyaraben/integrations/nextui/internal/app"
	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/integrations/nextui/internal/service"
	"github.com/fnune/kyaraben/integrations/nextui/internal/sync"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui/fake"
	"github.com/fnune/kyaraben/internal/syncguest"
	"github.com/twpayne/go-vfs/v5/vfst"
)

type testHarness struct {
	env      app.Env
	cfg      *config.Config
	cfgStore *config.ConfigStore
	syncMgr  *sync.FakeManager
	svcMgr   *service.FakeManager
	fakeUI   *fake.UI
}

func setupTest(t *testing.T) *testHarness {
	t.Helper()

	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/data":   &vfst.Dir{Perm: 0755},
		"/sdcard": &vfst.Dir{Perm: 0755},
		"/pak":    &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	env := app.Env{
		SDCardPath: "/sdcard",
		Platform:   "tg5040",
		PakPath:    "/pak",
	}

	cfgStore := config.NewConfigStore(fs, "/data")
	cfg := config.DefaultConfig()
	cfg.Service.Autostart = false

	return &testHarness{
		env:      env,
		cfg:      &cfg,
		cfgStore: cfgStore,
		syncMgr:  sync.NewFakeManager(),
		svcMgr:   service.NewFakeManager(),
		fakeUI:   fake.New(),
	}
}

func (h *testHarness) newApp() *app.App {
	return app.New(h.env, h.cfg, h.cfgStore, h.syncMgr, h.svcMgr, h.fakeUI)
}

func TestAppShowsMainMenu(t *testing.T) {
	h := setupTest(t)
	h.fakeUI.MenuUI.SelectAction = ui.ActionBack

	application := h.newApp()

	err := application.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(h.fakeUI.MenuUI.ShowCalls) == 0 {
		t.Fatal("expected menu to be shown")
	}

	call := h.fakeUI.MenuUI.ShowCalls[0]
	if call.Options.Title != "Kyaraben" {
		t.Errorf("expected title 'Kyaraben', got %q", call.Options.Title)
	}

	if len(call.Items) < 3 {
		t.Errorf("expected at least 3 menu items, got %d", len(call.Items))
	}
}

func TestAppMenuHasExpectedItems(t *testing.T) {
	h := setupTest(t)
	h.fakeUI.MenuUI.SelectAction = ui.ActionBack

	application := h.newApp()
	_ = application.Run(context.Background())

	call := h.fakeUI.MenuUI.ShowCalls[0]

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
	h := setupTest(t)
	h.cfg.Saves["gba"] = "Saves/MGBA"
	h.fakeUI.MenuUI.SelectAction = ui.ActionBack

	application := h.newApp()

	err := application.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToggleAutostartSavesConfig(t *testing.T) {
	h := setupTest(t)

	callCount := 0
	h.fakeUI.MenuUI.SelectFunc = func(items []ui.MenuItem) (int, ui.Action) {
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

	application := h.newApp()
	_ = application.Run(context.Background())

	if !h.cfg.Service.Autostart {
		t.Error("expected Service.Autostart to be true after toggle")
	}

	if !h.svcMgr.IsAutostartEnabled() {
		t.Error("expected autostart to be enabled")
	}
}

func TestToggleSyncStatesConfiguresFoldersAndShares(t *testing.T) {
	h := setupTest(t)
	h.svcMgr.Running = true

	callCount := 0
	h.fakeUI.MenuUI.SelectFunc = func(items []ui.MenuItem) (int, ui.Action) {
		callCount++
		switch callCount {
		case 1:
			for i, item := range items {
				if item.Value == "toggle_sync_states" {
					return i, ui.ActionSelect
				}
			}
		case 2:
			for i, item := range items {
				if item.Value == "yes" {
					return i, ui.ActionSelect
				}
			}
		}
		return 0, ui.ActionBack
	}

	application := h.newApp()
	_ = application.Run(context.Background())

	if !h.cfg.Service.SyncStates {
		t.Error("expected SyncStates to be enabled")
	}

	if len(h.syncMgr.ConfigureFoldersCalls) == 0 {
		t.Error("expected ConfigureFolders to be called")
	}

	if h.syncMgr.ShareFoldersWithAllCalls == 0 {
		t.Error("expected ShareFoldersWithAllDevices to be called")
	}

	folders := h.syncMgr.ConfigureFoldersCalls[len(h.syncMgr.ConfigureFoldersCalls)-1]
	hasStateFolders := false
	for _, f := range folders {
		if f.ID == "kyaraben-states-retroarch:snes9x" {
			hasStateFolders = true
			break
		}
	}
	if !hasStateFolders {
		t.Error("expected state folders to be configured")
	}
}

func TestToggleSyncStartsSyncthing(t *testing.T) {
	h := setupTest(t)
	h.svcMgr.Running = false

	callCount := 0
	h.fakeUI.MenuUI.SelectFunc = func(items []ui.MenuItem) (int, ui.Action) {
		callCount++
		if callCount == 1 {
			for i, item := range items {
				if item.Value == "toggle_sync" {
					return i, ui.ActionSelect
				}
			}
		}
		return 0, ui.ActionBack
	}

	application := h.newApp()
	_ = application.Run(context.Background())

	if h.svcMgr.StartCalls == 0 {
		t.Error("expected Start to be called")
	}
}

func TestToggleSyncStopsSyncthing(t *testing.T) {
	h := setupTest(t)
	h.svcMgr.Running = true

	callCount := 0
	h.fakeUI.MenuUI.SelectFunc = func(items []ui.MenuItem) (int, ui.Action) {
		callCount++
		if callCount == 1 {
			for i, item := range items {
				if item.Value == "toggle_sync" {
					return i, ui.ActionSelect
				}
			}
		}
		return 0, ui.ActionBack
	}

	application := h.newApp()
	_ = application.Run(context.Background())

	if h.svcMgr.StopCalls == 0 {
		t.Error("expected Stop to be called")
	}
}

func TestGeneratePairingCode(t *testing.T) {
	h := setupTest(t)
	h.svcMgr.Running = true

	callCount := 0
	h.fakeUI.MenuUI.SelectFunc = func(items []ui.MenuItem) (int, ui.Action) {
		callCount++
		switch callCount {
		case 1:
			for i, item := range items {
				if item.Value == "pair" {
					return i, ui.ActionSelect
				}
			}
		case 2:
			return 0, ui.ActionSelect
		}
		return 0, ui.ActionBack
	}

	application := h.newApp()
	_ = application.Run(context.Background())

	if h.syncMgr.CreatePairingSessionCalls == 0 {
		t.Error("expected CreatePairingSession to be called")
	}

	if len(h.syncMgr.WaitForPeerCalls) == 0 {
		t.Error("expected WaitForPeer to be called")
	}

	if h.syncMgr.WaitForPeerCalls[0] != "ABC123" {
		t.Errorf("expected WaitForPeer to be called with ABC123, got %s", h.syncMgr.WaitForPeerCalls[0])
	}

	if len(h.syncMgr.AddPeerCalls) == 0 {
		t.Error("expected AddPeer to be called")
	}

	if len(h.fakeUI.PresenterUI.Messages) == 0 {
		t.Error("expected pairing code to be shown")
	}
}

func TestEnterPairingCode(t *testing.T) {
	h := setupTest(t)
	h.svcMgr.Running = true
	h.fakeUI.KeyboardUI.Input = "XYZ789"

	callCount := 0
	h.fakeUI.MenuUI.SelectFunc = func(items []ui.MenuItem) (int, ui.Action) {
		callCount++
		switch callCount {
		case 1:
			for i, item := range items {
				if item.Value == "pair" {
					return i, ui.ActionSelect
				}
			}
		case 2:
			return 1, ui.ActionSelect
		}
		return 0, ui.ActionBack
	}

	application := h.newApp()
	_ = application.Run(context.Background())

	if len(h.syncMgr.JoinPairingSessionCalls) == 0 {
		t.Error("expected JoinPairingSession to be called")
	}

	if h.syncMgr.JoinPairingSessionCalls[0] != "XYZ789" {
		t.Errorf("expected JoinPairingSession with XYZ789, got %s", h.syncMgr.JoinPairingSessionCalls[0])
	}

	if len(h.syncMgr.AddPeerCalls) == 0 {
		t.Error("expected AddPeer to be called")
	}
}

func TestPairingRequiresSyncRunning(t *testing.T) {
	h := setupTest(t)
	h.svcMgr.Running = false

	callCount := 0
	h.fakeUI.MenuUI.SelectFunc = func(items []ui.MenuItem) (int, ui.Action) {
		callCount++
		if callCount == 1 {
			for i, item := range items {
				if item.Value == "pair" {
					return i, ui.ActionSelect
				}
			}
		}
		return 0, ui.ActionBack
	}

	application := h.newApp()
	_ = application.Run(context.Background())

	if h.syncMgr.CreatePairingSessionCalls != 0 {
		t.Error("expected CreatePairingSession NOT to be called when sync not running")
	}

	hasErrorMessage := false
	for _, msg := range h.fakeUI.PresenterUI.Messages {
		if msg.Title == "syncing is not running - enable it first" {
			hasErrorMessage = true
			break
		}
	}
	if !hasErrorMessage {
		t.Error("expected error message about sync not running")
	}
}

func TestShowDevices(t *testing.T) {
	h := setupTest(t)
	h.svcMgr.Running = true
	h.syncMgr.StatusValue = &syncguest.Status{
		Running: true,
		Peers: []syncguest.PeerStatus{
			{ID: "DEVICE-1", Name: "steamdeck", Connected: true},
			{ID: "DEVICE-2", Name: "desktop", Connected: false},
		},
	}

	callCount := 0
	h.fakeUI.MenuUI.SelectFunc = func(items []ui.MenuItem) (int, ui.Action) {
		callCount++
		if callCount == 1 {
			for i, item := range items {
				if item.Value == "devices" {
					return i, ui.ActionSelect
				}
			}
		}
		return 0, ui.ActionBack
	}

	application := h.newApp()
	_ = application.Run(context.Background())

	if len(h.fakeUI.MenuUI.ShowCalls) < 2 {
		t.Fatal("expected devices menu to be shown")
	}

	devicesCall := h.fakeUI.MenuUI.ShowCalls[1]
	if len(devicesCall.Items) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devicesCall.Items))
	}
}

func TestStatusDisplayPluralization(t *testing.T) {
	tests := []struct {
		name           string
		connectedPeers int
		wantContains   string
	}{
		{"one device", 1, "1 device"},
		{"multiple devices", 3, "3 devices"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setupTest(t)
			h.svcMgr.Running = true

			peers := make([]syncguest.PeerStatus, tt.connectedPeers)
			for i := 0; i < tt.connectedPeers; i++ {
				peers[i] = syncguest.PeerStatus{Connected: true}
			}

			h.syncMgr.StatusValue = &syncguest.Status{
				Running:        true,
				Syncing:        false,
				Progress:       100,
				ConnectedPeers: tt.connectedPeers,
				Peers:          peers,
			}
			h.fakeUI.MenuUI.SelectAction = ui.ActionBack

			application := h.newApp()
			_ = application.Run(context.Background())

			call := h.fakeUI.MenuUI.ShowCalls[0]
			statusItem := call.Items[0]

			found := false
			for _, opt := range statusItem.Options {
				if opt == "Synced ("+tt.wantContains+")" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected status option containing %q, got %v", tt.wantContains, statusItem.Options)
			}
		})
	}
}

func TestAutostartEnablesSyncOnRun(t *testing.T) {
	h := setupTest(t)
	h.cfg.Service.Autostart = true
	h.fakeUI.MenuUI.SelectAction = ui.ActionBack

	application := h.newApp()
	_ = application.Run(context.Background())

	if h.svcMgr.StartCalls == 0 {
		t.Error("expected Start to be called when autostart is enabled")
	}

	if len(h.syncMgr.ConfigureFoldersCalls) == 0 {
		t.Error("expected ConfigureFolders to be called on autostart")
	}
}
