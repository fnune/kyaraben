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
	DryRun   bool `help:"Show what would be done without making changes."`
	ShowDiff bool `help:"Show config changes before applying." default:"true" negatable:""`
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

	// Collect emulators and generate patches first (for diff display)
	emulatorsToInstall := make([]model.EmulatorID, 0, len(cfg.Systems))
	allPatches := make([]model.ConfigPatch, 0)

	for sys, sysConf := range cfg.Systems {
		emulatorsToInstall = append(emulatorsToInstall, sysConf.Emulator)

		gen := emulators.GetConfigGenerator(sysConf.Emulator)
		if gen == nil {
			continue
		}

		patches, err := gen.Generate(userStore, []model.SystemID{sys})
		if err != nil {
			return fmt.Errorf("generating config for %s: %w", sysConf.Emulator, err)
		}
		allPatches = append(allPatches, patches...)
	}

	// Show diff if requested or in dry-run mode
	if cmd.ShowDiff || cmd.DryRun {
		hasChanges := false
		fmt.Println("Config changes:")

		for _, patch := range allPatches {
			diff, err := emulators.ComputeDiff(patch)
			if err != nil {
				fmt.Printf("  Warning: could not compute diff for %s: %v\n", patch.Config.Path, err)
				continue
			}

			if diff.HasChanges() {
				hasChanges = true
				fmt.Print(diff.Format())
			}
		}

		if !hasChanges {
			fmt.Println("  No config changes needed.")
		}
		fmt.Println()

		if cmd.DryRun {
			fmt.Println("Dry run - no changes applied.")
			return nil
		}
	}

	// Step 1: Create UserStore directory structure
	fmt.Println("Creating directory structure...")
	if err := userStore.Initialize(); err != nil {
		return fmt.Errorf("initializing user store: %w", err)
	}

	for sys := range cfg.Systems {
		if err := userStore.InitializeSystem(sys); err != nil {
			return fmt.Errorf("initializing system %s: %w", sys, err)
		}
		fmt.Printf("  Created directories for %s\n", sys)
	}
	fmt.Println()

	// Step 2: Generate and build flake
	fmt.Println("Generating Nix flake...")
	flakeGen := nix.NewFlakeGenerator()

	if err := nixClient.EnsureFlakeDir(); err != nil {
		return fmt.Errorf("creating flake directory: %w", err)
	}

	if err := flakeGen.Generate(nixClient.FlakePath, emulatorsToInstall); err != nil {
		return fmt.Errorf("generating flake: %w", err)
	}
	fmt.Printf("  Flake written to %s\n", nixClient.FlakePath)
	fmt.Println()

	// Step 3: Build emulators
	fmt.Println("Building emulators (this may take a while on first run)...")
	buildCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	flakeRef := flakeGen.DefaultFlakeRef(nixClient.FlakePath)
	storePath, err := nixClient.Build(buildCtx, flakeRef)
	if err != nil {
		return fmt.Errorf("building emulators: %w", err)
	}
	fmt.Printf("  Built: %s\n", storePath)
	fmt.Println()

	// Step 4: Apply emulator configs
	fmt.Println("Applying emulator configurations...")
	configWriter := emulators.NewConfigWriter()

	for _, patch := range allPatches {
		if err := configWriter.Apply(patch); err != nil {
			return fmt.Errorf("applying config %s: %w", patch.Config.Path, err)
		}
		fmt.Printf("  Applied: %s\n", patch.Config.Path)
	}
	fmt.Println()

	// Step 5: Update manifest
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
			Version:   "latest",
			StorePath: storePath,
			Installed: time.Now(),
		})
	}

	// Track managed configs
	for _, patch := range allPatches {
		manifest.AddManagedConfig(model.ManagedConfig{
			Path:         patch.Config.Path,
			LastModified: time.Now(),
			EmulatorID:   patch.Config.EmulatorID,
		})
	}

	if err := manifest.Save(manifestPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	fmt.Println("Done!")
	fmt.Println()
	fmt.Printf("Your emulation directory is ready at: %s\n", userStore.Root)
	fmt.Println("Place your ROMs in the appropriate subdirectories.")

	return nil
}
