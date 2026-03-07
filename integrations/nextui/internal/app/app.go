package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/integrations/nextui/internal/mapping"
	"github.com/fnune/kyaraben/integrations/nextui/internal/service"
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
	env     Env
	cfg     *config.Config
	dataDir string
	mapper  *mapping.Mapper
	syncMgr *syncguest.Manager
	svcMgr  *service.Manager
	ui      ui.UI
}

func New(env Env, cfg *config.Config, dataDir string, syncMgr *syncguest.Manager, svcMgr *service.Manager, appUI ui.UI) *App {
	return &App{
		env:     env,
		cfg:     cfg,
		dataDir: dataDir,
		mapper:  mapping.NewMapper(env.SDCardPath, *cfg),
		syncMgr: syncMgr,
		svcMgr:  svcMgr,
		ui:      appUI,
	}
}

func (a *App) Run(ctx context.Context) error {
	if a.cfg.Service.Enabled {
		if err := a.startSyncthing(ctx); err != nil {
			a.showError(fmt.Errorf("start syncthing: %w", err))
		}
	}

	if a.cfg.Service.StartOnBoot {
		if err := a.svcMgr.EnableAutostart(); err != nil {
			a.showError(fmt.Errorf("enable autostart: %w", err))
		}
	}

	if err := a.syncMgr.ConfigureFolders(a.mapper.SyncguestFolderMappings()); err != nil {
		return fmt.Errorf("configure folders: %w", err)
	}

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

func (a *App) startSyncthing(ctx context.Context) error {
	if a.svcMgr.IsRunning(ctx) {
		return nil
	}
	return a.svcMgr.Start(ctx)
}

func (a *App) showMainMenu(ctx context.Context) (string, error) {
	status := a.getSyncStatus(ctx)
	syncToggle := a.syncToggleLabel()
	bootToggle := a.bootToggleLabel()

	guiPort := a.syncMgr.Client().Config().GUIPort
	items := []ui.MenuItem{
		{Label: fmt.Sprintf("Status: %s", status), Value: "status"},
		{Label: syncToggle, Value: "toggle_sync"},
		{Label: bootToggle, Value: "toggle_boot"},
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
	case "toggle_sync":
		if err := a.toggleSync(ctx); err != nil {
			a.showError(err)
		}
	case "toggle_boot":
		if err := a.toggleBoot(); err != nil {
			a.showError(err)
		}
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

func (a *App) syncToggleLabel() string {
	if a.cfg.Service.Enabled {
		return "Syncing: enabled"
	}
	return "Syncing: disabled"
}

func (a *App) bootToggleLabel() string {
	if a.cfg.Service.StartOnBoot {
		return "Start on boot: enabled"
	}
	return "Start on boot: disabled"
}

func (a *App) toggleSync(ctx context.Context) error {
	if a.cfg.Service.Enabled {
		a.cfg.Service.Enabled = false
		if err := a.svcMgr.Stop(); err != nil {
			return fmt.Errorf("stop syncthing: %w", err)
		}
	} else {
		a.cfg.Service.Enabled = true
		if err := a.svcMgr.Start(ctx); err != nil {
			return fmt.Errorf("start syncthing: %w", err)
		}
	}
	return a.cfg.Save(a.dataDir)
}

func (a *App) toggleBoot() error {
	if a.cfg.Service.StartOnBoot {
		a.cfg.Service.StartOnBoot = false
		if err := a.svcMgr.DisableAutostart(); err != nil {
			return fmt.Errorf("disable autostart: %w", err)
		}
	} else {
		a.cfg.Service.StartOnBoot = true
		if err := a.svcMgr.EnableAutostart(); err != nil {
			return fmt.Errorf("enable autostart: %w", err)
		}
	}
	return a.cfg.Save(a.dataDir)
}

func (a *App) getSyncStatus(ctx context.Context) string {
	if !a.cfg.Service.Enabled {
		return "Disabled"
	}

	status, err := a.syncMgr.GetStatus(ctx)
	if err != nil {
		return "Error"
	}

	if !status.Running {
		return "Not running"
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
	session, err := a.syncMgr.CreatePairingSession(ctx)
	if err != nil {
		return fmt.Errorf("create pairing session: %w", err)
	}

	if err := a.ui.Presenter().ShowMessage("Pairing code", session.Code); err != nil {
		return err
	}

	peerID, err := a.syncMgr.WaitForPeer(ctx, session.Code)
	_ = a.ui.Presenter().Close()
	if err != nil {
		return fmt.Errorf("waiting for peer: %w", err)
	}

	if err := a.syncMgr.AddPeer(ctx, peerID); err != nil {
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

	peerID, err := a.syncMgr.JoinPairingSession(ctx, code)
	if err != nil {
		return fmt.Errorf("join session: %w", err)
	}

	if err := a.syncMgr.AddPeer(ctx, peerID); err != nil {
		return fmt.Errorf("add peer: %w", err)
	}

	_ = a.ui.Presenter().ShowMessage("Paired", "Device paired successfully")
	return nil
}

func (a *App) showDevices(ctx context.Context) error {
	status, err := a.syncMgr.GetStatus(ctx)
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
