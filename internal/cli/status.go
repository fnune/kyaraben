package cli

import (
	"fmt"
	"strings"

	"github.com/fnune/kyaraben/internal/status"
)

type StatusCmd struct{}

func (cmd *StatusCmd) Run(ctx *Context) error {
	configPath, err := ctx.GetConfigPath()
	if err != nil {
		return err
	}

	cfg, err := ctx.LoadConfig()
	if err != nil {
		fmt.Printf("Config: %s (not found or invalid)\n", configPath)
		fmt.Println()
		fmt.Println("Run 'kyaraben init' to create a configuration.")
		return nil
	}

	registry := ctx.NewRegistry()

	result, err := status.Get(cfg, configPath, registry)
	if err != nil {
		return err
	}

	fmt.Printf("Config: %s\n", result.ConfigPath)
	fmt.Printf("Emulation folder: %s", result.UserStorePath)
	if result.UserStoreInitialized {
		fmt.Println(" (initialized)")
	} else {
		fmt.Println(" (not initialized)")
	}
	fmt.Println()

	if len(result.EnabledSystems) == 0 {
		fmt.Println("Enabled systems: none")
	} else {
		names := make([]string, len(result.EnabledSystems))
		for i, sys := range result.EnabledSystems {
			names[i] = sys.Name
		}
		fmt.Printf("Enabled systems: %s\n", strings.Join(names, ", "))
	}
	fmt.Println()

	if len(result.InstalledEmulators) == 0 {
		fmt.Println("Managed emulators: none")
		fmt.Println()
		fmt.Println("Run 'kyaraben apply' to install emulators.")
	} else {
		fmt.Println("Managed emulators:")
		for _, emu := range result.InstalledEmulators {
			fmt.Printf("  %-20s %s\n", emu.Name, emu.Version)
		}
		fmt.Println()

		if !result.LastApplied.IsZero() {
			fmt.Printf("Last applied: %s\n", result.LastApplied.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	if result.MissingRequiredCount > 0 {
		fmt.Printf("Provision status: %d system(s) missing required files (run 'kyaraben doctor')\n", result.MissingRequiredCount)
	} else if len(result.EnabledSystems) > 0 {
		fmt.Println("Provision status: all required files present")
	}

	return nil
}
