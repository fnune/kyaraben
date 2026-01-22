package cli

import (
	"fmt"

	"github.com/fnune/kyaraben/internal/apply"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
)

type ApplyCmd struct {
	DryRun     bool `help:"Show what would be done without making changes."`
	ShowDiff   bool `help:"Show config changes before applying." default:"true" negatable:""`
	ConfigOnly bool `help:"Only apply configs, skip Nix build (for home-manager integration)."`
}

func (cmd *ApplyCmd) Run(ctx *Context) error {
	cfg, err := ctx.LoadConfig()
	if err != nil {
		return err
	}

	userStore, err := ctx.NewUserStore(cfg)
	if err != nil {
		return err
	}

	registry := ctx.NewRegistry()
	nixClient, err := ctx.NewNixClient()
	if err != nil {
		return fmt.Errorf("creating nix client: %w", err)
	}
	flakeGenerator := nix.NewFlakeGenerator(registry)
	configWriter := emulators.NewConfigWriter()
	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return err
	}

	applier := &apply.Applier{
		NixClient:      nixClient,
		FlakeGenerator: flakeGenerator,
		ConfigWriter:   configWriter,
		Registry:       registry,
		ManifestPath:   manifestPath,
	}

	fmt.Println("Applying kyaraben configuration...")
	fmt.Println()

	opts := apply.Options{
		DryRun:     cmd.DryRun,
		ShowDiff:   cmd.ShowDiff,
		ConfigOnly: cmd.ConfigOnly,
		OnProgress: func(p apply.Progress) {
			switch p.Step {
			case "directories":
				fmt.Println("Creating directory structure...")
			case "flake":
				fmt.Println("Generating Nix flake...")
			case "build":
				fmt.Println("Building emulators (this may take a while on first run)...")
			case "configs":
				fmt.Println("Applying emulator configurations...")
			}
		},
	}

	if cmd.DryRun || cmd.ShowDiff {
		dryOpts := apply.Options{DryRun: true}
		dryResult, err := applier.Apply(cfg, userStore, dryOpts)
		if err != nil {
			return err
		}

		fmt.Println("Config changes:")
		fmt.Println()

		var totalAdds, totalModifies, totalRemoves int
		var filesCreated, filesModified, filesUnchanged int

		for _, patch := range dryResult.Patches {
			diff, err := emulators.ComputeDiff(patch)
			if err != nil {
				fmt.Printf("  Warning: could not compute diff for %s: %v\n", patch.Config.Path, err)
				continue
			}

			fmt.Print(diff.Format())

			adds, modifies, removes := diff.Stats()
			totalAdds += adds
			totalModifies += modifies
			totalRemoves += removes

			if diff.IsNewFile {
				filesCreated++
			} else if diff.HasChanges() {
				filesModified++
			} else {
				filesUnchanged++
			}
		}

		// Summary
		fmt.Println()
		fmt.Printf("  Summary: %d file(s) to create, %d to modify, %d unchanged\n",
			filesCreated, filesModified, filesUnchanged)
		if totalAdds > 0 || totalModifies > 0 || totalRemoves > 0 {
			fmt.Printf("  Changes: %d additions, %d modifications, %d removals\n",
				totalAdds, totalModifies, totalRemoves)
		}
		fmt.Println()

		if cmd.DryRun {
			fmt.Println("Dry run - no changes applied.")
			return nil
		}
	}

	result, err := applier.Apply(cfg, userStore, opts)
	if err != nil {
		return err
	}

	if result.StorePath != "" {
		fmt.Printf("  Built: %s\n", result.StorePath)
		fmt.Println()
	}

	for _, patch := range result.Patches {
		fmt.Printf("  Applied: %s\n", patch.Config.Path)
	}
	fmt.Println()

	fmt.Println("Done!")
	fmt.Println()
	fmt.Printf("Your emulation directory is ready at: %s\n", userStore.Root)
	if !cmd.ConfigOnly {
		fmt.Println("Place your ROMs in the appropriate subdirectories.")
	}

	return nil
}
