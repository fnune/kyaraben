package cli

import (
	"fmt"
	"os"

	"github.com/fnune/kyaraben/internal/model"
)

// InitCmd initializes a new kyaraben configuration.
type InitCmd struct {
	UserStore string   `short:"u" help:"Path to emulation directory." default:"~/Emulation"`
	Systems   []string `short:"s" help:"Systems to enable (e.g., snes, psx, tic80)."`
	Force     bool     `short:"f" help:"Overwrite existing configuration."`
}

// Run executes the init command.
func (cmd *InitCmd) Run(ctx *Context) error {
	configPath, err := ctx.GetConfigPath()
	if err != nil {
		return err
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !cmd.Force {
		return fmt.Errorf("config already exists at %s. Use --force to overwrite", configPath)
	}

	registry := ctx.NewRegistry()

	// Create config
	cfg, err := model.NewDefaultConfig()
	if err != nil {
		return err
	}

	cfg.Global.UserStore = cmd.UserStore

	// Add systems
	if len(cmd.Systems) == 0 {
		// Default to tic80 for easy testing
		cmd.Systems = []string{"tic80"}
	}

	for _, sysName := range cmd.Systems {
		sysID := model.SystemID(sysName)
		_, err := registry.GetSystem(sysID)
		if err != nil {
			return fmt.Errorf("unknown system: %s", sysName)
		}

		emu, err := registry.GetDefaultEmulator(sysID)
		if err != nil {
			return fmt.Errorf("no default emulator for system %s", sysName)
		}

		cfg.Systems[sysID] = model.SystemConf{
			Emulator: emu.ID,
		}
	}

	// Save config
	if err := model.SaveConfig(cfg, configPath); err != nil {
		return err
	}

	fmt.Printf("Created configuration at %s\n", configPath)
	fmt.Println()
	fmt.Println("Enabled systems:")
	for sys, sysConf := range cfg.Systems {
		s, _ := registry.GetSystem(sys)
		e, _ := registry.GetEmulator(sysConf.Emulator)
		fmt.Printf("  %s (%s) - %s\n", s.Name, sys, e.Name)
	}
	fmt.Println()
	fmt.Println("Run 'kyaraben apply' to install emulators and create directories.")

	return nil
}
