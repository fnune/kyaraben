package cli

import (
	"github.com/fnune/kyaraben/internal/frontends/esde"
	"github.com/fnune/kyaraben/internal/logging"
)

type SystemCmd struct {
	ESDEImport SystemESDEImportCmd `cmd:"" name:"esde-import" help:"Import ES-DE gamelists from sync directory."`
	ESDEExport SystemESDEExportCmd `cmd:"" name:"esde-export" help:"Export ES-DE gamelists to sync directory."`
}

type SystemESDEImportCmd struct{}

func (cmd *SystemESDEImportCmd) Run(ctx *Context) error {
	log := logging.New("system")

	cfg, err := ctx.LoadConfig()
	if err != nil {
		log.Warn("failed to load config: %v", err)
		return nil
	}

	userStore, err := ctx.NewUserStore(cfg)
	if err != nil {
		log.Warn("failed to create user store: %v", err)
		return nil
	}

	sync, err := esde.NewDefaultGamelistSync(userStore)
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

	userStore, err := ctx.NewUserStore(cfg)
	if err != nil {
		log.Warn("failed to create user store: %v", err)
		return nil
	}

	sync, err := esde.NewDefaultGamelistSync(userStore)
	if err != nil {
		log.Warn("failed to create gamelist sync: %v", err)
		return nil
	}

	systems := cfg.EnabledSystems()
	log.Info("exporting gamelists for %d systems", len(systems))
	sync.ExportAll(systems)

	return nil
}
