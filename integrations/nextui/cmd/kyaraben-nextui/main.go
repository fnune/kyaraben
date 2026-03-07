package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fnune/kyaraben/integrations/nextui/internal/app"
	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui/minui"
	"github.com/fnune/kyaraben/internal/syncguest"
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

	dataDir := filepath.Join(env.UserdataPath, env.Platform, "kyaraben")
	syncConfig := syncguest.DefaultConfig(dataDir)
	syncConfig.SyncthingPath = filepath.Join(env.PakPath, "syncthing")
	mgr := syncguest.New(syncConfig)

	appUI := minui.New(env.PakPath)

	application := app.New(env, cfg, mgr, appUI)

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
