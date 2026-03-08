package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/sync"
)

type SyncCmd struct {
	Status       SyncStatusCmd       `cmd:"" help:"Show sync status."`
	Pair         SyncPairCmd         `cmd:"" help:"Pair with another device."`
	AddDevice    SyncAddDeviceCmd    `cmd:"" help:"Add a device by ID (for manual pairing)."`
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
		return nil
	}

	client := sync.NewClient(cfg.Sync)
	if stateDir, err := cliCtx.GetPaths().StateDir(); err == nil {
		if apiKey := loadSyncAPIKey(stateDir); apiKey != "" {
			client.SetAPIKey(apiKey)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !client.IsRunning(ctx) {
		fmt.Println("Sync: enabled")
		fmt.Println("Status: not running")
		fmt.Println()
		fmt.Println("Syncthing is not running. It will start when the kyaraben daemon runs.")
		return nil
	}

	status, err := client.GetStatus(ctx, vfs.OSFS)
	if err != nil {
		return fmt.Errorf("getting sync status: %w", err)
	}

	fmt.Println("Sync: enabled")
	fmt.Printf("Status: %s\n", status.OverallState())
	fmt.Printf("Device ID: %s\n", status.DeviceID)
	fmt.Printf("UI: %s\n", status.GUIURL)
	fmt.Println()

	if status.LocalConnectivityIssue != "" {
		fmt.Println()
		switch status.LocalConnectivityIssue {
		case "listen_error":
			fmt.Println("! Warning: Failed to listen on sync port")
		case "no_lan_address":
			fmt.Println("! Warning: Other devices may not be able to connect")
		default:
			fmt.Printf("! Warning: %s\n", status.LocalConnectivityIssue)
		}
		fmt.Println("  Check your firewall and network settings.")
		fmt.Println("  See: https://docs.syncthing.net/users/firewall.html")
	}

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
				if dev.ConnectionType != "" {
					state = fmt.Sprintf("connected (%s)", dev.ConnectionType)
				}
			}
			fmt.Printf("  %-30s %s\n", dev.Name, state)
			if dev.ConnectivityIssue == "port_unreachable" {
				fmt.Printf("    ! Port unreachable - check firewall on peer device\n")
				fmt.Printf("      See: https://docs.syncthing.net/users/firewall.html\n")
			}
		}
	}

	fmt.Println()
	if len(status.Folders) == 0 {
		fmt.Println("Synced folders: none")
	} else {
		fmt.Println("Synced folders:")
		for _, f := range status.Folders {
			if f.State == "error" && f.Error != "" {
				fmt.Printf("  %-20s %s: %s\n", f.Label, f.State, f.Error)
			} else {
				fmt.Printf("  %-20s %s\n", f.Label, f.State)
			}
			if f.ConflictCount > 0 {
				fmt.Printf("    %d conflict file(s)\n", f.ConflictCount)
			}
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
	if stateDir, err := cliCtx.GetPaths().StateDir(); err == nil {
		if apiKey := loadSyncAPIKey(stateDir); apiKey != "" {
			client.SetAPIKey(apiKey)
		}
	}
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
	if stateDir, err := cliCtx.GetPaths().StateDir(); err == nil {
		if apiKey := loadSyncAPIKey(stateDir); apiKey != "" {
			client.SetAPIKey(apiKey)
		}
	}
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
	CodeOrDeviceID string `arg:"" optional:"" help:"6-character pairing code or full device ID. Omit to start as initiator."`
	DeviceID       bool   `help:"Show full device ID instead of pairing code (for manual pairing)."`
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
	if stateDir, err := cliCtx.GetPaths().StateDir(); err == nil {
		if apiKey := loadSyncAPIKey(stateDir); apiKey != "" {
			client.SetAPIKey(apiKey)
		}
	}
	ctx := context.Background()

	if !client.IsRunning(ctx) {
		return fmt.Errorf("syncthing is not running; run 'kyaraben apply' first")
	}

	progress := func(msg string) {
		fmt.Println(msg)
	}

	if cmd.CodeOrDeviceID == "" {
		return cmd.runInitiator(ctx, cfg, configPath, client, progress, cliCtx.SaveConfig)
	}

	input := strings.ToUpper(strings.TrimSpace(cmd.CodeOrDeviceID))
	if sync.IsRelayCode(input) {
		return cmd.runJoinerWithRelay(ctx, cfg, configPath, client, input, progress, cliCtx.SaveConfig)
	}
	return cmd.runJoiner(ctx, cfg, configPath, client, input, progress, cliCtx.SaveConfig)
}

func (cmd *SyncPairCmd) runInitiator(ctx context.Context, cfg *model.KyarabenConfig, configPath string, client *sync.Client, progress func(string), saveConfig func(*model.KyarabenConfig, string) error) error {
	if cmd.DeviceID {
		localID, err := client.GetDeviceID(ctx)
		if err != nil {
			return fmt.Errorf("getting device ID: %w", err)
		}
		fmt.Printf("Device ID: %s\n", localID)
		fmt.Println()
		fmt.Println("On the other device, run:")
		fmt.Printf("  kyaraben sync pair %s\n", localID)
		return nil
	}

	flow := sync.NewRelayInitiatorPairingFlow(sync.RelayPairingFlowConfig{
		SyncConfig: cfg.Sync,
		Client:     client,
		RelayURLs:  cfg.Sync.Relays,
		OnProgress: progress,
		OnCode: func(code string, expiresIn int) {
			fmt.Printf("Pairing code: %s\n", code)
			fmt.Printf("Waiting for devices... (expires in %d minutes)\n", expiresIn/60)
			fmt.Println()
		},
	})

	result, err := flow.Run(ctx)
	if err != nil {
		return fmt.Errorf("pairing: %w", err)
	}

	if result.Code == "" {
		fmt.Printf("Device ID: %s\n", result.DeviceID)
		fmt.Println()
		fmt.Println("Relay unavailable. On the other device, run:")
		fmt.Printf("  kyaraben sync pair %s\n", result.DeviceID)
		return nil
	}

	persistSyncEnabled(cfg, configPath, saveConfig)
	fmt.Printf("Paired with %s (%s)\n", result.PeerName, truncateDeviceID(result.PeerDeviceID))
	return nil
}

func (cmd *SyncPairCmd) runJoinerWithRelay(ctx context.Context, cfg *model.KyarabenConfig, configPath string, client *sync.Client, code string, progress func(string), saveConfig func(*model.KyarabenConfig, string) error) error {
	flow := sync.NewRelayJoinerPairingFlow(sync.RelayPairingFlowConfig{
		SyncConfig: cfg.Sync,
		Client:     client,
		RelayURLs:  cfg.Sync.Relays,
		OnProgress: progress,
	})

	result, err := flow.Run(ctx, code)
	if err != nil {
		return fmt.Errorf("pairing: %w", err)
	}

	persistSyncEnabled(cfg, configPath, saveConfig)
	fmt.Printf("Paired with %s (%s)\n", result.PeerName, truncateDeviceID(result.PeerDeviceID))
	return nil
}

func (cmd *SyncPairCmd) runJoiner(ctx context.Context, cfg *model.KyarabenConfig, configPath string, client *sync.Client, peerDeviceID string, progress func(string), saveConfig func(*model.KyarabenConfig, string) error) error {
	flow := sync.NewJoinerPairingFlow(sync.PairingFlowConfig{
		SyncConfig: cfg.Sync,
		Client:     client,
		OnProgress: progress,
	})

	result, err := flow.Run(ctx, peerDeviceID)
	if err != nil {
		return fmt.Errorf("pairing: %w", err)
	}

	persistSyncEnabled(cfg, configPath, saveConfig)
	fmt.Printf("Paired with %s (%s)\n", result.PeerName, truncateDeviceID(result.PeerDeviceID))
	return nil
}

func persistSyncEnabled(cfg *model.KyarabenConfig, configPath string, saveConfig func(*model.KyarabenConfig, string) error) {
	cfg.Sync.Enabled = true

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

func loadSyncAPIKey(stateDir string) string {
	keyPath := filepath.Join(stateDir, "syncthing", "config", ".apikey")
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
