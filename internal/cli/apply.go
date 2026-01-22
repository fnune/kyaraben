package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
)

// ApplyCmd applies the kyaraben configuration.
type ApplyCmd struct {
	DryRun bool `help:"Show what would be done without making changes."`
}

// Run executes the apply command.
func (cmd *ApplyCmd) Run(ctx *Context) error {
	cfg, err := ctx.LoadConfig()
	if err != nil {
		return err
	}

	nixClient, err := ctx.NewNixClient()
	if err != nil {
		return err
	}

	userStore, err := ctx.NewUserStore(cfg)
	if err != nil {
		return err
	}

	// Check if nix is available (skip in dry-run mode)
	if !cmd.DryRun && !nixClient.IsAvailable() {
		return fmt.Errorf("nix is not available. Please install nix or run from the development shell")
	}

	fmt.Println("Applying kyaraben configuration...")
	fmt.Println()

	// Step 1: Create UserStore directory structure
	fmt.Println("Creating directory structure...")
	if !cmd.DryRun {
		if err := userStore.Initialize(); err != nil {
			return fmt.Errorf("initializing user store: %w", err)
		}
	}

	// Collect emulators to install
	emulatorsToInstall := make([]model.EmulatorID, 0, len(cfg.Systems))
	for sys, sysConf := range cfg.Systems {
		// Initialize system directories
		if !cmd.DryRun {
			if err := userStore.InitializeSystem(sys); err != nil {
				return fmt.Errorf("initializing system %s: %w", sys, err)
			}
		}
		fmt.Printf("  Created directories for %s\n", sys)

		emulatorsToInstall = append(emulatorsToInstall, sysConf.Emulator)
	}
	fmt.Println()

	// Step 2: Generate and build flake
	fmt.Println("Generating Nix flake...")
	flakeGen := nix.NewFlakeGenerator()

	if !cmd.DryRun {
		if err := nixClient.EnsureFlakeDir(); err != nil {
			return fmt.Errorf("creating flake directory: %w", err)
		}

		if err := flakeGen.Generate(nixClient.FlakePath, emulatorsToInstall); err != nil {
			return fmt.Errorf("generating flake: %w", err)
		}
	}
	fmt.Printf("  Flake written to %s\n", nixClient.FlakePath)
	fmt.Println()

	// Step 3: Build emulators
	fmt.Println("Building emulators (this may take a while on first run)...")
	buildCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	var storePath string
	if !cmd.DryRun {
		flakeRef := flakeGen.DefaultFlakeRef(nixClient.FlakePath)
		storePath, err = nixClient.Build(buildCtx, flakeRef)
		if err != nil {
			return fmt.Errorf("building emulators: %w", err)
		}
		fmt.Printf("  Built: %s\n", storePath)
	} else {
		fmt.Println("  (dry run - skipping build)")
	}
	fmt.Println()

	// Step 4: Generate emulator configs
	fmt.Println("Generating emulator configurations...")
	configWriter := emulators.NewConfigWriter()

	for sys, sysConf := range cfg.Systems {
		gen := emulators.GetConfigGenerator(sysConf.Emulator)
		if gen == nil {
			fmt.Printf("  Warning: no config generator for %s\n", sysConf.Emulator)
			continue
		}

		patches, err := gen.Generate(userStore, []model.SystemID{sys})
		if err != nil {
			return fmt.Errorf("generating config for %s: %w", sysConf.Emulator, err)
		}

		for _, patch := range patches {
			if !cmd.DryRun {
				if err := configWriter.Apply(patch); err != nil {
					return fmt.Errorf("applying config %s: %w", patch.Config.Path, err)
				}
			}
			fmt.Printf("  Configured: %s\n", patch.Config.Path)
		}
	}
	fmt.Println()

	// Step 5: Update manifest
	if !cmd.DryRun {
		manifestPath, err := model.DefaultManifestPath()
		if err != nil {
			return fmt.Errorf("getting manifest path: %w", err)
		}

		manifest, err := model.LoadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("loading manifest: %w", err)
		}

		manifest.LastApplied = time.Now()

		for _, emuID := range emulatorsToInstall {
			manifest.AddEmulator(model.InstalledEmulator{
				ID:        emuID,
				Version:   "latest", // TODO: get actual version from nix
				StorePath: storePath,
				Installed: time.Now(),
			})
		}

		if err := manifest.Save(manifestPath); err != nil {
			return fmt.Errorf("saving manifest: %w", err)
		}
	}

	fmt.Println("Done!")
	fmt.Println()
	fmt.Printf("Your emulation directory is ready at: %s\n", userStore.Root)
	fmt.Println("Place your ROMs in the appropriate subdirectories.")

	return nil
}
