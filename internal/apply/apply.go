package apply

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/emulators/symlink"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/packages"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
	"github.com/fnune/kyaraben/internal/version"
	"github.com/fnune/kyaraben/internal/versions"
)

var versionsGet = versions.Get

const installTimeout = 30 * time.Minute

type Progress struct {
	Step            string
	Message         string
	Output          string
	BuildPhase      string // "downloading", "extracting", "installed", "skipped"
	PackageName     string
	ProgressPercent int
}

type Result struct {
	Patches []model.ConfigPatch
	Backups []BackupInfo
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
	Installer       packages.Installer
	ConfigWriter    *emulators.ConfigWriter
	Registry        *registry.Registry
	ManifestPath    string
	LauncherManager *launcher.Manager
	BaseDirResolver model.BaseDirResolver
	SymlinkCreator  model.SymlinkCreator
}

func (a *Applier) Apply(ctx context.Context, cfg *model.KyarabenConfig, userStore *store.UserStore, opts Options) (*Result, error) {
	if opts.OnProgress == nil {
		opts.OnProgress = func(Progress) {}
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
			EnabledSystems:     cfg.EnabledSystems(),
			SystemEmulators:    cfg.Systems,
			GetSystem:          a.Registry.GetSystem,
			GetEmulator:        a.Registry.GetEmulator,
			GetConfigGenerator: a.Registry.GetConfigGenerator,
			Store:              userStore,
			BinDir:             binDir,
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

	installCtx, cancel := context.WithTimeout(ctx, installTimeout)
	defer cancel()

	opts.OnProgress(Progress{Step: "build", Message: "This may take a while on first run"})

	installedBinaries, installedCores, installedIcons, err := a.installPackages(installCtx, emulatorsToInstall, frontendsToInstall, opts)
	if err != nil {
		return nil, err
	}

	opts.OnProgress(Progress{Step: "gc"})
	if err := a.garbageCollect(emulatorsToInstall, frontendsToInstall); err != nil {
		opts.OnProgress(Progress{Step: "gc", Message: "Skipped (cleanup failed, will retry next time)"})
	}

	manifest, err := model.LoadManifest(a.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	if a.LauncherManager != nil {
		opts.OnProgress(Progress{Step: "desktop"})

		launcherBinaries := toLauncherBinaries(installedBinaries)
		if err := a.LauncherManager.GenerateWrappers(launcherBinaries); err != nil {
			return nil, fmt.Errorf("generating launcher wrappers: %w", err)
		}

		launcherCores := toLauncherCores(installedCores)
		if err := a.LauncherManager.GenerateCoreSymlinks(launcherCores); err != nil {
			return nil, fmt.Errorf("generating core symlinks: %w", err)
		}

		previousFiles := &launcher.GeneratedFiles{
			DesktopFiles: manifest.DesktopFiles,
			IconFiles:    manifest.IconFiles,
		}
		desktopEntries := a.buildDesktopEntries(emulatorsToInstall, frontendsToInstall, userStore)
		launcherIcons := toLauncherIcons(installedIcons)
		generatedFiles, err := a.LauncherManager.GenerateDesktopFiles(desktopEntries, launcherIcons, previousFiles)
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

	symlinkCreator := a.SymlinkCreator
	if symlinkCreator == nil {
		symlinkCreator = symlink.OSCreator{}
	}
	for emuID := range enabledEmulators {
		gen := a.Registry.GetConfigGenerator(emuID)
		if provider, ok := gen.(model.SymlinkProvider); ok {
			specs, err := provider.Symlinks(userStore, a.BaseDirResolver)
			if err != nil {
				return nil, fmt.Errorf("getting symlink specs for %s: %w", emuID, err)
			}
			if err := symlink.CreateAll(symlinkCreator, specs); err != nil {
				return nil, fmt.Errorf("creating symlinks for %s: %w", emuID, err)
			}
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	newInstalledEmulators := make(map[model.EmulatorID]model.InstalledEmulator)
	now := time.Now()
	for _, emuID := range emulatorsToInstall {
		emu, err := a.Registry.GetEmulator(emuID)
		if err != nil {
			continue
		}
		resolvedVersion := a.Installer.ResolveVersion(emu.Package.PackageName())
		if resolvedVersion == "" {
			resolvedVersion = "unknown"
		}
		newInstalledEmulators[emuID] = model.InstalledEmulator{
			ID:          emuID,
			Version:     resolvedVersion,
			PackagePath: a.Installer.PackagesDir(),
			Installed:   now,
		}
	}

	newInstalledFrontends := make(map[model.FrontendID]model.InstalledFrontend)
	for _, feID := range frontendsToInstall {
		fe, err := a.Registry.GetFrontend(feID)
		if err != nil {
			continue
		}
		resolvedVersion := a.Installer.ResolveVersion(fe.Package.PackageName())
		if resolvedVersion == "" {
			resolvedVersion = "unknown"
		}
		newInstalledFrontends[feID] = model.InstalledFrontend{
			ID:          feID,
			Version:     resolvedVersion,
			PackagePath: a.Installer.PackagesDir(),
			Installed:   now,
		}
	}

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

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	enabledEmuSet := make(map[model.EmulatorID]bool)
	for _, emuID := range emulatorsToInstall {
		enabledEmuSet[emuID] = true
	}

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
		Patches: allPatches,
		Backups: backups,
	}, nil
}

func (a *Applier) installPackages(ctx context.Context, emulatorIDs []model.EmulatorID, frontendIDs []model.FrontendID, opts Options) ([]packages.InstalledBinary, []packages.InstalledCore, []packages.InstalledIcon, error) {
	seenPackages := make(map[string]bool)
	var binaries []packages.InstalledBinary
	var coreNames []string
	var icons []packages.InstalledIcon

	progressFn := func(p packages.InstallProgress) {
		opts.OnProgress(Progress{
			Step:        "build",
			BuildPhase:  p.Phase,
			PackageName: p.PackageName,
		})
	}

	for _, emuID := range emulatorIDs {
		emu, err := a.Registry.GetEmulator(emuID)
		if err != nil {
			continue
		}

		if strings.HasPrefix(string(emuID), "retroarch:") {
			coreName := strings.TrimPrefix(string(emuID), "retroarch:")
			coreNames = append(coreNames, coreName)
			continue
		}

		pkgName := emu.Package.PackageName()
		if seenPackages[pkgName] {
			continue
		}
		seenPackages[pkgName] = true

		binary, err := a.Installer.InstallEmulator(ctx, pkgName, progressFn)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("installing %s: %w", pkgName, err)
		}
		binaries = append(binaries, *binary)

		if emu.Launcher.Binary != "" {
			icon, err := a.installEmulatorIcon(ctx, emu)
			if err == nil && icon != nil {
				icons = append(icons, *icon)
			}
		}
	}

	for _, feID := range frontendIDs {
		fe, err := a.Registry.GetFrontend(feID)
		if err != nil {
			continue
		}

		pkgName := fe.Package.PackageName()
		if seenPackages[pkgName] {
			continue
		}
		seenPackages[pkgName] = true

		binary, err := a.Installer.InstallEmulator(ctx, pkgName, progressFn)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("installing frontend %s: %w", pkgName, err)
		}
		binaries = append(binaries, *binary)
	}

	if len(coreNames) > 0 {
		if pkgName := "retroarch"; !seenPackages[pkgName] {
			seenPackages[pkgName] = true
			binary, err := a.Installer.InstallEmulator(ctx, pkgName, progressFn)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("installing retroarch: %w", err)
			}
			binaries = append(binaries, *binary)

			emu, err := a.Registry.GetEmulator(model.EmulatorIDRetroArch)
			if err == nil {
				icon, err := a.installEmulatorIcon(ctx, emu)
				if err == nil && icon != nil {
					icons = append(icons, *icon)
				}
			}
		}

		cores, err := a.Installer.InstallCores(ctx, coreNames, progressFn)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("installing cores: %w", err)
		}
		return binaries, cores, icons, nil
	}

	return binaries, nil, icons, nil
}

