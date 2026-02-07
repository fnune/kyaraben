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
	"github.com/fnune/kyaraben/internal/version"
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
	patchEmulators := make([]model.EmulatorID, 0)

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
		for range patches {
			patchEmulators = append(patchEmulators, emuID)
		}
		allPatches = append(allPatches, patches...)
	}

	enabledFrontends := cfg.EnabledFrontends()
	frontendsToInstall := make([]model.FrontendID, 0, len(enabledFrontends))

	var binDir string
	if a.LauncherManager != nil {
		binDir = a.LauncherManager.BinDir()
	}

	for _, feID := range enabledFrontends {
		frontendsToInstall = append(frontendsToInstall, feID)

		gen := a.Registry.GetFrontendConfigGenerator(feID)
		if gen == nil {
			continue
		}

		frontendCtx := model.FrontendContext{
			EnabledSystems:  cfg.EnabledSystems(),
			SystemEmulators: cfg.Systems,
			GetSystem:       a.Registry.GetSystem,
			GetEmulator:     a.Registry.GetEmulator,
			Store:           userStore,
			BinDir:          binDir,
		}

		patches, err := gen.Generate(frontendCtx)
		if err != nil {
			return nil, fmt.Errorf("generating config for frontend %s: %w", feID, err)
		}
		allPatches = append(allPatches, patches...)
	}

	var summaryParts []string
	if len(emulatorsToInstall) > 0 {
		var emulatorNames []string
		for _, emuID := range emulatorsToInstall {
			if emu, err := a.Registry.GetEmulator(emuID); err == nil {
				emulatorNames = append(emulatorNames, emu.Name)
			}
		}
		summaryParts = append(summaryParts, strings.Join(emulatorNames, ", "))
	}
	if len(frontendsToInstall) > 0 {
		var frontendNames []string
		for _, feID := range frontendsToInstall {
			if fe, err := a.Registry.GetFrontend(feID); err == nil {
				frontendNames = append(frontendNames, fe.Name)
			}
		}
		summaryParts = append(summaryParts, strings.Join(frontendNames, ", "))
	}
	if len(summaryParts) > 0 {
		opts.OnProgress(Progress{
			Step:    "summary",
			Message: "Enabling " + strings.Join(summaryParts, ", "),
		})
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

	for sys, emulatorIDs := range cfg.Systems {
		for _, emuID := range emulatorIDs {
			emu, err := a.Registry.GetEmulator(emuID)
			if err != nil {
				continue
			}
			if err := userStore.InitializeForEmulator(sys, emuID, emu.PathUsage); err != nil {
				return nil, fmt.Errorf("initializing %s for %s: %w", sys, emuID, err)
			}
		}
	}

	if err := a.NixClient.EnsureFlakeDir(); err != nil {
		return nil, fmt.Errorf("creating flake directory: %w", err)
	}

	genResult, err := a.FlakeGenerator.Generate(a.NixClient.GetFlakePath(), emulatorsToInstall, frontendsToInstall)
	if err != nil {
		return nil, fmt.Errorf("generating flake: %w", err)
	}

	for _, skipped := range genResult.SkippedEmulators {
		opts.OnProgress(Progress{Step: "build", Message: fmt.Sprintf("Warning: emulator '%s' is no longer supported and will be skipped", skipped)})
	}
	for _, skipped := range genResult.SkippedFrontends {
		opts.OnProgress(Progress{Step: "build", Message: fmt.Sprintf("Warning: frontend '%s' is no longer supported and will be skipped", skipped)})
	}

	resolvedVersions := a.FlakeGenerator.GetResolvedVersions(emulatorsToInstall)
	resolvedFrontendVersions := a.FlakeGenerator.GetResolvedFrontendVersions(frontendsToInstall)

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

	opts.OnProgress(Progress{Step: "build", Message: "This may take a while on first run"})

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
		packageInfo := a.buildPackageInfo(emulatorsToInstall, frontendsToInstall)
		if err := a.LauncherManager.GenerateWrappers(packageInfo); err != nil {
			return nil, fmt.Errorf("generating launcher wrappers: %w", err)
		}
		if err := a.LauncherManager.GenerateCoreSymlinks(); err != nil {
			return nil, fmt.Errorf("generating core symlinks: %w", err)
		}

		previousFiles := &launcher.GeneratedFiles{
			DesktopFiles: manifest.DesktopFiles,
			IconFiles:    manifest.IconFiles,
		}
		desktopEntries := a.buildDesktopEntries(emulatorsToInstall, frontendsToInstall, userStore)
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

	// Check for cancellation before committing manifest changes.
	// This prevents saving partial state if the user cancelled during config application.
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	newInstalledEmulators := make(map[model.EmulatorID]model.InstalledEmulator)
	now := time.Now()
	for _, emuID := range emulatorsToInstall {
		version := resolvedVersions[emuID]
		if version == "" {
			version = "unknown"
		}
		newInstalledEmulators[emuID] = model.InstalledEmulator{
			ID:        emuID,
			Version:   version,
			StorePath: storePath,
			Installed: now,
		}
	}

	newInstalledFrontends := make(map[model.FrontendID]model.InstalledFrontend)
	for _, feID := range frontendsToInstall {
		version := resolvedFrontendVersions[feID]
		if version == "" {
			version = "unknown"
		}
		newInstalledFrontends[feID] = model.InstalledFrontend{
			ID:        feID,
			Version:   version,
			StorePath: storePath,
			Installed: now,
		}
	}

	// Build new managed configs list (emulator patches only, not frontend patches)
	emulatorPatchCount := len(patchEmulators)
	newManagedConfigs := make([]model.ManagedConfig, 0, emulatorPatchCount)
	for i, patch := range allPatches {
		if i >= emulatorPatchCount {
			break
		}
		if patch.Target.BaseDir == model.ConfigBaseDirOpaqueDir {
			continue
		}

		var managedKeys []model.ManagedKey
		for _, entry := range patch.Entries {
			if entry.Unmanaged {
				continue
			}
			managedKeys = append(managedKeys, model.ManagedKey{
				Path:  entry.Path,
				Value: entry.Value,
			})
		}

		newManagedConfigs = append(newManagedConfigs, model.ManagedConfig{
			EmulatorIDs:  []model.EmulatorID{patchEmulators[i]},
			Target:       patch.Target,
			BaselineHash: configResults[i].BaselineHash,
			LastModified: now,
			ManagedKeys:  managedKeys,
		})
	}

	// Final cancellation check before committing to disk
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Build set of enabled emulators for filtering stale configs
	enabledEmuSet := make(map[model.EmulatorID]bool)
	for _, emuID := range emulatorsToInstall {
		enabledEmuSet[emuID] = true
	}

	// Filter out ManagedConfigs where no emulator is still enabled
	filteredConfigs := make([]model.ManagedConfig, 0, len(manifest.ManagedConfigs))
	for _, cfg := range manifest.ManagedConfigs {
		for _, emuID := range cfg.EmulatorIDs {
			if enabledEmuSet[emuID] {
				filteredConfigs = append(filteredConfigs, cfg)
				break
			}
		}
	}
	manifest.ManagedConfigs = filteredConfigs

	manifest.LastApplied = now
	manifest.KyarabenVersion = version.Get()
	manifest.InstalledEmulators = newInstalledEmulators
	manifest.InstalledFrontends = newInstalledFrontends
	for _, cfg := range newManagedConfigs {
		if err := manifest.AddManagedConfig(cfg); err != nil {
			return nil, fmt.Errorf("adding managed config: %w", err)
		}
	}

	if err := manifest.SaveWithBackup(a.ManifestPath); err != nil {
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

func (a *Applier) buildPackageInfo(emulatorIDs []model.EmulatorID, frontendIDs []model.FrontendID) []launcher.EmulatorPackageInfo {
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

	for _, feID := range frontendIDs {
		fe, err := a.Registry.GetFrontend(feID)
		if err != nil || fe.Launcher.Binary == "" {
			continue
		}

		if seenBinaries[fe.Launcher.Binary] {
			continue
		}
		seenBinaries[fe.Launcher.Binary] = true

		info = append(info, launcher.EmulatorPackageInfo{
			BinaryName: fe.Launcher.Binary,
		})
	}

	return info
}

func (a *Applier) buildDesktopEntries(emulatorIDs []model.EmulatorID, frontendIDs []model.FrontendID, store model.StoreReader) []launcher.GeneratedDesktop {
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

	for _, feID := range frontendIDs {
		fe, err := a.Registry.GetFrontend(feID)
		if err != nil || fe.Launcher.Binary == "" {
			continue
		}

		if seenBinaries[fe.Launcher.Binary] {
			continue
		}
		seenBinaries[fe.Launcher.Binary] = true

		displayName := fe.Launcher.DisplayName
		if displayName == "" {
			displayName = fe.Name
		}

		entries = append(entries, launcher.GeneratedDesktop{
			BinaryName:    fe.Launcher.Binary,
			Name:          displayName,
			GenericName:   fe.Launcher.GenericName,
			CategoriesStr: strings.Join(fe.Launcher.Categories, ";"),
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
