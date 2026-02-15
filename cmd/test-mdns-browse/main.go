package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fnune/kyaraben/internal/sync"
)

func init() {
	log.SetOutput(io.Discard)
}

func timestamp() string {
	return time.Now().Format("15:04:05")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Printf("[%s] Stopping...\n", timestamp())
		cancel()
	}()

	fmt.Printf("[%s] Watching for kyaraben devices... (Ctrl+C to stop)\n", timestamp())

	current := make(map[string]sync.PairingOffer)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			queryCtx, queryCancel := context.WithTimeout(ctx, time.Second)
			browser := sync.NewMDNSBrowser()
			offers, err := browser.Browse(queryCtx)
			if err != nil {
				queryCancel()
				continue
			}

			found := make(map[string]sync.PairingOffer)
			for offer := range offers {
				key := offer.Hostname + offer.PairingAddr
				found[key] = offer
				if _, exists := current[key]; !exists {
					fmt.Printf("[%s] FOUND: %s at %s\n", timestamp(), offer.Hostname, offer.PairingAddr)
				}
			}

			for key, offer := range current {
				if _, exists := found[key]; !exists {
					fmt.Printf("[%s] LOST: %s at %s\n", timestamp(), offer.Hostname, offer.PairingAddr)
				}
			}

			current = found
			queryCancel()
		}
	}
}
