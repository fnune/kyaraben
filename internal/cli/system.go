package cli

import (
	"fmt"
	"os"

	"github.com/fnune/kyaraben/internal/frontends/esde"
	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/sdlmappings"
)

type SystemCmd struct {
	ESDEImport  SystemESDEImportCmd  `cmd:"" name:"esde-import" help:"Import ES-DE gamelists from sync directory."`
	ESDEExport  SystemESDEExportCmd  `cmd:"" name:"esde-export" help:"Export ES-DE gamelists to sync directory."`
	SDLMappings SystemSDLMappingsCmd `cmd:"" name:"sdl-mappings" help:"Print SDL controller mappings for RetroArch's wrapper."`
}

type SystemESDEImportCmd struct{}

func (cmd *SystemESDEImportCmd) Run(ctx *Context) error {
	log := logging.New("system")

	cfg, err := ctx.LoadConfig()
	if err != nil {
		log.Warn("failed to load config: %v", err)
		return nil
	}

	collection, err := ctx.NewCollection(cfg)
	if err != nil {
		log.Warn("failed to create user store: %v", err)
		return nil
	}

	sync, err := esde.NewDefaultGamelistSync(collection)
	if err != nil {
		log.Warn("failed to create gamelist sync: %v", err)
		return nil
	}

	systems := cfg.EnabledSystems()
	log.Info("importing gamelists for %d systems", len(systems))
	sync.ImportAll(systems)

	return nil
}

type SystemESDEExportCmd struct{}

func (cmd *SystemESDEExportCmd) Run(ctx *Context) error {
	log := logging.New("system")

	cfg, err := ctx.LoadConfig()
	if err != nil {
		log.Warn("failed to load config: %v", err)
		return nil
	}

	collection, err := ctx.NewCollection(cfg)
	if err != nil {
		log.Warn("failed to create user store: %v", err)
		return nil
	}

	sync, err := esde.NewDefaultGamelistSync(collection)
	if err != nil {
		log.Warn("failed to create gamelist sync: %v", err)
		return nil
	}

	systems := cfg.EnabledSystems()
	log.Info("exporting gamelists for %d systems", len(systems))
	sync.ExportAll(systems)

	return nil
}

type SystemSDLMappingsCmd struct{}

func (cmd *SystemSDLMappingsCmd) Run(_ *Context) error {
	devices, err := os.Open("/proc/bus/input/devices")
	if err != nil {
		return nil
	}
	defer func() { _ = devices.Close() }()
	mappings, err := sdlmappings.ForDevices(devices, sdlmappings.Budget)
	if err != nil {
		return nil
	}
	fmt.Print(mappings)
	return nil
}