func (a *Applier) installEmulatorIcon(ctx context.Context, emu model.Emulator) (*packages.InstalledIcon, error) {
	pkgName := emu.Package.PackageName()
	v, err := versionsGet()
	if err != nil {
		return nil, nil
	}
	spec, ok := v.GetEmulator(pkgName)
	if !ok || spec.IconURL == "" {
		return nil, nil
	}
	return a.Installer.InstallIcon(ctx, emu.Launcher.Binary, spec.IconURL, spec.IconSHA256)
}

func (a *Applier) garbageCollect(emulatorIDs []model.EmulatorID, frontendIDs []model.FrontendID) error {
	keep := make(map[string]string)
	for _, emuID := range emulatorIDs {
		emu, err := a.Registry.GetEmulator(emuID)
		if err != nil {
			continue
		}
		pkgName := emu.Package.PackageName()
		keep[pkgName] = a.Installer.ResolveVersion(pkgName)
	}
	for _, feID := range frontendIDs {
		fe, err := a.Registry.GetFrontend(feID)
		if err != nil {
			continue
		}
		pkgName := fe.Package.PackageName()
		keep[pkgName] = a.Installer.ResolveVersion(pkgName)
	}
	if len(keep) > 0 {
		keep["retroarch-cores"] = a.Installer.ResolveVersion("retroarch-cores")
	}
	return a.Installer.GarbageCollect(keep)
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

func (a *Applier) collectEnabledEmulators(cfg *model.KyarabenConfig) map[model.EmulatorID]bool {
	enabled := make(map[model.EmulatorID]bool)
	for _, emulators := range cfg.Systems {
		for _, emuID := range emulators {
			enabled[emuID] = true
		}
	}
	return enabled
}

func toLauncherBinaries(binaries []packages.InstalledBinary) []launcher.InstalledBinary {
	result := make([]launcher.InstalledBinary, len(binaries))
	for i, b := range binaries {
		result[i] = launcher.InstalledBinary{Name: b.Name, Path: b.Path}
	}
	return result
}

func toLauncherCores(cores []packages.InstalledCore) []launcher.InstalledCore {
	result := make([]launcher.InstalledCore, len(cores))
	for i, c := range cores {
		result[i] = launcher.InstalledCore{Filename: c.Filename, Path: c.Path}
	}
	return result
}

func toLauncherIcons(icons []packages.InstalledIcon) []launcher.InstalledIcon {
	result := make([]launcher.InstalledIcon, len(icons))
	for i, ic := range icons {
		result[i] = launcher.InstalledIcon{Name: ic.Name, Filename: ic.Filename, Path: ic.Path}
	}
	return result
}
