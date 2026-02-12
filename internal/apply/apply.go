package apply

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/emulators/symlink"
	"github.com/fnune/kyaraben/internal/hardware"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/logging"
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
	BytesDownloaded int64
	BytesTotal      int64
	BytesPerSecond  int64
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

	logging.SetOutputHook(func(line string) {
		opts.OnProgress(Progress{Step: "build", Output: line})
	})
	defer logging.SetOutputHook(nil)

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
			if err := ensureProvisionDirs(userStore, sys, emu); err != nil {
				return nil, fmt.Errorf("preparing provision directories for %s/%s: %w", sys, emuID, err)
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

	opts.OnProgress(Progress{Step: "cleanup", Message: "Removing unused package versions"})
	if err := a.garbageCollect(emulatorsToInstall, frontendsToInstall); err != nil {
		opts.OnProgress(Progress{Step: "cleanup", Message: "Skipped (cleanup failed, will retry next time)"})
	}

	manifest, err := model.LoadManifest(a.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	opts.OnProgress(Progress{Step: "finalize", Message: "Creating desktop entries and emulator configs"})

	if a.LauncherManager != nil {

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
	installPlan := a.buildInstallPlan(emulatorIDs, frontendIDs)
	packageDownloadSizes, packageTotals, totalBytes, packageArchiveTypes := a.buildPackageSizes(installPlan)
	aggregator := newProgressAggregator(packageTotals, packageDownloadSizes, packageArchiveTypes, totalBytes, opts.OnProgress)

	seenPackages := make(map[string]bool)
	var binaries []packages.InstalledBinary
	var coreNames []string
	var cores []packages.InstalledCore
	var icons []packages.InstalledIcon

	progressFn := func(p packages.InstallProgress) {
		aggregator.onPackageProgress(p)
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

		icon, err := a.installIconForPackage(ctx, pkgName, fe.Launcher.Binary)
		if err == nil && icon != nil {
			icons = append(icons, *icon)
		}
	}

	if len(coreNames) > 0 {
		if pkgName := "retroarch"; !seenPackages[pkgName] {
			seenPackages[pkgName] = true
			binary, err := a.Installer.InstallEmulator(ctx, pkgName, progressFn)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("installing retroarch: %w", err)
			}
			binaries = append(binaries, *binary)

			icon, err := a.installIconForPackage(ctx, "retroarch", "retroarch")
			if err == nil && icon != nil {
				icons = append(icons, *icon)
			}
		}

		var coreErr error
		cores, coreErr = a.Installer.InstallCores(ctx, coreNames, progressFn)
		if coreErr != nil {
			return nil, nil, nil, fmt.Errorf("installing cores: %w", coreErr)
		}
	}

	return binaries, cores, icons, nil
}

func ensureProvisionDirs(userStore *store.UserStore, sys model.SystemID, emu model.Emulator) error {
	for _, group := range emu.ProvisionGroups {
		baseDir := group.BaseDirFor(userStore, sys)
		if baseDir == "" {
			continue
		}
		if err := os.MkdirAll(baseDir, 0o755); err != nil {
			return fmt.Errorf("creating %s: %w", baseDir, err)
		}
	}
	return nil
}

type installPlan struct {
	packageNames []string
	coreNames    []string
}

func (a *Applier) buildInstallPlan(emulatorIDs []model.EmulatorID, frontendIDs []model.FrontendID) installPlan {
	plan := installPlan{
		packageNames: make([]string, 0),
		coreNames:    make([]string, 0),
	}
	seenPackages := make(map[string]bool)

	for _, emuID := range emulatorIDs {
		emu, err := a.Registry.GetEmulator(emuID)
		if err != nil {
			continue
		}

		if strings.HasPrefix(string(emuID), "retroarch:") {
			coreName := strings.TrimPrefix(string(emuID), "retroarch:")
			plan.coreNames = append(plan.coreNames, coreName)
			continue
		}

		pkgName := emu.Package.PackageName()
		if seenPackages[pkgName] {
			continue
		}
		seenPackages[pkgName] = true
		plan.packageNames = append(plan.packageNames, pkgName)
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
		plan.packageNames = append(plan.packageNames, pkgName)
	}

	if len(plan.coreNames) > 0 {
		if !seenPackages["retroarch"] {
			seenPackages["retroarch"] = true
			plan.packageNames = append(plan.packageNames, "retroarch")
		}
	}

	return plan
}

func (a *Applier) buildPackageSizes(plan installPlan) (map[string]int64, map[string]int64, int64, map[string]string) {
	v := versions.MustGet()
	targetName := hardware.DetectTarget().Name

	installed := make(map[string]bool)
	packageDownloadSizes := make(map[string]int64)
	packageTotals := make(map[string]int64)
	packageArchiveTypes := make(map[string]string)
	installList := make([]string, 0, len(plan.packageNames)+1)

	for _, pkgName := range plan.packageNames {
		installList = append(installList, pkgName)
		installed[pkgName] = a.Installer.IsEmulatorInstalled(pkgName)
		downloadSize := packageDownloadSize(pkgName, targetName, v, a.Installer)
		packageDownloadSizes[pkgName] = downloadSize
		packageArchiveTypes[pkgName] = packageArchiveType(pkgName, targetName, v, a.Installer)
	}

	if len(plan.coreNames) > 0 {
		installList = append(installList, "retroarch-cores")
		installed["retroarch-cores"] = packages.RetroArchCoresInstalled(a.Installer.PackagesDir(), plan.coreNames, v)
		packageDownloadSizes["retroarch-cores"] = coreDownloadSize(plan.coreNames, targetName, v)
		packageArchiveTypes["retroarch-cores"] = coreArchiveType(targetName, v)
	}

	summary := packages.CalculateChangeSummary(installList, nil, installed, func(pkgName string) int64 {
		return packageDownloadSizes[pkgName]
	})

	downloadSizes := make(map[string]int64, len(summary.PackagesToDownload))
	var totalBytes int64
	for _, pkgName := range summary.PackagesToDownload {
		downloadSize := packageDownloadSizes[pkgName]
		downloadSizes[pkgName] = downloadSize
		archiveType := packageArchiveTypes[pkgName]
		totalSize := downloadSize
		if archiveType != "" {
			totalSize += extractionWeight(downloadSize)
		}
		packageTotals[pkgName] = totalSize
		totalBytes += totalSize
	}

	return downloadSizes, packageTotals, totalBytes, packageArchiveTypes
}

func packageDownloadSize(pkgName string, targetName string, v *versions.Versions, installer packages.Installer) int64 {
	spec, ok := v.GetEmulator(pkgName)
	if !ok {
		return 0
	}
	version := installer.ResolveVersion(pkgName)
	entry := spec.GetVersion(version)
	if entry == nil {
		entry = spec.GetDefault()
	}
	if entry == nil {
		return 0
	}
	target := entry.SelectTarget(targetName)
	if target == "" {
		return 0
	}
	build := entry.Target(target)
	if build == nil {
		return 0
	}
	return build.Size
}

func coreDownloadSize(coreNames []string, targetName string, v *versions.Versions) int64 {
	if len(coreNames) == 0 {
		return 0
	}

	version := v.RetroArchCores.Default
	if version == "" {
		return 0
	}
	build, ok := v.RetroArchCores.Versions[version]
	if !ok {
		return 0
	}

	selectedTarget := selectCoresTargetName(build, targetName)
	if selectedTarget != "" {
		if targetBuild, ok := build.Targets[selectedTarget]; ok && targetBuild.Size > 0 {
			return targetBuild.Size
		}
	}

	var total int64
	for _, coreName := range coreNames {
		total += v.GetCoreSize(coreName)
	}
	return total
}

func packageArchiveType(pkgName string, targetName string, v *versions.Versions, installer packages.Installer) string {
	spec, ok := v.GetEmulator(pkgName)
	if !ok {
		return ""
	}
	version := installer.ResolveVersion(pkgName)
	entry := spec.GetVersion(version)
	if entry == nil {
		entry = spec.GetDefault()
	}
	if entry == nil {
		return ""
	}
	target := entry.SelectTarget(targetName)
	if target == "" {
		return ""
	}
	return entry.ArchiveType(target, spec)
}

func coreArchiveType(targetName string, v *versions.Versions) string {
	version := v.RetroArchCores.Default
	if version == "" {
		return ""
	}
	build, ok := v.RetroArchCores.Versions[version]
	if !ok {
		return ""
	}

	selectedTarget := selectCoresTargetName(build, targetName)
	if selectedTarget == "" {
		return ""
	}

	url, _, ok := v.RetroArchCores.GetCoresURL(selectedTarget)
	if !ok {
		return ""
	}
	return archiveTypeFromURL(url)
}

func archiveTypeFromURL(url string) string {
	switch {
	case strings.HasSuffix(url, ".tar.zst"):
		return "tar.zst"
	case strings.HasSuffix(url, ".tar.gz"), strings.HasSuffix(url, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(url, ".tar.xz"):
		return "tar.xz"
	case strings.HasSuffix(url, ".zip"):
		return "zip"
	case strings.HasSuffix(url, ".7z"):
		return "7z"
	default:
		return ""
	}
}

func selectCoresTargetName(build versions.RetroArchCoresBuild, detected string) string {
	if _, ok := build.Targets[detected]; ok {
		return detected
	}

	if fallback, ok := versions.TargetFallback[detected]; ok {
		if _, ok := build.Targets[fallback.String()]; ok {
			return fallback.String()
		}
	}

	return ""
}

func extractionWeight(downloadSize int64) int64 {
	const (
		minWeight = int64(12 * 1024 * 1024)
		maxWeight = int64(160 * 1024 * 1024)
	)
	if downloadSize <= 0 {
		return minWeight
	}
	weight := downloadSize / 2
	if weight < minWeight {
		return minWeight
	}
	if weight > maxWeight {
		return maxWeight
	}
	return weight
}

type progressAggregator struct {
	packageTotals     map[string]int64
	downloadTotals    map[string]int64
	extractWeights    map[string]int64
	completedBytes    int64
	packageBytes      map[string]int64
	extractBytes      map[string]int64
	completedPackages map[string]bool
	totalBytes        int64
	lastReportedBytes int64
	speedTracker      *packages.SpeedTracker
	onProgress        func(Progress)
	extracting        map[string]*extractionProgress
	extractTickerStop chan struct{}
	mu                sync.Mutex
}

type extractionProgress struct {
	start    time.Time
	weight   int64
	duration time.Duration
}

func newProgressAggregator(packageTotals map[string]int64, downloadTotals map[string]int64, packageArchiveTypes map[string]string, totalBytes int64, onProgress func(Progress)) *progressAggregator {
	extractWeights := make(map[string]int64, len(packageTotals))
	for pkgName, total := range packageTotals {
		downloadSize := downloadTotals[pkgName]
		weight := total - downloadSize
		if weight <= 0 && packageArchiveTypes[pkgName] != "" {
			weight = extractionWeight(downloadSize)
		}
		if weight > 0 {
			extractWeights[pkgName] = weight
		}
	}
	return &progressAggregator{
		packageTotals:     packageTotals,
		downloadTotals:    downloadTotals,
		extractWeights:    extractWeights,
		completedBytes:    0,
		packageBytes:      make(map[string]int64),
		extractBytes:      make(map[string]int64),
		completedPackages: make(map[string]bool),
		totalBytes:        totalBytes,
		lastReportedBytes: 0,
		speedTracker:      packages.NewSpeedTracker(3 * time.Second),
		onProgress:        onProgress,
		extracting:        make(map[string]*extractionProgress),
	}
}

func (a *progressAggregator) onPackageProgress(p packages.InstallProgress) {
	if a == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if p.PackageName == "" {
		return
	}

	if p.BytesDownloaded > 0 {
		a.packageBytes[p.PackageName] = p.BytesDownloaded
	}

	if p.Phase == "skipped" {
		a.markSkipped(p.PackageName)
		a.emitProgressLocked(p)
		return
	}

	if p.Phase == "extracting" {
		a.startExtractionLocked(p.PackageName)
	}

	if p.Phase == "installed" {
		a.markCompleted(p.PackageName)
	}

	a.emitProgressLocked(p)
}

func (a *progressAggregator) markSkipped(pkgName string) {
	if a.completedPackages[pkgName] {
		return
	}
	a.completedPackages[pkgName] = true
	if expected, ok := a.packageTotals[pkgName]; ok {
		a.totalBytes -= expected
		delete(a.packageTotals, pkgName)
	}
	delete(a.packageBytes, pkgName)
	delete(a.downloadTotals, pkgName)
	delete(a.extractBytes, pkgName)
	delete(a.extractWeights, pkgName)
	delete(a.extracting, pkgName)
	a.maybeStopExtractionTickerLocked()
}

func (a *progressAggregator) markCompleted(pkgName string) {
	if a.completedPackages[pkgName] {
		return
	}
	a.completedPackages[pkgName] = true
	a.completedBytes += a.packageBytes[pkgName] + a.extractBytes[pkgName]
	delete(a.packageBytes, pkgName)
	delete(a.extractBytes, pkgName)
	delete(a.extracting, pkgName)
	a.maybeStopExtractionTickerLocked()
}

func (a *progressAggregator) emitProgressLocked(p packages.InstallProgress) {
	var inFlight int64
	for pkgName, bytes := range a.packageBytes {
		inFlight += bytes + a.extractBytes[pkgName]
	}
	bytesDownloaded := a.completedBytes + inFlight
	if a.totalBytes > 0 && bytesDownloaded > a.totalBytes {
		bytesDownloaded = a.totalBytes
	}

	if bytesDownloaded > a.lastReportedBytes {
		a.speedTracker.Record(bytesDownloaded - a.lastReportedBytes)
		a.lastReportedBytes = bytesDownloaded
	}

	percent := 0
	if a.totalBytes > 0 {
		percent = int(bytesDownloaded * 100 / a.totalBytes)
		if percent > 100 {
			percent = 100
		}
	}

	a.onProgress(Progress{
		Step:            "build",
		BuildPhase:      p.Phase,
		PackageName:     p.PackageName,
		ProgressPercent: percent,
		BytesDownloaded: bytesDownloaded,
		BytesTotal:      a.totalBytes,
		BytesPerSecond:  a.speedTracker.BytesPerSecond(),
	})

}

func (a *progressAggregator) startExtractionLocked(pkgName string) {
	if a.completedPackages[pkgName] {
		return
	}
	if _, ok := a.extracting[pkgName]; ok {
		return
	}
	weight := a.extractWeights[pkgName]
	if weight <= 0 {
		return
	}
	duration := extractionDuration(a.downloadTotals[pkgName])
	a.extracting[pkgName] = &extractionProgress{
		start:    time.Now(),
		weight:   weight,
		duration: duration,
	}
	a.startExtractionTickerLocked()
}

func (a *progressAggregator) startExtractionTickerLocked() {
	if a.extractTickerStop != nil {
		return
	}
	stop := make(chan struct{})
	a.extractTickerStop = stop
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.advanceExtraction()
			case <-stop:
				return
			}
		}
	}()
}

