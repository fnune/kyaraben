package main

import (
	"github.com/alecthomas/kong"

	"github.com/fnune/kyaraben/internal/cli"
	"github.com/fnune/kyaraben/internal/logging"
)

var CLI struct {
	Apply     cli.ApplyCmd     `cmd:"" help:"Apply configuration: install emulators and generate configs."`
	Doctor    cli.DoctorCmd    `cmd:"" help:"Check provision status (BIOS files, etc.)."`
	Status    cli.StatusCmd    `cmd:"" help:"Show current state."`
	Init      cli.InitCmd      `cmd:"" help:"Initialize a new kyaraben configuration."`
	Uninstall cli.UninstallCmd `cmd:"" help:"Remove kyaraben-managed files."`
	Daemon    cli.DaemonCmd    `cmd:"" help:"Run in daemon mode for UI communication."`
	Sync      cli.SyncCmd      `cmd:"" help:"Manage sync settings and status."`

	Config string `short:"c" help:"Path to config file." type:"path"`
}

func main() {
	_ = logging.Init()
	defer logging.Close()

	ctx := kong.Parse(&CLI,
		kong.Name("kyaraben"),
		kong.Description("Declarative emulation manager"),
		kong.UsageOnError(),
	)

	err := ctx.Run(&cli.Context{
		ConfigPath: CLI.Config,
	})
	if err != nil {
		logging.Error("command failed: %v", err)
		ctx.FatalIfErrorf(err)
	}
}
