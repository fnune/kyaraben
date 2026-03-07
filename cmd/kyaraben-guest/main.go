package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fnune/kyaraben/internal/syncguest"
)

var dataDir string

func main() {
	flag.StringVar(&dataDir, "data-dir", "", "Data directory for config and state")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	var err error
	switch args[0] {
	case "status":
		err = runStatus()
	case "pair":
		err = runPair()
	case "join":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: kyaraben-guest join <code>")
			os.Exit(1)
		}
		err = runJoin(args[1])
	case "start":
		err = runStart()
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`kyaraben-guest - Join a Kyaraben sync cluster as a guest device

Usage:
  kyaraben-guest [flags] <command>

Commands:
  status    Show sync status
  pair      Generate a pairing code for another device to join
  join      Join a sync cluster using a pairing code
  start     Start syncthing and keep running
  help      Show this help

Flags:
  --data-dir    Data directory for config and state`)
}

func runStatus() error {
	mgr := newManager()
	ctx := context.Background()

	status, err := mgr.GetStatus(ctx)
	if err != nil {
		return err
	}

	if !status.Running {
		fmt.Println("Syncthing is not running")
		return nil
	}

	fmt.Printf("Device ID: %s\n", status.DeviceID)
	fmt.Printf("Status: ")
	if status.Syncing {
		fmt.Printf("Syncing (%d%%)\n", status.Progress)
	} else {
		fmt.Println("Synced")
	}
	fmt.Printf("Connected peers: %d/%d\n", status.ConnectedPeers, len(status.Peers))

	if len(status.Peers) > 0 {
		fmt.Println("\nPeers:")
		for _, p := range status.Peers {
			state := "offline"
			if p.Connected {
				state = "connected"
			}
			name := p.Name
			if name == "" && len(p.ID) > 12 {
				name = p.ID[:12] + "..."
			}
			fmt.Printf("  %s (%s)\n", name, state)
		}
	}

	return nil
}

func runPair() error {
	mgr := newManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("start syncthing: %w", err)
	}
	defer func() { _ = mgr.Stop() }()

	session, err := mgr.CreatePairingSession(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Pairing code: %s\n", session.Code)
	fmt.Printf("Expires in: %s\n", session.ExpiresIn)
	fmt.Println("\nWaiting for peer to join...")

	waitCtx, waitCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer waitCancel()

	peerID, err := mgr.WaitForPeer(waitCtx, session.Code)
	if err != nil {
		return fmt.Errorf("waiting for peer: %w", err)
	}

	if err := mgr.AddPeer(ctx, peerID); err != nil {
		return fmt.Errorf("add peer: %w", err)
	}

	shortID := peerID
	if len(peerID) > 12 {
		shortID = peerID[:12] + "..."
	}
	fmt.Printf("Paired with device: %s\n", shortID)
	return nil
}

func runJoin(code string) error {
	mgr := newManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("start syncthing: %w", err)
	}
	defer func() { _ = mgr.Stop() }()

	peerID, err := mgr.JoinPairingSession(ctx, code)
	if err != nil {
		return err
	}

	if err := mgr.AddPeer(ctx, peerID); err != nil {
		return fmt.Errorf("add peer: %w", err)
	}

	shortID := peerID
	if len(peerID) > 12 {
		shortID = peerID[:12] + "..."
	}
	fmt.Printf("Joined cluster, paired with: %s\n", shortID)
	return nil
}

func runStart() error {
	mgr := newManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("start syncthing: %w", err)
	}
	defer func() { _ = mgr.Stop() }()

	fmt.Println("Syncthing started. Press Ctrl+C to stop.")

	<-ctx.Done()
	fmt.Println("\nStopping...")
	return nil
}

func newManager() *syncguest.Manager {
	dir := dataDir
	if dir == "" {
		dir = os.Getenv("KYARABEN_DATA_DIR")
	}
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = home + "/.kyaraben-guest"
	}

	config := syncguest.DefaultConfig(dir)
	return syncguest.New(config)
}

func setupSignalHandler(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()
}
