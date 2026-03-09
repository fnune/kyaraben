package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/apply"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/emulators/symlink"
	"github.com/fnune/kyaraben/internal/folders"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/packages"
	"github.com/fnune/kyaraben/internal/steam"
	"github.com/fnune/kyaraben/internal/store"
	syncpkg "github.com/fnune/kyaraben/internal/sync"
)

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

type ApplyCmd struct {
	DryRun bool `help:"Show what would be done without making changes."`
}

func (cmd *ApplyCmd) Run(ctx *Context) error {
	cfg, err := ctx.LoadConfig()
	if err != nil {
		return err
	}

	collection, err := ctx.NewCollection(cfg)
	if err != nil {
		return err
	}

	if cfg.Global.Headless {
		return cmd.runHeadless(ctx, cfg, collection)
	}

	registry := ctx.NewRegistry()
	installer, err := ctx.NewInstaller()
	if err != nil {
		return fmt.Errorf("creating installer: %w", err)
	}
	versionOverrides, err := cfg.BuildVersionOverrides(registry.GetEmulator, registry.GetFrontend)
	if err != nil {
		return err
	}
	installer.SetVersionOverrides(versionOverrides)
	resolver := model.NewDefaultResolver()
	configWriter := emulators.NewDefaultConfigWriter()
	manifestPath, err := ctx.GetPaths().ManifestPath()
	if err != nil {
		return err
	}

	lockDir := filepath.Dir(manifestPath)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}
	lockPath := filepath.Join(lockDir, "apply.lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("creating lock file: %w", err)
	}
	defer func() { _ = lockFile.Close() }()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return fmt.Errorf("another installation is already in progress")
	}
	defer func() { _ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN) }()

	launcherManager, err := launcher.NewManager(vfs.OSFS, ctx.GetPaths(), resolver)
	if err != nil {
		return fmt.Errorf("creating launcher manager: %w", err)
	}

	applier := apply.NewApplier(
		vfs.OSFS,
		installer,
		configWriter,
		registry,
		manifestPath,
		launcherManager,
		resolver,
		symlink.NewCreator(vfs.OSFS),
	)
	applier.SteamManager = steam.NewDefaultManager()

	fmt.Println("Applying kyaraben configuration...")
	fmt.Println()

	preflight, err := applier.Preflight(context.Background(), cfg, collection)
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
	lastProgressLine := ""
	isTTY := isTerminal()

	clearProgress := func() {
		if isTTY && lastProgressLine != "" {
			fmt.Print("\r" + strings.Repeat(" ", len(lastProgressLine)) + "\r")
			lastProgressLine = ""
		}
	}

	opts := apply.Options{
		DryRun:        cmd.DryRun,
		CreateBackups: createBackups,
		OnProgress: func(p apply.Progress) {
			switch p.Step {
			case "store":
				clearProgress()
				fmt.Println(p.Message)
			case "build":
				if !buildMsgPrinted {
					buildMsgPrinted = true
					fmt.Println("Installing emulators...")
				}

				var line string
				switch p.BuildPhase {
				case "downloading":
					if p.PackageName != "" {
						line = fmt.Sprintf("  Downloading %s...", p.PackageName)
					}
				case "extracting":
					if p.PackageName != "" {
						line = fmt.Sprintf("  Extracting %s...", p.PackageName)
					}
				case "installed":
					if p.PackageName != "" {
						line = fmt.Sprintf("  Installed %s", p.PackageName)
					}
				case "skipped":
					if p.PackageName != "" {
						line = fmt.Sprintf("  %s (already installed)", p.PackageName)
					}
				}

				if line != "" && isTTY {
					if len(line) < len(lastProgressLine) {
						padding := strings.Repeat(" ", len(lastProgressLine)-len(line))
						fmt.Print("\r" + line + padding)
					} else {
						fmt.Print("\r" + line)
					}
					lastProgressLine = line
				}
			case "desktop":
				clearProgress()
				fmt.Println("Adding to application menu...")
			case "config":
				clearProgress()
				fmt.Println("Configuring emulators...")
			}
		},
	}

	dryOpts := apply.Options{DryRun: true}
	dryResult, err := applier.Apply(context.Background(), cfg, collection, dryOpts)
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
	var hasKyarabenUpdates bool

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
		if diff.KyarabenChanged && len(diff.VersionUpgrades) > 0 {
			hasKyarabenUpdates = true
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

	if hasKyarabenUpdates {
		fmt.Println("\033[32mKyaraben has updated its defaults:\033[0m")
		fmt.Println()
		seenKyarabenPaths := make(map[string]bool)
		for _, diff := range diffs {
			if diff.KyarabenChanged && len(diff.VersionUpgrades) > 0 && !seenKyarabenPaths[diff.Path] {
				seenKyarabenPaths[diff.Path] = true
				fmt.Printf("  %s\n", diff.Path)
				seenKeys := make(map[string]bool)
				for _, vu := range diff.VersionUpgrades {
					if !seenKeys[vu.Key] {
						seenKeys[vu.Key] = true
						newDisplay := vu.NewValue
						if newDisplay == "" {
							newDisplay = "(removed)"
						}
						fmt.Printf("    %s: was: %s → becomes: %s\n", vu.Key, vu.OldValue, newDisplay)
					}
				}
			}
		}
		fmt.Println()
	}

	if hasOverwrittenUserChanges {
		fmt.Println("\033[33mYour changes to managed settings will be overwritten:\033[0m")
		fmt.Println()
		seenUserPaths := make(map[string]bool)
		for _, diff := range diffs {
			if diff.UserModified && len(diff.UserChanges) > 0 && !seenUserPaths[diff.Path] {
				seenUserPaths[diff.Path] = true
				fmt.Printf("  %s\n", diff.Path)
				seenKeys := make(map[string]bool)
				for _, uc := range diff.UserChanges {
					if !seenKeys[uc.Key] {
						seenKeys[uc.Key] = true
						currentDisplay := uc.CurrentValue
						if currentDisplay == "" {
							currentDisplay = "(deleted)"
						}
						fmt.Printf("    %s: yours: %s → kyaraben: %s\n", uc.Key, currentDisplay, uc.WrittenValue)
					}
				}
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

	result, err := applier.Apply(context.Background(), cfg, collection, opts)
	if err != nil {
		return err
	}

	clearProgress()
	fmt.Println()

	seenPaths := make(map[string]bool)
	for _, patch := range result.Patches {
		path, _ := patch.Target.ResolveWith(resolver)
		if !seenPaths[path] {
			seenPaths[path] = true
			fmt.Printf("  Applied: %s\n", path)
		}
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
	fmt.Printf("Your collection is ready at: %s\n", collection.Root())
	fmt.Println("Place your ROMs in the appropriate subdirectories.")

	return nil
}

func (cmd *ApplyCmd) runHeadless(ctx *Context, cfg *model.KyarabenConfig, collection *store.Collection) error {
	if cmd.DryRun {
		fmt.Println("Headless mode (sync hub only)")
		fmt.Println()
		fmt.Println("Would create directories for all systems and emulators.")
		if cfg.Sync.Enabled {
			fmt.Println("Would set up synchronization.")
		}
		fmt.Println()
		fmt.Println("Dry run - no changes applied.")
		return nil
	}

	fmt.Println("Setting up headless sync hub...")
	fmt.Println()

	if err := collection.Initialize(); err != nil {
		return fmt.Errorf("initializing collection: %w", err)
	}

	registry := ctx.NewRegistry()
	systemCount := 0
	for _, sys := range registry.AllSystems() {
		for _, emu := range registry.GetEmulatorsForSystem(sys.ID) {
			if err := collection.InitializeForEmulator(sys.ID, emu.ID, emu.PathUsage); err != nil {
				return fmt.Errorf("initializing %s for %s: %v", sys.ID, emu.ID, err)
			}
		}
		systemCount++
	}
	fmt.Printf("Created directories for %d systems\n", systemCount)

	if cfg.Sync.Enabled {
		fmt.Println("Setting up synchronization...")

		installer, err := ctx.NewInstaller()
		if err != nil {
			return fmt.Errorf("creating installer: %w", err)
		}

		stateDir, err := ctx.GetPaths().StateDir()
		if err != nil {
			return fmt.Errorf("getting state directory: %w", err)
		}

		allSystems := make([]model.SystemID, 0)
		for _, sys := range registry.AllSystems() {
			allSystems = append(allSystems, sys.ID)
		}

		allEmulators := make([]folders.EmulatorInfo, 0)
		for _, emu := range registry.AllEmulators() {
			allEmulators = append(allEmulators, folders.EmulatorInfo{
				ID:                 emu.ID,
				UsesStatesDir:      emu.PathUsage.UsesStatesDir,
				UsesScreenshotsDir: emu.PathUsage.UsesScreenshotsDir,
			})
		}

		defaultCfg := model.NewDefaultConfig()
		allFrontends := defaultCfg.EnabledFrontends()

		setup := syncpkg.NewDefaultSetup(installer, stateDir)
		result, err := setup.Install(
			context.Background(),
			cfg.Sync,
			collection.Root(),
			allSystems,
			allEmulators,
			allFrontends,
			func(p packages.InstallProgress) {
				switch p.Phase {
				case "downloading":
					fmt.Printf("  Downloading %s...\n", p.PackageName)
				case "extracting":
					fmt.Printf("  Extracting %s...\n", p.PackageName)
				}
			},
		)
		if err != nil {
			return fmt.Errorf("setting up sync: %w", err)
		}

		manifestPath, err := ctx.GetPaths().ManifestPath()
		if err != nil {
			return fmt.Errorf("getting manifest path: %w", err)
		}

		manifest, err := model.LoadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("loading manifest: %w", err)
		}

		manifest.SyncthingInstall = &model.SyncthingInstall{
			Version:             installer.ResolveVersion("syncthing"),
			ConfigSchemaVersion: syncpkg.ConfigSchemaVersion,
			BinaryPath:          result.SyncthingBinary,
			ConfigDir:           result.ConfigDir,
			DataDir:             result.DataDir,
			SystemdUnitPath:     result.SystemdUnitPath,
		}
		if err := manifest.SaveWithBackup(manifestPath); err != nil {
			return fmt.Errorf("saving manifest: %w", err)
		}

		fmt.Println("Synchronization ready")
	}

	fmt.Println()
	fmt.Println("Done!")
	fmt.Println()
	fmt.Printf("Your collection is ready at: %s\n", collection.Root())
	fmt.Println("Place your ROMs in the appropriate subdirectories.")
	fmt.Println()
	fmt.Println("Other devices can sync with this hub using:")
	fmt.Println("  kyaraben sync pair")

	return nil
}
