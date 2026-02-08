package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fnune/kyaraben/internal/apply"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
)

type ApplyCmd struct {
	DryRun bool `help:"Show what would be done without making changes."`
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
	flakeGenerator := nix.NewFlakeGenerator(registry, registry)
	versionOverrides, err := cfg.BuildVersionOverrides(registry.GetEmulator)
	if err != nil {
		return err
	}
	flakeGenerator.SetVersionOverrides(versionOverrides)
	configWriter := emulators.NewConfigWriter(model.OSBaseDirResolver{})
	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return err
	}

	launcherManager, err := launcher.NewManager()
	if err != nil {
		return fmt.Errorf("creating launcher manager: %w", err)
	}

	applier := &apply.Applier{
		NixClient:       nixClient,
		FlakeGenerator:  flakeGenerator,
		ConfigWriter:    configWriter,
		Registry:        registry,
		ManifestPath:    manifestPath,
		LauncherManager: launcherManager,
		BaseDirResolver: model.OSBaseDirResolver{},
	}

	fmt.Println("Applying kyaraben configuration...")
	fmt.Println()

	preflight, err := applier.Preflight(context.Background(), cfg, userStore)
	if err != nil {
		return fmt.Errorf("preflight check: %w", err)
	}

	createBackups := false
	if len(preflight.FilesToBackup) > 0 {
		fmt.Println("The following existing config files will be modified:")
		for _, path := range preflight.FilesToBackup {
			fmt.Printf("  %s\n", path)
		}
		fmt.Println()
		fmt.Print("Create backups before modifying? [Y/n] ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		createBackups = response == "" || response == "y" || response == "yes"
		fmt.Println()
	}

	buildMsgPrinted := false
	opts := apply.Options{
		DryRun:        cmd.DryRun,
		CreateBackups: createBackups,
		OnProgress: func(p apply.Progress) {
			switch p.Step {
			case "store":
				fmt.Println(p.Message)
			case "build":
				if !buildMsgPrinted {
					buildMsgPrinted = true
					fmt.Println("Installing emulators (this may take a while on first run)...")
				}
			case "desktop":
				fmt.Println("Adding to application menu...")
			case "config":
				fmt.Println("Configuring emulators...")
			}
		},
	}

	dryOpts := apply.Options{DryRun: true}
	dryResult, err := applier.Apply(context.Background(), cfg, userStore, dryOpts)
	if err != nil {
		return err
	}

	manifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	var totalAdds, totalModifies, totalRemoves int
	var filesCreated, filesModified, filesUnchanged int
	var hasOverwrittenUserChanges bool

	diffs := make([]*emulators.ConfigDiff, 0, len(dryResult.Patches))
	for _, patch := range dryResult.Patches {
		baseline, found := manifest.GetManagedConfig(patch.Target)
		var baselinePtr *model.ManagedConfig
		if found {
			baselinePtr = &baseline
		}

		diff, err := emulators.ComputeDiffWithBaseline(patch, baselinePtr)
		if err != nil {
			if cmd.DryRun {
				fmt.Printf("  Warning: could not compute diff: %v\n", err)
			}
			continue
		}

		diffs = append(diffs, diff)

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

		if diff.UserModified && len(diff.UserChanges) > 0 && diff.HasChanges() {
			hasOverwrittenUserChanges = true
		}
	}

	if cmd.DryRun {
		fmt.Println("Config changes:")
		fmt.Println()

		for _, diff := range diffs {
			fmt.Print(diff.Format())
		}

		fmt.Println()
		fmt.Printf("  Summary: %d file(s) to create, %d to modify, %d unchanged\n",
			filesCreated, filesModified, filesUnchanged)
		if totalAdds > 0 || totalModifies > 0 || totalRemoves > 0 {
			fmt.Printf("  Changes: %d additions, %d modifications, %d removals\n",
				totalAdds, totalModifies, totalRemoves)
		}
		fmt.Println()

		fmt.Println("Dry run - no changes applied.")
		return nil
	}

	if hasOverwrittenUserChanges {
		fmt.Println("Your changes to managed keys will be overwritten:")
		fmt.Println()
		for _, diff := range diffs {
			if diff.UserModified && len(diff.UserChanges) > 0 {
				fmt.Print(diff.Format())
			}
		}
		fmt.Println()
		fmt.Print("Proceed? [Y/n] ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "" && response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
		fmt.Println()
	}

	result, err := applier.Apply(context.Background(), cfg, userStore, opts)
	if err != nil {
		return err
	}

	fmt.Printf("  Built: %s\n", result.StorePath)
	fmt.Println()

	for _, patch := range result.Patches {
		path, _ := patch.Target.Resolve()
		fmt.Printf("  Applied: %s\n", path)
	}
	fmt.Println()

	if len(result.Backups) > 0 {
		fmt.Println("Backups created:")
		for _, backup := range result.Backups {
			fmt.Printf("  %s\n", backup.BackupPath)
		}
		fmt.Println()
	}

	fmt.Println("Done!")
	fmt.Println()
	fmt.Printf("Your emulation directory is ready at: %s\n", userStore.Root())
	fmt.Println("Place your ROMs in the appropriate subdirectories.")

	return nil
}
