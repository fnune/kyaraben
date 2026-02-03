package apply

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
)

const nixBuildTimeout = 30 * time.Minute

type Progress struct {
	Step    string
	Message string
	Output  string // Optional streaming output line (e.g., from nix build)
	Speed   string // Optional download speed (e.g., "12.3 MB/s")
}

func formatSpeed(bytesPerSec int64) string {
	const unit = 1024
	if bytesPerSec < unit {
		return fmt.Sprintf("%d B/s", bytesPerSec)
	}
	div, exp := int64(unit), 0
	for n := bytesPerSec / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB/s", float64(bytesPerSec)/float64(div), "KMGTPE"[exp])
}

type Result struct {
	StorePath string
	Patches   []model.ConfigPatch
	Backups   []BackupInfo
}

type BackupInfo struct {
	OriginalPath string
	BackupPath   string
}

type Options struct {
	DryRun        bool
	ShowDiff      bool
	CreateBackups bool
	OnProgress    func(Progress)
}

type PreflightResult struct {
	Patches       []model.ConfigPatch
	FilesToBackup []string
}

func (a *Applier) Preflight(ctx context.Context, cfg *model.KyarabenConfig, userStore *store.UserStore) (*PreflightResult, error) {
	allPatches := make([]model.ConfigPatch, 0)

	for emuID := range a.collectEnabledEmulators(cfg) {
		gen := a.Registry.GetConfigGenerator(emuID)
		if gen == nil {
			continue
		}

		patches, err := gen.Generate(userStore)
		if err != nil {
			return nil, fmt.Errorf("generating config for %s: %w", emuID, err)
		}
		allPatches = append(allPatches, patches...)
	}

	manifest, err := model.LoadManifest(a.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	var filesToBackup []string
	for _, patch := range allPatches {
		path, exists, err := a.ConfigWriter.NeedsBackup(patch)
		if err != nil {
			return nil, fmt.Errorf("checking config file: %w", err)
		}

		if !exists {
			continue
		}

		if _, managed := manifest.GetManagedConfig(patch.Target); managed {
			continue
		}

		diff, err := emulators.ComputeDiff(patch)
		if err != nil {
			return nil, fmt.Errorf("computing diff for %s: %w", path, err)
		}

		if diff.HasChanges() {
			filesToBackup = append(filesToBackup, path)
		}
	}

	return &PreflightResult{
		Patches:       allPatches,
		FilesToBackup: filesToBackup,
	}, nil
}

type Applier struct {
	NixClient       nix.NixClient
	FlakeGenerator  *nix.FlakeGenerator
	ConfigWriter    *emulators.ConfigWriter
	Registry        *registry.Registry
	ManifestPath    string
	LauncherManager *launcher.Manager
}

func (a *Applier) Apply(ctx context.Context, cfg *model.KyarabenConfig, userStore *store.UserStore, opts Options) (*Result, error) {
	if opts.OnProgress == nil {
		opts.OnProgress = func(Progress) {}
	}

	if !opts.DryRun && !a.NixClient.IsAvailable() {
		return nil, fmt.Errorf("nix is not available (nix-portable not found)")
	}

	enabledEmulators := a.collectEnabledEmulators(cfg)
	emulatorsToInstall := make([]model.EmulatorID, 0, len(enabledEmulators))
	allPatches := make([]model.ConfigPatch, 0)

	for emuID := range enabledEmulators {
		emulatorsToInstall = append(emulatorsToInstall, emuID)

		gen := a.Registry.GetConfigGenerator(emuID)
		if gen == nil {
			continue
		}

		patches, err := gen.Generate(userStore)
		if err != nil {
			return nil, fmt.Errorf("generating config for %s: %w", emuID, err)
		}
		allPatches = append(allPatches, patches...)
	}

	if opts.DryRun {
		return &Result{Patches: allPatches}, nil
	}

	var storeMsg string
	if userStore.Exists() {
		storeMsg = fmt.Sprintf("Using %s (existing data preserved)", userStore.Path())
	} else {
		storeMsg = fmt.Sprintf("Creating %s", userStore.Path())
	}
	opts.OnProgress(Progress{Step: "store", Message: storeMsg})

	if err := userStore.Initialize(); err != nil {
		return nil, fmt.Errorf("initializing user store: %w", err)
	}

	for sys := range cfg.Systems {
		if err := userStore.InitializeSystem(sys); err != nil {
			return nil, fmt.Errorf("initializing system %s: %w", sys, err)
		}
	}

	if err := a.NixClient.EnsureFlakeDir(); err != nil {
		return nil, fmt.Errorf("creating flake directory: %w", err)
	}

	genResult, err := a.FlakeGenerator.Generate(a.NixClient.GetFlakePath(), emulatorsToInstall)
	if err != nil {
		return nil, fmt.Errorf("generating flake: %w", err)
	}

	for _, skipped := range genResult.SkippedEmulators {
		opts.OnProgress(Progress{Step: "build", Message: fmt.Sprintf("Warning: emulator '%s' is no longer supported and will be skipped", skipped)})
	}

	resolvedVersions := a.FlakeGenerator.GetResolvedVersions(emulatorsToInstall)

	opts.OnProgress(Progress{Step: "build", Message: "This may take a few minutes on first run"})

	netMon := NewNetMonitor(func(bytesPerSec int64) {
		if bytesPerSec > 1024 { // Only report if >1 KB/s
			opts.OnProgress(Progress{Step: "build", Speed: formatSpeed(bytesPerSec)})
		}
	})
	netMon.Start()
	defer netMon.Stop()

	a.NixClient.SetOutputCallback(func(line string) {
		opts.OnProgress(Progress{Step: "build", Output: line})
	})
	defer a.NixClient.SetOutputCallback(nil)

	opts.OnProgress(Progress{Step: "build", Message: "Resolving package versions..."})

	buildCtx, cancel := context.WithTimeout(ctx, nixBuildTimeout)
	defer cancel()

	flakeRef := a.FlakeGenerator.DefaultFlakeRef(string(genResult.Path))

	var profileLink string
	if a.LauncherManager != nil {
		profileLink = a.LauncherManager.CurrentLink()
	}

	var storePath string
	if profileLink != "" {
		// Remove existing symlink before building - nix won't overwrite it
		if _, err := os.Lstat(profileLink); err == nil {
			if err := os.Remove(profileLink); err != nil {
				return nil, fmt.Errorf("removing existing profile link: %w", err)
			}
		}

		opts.OnProgress(Progress{Step: "build", Output: "$ nix build " + flakeRef})
		if err := a.NixClient.BuildWithLink(buildCtx, flakeRef, profileLink); err != nil {
			return nil, fmt.Errorf("building emulators: %w", err)
		}
		target, err := os.Readlink(profileLink)
		if err != nil {
			return nil, fmt.Errorf("reading profile link: %w", err)
		}

		// nix-portable virtualizes /nix/store, so the symlink target doesn't exist
		// on the real filesystem. Translate to the real store path.
		realTarget := a.NixClient.RealStorePath(target)
		if realTarget != target {
			if err := os.Remove(profileLink); err != nil {
				return nil, fmt.Errorf("removing old profile link: %w", err)
			}
			if err := os.Symlink(realTarget, profileLink); err != nil {
				return nil, fmt.Errorf("creating real profile link: %w", err)
			}
		}
		storePath = realTarget
	} else {
		opts.OnProgress(Progress{Step: "build", Output: "$ nix build " + flakeRef})
		var err error
		storePath, err = a.NixClient.Build(buildCtx, flakeRef)
		if err != nil {
			return nil, fmt.Errorf("building emulators: %w", err)
		}
	}

	manifest, err := model.LoadManifest(a.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	if a.LauncherManager != nil {
		opts.OnProgress(Progress{Step: "desktop"})

		persistentNixPortable, err := a.NixClient.EnsurePersistentNixPortable()
		if err != nil {
			return nil, fmt.Errorf("ensuring persistent nix-portable: %w", err)
		}
		a.LauncherManager.SetNixPortableBinary(persistentNixPortable)
		a.LauncherManager.SetNixPortableLocation(a.NixClient.GetNixPortableLocation())
		packageInfo := a.buildEmulatorPackageInfo(emulatorsToInstall)
		if err := a.LauncherManager.GenerateWrappers(packageInfo); err != nil {
			return nil, fmt.Errorf("generating launcher wrappers: %w", err)
		}

		previousFiles := &launcher.GeneratedFiles{
			DesktopFiles: manifest.DesktopFiles,
			IconFiles:    manifest.IconFiles,
		}
		desktopEntries := a.buildDesktopEntries(emulatorsToInstall, userStore)
		generatedFiles, err := a.LauncherManager.GenerateDesktopFiles(desktopEntries, previousFiles)
		if err != nil {
			return nil, fmt.Errorf("generating desktop files: %w", err)
		}
		manifest.DesktopFiles = generatedFiles.DesktopFiles
		manifest.IconFiles = generatedFiles.IconFiles
	}

	opts.OnProgress(Progress{Step: "config"})

	configResults := make([]emulators.ApplyResult, len(allPatches))
	var backups []BackupInfo
	for i, patch := range allPatches {
		createBackup := false
		if opts.CreateBackups {
			_, exists, err := a.ConfigWriter.NeedsBackup(patch)
			if err != nil {
				return nil, fmt.Errorf("checking config file: %w", err)
			}
			if exists {
				if _, managed := manifest.GetManagedConfig(patch.Target); !managed {
					createBackup = true
				}
			}
		}

		result, err := a.ConfigWriter.ApplyWithOptions(patch, emulators.ApplyOptions{
			CreateBackup: createBackup,
		})
		if err != nil {
			return nil, fmt.Errorf("applying config: %w", err)
		}
		configResults[i] = result

		if result.BackupPath != "" {
			backups = append(backups, BackupInfo{
				OriginalPath: result.Path,
				BackupPath:   result.BackupPath,
			})
		}
	}

	manifest.LastApplied = time.Now()

	for _, emuID := range emulatorsToInstall {
		version := resolvedVersions[emuID]
		if version == "" {
			version = "unknown"
		}
		manifest.AddEmulator(model.InstalledEmulator{
			ID:        emuID,
			Version:   version,
			StorePath: storePath,
			Installed: time.Now(),
		})
	}

	for i, patch := range allPatches {
		if patch.Target.BaseDir == model.ConfigBaseDirOpaqueDir {
			continue
		}

		managedKeys := make([]model.ManagedKey, len(patch.Entries))
		for j, entry := range patch.Entries {
			managedKeys[j] = model.ManagedKey(entry)
		}

		manifest.AddManagedConfig(model.ManagedConfig{
			Target:       patch.Target,
			BaselineHash: configResults[i].BaselineHash,
			LastModified: time.Now(),
			ManagedKeys:  managedKeys,
		})
	}

	if err := manifest.Save(a.ManifestPath); err != nil {
		return nil, fmt.Errorf("saving manifest: %w", err)
	}

	return &Result{
		StorePath: storePath,
		Patches:   allPatches,
		Backups:   backups,
	}, nil
}

func ComputeDiffs(patches []model.ConfigPatch) ([]*emulators.ConfigDiff, error) {
	diffs := make([]*emulators.ConfigDiff, 0, len(patches))
	for _, patch := range patches {
		diff, err := emulators.ComputeDiff(patch)
		if err != nil {
			return nil, fmt.Errorf("computing diff: %w", err)
		}
		diffs = append(diffs, diff)
	}
	return diffs, nil
}

func (a *Applier) buildEmulatorPackageInfo(emulatorIDs []model.EmulatorID) []launcher.EmulatorPackageInfo {
	seenBinaries := make(map[string]bool)
	var info []launcher.EmulatorPackageInfo

	for _, emuID := range emulatorIDs {
		emu, err := a.Registry.GetEmulator(emuID)
		if err != nil || emu.Launcher.Binary == "" {
			continue
		}

		if seenBinaries[emu.Launcher.Binary] {
			continue
		}
		seenBinaries[emu.Launcher.Binary] = true

		info = append(info, launcher.EmulatorPackageInfo{
			BinaryName: emu.Launcher.Binary,
		})
	}

	return info
}

func (a *Applier) buildDesktopEntries(emulatorIDs []model.EmulatorID, store model.StoreReader) []launcher.GeneratedDesktop {
	seenBinaries := make(map[string]bool)
	var entries []launcher.GeneratedDesktop

	for _, emuID := range emulatorIDs {
		emu, err := a.Registry.GetEmulator(emuID)
		if err != nil || emu.Launcher.Binary == "" {
			continue
		}

		if seenBinaries[emu.Launcher.Binary] {
			continue
		}
		seenBinaries[emu.Launcher.Binary] = true

		displayName := emu.Launcher.DisplayName
		if displayName == "" {
			displayName = emu.Name
		}

		// Check if the config generator provides launch arguments
		var launchArgs string
		if gen := a.Registry.GetConfigGenerator(emuID); gen != nil {
			if provider, ok := gen.(model.LaunchArgsProvider); ok {
				args := provider.LaunchArgs(store)
				launchArgs = strings.Join(args, " ")
			}
		}

		entries = append(entries, launcher.GeneratedDesktop{
			BinaryName:    emu.Launcher.Binary,
			Name:          displayName,
			GenericName:   emu.Launcher.GenericName,
			CategoriesStr: strings.Join(emu.Launcher.Categories, ";"),
			LaunchArgs:    launchArgs,
		})
	}

	return entries
}

// collectEnabledEmulators returns a deduplicated set of emulator IDs from the config.
func (a *Applier) collectEnabledEmulators(cfg *model.KyarabenConfig) map[model.EmulatorID]bool {
	enabled := make(map[model.EmulatorID]bool)
	for _, emulators := range cfg.Systems {
		for _, emuID := range emulators {
			enabled[emuID] = true
		}
	}
	return enabled
}
