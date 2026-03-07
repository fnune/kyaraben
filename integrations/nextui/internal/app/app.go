package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/integrations/nextui/internal/mapping"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui"
	"github.com/fnune/kyaraben/internal/syncguest"
)

type Env struct {
	SDCardPath   string
	SavesPath    string
	BIOSPath     string
	UserdataPath string
	LogsPath     string
	Platform     string
	PakPath      string
}

func EnvFromOS() Env {
	pakPath := os.Getenv("PAK_PATH")
	if pakPath == "" {
		exe, _ := os.Executable()
		pakPath = filepath.Dir(exe)
	}

	return Env{
		SDCardPath:   os.Getenv("SDCARD_PATH"),
		SavesPath:    os.Getenv("SAVES_PATH"),
		BIOSPath:     os.Getenv("BIOS_PATH"),
		UserdataPath: os.Getenv("USERDATA_PATH"),
		LogsPath:     os.Getenv("LOGS_PATH"),
		Platform:     os.Getenv("PLATFORM"),
		PakPath:      pakPath,
	}
}

type App struct {
	env    Env
	cfg    config.Config
	mapper *mapping.Mapper
	mgr    *syncguest.Manager
	ui     ui.UI
}

func New(env Env, cfg config.Config, mgr *syncguest.Manager, appUI ui.UI) *App {
	return &App{
		env:    env,
		cfg:    cfg,
		mapper: mapping.NewMapper(env.SDCardPath, cfg.TagOverrides),
		mgr:    mgr,
		ui:     appUI,
	}
}

func (a *App) Run(ctx context.Context) error {
	for {
		action, err := a.showMainMenu(ctx)
		if err != nil {
			return err
		}
		if action == "exit" {
			return nil
		}
	}
}

func (a *App) showMainMenu(ctx context.Context) (string, error) {
	status := a.getSyncStatus(ctx)

	guiPort := a.mgr.Client().Config().GUIPort
	items := []ui.MenuItem{
		{Label: fmt.Sprintf("Status: %s", status), Value: "status"},
		{Label: "Pair new device", Value: "pair"},
		{Label: "View paired devices", Value: "devices"},
		{Label: fmt.Sprintf("Syncthing UI: http://localhost:%d", guiPort), Value: "url"},
	}

	idx, action, err := a.ui.Menu().Show(items, ui.MenuOptions{
		Title:    "Kyaraben",
		ShowBack: true,
	})
	if err != nil {
		return "", err
	}
	if action == ui.ActionBack {
		return "exit", nil
	}

	switch items[idx].Value {
	case "pair":
		if err := a.runPairing(ctx); err != nil {
			a.showError(err)
		}
	case "devices":
		if err := a.showDevices(ctx); err != nil {
			a.showError(err)
		}
	}

	return "", nil
}

func (a *App) getSyncStatus(ctx context.Context) string {
	status, err := a.mgr.GetStatus(ctx)
	if err != nil {
		return "Error"
	}

	if !status.Running {
		return "Syncthing not running"
	}

	if status.Syncing {
		return fmt.Sprintf("Syncing %d%%", status.Progress)
	}

	if status.ConnectedPeers == 0 {
		return "Disconnected"
	}

	return fmt.Sprintf("Synced (%d devices)", status.ConnectedPeers)
}

func (a *App) runPairing(ctx context.Context) error {
	items := []ui.MenuItem{
		{Label: "Generate pairing code", Value: "generate"},
		{Label: "Enter pairing code", Value: "enter"},
	}

	idx, action, err := a.ui.Menu().Show(items, ui.MenuOptions{
		Title:    "Pair new device",
		ShowBack: true,
	})
	if err != nil {
		return err
	}
	if action == ui.ActionBack {
		return nil
	}

	switch items[idx].Value {
	case "generate":
		return a.generateCode(ctx)
	case "enter":
		return a.enterCode(ctx)
	}

	return nil
}

func (a *App) generateCode(ctx context.Context) error {
	session, err := a.mgr.CreatePairingSession(ctx)
	if err != nil {
		return fmt.Errorf("create pairing session: %w", err)
	}

	if err := a.ui.Presenter().ShowMessage("Pairing code", session.Code); err != nil {
		return err
	}

	peerID, err := a.mgr.WaitForPeer(ctx, session.Code)
	_ = a.ui.Presenter().Close()
	if err != nil {
		return fmt.Errorf("waiting for peer: %w", err)
	}

	if err := a.mgr.AddPeer(ctx, peerID); err != nil {
		return fmt.Errorf("add peer: %w", err)
	}

	_ = a.ui.Presenter().ShowMessage("Paired", "Device paired successfully")
	return nil
}

func (a *App) enterCode(ctx context.Context) error {
	code, err := a.ui.Keyboard().GetInput(ui.KeyboardOptions{
		Title:     "Enter pairing code",
		MaxLength: 6,
		Uppercase: true,
	})
	if err != nil {
		return err
	}
	if code == "" {
		return nil
	}

	peerID, err := a.mgr.JoinPairingSession(ctx, code)
	if err != nil {
		return fmt.Errorf("join session: %w", err)
	}

	if err := a.mgr.AddPeer(ctx, peerID); err != nil {
		return fmt.Errorf("add peer: %w", err)
	}

	_ = a.ui.Presenter().ShowMessage("Paired", "Device paired successfully")
	return nil
}

func (a *App) showDevices(ctx context.Context) error {
	status, err := a.mgr.GetStatus(ctx)
	if err != nil {
		return err
	}

	if len(status.Peers) == 0 {
		_ = a.ui.Presenter().ShowMessage("Devices", "No paired devices")
		return nil
	}

	var items []ui.MenuItem
	for _, p := range status.Peers {
		state := "offline"
		if p.Connected {
			state = "connected"
		}

		name := p.Name
		if name == "" {
			name = p.ID[:8] + "..."
		}

		items = append(items, ui.MenuItem{
			Label: fmt.Sprintf("%s (%s)", name, state),
			Value: p.ID,
		})
	}

	_, action, err := a.ui.Menu().Show(items, ui.MenuOptions{
		Title:    "Paired devices",
		ShowBack: true,
	})
	if err != nil {
		return err
	}
	if action == ui.ActionBack {
		return nil
	}

	return nil
}

func (a *App) showError(err error) {
	_ = a.ui.Presenter().ShowMessage("Error", err.Error())
}
