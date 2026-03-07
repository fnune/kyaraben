package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	fmt.Fprintf(os.Stderr, "kyaraben-nextui: Run starting\n")

	if a.cfg.Service.Autostart {
		fmt.Fprintf(os.Stderr, "Run: autostart enabled, starting syncthing\n")
		if err := a.startAndConfigureSyncthing(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Run: startAndConfigureSyncthing error: %v\n", err)
			a.showError(fmt.Errorf("start syncthing: %w", err))
		}
		if err := a.svcMgr.EnableAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "Run: EnableAutostart error: %v\n", err)
		}
	} else {
		if err := a.svcMgr.DisableAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "Run: DisableAutostart error: %v\n", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Run: entering main menu loop\n")
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

func (a *App) startAndConfigureSyncthing(ctx context.Context) error {
	if a.svcMgr.IsRunning(ctx) {
		fmt.Fprintf(os.Stderr, "startAndConfigureSyncthing: already running, configuring folders\n")
		if err := a.configureFoldersWithRetry(ctx); err != nil {
			return fmt.Errorf("configure folders: %w", err)
		}
		return nil
	}

	fmt.Fprintf(os.Stderr, "startAndConfigureSyncthing: starting syncthing\n")
	if err := a.svcMgr.Start(ctx); err != nil {
		return fmt.Errorf("start syncthing: %w", err)
	}

	fmt.Fprintf(os.Stderr, "startAndConfigureSyncthing: configuring folders via API\n")
	if err := a.configureFoldersWithRetry(ctx); err != nil {
		return fmt.Errorf("configure folders: %w", err)
	}

	fmt.Fprintf(os.Stderr, "startAndConfigureSyncthing: done\n")
	return nil
}

func (a *App) configureFoldersWithRetry(ctx context.Context) error {
	var lastErr error
	for i := 0; i < 5; i++ {
		if err := a.syncMgr.ConfigureFolders(a.mapper.SyncguestFolderMappings()); err != nil {
			lastErr = err
			fmt.Fprintf(os.Stderr, "configureFoldersWithRetry: attempt %d failed: %v\n", i+1, err)
			time.Sleep(time.Second)
			continue
		}
		return nil
	}
	return lastErr
}

func (a *App) showMainMenu(ctx context.Context) (string, error) {
	statusOpts, statusIdx, statusColor := a.getSyncStatus(ctx)
	guiPort := a.syncMgr.Client().Config().GUIPort
	isRunning := a.svcMgr.IsRunning(ctx)

	syncSelected := 0
	if !isRunning {
		syncSelected = 1
	}
	autostartSelected := 0
	if !a.cfg.Service.Autostart {
		autostartSelected = 1
	}
	syncStatesSelected := 0
	if !a.cfg.Service.SyncStates {
		syncStatesSelected = 1
	}

	items := []ui.MenuItem{
		{Label: "Status", Value: "status", Options: statusOpts, Selected: statusIdx, Unselectable: true, BackgroundColor: statusColor},
		{Label: "Syncing", Value: "toggle_sync", Options: []string{"Enabled", "Disabled"}, Selected: syncSelected, ConfirmText: "Confirm"},
		{Label: "Autostart", Value: "toggle_autostart", Options: []string{"Enabled", "Disabled"}, Selected: autostartSelected, ConfirmText: "Confirm"},
		{Label: "Sync states", Value: "toggle_sync_states", Options: []string{"Enabled", "Disabled"}, Selected: syncStatesSelected, ConfirmText: "Confirm"},
		{Label: "Pair new device", Value: "pair"},
		{Label: "View paired devices", Value: "devices"},
		{Label: fmt.Sprintf("Syncthing UI: http://%s:%d", getLocalIP(), guiPort), Value: "url", Unselectable: true},
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
	case "toggle_autostart":
		if err := a.toggleAutostart(); err != nil {
			a.showError(err)
		}
	case "toggle_sync_states":
		if err := a.toggleSyncStates(ctx); err != nil {
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

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return "unknown"
}

func (a *App) toggleSync(ctx context.Context) error {
	if a.svcMgr.IsRunning(ctx) {
		return a.svcMgr.Stop()
	}
	return a.startAndConfigureSyncthing(ctx)
}

func (a *App) toggleAutostart() error {
	a.cfg.Service.Autostart = !a.cfg.Service.Autostart
	if a.cfg.Service.Autostart {
		if err := a.svcMgr.EnableAutostart(); err != nil {
			return fmt.Errorf("enable autostart: %w", err)
		}
	} else {
		if err := a.svcMgr.DisableAutostart(); err != nil {
			return fmt.Errorf("disable autostart: %w", err)
		}
	}
	return a.cfg.Save(a.dataDir)
}

func (a *App) toggleSyncStates(ctx context.Context) error {
	if !a.cfg.Service.SyncStates {
		confirmed, err := a.confirmSyncStates()
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	a.cfg.Service.SyncStates = !a.cfg.Service.SyncStates
	if err := a.cfg.Save(a.dataDir); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	a.mapper = mapping.NewMapper(a.env.SDCardPath, *a.cfg)

	if a.svcMgr.IsRunning(ctx) {
		if err := a.syncMgr.ConfigureFolders(a.mapper.SyncguestFolderMappings()); err != nil {
			return fmt.Errorf("configure folders: %w", err)
		}
		if err := a.syncMgr.ShareFoldersWithAllDevices(ctx); err != nil {
			return fmt.Errorf("share folders: %w", err)
		}
	}

	return nil
}

func (a *App) confirmSyncStates() (bool, error) {
	items := []ui.MenuItem{
		{Label: "Enable (states may not be compatible)", Value: "yes"},
		{Label: "Cancel", Value: "no"},
	}

	idx, action, err := a.ui.Menu().Show(items, ui.MenuOptions{
		Title:    "Sync states?",
		ShowBack: true,
	})
	if err != nil {
		return false, err
	}
	if action == ui.ActionBack {
		return false, nil
	}

	return items[idx].Value == "yes", nil
}

const (
	colorSecondaryText = "#484949"
	colorLink          = "#1a65a6"
	colorError         = "#ca3535"
	colorSuccess       = "#0f7a52"
)

var statusOptions = []string{"Not running", "Error", "Syncing", "Idle", "Synced"}

const (
	statusNotRunning = 0
	statusError      = 1
	statusSyncing    = 2
	statusIdle       = 3
	statusSynced     = 4
)

func (a *App) getSyncStatus(ctx context.Context) ([]string, int, string) {
	if !a.svcMgr.IsRunning(ctx) {
		return statusOptions, statusNotRunning, colorSecondaryText
	}

	status, err := a.syncMgr.GetStatus(ctx)
	if err != nil {
		return statusOptions, statusError, colorError
	}

	if status.Syncing {
		opts := make([]string, len(statusOptions))
		copy(opts, statusOptions)
		opts[statusSyncing] = fmt.Sprintf("Syncing %d%%", status.Progress)
		return opts, statusSyncing, colorLink
	}

	if len(status.Peers) == 0 {
		opts := make([]string, len(statusOptions))
		copy(opts, statusOptions)
		opts[statusIdle] = "No devices paired"
		return opts, statusIdle, colorSecondaryText
	}

	if status.ConnectedPeers == 0 {
		opts := make([]string, len(statusOptions))
		copy(opts, statusOptions)
		opts[statusIdle] = "No online devices"
		return opts, statusIdle, colorError
	}

	opts := make([]string, len(statusOptions))
	copy(opts, statusOptions)
	if status.ConnectedPeers == 1 {
		opts[statusSynced] = "Synced (1 device)"
	} else {
		opts[statusSynced] = fmt.Sprintf("Synced (%d devices)", status.ConnectedPeers)
	}
	return opts, statusSynced, colorSuccess
}

var errSyncNotRunning = fmt.Errorf("syncing is not running - enable it first")

func (a *App) runPairing(ctx context.Context) error {
	if !a.svcMgr.IsRunning(ctx) {
		return errSyncNotRunning
	}

	items := []ui.MenuItem{
		{Label: "Generate pairing code", Value: "generate"},
		{Label: "Enter pairing code", Value: "enter"},
	}

	fmt.Fprintf(os.Stderr, "runPairing: showing pairing menu\n")
	idx, action, err := a.ui.Menu().Show(items, ui.MenuOptions{
		Title:    "Pair new device",
		ShowBack: true,
	})
	fmt.Fprintf(os.Stderr, "runPairing: menu returned idx=%d, action=%v, err=%v\n", idx, action, err)
	if err != nil {
		return err
	}
	if action == ui.ActionBack {
		return nil
	}

	switch items[idx].Value {
	case "generate":
		fmt.Fprintf(os.Stderr, "runPairing: calling generateCode\n")
		return a.generateCode(ctx)
	case "enter":
		fmt.Fprintf(os.Stderr, "runPairing: calling enterCode\n")
		return a.enterCode(ctx)
	}

	return nil
}

func (a *App) generateCode(ctx context.Context) error {
	session, err := a.syncMgr.CreatePairingSession(ctx)
	if err != nil {
		return friendlyError(err)
	}

	if err := a.ui.Presenter().ShowMessageAsync(fmt.Sprintf("Pairing code: %s", session.Code), ""); err != nil {
		return err
	}

	peerID, err := a.syncMgr.WaitForPeer(ctx, session.Code)
	_ = a.ui.Presenter().Close()
	if err != nil {
		return friendlyError(err)
	}

	if err := a.syncMgr.AddPeer(ctx, peerID); err != nil {
		return friendlyError(err)
	}

	_ = a.ui.Presenter().ShowMessage("Device paired successfully", "")
	return nil
}

func (a *App) enterCode(ctx context.Context) error {
	code, err := a.ui.Keyboard().GetInput(ui.KeyboardOptions{
		Title: "Enter pairing code",
	})
	if err != nil {
		return err
	}
	if code == "" {
		return nil
	}

	peerID, err := a.syncMgr.JoinPairingSession(ctx, code)
	if err != nil {
		return friendlyError(err)
	}

	if err := a.syncMgr.AddPeer(ctx, peerID); err != nil {
		return friendlyError(err)
	}

	_ = a.ui.Presenter().ShowMessage("Device paired successfully", "")
	return nil
}

func (a *App) showDevices(ctx context.Context) error {
	if !a.svcMgr.IsRunning(ctx) {
		return errSyncNotRunning
	}

	status, err := a.syncMgr.GetStatus(ctx)
	if err != nil {
		return err
	}

	if len(status.Peers) == 0 {
		_ = a.ui.Presenter().ShowMessage("No paired devices", "")
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
	fmt.Fprintf(os.Stderr, "showError called with: %v\n", err)
	if presenterErr := a.ui.Presenter().ShowMessage(err.Error(), ""); presenterErr != nil {
		fmt.Fprintf(os.Stderr, "presenter failed: %v\n", presenterErr)
	}
}

var (
	errInvalidCode    = errors.New("Invalid pairing code")       //nolint:staticcheck
	errCodeExpired    = errors.New("Pairing code expired")       //nolint:staticcheck
	errConnectionLost = errors.New("Connection lost, try again") //nolint:staticcheck
)

func friendlyError(err error) error {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "session not found"):
		return errInvalidCode
	case strings.Contains(msg, "session expired"):
		return errCodeExpired
	case strings.Contains(msg, "context deadline exceeded"):
		return errConnectionLost
	case strings.Contains(msg, "connection refused"):
		return errConnectionLost
	}

	return err
}
