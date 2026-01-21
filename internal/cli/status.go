package cli

import (
	"fmt"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

// StatusCmd shows the current state.
type StatusCmd struct{}

// Run executes the status command.
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

	userStorePath, err := cfg.ExpandUserStore()
	if err != nil {
		return err
	}
	userStore := store.NewUserStore(userStorePath)

	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("Emulation folder: %s", userStorePath)
	if userStore.IsInitialized() {
		fmt.Println(" (initialized)")
	} else {
		fmt.Println(" (not initialized)")
	}
	fmt.Println()

	// Enabled systems
	systems := cfg.EnabledSystems()
	if len(systems) == 0 {
		fmt.Println("Enabled systems: none")
	} else {
		systemNames := make([]string, 0, len(systems))
		for _, sys := range systems {
			s, err := registry.GetSystem(sys)
			if err != nil {
				systemNames = append(systemNames, string(sys))
			} else {
				systemNames = append(systemNames, s.Name)
			}
		}
		fmt.Printf("Enabled systems: %s\n", strings.Join(systemNames, ", "))
	}
	fmt.Println()

	// Installed emulators
	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return err
	}

	manifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	if len(manifest.InstalledEmulators) == 0 {
		fmt.Println("Managed emulators: none")
		fmt.Println()
		fmt.Println("Run 'kyaraben apply' to install emulators.")
	} else {
		fmt.Println("Managed emulators:")
		for _, emu := range manifest.InstalledEmulators {
			e, err := registry.GetEmulator(emu.ID)
			name := string(emu.ID)
			if err == nil {
				name = e.Name
			}
			fmt.Printf("  %-20s %s\n", name, emu.Version)
		}
		fmt.Println()

		if !manifest.LastApplied.IsZero() {
			fmt.Printf("Last applied: %s\n", manifest.LastApplied.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	// Provision status summary
	checker := store.NewProvisionChecker(userStore)
	missingRequired := 0
	for sys, sysConf := range cfg.Systems {
		emu, err := registry.GetEmulator(sysConf.Emulator)
		if err != nil {
			continue
		}
		results := checker.Check(emu, sys)
		if store.HasMissingRequired(results) {
			missingRequired++
		}
	}

	if missingRequired > 0 {
		fmt.Printf("Provision status: %d system(s) missing required files (run 'kyaraben doctor')\n", missingRequired)
	} else if len(cfg.Systems) > 0 {
		fmt.Println("Provision status: all required files present")
	}

	return nil
}
