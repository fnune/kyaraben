package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fnune/kyaraben/integrations/batocera/internal/config"
	"github.com/fnune/kyaraben/integrations/batocera/internal/mapping"
	"github.com/fnune/kyaraben/integrations/batocera/internal/service"
	"github.com/fnune/kyaraben/internal/syncguest"
	"github.com/fnune/kyaraben/internal/syncthing"
)

const (
	basePath       = "/userdata"
	dataDir        = "/userdata/system/configs"
	kyarabenData   = "/userdata/system/kyaraben"
	defaultGUIPort = 8384
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := run(ctx, os.Args[1], os.Args[2:]); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `kyaraben-batocera - Kyaraben integration for Batocera/KNULLI

Commands:
  start       Start syncing and configure folders
  stop        Stop syncing
  status      Show sync status
  enable      Enable autostart
  disable     Disable autostart
  pair        Start pairing flow`)
}

func run(ctx context.Context, cmd string, args []string) error {
	cfgStore := config.NewConfigStore(kyarabenData)
	cfg, err := cfgStore.Load(config.DefaultConfig())
	if err != nil {
		log.Printf("load config: %v (using defaults)", err)
		defaultCfg := config.DefaultConfig()
		cfg = &defaultCfg
	}

	stConfig := syncthing.DefaultConfig()
	stConfig.GUIPort = defaultGUIPort
	stConfig.BaseURL = fmt.Sprintf("http://localhost:%d", defaultGUIPort)
	client := syncthing.NewClient(stConfig)

	syncConfig := syncguest.DefaultConfig(dataDir)
	syncConfig.Syncthing.GUIPort = defaultGUIPort
	syncMgr := syncguest.NewWithClient(syncConfig, client)

	svcMgr := service.NewManager(client)
	mapper := mapping.NewMapper(basePath, *cfg)

	switch cmd {
	case "start":
		return cmdStart(ctx, svcMgr, syncMgr, mapper)
	case "stop":
		return cmdStop(svcMgr)
	case "status":
		return cmdStatus(ctx, svcMgr, syncMgr)
	case "enable":
		return cmdEnable(svcMgr)
	case "disable":
		return cmdDisable(svcMgr)
	case "pair":
		return cmdPair(ctx, svcMgr, syncMgr)
	default:
		usage()
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func cmdStart(ctx context.Context, svcMgr *service.Manager, syncMgr *syncguest.Manager, mapper *mapping.Mapper) error {
	if !svcMgr.IsRunning(ctx) {
		fmt.Println("Starting Syncthing...")
		if err := svcMgr.Start(ctx); err != nil {
			return fmt.Errorf("start syncthing: %w", err)
		}
	}

	fmt.Println("Configuring folders...")
	if err := syncMgr.ConfigureFolders(mapper.SyncguestFolderMappings()); err != nil {
		return fmt.Errorf("configure folders: %w", err)
	}

	fmt.Println("Syncthing is running")
	return nil
}

func cmdStop(svcMgr *service.Manager) error {
	fmt.Println("Stopping Syncthing...")
	if err := svcMgr.Stop(); err != nil {
		return fmt.Errorf("stop syncthing: %w", err)
	}
	fmt.Println("Syncthing stopped")
	return nil
}

func cmdStatus(ctx context.Context, svcMgr *service.Manager, syncMgr *syncguest.Manager) error {
	if !svcMgr.IsRunning(ctx) {
		fmt.Println("Status: Not running")
		return nil
	}

	status, err := syncMgr.GetStatus(ctx)
	if err != nil {
		fmt.Println("Status: Running (unable to get details)")
		return nil
	}

	if status.Syncing {
		fmt.Printf("Status: Syncing %d%%\n", status.Progress)
	} else if len(status.Peers) == 0 {
		fmt.Println("Status: Idle (no paired devices)")
	} else if status.ConnectedPeers == 0 {
		fmt.Println("Status: Idle (no devices online)")
	} else {
		fmt.Printf("Status: Synced (%d device(s) connected)\n", status.ConnectedPeers)
	}

	if svcMgr.IsAutostartEnabled() {
		fmt.Println("Autostart: Enabled")
	} else {
		fmt.Println("Autostart: Disabled")
	}

	return nil
}

func cmdEnable(svcMgr *service.Manager) error {
	if err := svcMgr.EnableAutostart(); err != nil {
		return fmt.Errorf("enable autostart: %w", err)
	}
	fmt.Println("Autostart enabled")
	return nil
}

func cmdDisable(svcMgr *service.Manager) error {
	if err := svcMgr.DisableAutostart(); err != nil {
		return fmt.Errorf("disable autostart: %w", err)
	}
	fmt.Println("Autostart disabled")
	return nil
}

func cmdPair(ctx context.Context, svcMgr *service.Manager, syncMgr *syncguest.Manager) error {
	if !svcMgr.IsRunning(ctx) {
		return fmt.Errorf("syncthing is not running - run 'kyaraben-batocera start' first")
	}

	fmt.Println("Generating pairing code...")
	session, err := syncMgr.CreatePairingSession(ctx)
	if err != nil {
		return syncguest.FriendlyPairingError(err)
	}

	fmt.Printf("\nPairing code: %s\n\n", session.Code)
	fmt.Println("Enter this code on the other device, or wait for them to enter theirs.")
	fmt.Println("Press Ctrl+C to cancel.")

	peerID, err := syncMgr.WaitForPeer(ctx, session.Code)
	if err != nil {
		return syncguest.FriendlyPairingError(err)
	}

	if err := syncMgr.AddPeer(ctx, peerID); err != nil {
		return syncguest.FriendlyPairingError(err)
	}

	fmt.Println("Device paired successfully!")
	return nil
}
