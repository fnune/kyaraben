package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/integrations/nextui/internal/mapping"
	"github.com/fnune/kyaraben/integrations/nextui/internal/pairing"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui"
	"github.com/fnune/kyaraben/internal/syncthing"
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
	env         Env
	cfg         config.Config
	mapper      *mapping.Mapper
	client      syncthing.SyncClient
	relayClient *syncthing.RelayClient
	ui          ui.UI
}

func New(env Env, cfg config.Config, stClient syncthing.SyncClient, relayClient *syncthing.RelayClient, appUI ui.UI) *App {
	return &App{
		env:         env,
		cfg:         cfg,
		mapper:      mapping.NewMapper(env.SDCardPath, cfg.TagOverrides),
		client:      stClient,
		relayClient: relayClient,
		ui:          appUI,
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

	items := []ui.MenuItem{
		{Label: fmt.Sprintf("Status: %s", status), Value: "status"},
		{Label: "Pair new device", Value: "pair"},
		{Label: "View paired devices", Value: "devices"},
		{Label: fmt.Sprintf("Syncthing UI: http://localhost:%d", a.client.Config().GUIPort), Value: "url"},
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
		flow := pairing.NewFlow(a.client, a.relayClient, a.ui)
		if err := flow.Run(ctx); err != nil {
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
	if !a.client.IsRunning(ctx) {
		return "Syncthing not running"
	}

	progress, err := a.client.GetSyncProgress(ctx)
	if err != nil {
		return "Error"
	}

	if progress.Percent < 100 {
		return fmt.Sprintf("Syncing %d%%", progress.Percent)
	}

	conns, err := a.client.GetConnections(ctx)
	if err != nil {
		return "Error"
	}

	connected := 0
	for _, c := range conns {
		if c.Connected {
			connected++
		}
	}

	if connected == 0 {
		return "Disconnected"
	}

	return fmt.Sprintf("Synced (%d devices)", connected)
}

func (a *App) showDevices(ctx context.Context) error {
	devices, err := a.client.GetConfiguredDevices(ctx)
	if err != nil {
		return err
	}

	conns, err := a.client.GetConnections(ctx)
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		_ = a.ui.Presenter().ShowMessage("Devices", "No paired devices")
		return nil
	}

	var items []ui.MenuItem
	for _, d := range devices {
		status := "offline"
		if conn, ok := conns[d.ID]; ok && conn.Connected {
			status = "connected"
		}

		name := d.Name
		if name == "" {
			name = d.ID[:8] + "..."
		}

		items = append(items, ui.MenuItem{
			Label: fmt.Sprintf("%s (%s)", name, status),
			Value: d.ID,
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
