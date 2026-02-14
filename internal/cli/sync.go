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
	Pause        SyncPauseCmd        `cmd:"" help:"Pause sync."`
	Resume       SyncResumeCmd       `cmd:"" help:"Resume sync."`
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

	if err := cliCtx.SaveConfig(cfg, configPath); err != nil {
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

	if err := cliCtx.SaveConfig(cfg, configPath); err != nil {
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
		isUpperAlpha := c >= 'A' && c <= 'Z'
		isBase32Digit := c >= '2' && c <= '7'
		if !isUpperAlpha && !isBase32Digit {
			return false
		}
	}
	return true
}

type SyncPairCmd struct {
	Code string `arg:"" optional:"" help:"Pairing code from the primary device. Omit to start as primary."`
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

	if cmd.Code == "" {
		return cmd.runPrimary(ctx, cfg, configPath, client, progress, cliCtx.SaveConfig)
	}
	return cmd.runSecondary(ctx, cfg, configPath, client, strings.ToUpper(strings.TrimSpace(cmd.Code)), progress, cliCtx.SaveConfig)
}

func (cmd *SyncPairCmd) runPrimary(ctx context.Context, cfg *model.KyarabenConfig, configPath string, client *sync.Client, progress func(string), saveConfig func(*model.KyarabenConfig, string) error) error {
	flow := sync.NewPrimaryPairingFlow(sync.PairingFlowConfig{
		SyncConfig: cfg.Sync,
		Advertiser: sync.NewMDNSAdvertiser(),
		Client:     client,
		OnProgress: progress,
	})

	result, _, err := flow.Run(ctx)
	if err != nil {
		return fmt.Errorf("pairing: %w", err)
	}

	persistPairedDevice(cfg, configPath, result.PeerDeviceID, result.PeerName, model.SyncModePrimary, saveConfig)
	fmt.Printf("Paired with %s (%s)\n", result.PeerName, truncateDeviceID(result.PeerDeviceID))
	return nil
}

func (cmd *SyncPairCmd) runSecondary(ctx context.Context, cfg *model.KyarabenConfig, configPath string, client *sync.Client, code string, progress func(string), saveConfig func(*model.KyarabenConfig, string) error) error {
	flow := sync.NewSecondaryPairingFlow(sync.PairingFlowConfig{
		SyncConfig: cfg.Sync,
		Browser:    sync.NewMDNSBrowser(),
		Client:     client,
		OnProgress: progress,
	})

	result, err := flow.Run(ctx, code)
	if err != nil {
		return fmt.Errorf("pairing: %w", err)
	}

	persistPairedDevice(cfg, configPath, result.PeerDeviceID, result.PeerName, model.SyncModeSecondary, saveConfig)
	fmt.Printf("Paired with %s (%s)\n", result.PeerName, truncateDeviceID(result.PeerDeviceID))
	return nil
}

func persistPairedDevice(cfg *model.KyarabenConfig, configPath, peerDeviceID, peerName string, mode model.SyncMode, saveConfig func(*model.KyarabenConfig, string) error) {
	for _, dev := range cfg.Sync.Devices {
		if dev.ID == peerDeviceID {
			return
		}
	}

	cfg.Sync.Devices = append(cfg.Sync.Devices, model.SyncDevice{
		ID:   peerDeviceID,
		Name: peerName,
	})
	cfg.Sync.Enabled = true
	cfg.Sync.Mode = mode

	if err := saveConfig(cfg, configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save config: %v\n", err)
	}
}

type SyncPauseCmd struct{}

func (cmd *SyncPauseCmd) Run(cliCtx *Context) error {
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
		return fmt.Errorf("syncthing is not running")
	}

	if err := client.PauseSync(ctx); err != nil {
		return fmt.Errorf("pausing sync: %w", err)
	}

	fmt.Println("Sync paused.")
	return nil
}

type SyncResumeCmd struct{}

func (cmd *SyncResumeCmd) Run(cliCtx *Context) error {
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
		return fmt.Errorf("syncthing is not running")
	}

	if err := client.ResumeSync(ctx); err != nil {
		return fmt.Errorf("resuming sync: %w", err)
	}

	fmt.Println("Sync resumed.")
	return nil
}

func truncateDeviceID(id string) string {
	if len(id) > 15 {
		return id[:7] + "..." + id[len(id)-7:]
	}
	return id
}
