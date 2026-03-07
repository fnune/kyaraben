package apply

import (
	"context"
	"fmt"
	"hash/crc32"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/cleanup"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/emulators/symlink"
	"github.com/fnune/kyaraben/internal/hardware"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/packages"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/steam"
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
	LogEntry        *logging.LogEntry
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

func (a *Applier) Preflight(ctx context.Context, cfg *model.KyarabenConfig, collection *store.Collection) (*PreflightResult, error) {
	allPatches := make([]model.ConfigPatch, 0)

	controllerConfig, err := cfg.ResolveControllerConfig()
	if err != nil {
		return nil, fmt.Errorf("resolving controller config: %w", err)
	}

	systemDisplayTypes := a.buildSystemDisplayTypes()

	genCtx := model.GenerateContext{
		Store:              collection,
		BaseDirResolver:    a.BaseDirResolver,
		ControllerConfig:   controllerConfig,
		SystemDisplayTypes: systemDisplayTypes,
		TargetDevice:       cfg.GraphicsTarget(),
	}

	for emuID := range a.collectEnabledEmulators(cfg) {
		gen := a.Registry.GetConfigGenerator(emuID)
		if gen == nil {
			continue
		}

		emu, _ := a.Registry.GetEmulator(emuID)
		genCtx.Preset = cfg.EmulatorPreset(emuID)
		genCtx.Resume = cfg.EmulatorResume(emuID, emu.ResumeRecommended)
		result, err := gen.Generate(genCtx)
		if err != nil {
			return nil, fmt.Errorf("generating config for %s: %w", emuID, err)
		}
		allPatches = append(allPatches, result.Patches...)
	}

	manifest, err := a.manifestStore.Load(a.ManifestPath)
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
	fs              vfs.FS
	manifestStore   *model.ManifestStore
	Installer       packages.Installer
	ConfigWriter    *emulators.ConfigWriter
	Registry        *registry.Registry
	ManifestPath    string
	LauncherManager *launcher.Manager
	BaseDirResolver model.BaseDirResolver
	SymlinkCreator  model.SymlinkCreator
	SteamManager    *steam.Manager
}

func NewApplier(fs vfs.FS, installer packages.Installer, configWriter *emulators.ConfigWriter, reg *registry.Registry, manifestPath string, launcherManager *launcher.Manager, resolver model.BaseDirResolver, symlinkCreator model.SymlinkCreator) *Applier {
	return &Applier{
		fs:              fs,
		manifestStore:   model.NewManifestStore(fs),
		Installer:       installer,
		ConfigWriter:    configWriter,
		Registry:        reg,
		ManifestPath:    manifestPath,
		LauncherManager: launcherManager,
		BaseDirResolver: resolver,
		SymlinkCreator:  symlinkCreator,
	}
}

func NewDefaultApplier(installer packages.Installer, configWriter *emulators.ConfigWriter, reg *registry.Registry, manifestPath string, launcherManager *launcher.Manager, resolver model.BaseDirResolver, symlinkCreator model.SymlinkCreator) *Applier {
	return NewApplier(vfs.OSFS, installer, configWriter, reg, manifestPath, launcherManager, resolver, symlinkCreator)
}

func (a *Applier) Apply(ctx context.Context, cfg *model.KyarabenConfig, collection *store.Collection, opts Options) (*Result, error) {
	if opts.OnProgress == nil {
		opts.OnProgress = func(Progress) {}
	}

	logging.SetUICallback(func(entry logging.LogEntry) {
		opts.OnProgress(Progress{Step: "build", Output: entry.Message, LogEntry: &entry})
	})
	defer logging.SetUICallback(nil)

	ctx = logging.WithUISession(ctx)

	controllerConfig, err := cfg.ResolveControllerConfig()
	if err != nil {
		return nil, fmt.Errorf("resolving controller config: %w", err)
	}

	systemDisplayTypes := a.buildSystemDisplayTypes()

	genCtx := model.GenerateContext{
		Store:              collection,
		BaseDirResolver:    a.BaseDirResolver,
		ControllerConfig:   controllerConfig,
		SystemDisplayTypes: systemDisplayTypes,
		TargetDevice:       cfg.GraphicsTarget(),
	}

	enabledEmulators := a.collectEnabledEmulators(cfg)
	emulatorsToInstall := make([]model.EmulatorID, 0, len(enabledEmulators))
	allPatches := make([]model.ConfigPatch, 0)
	patchEmulators := make([]model.EmulatorID, 0)
	allSymlinks := make(map[model.EmulatorID][]model.SymlinkSpec)
	allLaunchArgs := make(map[model.EmulatorID][]string)
	allInitialDownloads := make([]model.InitialDownload, 0)
	allEmbeddedFiles := make([]model.EmbeddedFile, 0)
	binaryEnvs := make(map[string]map[string]string)

	for emuID := range enabledEmulators {
		emulatorsToInstall = append(emulatorsToInstall, emuID)

		if emu, err := a.Registry.GetEmulator(emuID); err == nil && len(emu.Launcher.Env) > 0 {
			binaryEnvs[emu.Launcher.Binary] = emu.Launcher.Env
		}

		gen := a.Registry.GetConfigGenerator(emuID)
		if gen == nil {
			continue
		}

		emu, _ := a.Registry.GetEmulator(emuID)
		genCtx.Preset = cfg.EmulatorPreset(emuID)
		genCtx.Resume = cfg.EmulatorResume(emuID, emu.ResumeRecommended)
		result, err := gen.Generate(genCtx)
		if err != nil {
			return nil, fmt.Errorf("generating config for %s: %w", emuID, err)
		}
		for range result.Patches {
			patchEmulators = append(patchEmulators, emuID)
		}
		allPatches = append(allPatches, result.Patches...)
		if len(result.Symlinks) > 0 {
			allSymlinks[emuID] = result.Symlinks
		}
		if len(result.LaunchArgs) > 0 {
			allLaunchArgs[emuID] = result.LaunchArgs
		}
		allInitialDownloads = append(allInitialDownloads, result.InitialDownloads...)
		allEmbeddedFiles = append(allEmbeddedFiles, result.EmbeddedFiles...)
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
			GetLaunchArgs: func(emuID model.EmulatorID) []string {
				return allLaunchArgs[emuID]
			},
			Store:  collection,
			BinDir: binDir,
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
	if collection.Exists() {
		storeMsg = fmt.Sprintf("Using %s (existing data preserved)", collection.Path())
	} else {
		storeMsg = fmt.Sprintf("Creating %s", collection.Path())
	}
	opts.OnProgress(Progress{Step: "store", Message: storeMsg})

	if err := collection.Initialize(); err != nil {
		return nil, fmt.Errorf("initializing user store: %w", err)
	}

	for sys, emulatorIDs := range cfg.Systems {
		for _, emuID := range emulatorIDs {
			emu, err := a.Registry.GetEmulator(emuID)
			if err != nil {
				continue
			}
			if err := collection.InitializeForEmulator(sys, emuID, emu.PathUsage); err != nil {
				return nil, fmt.Errorf("initializing %s for %s: %w", sys, emuID, err)
			}
			if err := a.ensureProvisionDirs(collection, sys, emu); err != nil {
				return nil, fmt.Errorf("preparing provision directories for %s/%s: %w", sys, emuID, err)
			}
			if err := a.downloadMissingProvisions(ctx, collection, sys, emu, opts.OnProgress); err != nil {
				return nil, fmt.Errorf("downloading provisions for %s/%s: %w", sys, emuID, err)
			}
		}
	}

	if err := a.downloadInitialFiles(ctx, allInitialDownloads); err != nil {
		return nil, fmt.Errorf("downloading initial files: %w", err)
	}

	if err := a.writeEmbeddedFiles(allEmbeddedFiles); err != nil {
		return nil, fmt.Errorf("writing embedded files: %w", err)
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

	manifest, err := a.manifestStore.Load(a.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	opts.OnProgress(Progress{Step: "finalize", Message: "Creating desktop entries and emulator configs"})

	var installedIconPaths map[string]string
	if a.LauncherManager != nil {
		launcherBinaries := toLauncherBinaries(installedBinaries, binaryEnvs)
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
		desktopEntries := a.buildDesktopEntries(emulatorsToInstall, frontendsToInstall, allLaunchArgs)
		launcherIcons := toLauncherIcons(installedIcons)
		generatedFiles, err := a.LauncherManager.GenerateDesktopFiles(desktopEntries, launcherIcons, previousFiles)
		if err != nil {
			return nil, fmt.Errorf("generating desktop files: %w", err)
		}
		manifest.DesktopFiles = generatedFiles.DesktopFiles
		manifest.IconFiles = generatedFiles.IconFiles
		installedIconPaths = generatedFiles.InstalledIconPaths
	}

	steamShortcuts := a.syncSteamShortcuts(ctx, frontendsToInstall, binDir, installedIconPaths, manifest)
	manifest.SteamShortcuts = steamShortcuts

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
		symlinkCreator = symlink.NewDefaultCreator()
	}

	var newSymlinks []model.SymlinkRecord
	for emuID, specs := range allSymlinks {
		if err := symlink.CreateAll(symlinkCreator, specs); err != nil {
			return nil, fmt.Errorf("creating symlinks for %s: %w", emuID, err)
		}
		for _, spec := range specs {
			newSymlinks = append(newSymlinks, model.SymlinkRecord{
				Source:     spec.Source,
				Target:     spec.Target,
				EmulatorID: emuID,
			})
		}
	}

	for _, old := range manifest.Symlinks {
		if !enabledEmulators[old.EmulatorID] {
			_ = symlink.Remove(a.fs, old.Source)
		}
	}

	var disabledConfigs []model.ManagedConfig
	for _, cfg := range manifest.ManagedConfigs {
		allDisabled := true
		for _, emuID := range cfg.EmulatorIDs {
			if enabledEmulators[emuID] {
				allDisabled = false
				break
			}
		}
		if allDisabled {
			disabledConfigs = append(disabledConfigs, cfg)
		}
	}
	cleaner := cleanup.New(a.fs, a.BaseDirResolver)
	cleaner.RemoveConfigDirs(disabledConfigs)

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
		var resolvedVersion string
		if coreName := emuID.RetroArchCoreName(); coreName != "" {
			resolvedVersion = a.Installer.ResolveVersion(coreName)
		} else {
			resolvedVersion = a.Installer.ResolveVersion(emu.Package.PackageName())
		}
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

		var regions model.ManagedRegions
		for _, r := range patch.ManagedRegions {
			regions = append(regions, r)
		}

		configInputs := collectConfigInputs(patch.Entries, cfg, controllerConfig, collection.Root())

		newManagedConfigs = append(newManagedConfigs, model.ManagedConfig{
			EmulatorIDs:             []model.EmulatorID{patchEmulators[i]},
			Target:                  patch.Target,
			WrittenEntries:          configResults[i].WrittenEntries,
			ConfigInputsWhenWritten: configInputs,
			LastModified:            now,
			ManagedRegions:          regions,
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
	manifest.Symlinks = newSymlinks
	for _, cfg := range newManagedConfigs {
		if err := manifest.AddManagedConfig(cfg); err != nil {
			return nil, fmt.Errorf("adding managed config: %w", err)
		}
	}

	if err := a.manifestStore.SaveWithBackup(manifest, a.ManifestPath); err != nil {
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

func (a *Applier) ensureProvisionDirs(collection *store.Collection, sys model.SystemID, emu model.Emulator) error {
	for _, group := range emu.ProvisionGroups {
		baseDir := group.BaseDirFor(collection, sys)
		if baseDir == "" {
			continue
		}
		if err := vfs.MkdirAll(a.fs, baseDir, 0o755); err != nil {
			return fmt.Errorf("creating %s: %w", baseDir, err)
		}
	}
	return nil
}

func (a *Applier) downloadMissingProvisions(ctx context.Context, collection *store.Collection, sys model.SystemID, emu model.Emulator, onProgress func(Progress)) error {
	downloader := packages.NewDownloader(a.fs)
	extractor := packages.NewExtractor(a.fs)

	for _, group := range emu.ProvisionGroups {
		baseDir := group.BaseDirFor(collection, sys)
		if baseDir == "" {
			continue
		}

		for _, prov := range group.Provisions {
			if !prov.CanDownload() {
				continue
			}

			result := prov.Check(a.fs, baseDir)
			if result.Status == model.ProvisionFound {
				continue
			}

			dl := prov.Download
			hints := prov.Hints()

			tmpDir := filepath.Join(baseDir, ".kyaraben-provision-tmp")
			if err := vfs.MkdirAll(a.fs, tmpDir, 0o755); err != nil {
				return fmt.Errorf("creating temp dir: %w", err)
			}
			defer func() { _ = a.fs.RemoveAll(tmpDir) }()

			archivePath := filepath.Join(tmpDir, "download")
			if err := downloader.Download(ctx, packages.DownloadRequest{
				URLs:     []string{dl.URL},
				SHA256:   dl.SHA256,
				DestPath: archivePath,
			}); err != nil {
				return fmt.Errorf("downloading %s: %w", hints.DisplayName, err)
			}

			if dl.ArchiveType != "" {
				extractDir := filepath.Join(tmpDir, "extracted")
				if err := extractor.Extract(archivePath, extractDir, dl.ArchiveType); err != nil {
					return fmt.Errorf("extracting %s: %w", hints.DisplayName, err)
				}

				srcFile := filepath.Join(extractDir, dl.FilenameInArchive)
				if dl.FilenameInArchive == "" {
					srcFile = filepath.Join(extractDir, hints.DisplayName)
				}
				destFile := filepath.Join(baseDir, hints.DisplayName)
				if err := copyFile(a.fs, srcFile, destFile); err != nil {
					return fmt.Errorf("copying %s: %w", hints.DisplayName, err)
				}
			} else {
				destFile := filepath.Join(baseDir, hints.DisplayName)
				if err := a.fs.Rename(archivePath, destFile); err != nil {
					return fmt.Errorf("moving %s: %w", hints.DisplayName, err)
				}
			}
		}
	}
	return nil
}

func copyFile(fs vfs.FS, src, dst string) error {
	srcFile, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := fs.Create(dst)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(dstFile, srcFile)
	closeErr := dstFile.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func (a *Applier) downloadInitialFiles(ctx context.Context, downloads []model.InitialDownload) error {
	if len(downloads) == 0 {
		return nil
	}

	downloader := packages.NewDownloader(a.fs)
	extractor := packages.NewExtractor(a.fs)

	for _, dl := range downloads {
		if dl.ArchiveType != "" {
			if _, err := a.fs.Stat(dl.ExtractDir); err == nil {
				continue
			}
			if err := a.downloadAndExtract(ctx, dl, downloader, extractor); err != nil {
				return err
			}
		} else {
			if _, err := a.fs.Stat(dl.DestPath); err == nil {
				continue
			}
			if err := a.downloadFile(ctx, dl, downloader); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *Applier) downloadFile(ctx context.Context, dl model.InitialDownload, downloader *packages.HTTPDownloader) error {
	destDir := filepath.Dir(dl.DestPath)
	if err := vfs.MkdirAll(a.fs, destDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", destDir, err)
	}

	tmpPath := dl.DestPath + ".download"
	if err := downloader.Download(ctx, packages.DownloadRequest{
		URLs:     []string{dl.URL},
		SHA256:   dl.SHA256,
		DestPath: tmpPath,
	}); err != nil {
		return fmt.Errorf("downloading %s: %w", filepath.Base(dl.DestPath), err)
	}

	if err := a.fs.Rename(tmpPath, dl.DestPath); err != nil {
		return fmt.Errorf("moving %s: %w", filepath.Base(dl.DestPath), err)
	}
	return nil
}

func (a *Applier) downloadAndExtract(ctx context.Context, dl model.InitialDownload, downloader *packages.HTTPDownloader, extractor *packages.OSExtractor) error {
	parentDir := filepath.Dir(dl.ExtractDir)
	if err := vfs.MkdirAll(a.fs, parentDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", parentDir, err)
	}

	tmpDir := dl.ExtractDir + ".download-tmp"
	if err := vfs.MkdirAll(a.fs, tmpDir, 0o755); err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer func() { _ = a.fs.RemoveAll(tmpDir) }()

	archivePath := filepath.Join(tmpDir, "archive")
	if err := downloader.Download(ctx, packages.DownloadRequest{
		URLs:     []string{dl.URL},
		SHA256:   dl.SHA256,
		DestPath: archivePath,
	}); err != nil {
		return fmt.Errorf("downloading %s: %w", filepath.Base(dl.URL), err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := extractor.Extract(archivePath, extractDir, dl.ArchiveType); err != nil {
		return fmt.Errorf("extracting %s: %w", filepath.Base(dl.URL), err)
	}

	srcDir := extractDir
	if dl.StripPrefix != "" {
		srcDir = filepath.Join(extractDir, dl.StripPrefix)
	}

	if err := a.fs.Rename(srcDir, dl.ExtractDir); err != nil {
		return fmt.Errorf("moving extracted files: %w", err)
	}

	return nil
}

func (a *Applier) writeEmbeddedFiles(files []model.EmbeddedFile) error {
	for _, f := range files {
		destDir := filepath.Dir(f.DestPath)
		if err := vfs.MkdirAll(a.fs, destDir, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", destDir, err)
		}

		if err := a.fs.WriteFile(f.DestPath, f.Content, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", filepath.Base(f.DestPath), err)
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
	target := hardware.DetectTarget()

	installed := make(map[string]bool)
	packageDownloadSizes := make(map[string]int64)
	packageTotals := make(map[string]int64)
	packageArchiveTypes := make(map[string]string)
	installList := make([]string, 0, len(plan.packageNames)+1)

	for _, pkgName := range plan.packageNames {
		installList = append(installList, pkgName)
		installed[pkgName] = a.Installer.IsEmulatorInstalled(pkgName)
		downloadSize := packageDownloadSize(pkgName, target.Name, target.Arch, v, a.Installer)
		packageDownloadSizes[pkgName] = downloadSize
		packageArchiveTypes[pkgName] = packageArchiveType(pkgName, target.Name, target.Arch, v, a.Installer)
	}

	if len(plan.coreNames) > 0 {
		installList = append(installList, "retroarch-cores")
		installed["retroarch-cores"] = packages.RetroArchCoresInstalled(a.Installer, plan.coreNames, v)
		packageDownloadSizes["retroarch-cores"] = coreDownloadSize(plan.coreNames, target.Name, target.Arch, v, a.Installer)
		packageArchiveTypes["retroarch-cores"] = coreArchiveType(plan.coreNames, target.Name, target.Arch, v, a.Installer)
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

func packageDownloadSize(pkgName string, targetName string, arch string, v *versions.Versions, installer packages.Installer) int64 {
	spec, ok := v.GetPackage(pkgName)
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
	target := entry.SelectTarget(targetName, arch)
	if target == "" {
		return 0
	}
	build := entry.Target(target)
	if build == nil {
		return 0
	}
	return build.Size
}

func coreDownloadSize(coreNames []string, targetName string, arch string, v *versions.Versions, installer packages.Installer) int64 {
	if len(coreNames) == 0 {
		return 0
	}

	var total int64
	seenURLs := make(map[string]bool)
	for _, coreName := range coreNames {
		spec, ok := v.GetPackage(coreName)
		if !ok {
			continue
		}
		version := installer.ResolveVersion(coreName)
		entry := spec.GetVersion(version)
		if entry == nil {
			entry = spec.GetDefault()
		}
		if entry == nil {
			continue
		}
		target := entry.SelectTarget(targetName, arch)
		if target == "" {
			continue
		}
		url := entry.URL(target, spec)
		if seenURLs[url] {
			continue
		}
		seenURLs[url] = true
		if spec.BundleSize > 0 {
			total += spec.BundleSize
		} else {
			build := entry.Target(target)
			if build != nil && build.Size > 0 {
				total += build.Size
			}
		}
	}
	return total
}

func packageArchiveType(pkgName string, targetName string, arch string, v *versions.Versions, installer packages.Installer) string {
	spec, ok := v.GetPackage(pkgName)
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
	target := entry.SelectTarget(targetName, arch)
	if target == "" {
		return ""
	}
	return entry.ArchiveType(target, spec)
}

func coreArchiveType(coreNames []string, targetName string, arch string, v *versions.Versions, installer packages.Installer) string {
	if len(coreNames) == 0 {
		return ""
	}
	coreName := coreNames[0]
	spec, ok := v.GetPackage(coreName)
	if !ok {
		return ""
	}
	version := installer.ResolveVersion(coreName)
	entry := spec.GetVersion(version)
	if entry == nil {
		entry = spec.GetDefault()
	}
	if entry == nil {
		return ""
	}
	target := entry.SelectTarget(targetName, arch)
	if target == "" {
		return ""
	}
	return entry.ArchiveType(target, spec)
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
	spec, ok := v.GetPackage(pkgName)
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
		if coreName := emuID.RetroArchCoreName(); coreName != "" {
			keep["retroarch"] = a.Installer.ResolveVersion("retroarch")
			pkgName = coreName
		}
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
	keep["syncthing"] = a.Installer.ResolveVersion("syncthing")
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

func (a *Applier) buildDesktopEntries(emulatorIDs []model.EmulatorID, frontendIDs []model.FrontendID, launchArgsMap map[model.EmulatorID][]string) []launcher.GeneratedDesktop {
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
		if args, ok := launchArgsMap[emuID]; ok {
			launchArgs = strings.Join(args, " ")
		}

		entries = append(entries, launcher.GeneratedDesktop{
			BinaryName:    emu.Launcher.Binary,
			Name:          displayName,
			GenericName:   emu.Launcher.GenericName,
			CategoriesStr: strings.Join(emu.Launcher.Categories, ";"),
			LaunchArgs:    launchArgs,
			Keywords:      emu.Launcher.Keywords,
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
			Keywords:      fe.Launcher.Keywords,
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

func (a *Applier) buildSystemDisplayTypes() map[model.SystemID]model.DisplayType {
	systems := a.Registry.AllSystems()
	result := make(map[model.SystemID]model.DisplayType, len(systems))
	for _, sys := range systems {
		result[sys.ID] = sys.DisplayType
	}
	return result
}

func toLauncherBinaries(binaries []packages.InstalledBinary, binaryEnvs map[string]map[string]string) []launcher.InstalledBinary {
	result := make([]launcher.InstalledBinary, len(binaries))
	for i, b := range binaries {
		result[i] = launcher.InstalledBinary{Name: b.Name, Path: b.Path, Env: binaryEnvs[b.Name]}
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

func (a *Applier) syncSteamShortcuts(ctx context.Context, frontendIDs []model.FrontendID, binDir string, iconPaths map[string]string, manifest *model.Manifest) []model.SteamShortcutRecord {
	if a.SteamManager == nil || !a.SteamManager.IsAvailable() {
		return nil
	}

	var entries []steam.ShortcutEntry
	var records []model.SteamShortcutRecord

	for _, feID := range frontendIDs {
		def := a.Registry.GetFrontendDefinition(feID)
		if def == nil {
			continue
		}

		provider, ok := def.(model.SteamShortcutProvider)
		if !ok {
			continue
		}

		info := provider.SteamShortcut(binDir)
		if info == nil {
			continue
		}

		fe, err := a.Registry.GetFrontend(feID)
		if err != nil {
			continue
		}

		exe := filepath.Join(binDir, fe.Launcher.Binary)
		desktopPath := a.LauncherManager.EmulatorDesktopPath(fe.Launcher.Binary)

		entry := steam.ShortcutEntry{
			AppName:       info.AppName,
			Exe:           exe,
			StartDir:      binDir,
			Icon:          iconPaths[fe.Launcher.Binary],
			ShortcutPath:  desktopPath,
			LaunchOptions: info.LaunchOptions,
			Tags:          info.Tags,
		}

		if info.GridAssets != nil {
			entry.GridAssets = &steam.GridAssets{
				Grid:    info.GridAssets.Grid,
				Hero:    info.GridAssets.Hero,
				Logo:    info.GridAssets.Logo,
				Capsule: info.GridAssets.Capsule,
			}
		}

		entries = append(entries, entry)

		appID := generateSteamAppID(exe, info.AppName)
		records = append(records, model.SteamShortcutRecord{
			AppID:      appID,
			AppName:    info.AppName,
			FrontendID: feID,
		})
	}

	if len(entries) == 0 {
		return nil
	}

	changed, err := a.SteamManager.Sync(entries)
	if err != nil {
		log.Debug("Failed to sync Steam shortcuts: %v", err)
		return manifest.SteamShortcuts
	}

	log.Info("Synced %d Steam shortcuts", len(entries))

	if changed {
		if err := a.SteamManager.Restart(ctx); err != nil {
			log.Debug("Failed to restart Steam: %v", err)
		}
	}

	return records
}

func generateSteamAppID(exe, appName string) uint32 {
	input := exe + appName
	crc := crc32.ChecksumIEEE([]byte(input))
	return crc | 0x80000000
}

func collectConfigInputs(entries []model.ConfigEntry, cfg *model.KyarabenConfig, controllerConfig *model.ControllerConfig, collectionRoot string) map[string]string {
	deps := make(map[model.ConfigInput]bool)
	for _, entry := range entries {
		for _, dep := range entry.DependsOn {
			deps[dep] = true
		}
	}

	if len(deps) == 0 {
		return nil
	}

	inputs := make(map[string]string)
	for dep := range deps {
		key := string(dep)
		switch dep {
		case model.ConfigInputNintendoConfirm:
			inputs[key] = string(controllerConfig.NintendoConfirm)
		case model.ConfigInputHotkeys:
			inputs[key] = controllerConfig.Hotkeys.Fingerprint()
		case model.ConfigInputCollection:
			inputs[key] = collectionRoot
		case model.ConfigInputPreset:
			inputs[key] = cfg.Graphics.Preset
		case model.ConfigInputResume:
			inputs[key] = cfg.Savestate.Resume
		}
	}
	return inputs
}
