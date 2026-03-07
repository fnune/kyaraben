package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fnune/kyaraben/integrations/nextui/internal/app"
	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/integrations/nextui/internal/service"
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui/minui"
	"github.com/fnune/kyaraben/internal/syncguest"
)

const guiPort = 8484

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	env := app.EnvFromOS()
	dataDir := filepath.Join(env.UserdataPath, "kyaraben")

	cfg, err := config.Load(dataDir)
	if err != nil {
		log.Printf("load config: %v (using defaults)", err)
		defaultCfg := config.DefaultConfig()
		cfg = &defaultCfg
	}

	syncConfig := syncguest.DefaultConfig(dataDir)
	syncConfig.SyncthingPath = filepath.Join(env.PakPath, "syncthing")
	syncMgr := syncguest.New(syncConfig)

	svcMgr := service.NewManager(service.Config{
		DataDir:       dataDir,
		PakPath:       env.PakPath,
		UserdataPath:  env.UserdataPath,
		Platform:      env.Platform,
		LogsPath:      env.LogsPath,
		SyncthingPath: filepath.Join(env.PakPath, "syncthing"),
		GUIPort:       guiPort,
	})

	appUI := minui.New(env.PakPath)

	application := app.New(env, cfg, dataDir, syncMgr, svcMgr, appUI)

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
