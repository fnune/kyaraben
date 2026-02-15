package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/sync"
)

type SyncCmd struct {
	Status       SyncStatusCmd       `cmd:"" help:"Show sync status."`
	Pair         SyncPairCmd         `cmd:"" help:"Pair with another device on the local network."`
	AddDevice    SyncAddDeviceCmd    `cmd:"" help:"Add a device by ID (for cross-network pairing)."`
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

	client := sync.NewClient(cfg.Sync)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !client.IsRunning(ctx) {
		return fmt.Errorf("syncthing is not running; run 'kyaraben apply' first")
	}

	devices, err := client.GetConfiguredDevices(ctx)
	if err != nil {
		return fmt.Errorf("getting configured devices: %w", err)
	}

	for _, existing := range devices {
		if strings.ToUpper(existing.ID) == deviceID {
			return fmt.Errorf("device %s is already added", deviceID)
		}
	}

	name := cmd.Name
	if name == "" {
		name = fmt.Sprintf("device-%d", len(devices)+1)
	}

	if err := client.AddDevice(ctx, deviceID, name); err != nil {
		return fmt.Errorf("adding device: %w", err)
	}

	if err := client.ShareFoldersWithDevice(ctx, deviceID); err != nil {
		return fmt.Errorf("sharing folders: %w", err)
	}

	fmt.Printf("Added device: %s (%s)\n", name, truncateDeviceID(deviceID))
	fmt.Println()
	fmt.Println("The other device also needs to add this device's ID.")
	fmt.Println("Run 'kyaraben sync status' to see this device's ID.")

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

	client := sync.NewClient(cfg.Sync)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !client.IsRunning(ctx) {
		return fmt.Errorf("syncthing is not running; run 'kyaraben apply' first")
	}

	query := strings.ToUpper(strings.TrimSpace(cmd.DeviceID))
	devices, err := client.GetConfiguredDevices(ctx)
	if err != nil {
		return fmt.Errorf("getting configured devices: %w", err)
	}

	var found *sync.ConfiguredDevice
	for _, dev := range devices {
		if strings.ToUpper(dev.ID) == query || strings.EqualFold(dev.Name, cmd.DeviceID) {
			found = &dev
			break
		}
	}

	if found == nil {
		return fmt.Errorf("device not found: %s", cmd.DeviceID)
	}

	if err := client.RemoveDevice(ctx, found.ID); err != nil {
		return fmt.Errorf("removing device: %w", err)
	}

	fmt.Printf("Removed device: %s (%s)\n", found.Name, truncateDeviceID(found.ID))

	return nil
}

func isValidDeviceID(id string) bool {
	id = strings.ReplaceAll(id, "-", "")
	if len(id) != 56 {
		return false
	}
	for _, c := range id {
		isUpperAlpha := c >= 'A' && c <= 'Z'
		isBase32Digit := c >= '2' && c <= '7'
		if !isUpperAlpha && !isBase32Digit {
			return false
		}
	}
	return true
}

type SyncPairCmd struct {
	PrimaryDeviceID string `arg:"" optional:"" help:"Device ID from the primary device. Omit to start as primary."`
}

func (cmd *SyncPairCmd) Run(cliCtx *Context) error {
	cfg, err := cliCtx.LoadConfig()
	if err != nil {
		return err
	}

	configPath, err := cliCtx.GetConfigPath()
	if err != nil {
		return err
	}

	client := sync.NewClient(cfg.Sync)
	ctx := context.Background()

	if !client.IsRunning(ctx) {
		return fmt.Errorf("syncthing is not running; run 'kyaraben apply' first")
	}

	progress := func(msg string) {
		fmt.Println(msg)
	}

	if cmd.PrimaryDeviceID == "" {
		return cmd.runPrimary(ctx, cfg, configPath, client, progress, cliCtx.SaveConfig)
	}
	return cmd.runSecondary(ctx, cfg, configPath, client, strings.ToUpper(strings.TrimSpace(cmd.PrimaryDeviceID)), progress, cliCtx.SaveConfig)
}

func (cmd *SyncPairCmd) runPrimary(ctx context.Context, cfg *model.KyarabenConfig, configPath string, client *sync.Client, progress func(string), saveConfig func(*model.KyarabenConfig, string) error) error {
	flow := sync.NewPrimaryPairingFlow(sync.PairingFlowConfig{
		SyncConfig: cfg.Sync,
		Client:     client,
		OnProgress: progress,
	})

	result, err := flow.Run(ctx)
	if err != nil {
		return fmt.Errorf("pairing: %w", err)
	}

	persistSyncEnabled(cfg, configPath, model.SyncModePrimary, saveConfig)
	fmt.Printf("Paired with %s (%s)\n", result.PeerName, truncateDeviceID(result.PeerDeviceID))
	return nil
}

func (cmd *SyncPairCmd) runSecondary(ctx context.Context, cfg *model.KyarabenConfig, configPath string, client *sync.Client, primaryDeviceID string, progress func(string), saveConfig func(*model.KyarabenConfig, string) error) error {
	flow := sync.NewSecondaryPairingFlow(sync.PairingFlowConfig{
		SyncConfig: cfg.Sync,
		Client:     client,
		OnProgress: progress,
	})

	result, err := flow.Run(ctx, primaryDeviceID)
	if err != nil {
		return fmt.Errorf("pairing: %w", err)
	}

	persistSyncEnabled(cfg, configPath, model.SyncModeSecondary, saveConfig)
	fmt.Printf("Paired with %s (%s)\n", result.PeerName, truncateDeviceID(result.PeerDeviceID))
	return nil
}

func persistSyncEnabled(cfg *model.KyarabenConfig, configPath string, mode model.SyncMode, saveConfig func(*model.KyarabenConfig, string) error) {
	cfg.Sync.Enabled = true
	cfg.Sync.Mode = mode

	if err := saveConfig(cfg, configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save config: %v\n", err)
	}
}

func truncateDeviceID(id string) string {
	if len(id) > 15 {
		return id[:7] + "..." + id[len(id)-7:]
	}
	return id
}
