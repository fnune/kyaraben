package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/fnune/kyaraben/internal/cli"
	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/versions"
)

var CLI struct {
	Apply          cli.ApplyCmd        `cmd:"" help:"Apply configuration: install emulators and generate configs."`
	Doctor         cli.DoctorCmd       `cmd:"" help:"Check provision status (BIOS files, etc.)."`
	Status         cli.StatusCmd       `cmd:"" help:"Show current state."`
	Init           cli.InitCmd         `cmd:"" help:"Initialize a new kyaraben configuration."`
	Import         cli.ImportCmd       `cmd:"" help:"Analyze existing collection for import."`
	Uninstall      cli.UninstallCmd    `cmd:"" help:"Remove kyaraben-managed files."`
	Update         cli.UpdateCmd       `cmd:"" help:"Check for and install CLI updates (does not update the desktop app)."`
	Daemon         cli.DaemonCmd       `cmd:"" help:"Run in daemon mode for UI communication."`
	Sync           cli.SyncCmd         `cmd:"" help:"Manage sync settings and status."`
	CheckDownloads cli.ValidateURLsCmd `cmd:"" help:"Validate download URLs and show sizes (CI check)."`
	System         cli.SystemCmd       `cmd:"" hidden:"" help:"Internal system utilities."`

	Config   string `short:"c" help:"Path to config file." type:"path"`
	Instance string `short:"i" help:"Instance name for running multiple isolated kyaraben instances (e.g. 'primary', 'secondary')."`
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name("kyaraben"),
		kong.Description("Declarative emulation manager"),
		kong.UsageOnError(),
	)

	_ = logging.InitWithPaths(paths.NewPaths(CLI.Instance))
	defer logging.Close()

	if err := versions.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load version data: %v\n", err)
		os.Exit(1)
	}

	installerFactory := cli.DefaultInstallerFactory
	if useFakeInstaller() {
		installerFactory = cli.FakeInstallerFactory
	}

	cliCtx := cli.NewContext(
		cli.DefaultFS(),
		paths.NewPaths(CLI.Instance),
		cli.DefaultResolver(),
		CLI.Config,
		installerFactory,
	)

	err := ctx.Run(cliCtx)
	if err != nil {
		logging.Error("command failed: %v", err)
		ctx.FatalIfErrorf(err)
	}
}

func useFakeInstaller() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("KYARABEN_E2E_FAKE_INSTALLER")))
	return value == "1" || value == "true" || value == "yes"
}
