package cli

import (
	"fmt"
	"os"

	"github.com/fnune/kyaraben/internal/model"
)

// InitCmd initializes a new kyaraben configuration.
type InitCmd struct {
	Collection string `short:"d" help:"Path to collection." default:"~/Emulation"`
	Force      bool   `short:"f" help:"Overwrite existing configuration."`
}

// Run executes the init command.
func (cmd *InitCmd) Run(ctx *Context) error {
	configPath, err := ctx.GetConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err == nil && !cmd.Force {
		return fmt.Errorf("config already exists at %s. Use --force to overwrite", configPath)
	}

	cfg := model.NewDefaultConfig()
	cfg.Global.Collection = cmd.Collection

	if ctx.Paths.Instance != "" {
		if cmd.Collection == "~/Emulation" {
			cfg.Global.Collection = "~/Emulation-" + ctx.Paths.Instance
		}
		offset := ctx.Paths.InstancePortOffset()
		cfg.Sync.Syncthing.ListenPort = 22100 + offset
		cfg.Sync.Syncthing.DiscoveryPort = 21127 + offset
		cfg.Sync.Syncthing.GUIPort = 8484 + offset
	}

	if err := ctx.SaveConfig(cfg, configPath); err != nil {
		return err
	}

	fmt.Printf("Created configuration at %s\n", configPath)
	fmt.Println()
	fmt.Printf("Enabled %d systems with default emulators.\n", len(cfg.Systems))
	fmt.Println("Run 'kyaraben apply' to install emulators and create directories.")

	return nil
}