func (a *progressAggregator) maybeStopExtractionTickerLocked() {
	if len(a.extracting) != 0 {
		return
	}
	if a.extractTickerStop != nil {
		close(a.extractTickerStop)
		a.extractTickerStop = nil
	}
}

func (a *progressAggregator) advanceExtraction() {
	a.mu.Lock()
	defer a.mu.Unlock()

	updated := false
	for pkgName, progress := range a.extracting {
		elapsed := time.Since(progress.start)
		if progress.duration <= 0 {
			progress.duration = time.Second
		}
		next := int64(float64(progress.weight) * elapsed.Seconds() / progress.duration.Seconds())
		if next > progress.weight {
			next = progress.weight
		}
		if next < 0 {
			next = 0
		}
		if next > a.extractBytes[pkgName] {
			a.extractBytes[pkgName] = next
			updated = true
		}
	}

	if updated {
		a.emitProgressLocked(packages.InstallProgress{Phase: "extracting"})
	}
}

func extractionDuration(downloadSize int64) time.Duration {
	const (
		minDuration = 6 * time.Second
		maxDuration = 45 * time.Second
	)
	if downloadSize <= 0 {
		return minDuration
	}
	seconds := time.Duration(downloadSize/(20*1024*1024)) * time.Second
	if seconds < minDuration {
		return minDuration
	}
	if seconds > maxDuration {
		return maxDuration
	}
	return seconds
}

var log = logging.New("apply")

func (a *Applier) installEmulatorIcon(ctx context.Context, emu model.Emulator) (*packages.InstalledIcon, error) {
	pkgName := emu.Package.PackageName()
	return a.installIconForPackage(ctx, pkgName, emu.Launcher.Binary)
}

func (a *Applier) installIconForPackage(ctx context.Context, pkgName, binaryName string) (*packages.InstalledIcon, error) {
	v, err := versionsGet()
	if err != nil {
		log.Debug("Failed to get versions for icon: %v", err)
		return nil, nil
	}
	spec, ok := v.GetEmulator(pkgName)
	if !ok {
		log.Debug("No version spec found for package %s", pkgName)
		return nil, nil
	}
	if spec.IconURL == "" {
		log.Debug("No icon URL configured for package %s", pkgName)
		return nil, nil
	}
	icon, err := a.Installer.InstallIcon(ctx, binaryName, spec.IconURL, spec.IconSHA256)
	if err != nil {
		log.Debug("Failed to install icon for %s: %v", pkgName, err)
		return nil, err
	}
	return icon, nil
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
