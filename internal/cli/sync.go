package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/sync"
)

type SyncCmd struct {
	Status       SyncStatusCmd       `cmd:"" help:"Show sync status."`
	AddDevice    SyncAddDeviceCmd    `cmd:"" help:"Add a device to sync with."`
	RemoveDevice SyncRemoveDeviceCmd `cmd:"" help:"Remove a paired device."`
}

type SyncStatusCmd struct{}

func (cmd *SyncStatusCmd) Run(cliCtx *Context) error {
	cfg, err := cliCtx.LoadConfig()
	if err != nil {
		return err
	}

	if !cfg.Sync.Enabled {
		fmt.Println("Sync: disabled")
		fmt.Println()
		fmt.Println("Enable sync in your config.toml:")
		fmt.Println()
		fmt.Println("  [sync]")
		fmt.Println("  enabled = true")
		fmt.Println("  mode = \"primary\"  # or \"secondary\"")
		return nil
	}

	client := sync.NewClient(cfg.Sync)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !client.IsRunning(ctx) {
		fmt.Printf("Sync: enabled (%s mode)\n", cfg.Sync.Mode)
		fmt.Println("Status: not running")
		fmt.Println()
		fmt.Println("Syncthing is not running. It will start when the kyaraben daemon runs.")
		return nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("getting sync status: %w", err)
	}

	fmt.Printf("Sync: enabled (%s mode)\n", status.Mode)
	fmt.Printf("Status: %s\n", status.OverallState())
	fmt.Printf("Device ID: %s\n", status.DeviceID)
	fmt.Printf("UI: %s\n", status.GUIURL)
	fmt.Println()

	if len(status.Devices) == 0 {
		fmt.Println("Paired devices: none")
		fmt.Println()
		fmt.Println("Add a device with:")
		fmt.Println("  kyaraben sync add-device <DEVICE-ID>")
	} else {
		fmt.Println("Paired devices:")
		for _, dev := range status.Devices {
			state := "disconnected"
			if dev.Connected {
				state = "connected"
			}
			fmt.Printf("  %-30s %s\n", dev.Name, state)
		}
	}

	return nil
}

type SyncAddDeviceCmd struct {
	DeviceID string `arg:"" help:"Device ID to add (from the other device's 'kyaraben sync status')."`
	Name     string `help:"Friendly name for this device." default:""`
}

func (cmd *SyncAddDeviceCmd) Run(cliCtx *Context) error {
	cfg, err := cliCtx.LoadConfig()
	if err != nil {
		return err
	}

	if !cfg.Sync.Enabled {
		return fmt.Errorf("sync is not enabled; enable it in config.toml first")
	}

	deviceID := strings.ToUpper(strings.TrimSpace(cmd.DeviceID))
	if !isValidDeviceID(deviceID) {
		return fmt.Errorf("invalid device ID format")
	}

	for _, existing := range cfg.Sync.Devices {
		if existing.ID == deviceID {
			return fmt.Errorf("device %s is already added", deviceID)
		}
	}

	name := cmd.Name
	if name == "" {
		name = fmt.Sprintf("device-%d", len(cfg.Sync.Devices)+1)
	}

	cfg.Sync.Devices = append(cfg.Sync.Devices, model.SyncDevice{
		ID:   deviceID,
		Name: name,
	})

	configPath, err := cliCtx.GetConfigPath()
	if err != nil {
		return err
	}

	if err := model.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Added device: %s (%s)\n", name, truncateDeviceID(deviceID))
	fmt.Println()
	fmt.Println("The other device also needs to add this device's ID.")
	fmt.Println("Run 'kyaraben sync status' to see this device's ID.")
	fmt.Println()
	fmt.Println("Run 'kyaraben apply' to apply the configuration.")

	return nil
}

type SyncRemoveDeviceCmd struct {
	DeviceID string `arg:"" help:"Device ID or name to remove."`
}

func (cmd *SyncRemoveDeviceCmd) Run(cliCtx *Context) error {
	cfg, err := cliCtx.LoadConfig()
	if err != nil {
		return err
	}

	if !cfg.Sync.Enabled {
		return fmt.Errorf("sync is not enabled")
	}

	query := strings.ToUpper(strings.TrimSpace(cmd.DeviceID))
	found := -1
	for i, dev := range cfg.Sync.Devices {
		if strings.ToUpper(dev.ID) == query || strings.EqualFold(dev.Name, cmd.DeviceID) {
			found = i
			break
		}
	}

	if found == -1 {
		return fmt.Errorf("device not found: %s", cmd.DeviceID)
	}

	removed := cfg.Sync.Devices[found]
	cfg.Sync.Devices = append(cfg.Sync.Devices[:found], cfg.Sync.Devices[found+1:]...)

	configPath, err := cliCtx.GetConfigPath()
	if err != nil {
		return err
	}

	if err := model.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Removed device: %s (%s)\n", removed.Name, truncateDeviceID(removed.ID))
	fmt.Println()
	fmt.Println("Run 'kyaraben apply' to apply the configuration.")

	return nil
}

func isValidDeviceID(id string) bool {
	id = strings.ReplaceAll(id, "-", "")
	if len(id) != 56 {
		return false
	}
	for _, c := range id {
		if !((c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7')) {
			return false
		}
	}
	return true
}

func truncateDeviceID(id string) string {
	if len(id) > 15 {
		return id[:7] + "..." + id[len(id)-7:]
	}
	return id
}
