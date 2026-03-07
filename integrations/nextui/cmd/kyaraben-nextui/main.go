package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fnune/kyaraben/integrations/nextui/internal/app"
	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui/minui"
	"github.com/fnune/kyaraben/internal/syncthing"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	env := app.EnvFromOS()

	cfg, err := config.Load(env.UserdataPath, env.Platform)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	stConfig := syncthing.Config{
		ListenPort:             22100,
		DiscoveryPort:          21127,
		GUIPort:                8484,
		RelayEnabled:           true,
		GlobalDiscoveryEnabled: true,
		BaseURL:                fmt.Sprintf("http://localhost:%d", 8484),
	}
	stClient := syncthing.NewClient(stConfig)

	relayClient, err := syncthing.NewRelayClient(syncthing.ProductionRelayURLs)
	if err != nil {
		return fmt.Errorf("create relay client: %w", err)
	}

	appUI := minui.New(env.PakPath)

	application := app.New(env, cfg, stClient, relayClient, appUI)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	return application.Run(ctx)
}
