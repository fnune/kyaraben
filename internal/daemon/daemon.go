package daemon

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/apply"
	"github.com/fnune/kyaraben/internal/doctor"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/emulators/symlink"
	"github.com/fnune/kyaraben/internal/folders"
	"github.com/fnune/kyaraben/internal/hardware"
	"github.com/fnune/kyaraben/internal/importscanner"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/packages"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/status"
	"github.com/fnune/kyaraben/internal/steam"
	"github.com/fnune/kyaraben/internal/store"
	syncpkg "github.com/fnune/kyaraben/internal/sync"
	"github.com/fnune/kyaraben/internal/version"
	"github.com/fnune/kyaraben/internal/versions"
)

var log = logging.New("daemon")
var pairingLog = log.WithPrefix("[pairing]")

type Deps struct {
	FS              vfs.FS
	Paths           *paths.Paths
	ConfigPath      string
	StateDir        string
	ManifestPath    string
	Registry        *registry.Registry
	Installer       packages.Installer
	ConfigWriter    *emulators.ConfigWriter
	LauncherManager *launcher.Manager
	Service         syncpkg.ServiceManager
	Resolver        model.BaseDirResolver
}

type Daemon struct {
	deps          Deps
	configStore   *model.ConfigStore
	manifestStore *model.ManifestStore

	mu              sync.Mutex
	applyCancelFunc context.CancelFunc

	pairingMu         sync.Mutex
	pairingCancelFunc context.CancelFunc
	pairingActive     bool

	autoAcceptMu         sync.Mutex
	autoAcceptCancelFunc context.CancelFunc

	syncReconfigMu          sync.Mutex
	syncReconfigRunning     bool
	syncReconfigLastFailure time.Time
	syncReconfigFailCount   int

	reconcileMu          sync.Mutex
	reconcileRunning     bool
	reconcileLastFailure time.Time
	reconcileFailCount   int

	completionLogMu   sync.Mutex
	completionLastLog time.Time

	removedDevicesMu sync.Mutex
	removedDevices   map[string]time.Time
}

func New(deps Deps) *Daemon {
	return &Daemon{
		deps:           deps,
		configStore:    model.NewConfigStore(deps.FS),
		manifestStore:  model.NewManifestStore(deps.FS),
		removedDevices: make(map[string]time.Time),
	}
}

func NewDefault(p *paths.Paths, configPath, stateDir, manifestPath string, reg *registry.Registry, installer packages.Installer, configWriter *emulators.ConfigWriter, launcherManager *launcher.Manager) *Daemon {
	return New(Deps{
		FS:              vfs.OSFS,
		Paths:           p,
		ConfigPath:      configPath,
		StateDir:        stateDir,
		ManifestPath:    manifestPath,
		Registry:        reg,
		Installer:       installer,
		ConfigWriter:    configWriter,
		LauncherManager: launcherManager,
		Service:         syncpkg.NewDefaultServiceManager(),
		Resolver:        model.NewDefaultResolver(),
	})
}

func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	return path
}

func computeFolderPath(collectionRoot, folderID string) string {
	id := strings.TrimPrefix(folderID, "kyaraben-")
	parts := strings.SplitN(id, "-", 2)
	if len(parts) == 2 {
		return filepath.Join(collectionRoot, parts[0], parts[1])
	}
	return filepath.Join(collectionRoot, id)
}

func (d *Daemon) Handle(cmd Command) []Event {
	return d.HandleWithEmit(cmd, nil)
}

func (d *Daemon) HandleWithEmit(cmd Command, emit func(Event)) []Event {
	switch cmd.Type {
	case CommandTypeStatus:
		return d.handleStatus()
	case CommandTypeDoctor:
		return d.handleDoctor()
	case CommandTypeApply:
		return d.handleApply(emit)
	case CommandTypeCancelApply:
		return d.handleCancelApply()
	case CommandTypeGetSystems:
		return d.handleGetSystems()
	case CommandTypeGetFrontends:
		return d.handleGetFrontends()
	case CommandTypeGetConfig:
		return d.handleGetConfig()
	case CommandTypeSetConfig:
		return d.handleSetConfig(nil)
	case CommandTypeSyncStatus:
		return d.handleSyncStatus()
	case CommandTypeSyncRemoveDevice:
		return d.handleSyncRemoveDevice(nil)
	case CommandTypeSyncCancelPairing:
		return d.handleSyncCancelPairing()
	case CommandTypeSyncAcceptDevice:
		return d.handleSyncAcceptDevice(nil)
	case CommandTypeSyncPending:
		return d.handleSyncPending()
	case CommandTypeUninstallPreview:
		return d.handleUninstallPreview()
	case CommandTypeUninstall:
		return d.handleUninstall()
	case CommandTypeInstallKyaraben:
		return d.handleInstallKyaraben(nil)
	case CommandTypeInstallStatus:
		return d.handleInstallStatus()
	case CommandTypeRefreshIconCaches:
		return d.handleRefreshIconCaches()
	case CommandTypePreflight:
		return d.handlePreflight()
	case CommandTypeSyncEnable:
		return d.handleSyncEnable(emit)
	case CommandTypeSyncRevertFolder:
		return d.handleSyncRevertFolder(nil)
	case CommandTypeSyncLocalChanges:
		return d.handleSyncLocalChanges(nil)
	case CommandTypeSyncReset:
		return d.handleSyncReset()
	case CommandTypeSyncDiscoveredDevices:
		return d.handleSyncDiscoveredDevices()
	case CommandTypeSyncSetSettings:
		return d.handleSyncSetSettings(nil)
	case CommandTypeGetStorageDevices:
		return d.handleGetStorageDevices()
	case CommandTypeImportScan:
		return d.handleImportScan(nil)
	default:
		return d.errorResponse(fmt.Sprintf("unknown command: %s", cmd.Type))
	}
}

func (d *Daemon) HandleSetConfig(cmd SetConfigCommand, emit func(Event)) []Event {
	return d.handleSetConfig(&cmd.Data)
}

func (d *Daemon) HandleSyncRemoveDevice(cmd SyncRemoveDeviceCommand, emit func(Event)) []Event {
	return d.handleSyncRemoveDevice(&cmd.Data)
}

func (d *Daemon) HandleSyncAcceptDevice(cmd SyncAcceptDeviceCommand, emit func(Event)) []Event {
	return d.handleSyncAcceptDevice(&cmd.Data)
}

func (d *Daemon) HandleInstallKyaraben(cmd InstallKyarabenCommand, emit func(Event)) []Event {
	return d.handleInstallKyaraben(&cmd.Data)
}

func (d *Daemon) HandleSyncEnable(_ SyncEnableCommand, emit func(Event)) []Event {
	return d.handleSyncEnable(emit)
}

func (d *Daemon) HandleSyncRevertFolder(cmd SyncRevertFolderCommand, emit func(Event)) []Event {
	return d.handleSyncRevertFolder(&cmd.Data)
}

func (d *Daemon) HandleSyncLocalChanges(cmd SyncLocalChangesCommand, emit func(Event)) []Event {
	return d.handleSyncLocalChanges(&cmd.Data)
}

func (d *Daemon) HandleSyncSetSettings(cmd SyncSetSettingsCommand, emit func(Event)) []Event {
	return d.handleSyncSetSettings(&cmd.Data)
}

func (d *Daemon) HandleImportScan(cmd ImportScanCommand, emit func(Event)) []Event {
	return d.handleImportScan(&cmd.Data)
}

func (d *Daemon) errorResponse(msg string) []Event {
	return []Event{{
		Type: EventTypeError,
		Data: ErrorResponse{Error: msg},
	}}
}

func (d *Daemon) loadManifest() (*model.Manifest, error) {
	manifest, err := d.manifestStore.Load(d.deps.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("manifest data appears corrupted: %w. Please report this as a bug and run 'kyaraben apply' to restore your configuration", err)
	}
	return manifest, nil
}

type loadConfigResult struct {
	Config   *model.KyarabenConfig
	Warnings model.ConfigWarnings
}

func (d *Daemon) loadConfig() (*model.KyarabenConfig, error) {
	result, err := d.loadConfigWithWarnings()
	if err != nil {
		return nil, err
	}
	return result.Config, nil
}

func (d *Daemon) loadConfigWithWarnings() (*loadConfigResult, error) {
	path := d.deps.ConfigPath
	if path == "" {
		var err error
		path, err = d.deps.Paths.ConfigPath()
		if err != nil {
			return nil, err
		}
	}
	validators := &model.ConfigValidators{
		GetEmulator: d.deps.Registry.GetEmulator,
		GetSystem:   d.deps.Registry.GetSystem,
		GetFrontend: d.deps.Registry.GetFrontend,
	}
	loadResult, err := d.configStore.LoadWithWarnings(path, validators)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := model.NewDefaultConfig()
			if d.deps.Paths.Instance != "" {
				offset := d.deps.Paths.InstancePortOffset()
				cfg.Sync.Syncthing.ListenPort = 22100 + offset
				cfg.Sync.Syncthing.DiscoveryPort = 21127 + offset
				cfg.Sync.Syncthing.GUIPort = 8484 + offset
				cfg.Global.Collection = "~/Emulation-" + d.deps.Paths.Instance
			}
			return &loadConfigResult{Config: cfg}, nil
		}
		return nil, err
	}

	ctrlResult := loadResult.Config.ResolveControllerConfigWithWarnings()
	warnings := append(loadResult.Warnings, ctrlResult.Warnings...)

	return &loadConfigResult{
		Config:   loadResult.Config,
		Warnings: warnings,
	}, nil
}

func (d *Daemon) handleStatus() []Event {
	loadResult, err := d.loadConfigWithWarnings()
	if err != nil {
		return d.errorResponse(err.Error())
	}
	cfg := loadResult.Config

	configPath := d.deps.ConfigPath
	if configPath == "" {
		configPath, _ = d.deps.Paths.ConfigPath()
	}

	collection, err := store.NewCollection(d.deps.FS, d.deps.Paths, cfg.Global.Collection)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	result, err := status.NewGetter(d.deps.FS, d.deps.Paths, d.deps.Resolver).Get(context.Background(), cfg, configPath, d.deps.Registry, collection, d.deps.ManifestPath)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	systems := make([]model.SystemID, len(result.EnabledSystems))
	for i, sys := range result.EnabledSystems {
		systems[i] = sys.ID
	}

	installedEmulators := make([]InstalledEmulator, len(result.InstalledEmulators))
	for i, emu := range result.InstalledEmulators {
		managedConfigs := make([]ManagedConfigInfo, len(emu.ManagedConfigs))
		for j, mc := range emu.ManagedConfigs {
			regions := make([]ManagedRegionInfo, len(mc.ManagedRegions))
			for k, r := range mc.ManagedRegions {
				regions[k] = ManagedRegionInfo{
					Type:      r.Type,
					Section:   r.Section,
					KeyPrefix: r.KeyPrefix,
				}
			}
			managedConfigs[j] = ManagedConfigInfo{
				Path:           shortenPath(mc.Path),
				ManagedRegions: regions,
			}
		}
		installed := InstalledEmulator{
			ID:             emu.ID,
			Version:        emu.Version,
			ManagedConfigs: managedConfigs,
		}
		if e, err := d.deps.Registry.GetEmulator(emu.ID); err == nil && e.Launcher.Binary != "" {
			execLine := fmt.Sprintf("%s/%s", d.deps.LauncherManager.BinDir(), e.Launcher.Binary)
			if e.Launcher.CoreName != "" {
				corePath := fmt.Sprintf("%s/%s.so", d.deps.LauncherManager.CoresDir(), e.Launcher.CoreName)
				execLine += " -L " + corePath + " --menu"
			}
			installed.ExecLine = execLine
		}

		installed.Paths = make(map[string]EmulatorPaths)
		emuDef, _ := d.deps.Registry.GetEmulator(emu.ID)
		for sysID, emuIDs := range cfg.Systems {
			for _, emuID := range emuIDs {
				if emuID == emu.ID {
					paths := EmulatorPaths{
						Roms: shortenPath(collection.SystemRomsDir(sysID)),
					}
					if emuDef.PathUsage.UsesBiosDir {
						paths.Bios = shortenPath(collection.SystemBiosDir(sysID))
					}
					if emuDef.PathUsage.UsesSavesDir {
						paths.Saves = shortenPath(collection.SystemSavesDir(sysID))
					}
					if emuDef.PathUsage.UsesStatesDir {
						paths.Savestates = shortenPath(collection.EmulatorStatesDir(emu.ID))
					}
					if emuDef.PathUsage.UsesScreenshotsDir {
						paths.Screenshots = shortenPath(collection.EmulatorScreenshotsDir(emu.ID))
					}
					installed.Paths[string(sysID)] = paths
					break
				}
			}
		}

		installedEmulators[i] = installed
	}

	installedFrontends := make([]InstalledFrontend, len(result.InstalledFrontends))
	for i, fe := range result.InstalledFrontends {
		installed := InstalledFrontend{
			ID:      fe.ID,
			Version: fe.Version,
		}
		if f, err := d.deps.Registry.GetFrontend(fe.ID); err == nil && f.Launcher.Binary != "" {
			installed.ExecLine = fmt.Sprintf("%s/%s", d.deps.LauncherManager.BinDir(), f.Launcher.Binary)
			log.Info("frontend exec line: %s -> %s", fe.ID, installed.ExecLine)
		} else if err != nil {
			log.Error("failed to get frontend %s: %v", fe.ID, err)
		} else {
			log.Info("frontend %s has no binary", fe.ID)
		}
		installedFrontends[i] = installed
	}

	symlinks := make([]SymlinkInfo, len(result.Symlinks))
	for i, s := range result.Symlinks {
		symlinks[i] = SymlinkInfo{
			Source:     shortenPath(s.Source),
			Target:     shortenPath(s.Target),
			EmulatorID: s.EmulatorID,
		}
	}

	manifest, _ := d.loadManifest()
	manifestVersion := ""
	if manifest != nil {
		manifestVersion = manifest.KyarabenVersion
	}

	var configWarnings []ConfigWarning
	for _, w := range loadResult.Warnings {
		configWarnings = append(configWarnings, ConfigWarning{
			Field:   w.Field,
			Message: w.Message,
		})
	}

	return []Event{{
		Type: EventTypeResult,
		Data: StatusResponse{
			Collection:              result.CollectionPath,
			EnabledSystems:          systems,
			InstalledEmulators:      installedEmulators,
			InstalledFrontends:      installedFrontends,
			Symlinks:                symlinks,
			LastApplied:             result.LastApplied.Format(time.RFC3339),
			HealthWarning:           result.HealthWarning,
			KyarabenVersion:         version.Get(),
			ManifestKyarabenVersion: manifestVersion,
			ConfigWarnings:          configWarnings,
		},
	}}
}

func (d *Daemon) handleDoctor() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	collection, err := store.NewCollection(d.deps.FS, d.deps.Paths, cfg.Global.Collection)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	result, err := doctor.Run(context.Background(), cfg, d.deps.Registry, collection)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	response := make(DoctorResponse)
	for _, sys := range result.Systems {
		provisions := make([]ProvisionResult, len(sys.Provisions))
		for i, prov := range sys.Provisions {
			provisions[i] = ProvisionResult{
				Filename:            prov.Filename,
				Kind:                string(prov.Kind),
				Description:         prov.Description,
				Status:              string(prov.Status),
				ExpectedPath:        shortenPath(prov.ExpectedPath),
				FoundPath:           prov.FoundPath,
				ImportViaUI:         prov.ImportViaUI,
				GroupMessage:        prov.GroupMessage,
				GroupRequired:       prov.GroupRequired,
				GroupSatisfied:      prov.GroupSatisfied,
				GroupSize:           prov.GroupSize,
				DisplayName:         prov.DisplayName,
				VerifiedDisplayName: prov.VerifiedDisplayName,
				Instructions:        prov.Instructions,
			}
		}
		key := string(sys.SystemID) + ":" + string(sys.EmulatorID)
		response[key] = provisions
	}

	return []Event{{
		Type: EventTypeResult,
		Data: response,
	}}
}

func (d *Daemon) handleApply(emit func(Event)) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	versionOverrides, err := cfg.BuildVersionOverrides(d.deps.Registry.GetEmulator, d.deps.Registry.GetFrontend)
	if err != nil {
		return d.errorResponse(err.Error())
	}
	d.deps.Installer.SetVersionOverrides(versionOverrides)

	collection, err := store.NewCollection(d.deps.FS, d.deps.Paths, cfg.Global.Collection)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if cfg.Global.Headless {
		return d.handleApplyHeadless(cfg, collection, emit)
	}

	// Acquire exclusive lock to prevent concurrent Apply operations
	lockDir := filepath.Dir(d.deps.ManifestPath)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return d.errorResponse(fmt.Sprintf("creating state directory: %v", err))
	}
	lockPath := filepath.Join(lockDir, "apply.lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("creating lock file: %v", err))
	}
	defer func() { _ = lockFile.Close() }()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return d.errorResponse("another installation is already in progress")
	}
	defer func() { _ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN) }()

	ctx, cancel := context.WithCancel(context.Background())
	d.mu.Lock()
	d.applyCancelFunc = cancel
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		d.applyCancelFunc = nil
		d.mu.Unlock()
	}()

	syncWasStopped := false
	if cfg.Sync.Enabled {
		if emit != nil {
			emit(Event{
				Type: EventTypeProgress,
				Data: ProgressEvent{Step: "sync-pause", Message: "Pausing synchronization"},
			})
		}
		syncWasStopped = d.stopSyncthing(cfg)
	}

	applier := apply.NewApplier(
		d.deps.FS,
		d.deps.Installer,
		d.deps.ConfigWriter,
		d.deps.Registry,
		d.deps.ManifestPath,
		d.deps.LauncherManager,
		d.deps.Resolver,
		symlink.NewCreator(d.deps.FS),
	)
	applier.SteamManager = steam.NewDefaultManager()

	logPosition := logging.CurrentPosition()

	opts := apply.Options{
		OnProgress: func(p apply.Progress) {
			event := Event{
				Type: EventTypeProgress,
				Data: ProgressEvent{
					Step:            p.Step,
					Message:         p.Message,
					Output:          p.Output,
					BuildPhase:      p.BuildPhase,
					PackageName:     p.PackageName,
					ProgressPercent: p.ProgressPercent,
					BytesDownloaded: p.BytesDownloaded,
					BytesTotal:      p.BytesTotal,
					BytesPerSecond:  p.BytesPerSecond,
					LogPosition:     logPosition,
					LogEntry:        p.LogEntry,
				},
			}
			if emit != nil {
				emit(event)
			}
		},
	}

	applyResult, err := applier.Apply(ctx, cfg, collection, opts)
	if ctx.Err() != nil {
		return []Event{{
			Type: EventTypeCancelled,
			Data: CancelledResponse{Message: "Installation cancelled"},
		}}
	}
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if _, err := d.deps.LauncherManager.InstallCLI(); err != nil {
		log.Debug("Failed to install Kyaraben to PATH: %v", err)
	}

	d.maybeCreateDesktopShortcut()

	if cfg.Sync.Enabled {
		if syncWasStopped && emit != nil {
			emit(Event{
				Type: EventTypeProgress,
				Data: ProgressEvent{Step: "sync-resume", Message: "Resuming synchronization"},
			})
		}
		if err := d.updateSyncConfig(cfg, collection.Root()); err != nil {
			log.Info("Failed to update sync config: %v", err)
		}
	}

	var failedPackages []FailedPackage
	for _, fp := range applyResult.FailedPackages {
		failedPackages = append(failedPackages, FailedPackage{
			Name:   fp.Name,
			Reason: fp.Reason,
		})
	}

	return []Event{{
		Type: EventTypeResult,
		Data: ApplyResult{
			Success:        true,
			FailedPackages: failedPackages,
		},
	}}
}

func (d *Daemon) handleApplyHeadless(cfg *model.KyarabenConfig, collection *store.Collection, emit func(Event)) []Event {
	if emit != nil {
		emit(Event{
			Type: EventTypeProgress,
			Data: ProgressEvent{Step: "headless", Message: "Setting up headless sync hub"},
		})
	}

	if err := collection.Initialize(); err != nil {
		return d.errorResponse(fmt.Sprintf("initializing collection: %v", err))
	}

	for _, sys := range d.deps.Registry.AllSystems() {
		for _, emu := range d.deps.Registry.GetEmulatorsForSystem(sys.ID) {
			if err := collection.InitializeForEmulator(sys.ID, emu.ID, emu.PathUsage); err != nil {
				return d.errorResponse(fmt.Sprintf("initializing %s for %s: %v", sys.ID, emu.ID, err))
			}
		}
	}

	if cfg.Sync.Enabled {
		if emit != nil {
			emit(Event{
				Type: EventTypeProgress,
				Data: ProgressEvent{Step: "sync", Message: "Setting up synchronization"},
			})
		}
		if err := d.updateSyncConfig(cfg, collection.Root()); err != nil {
			return d.errorResponse(fmt.Sprintf("setting up sync: %v", err))
		}
	}

	return []Event{{
		Type: EventTypeResult,
		Data: ApplyResult{
			Success: true,
		},
	}}
}

func (d *Daemon) handlePreflight() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	collection, err := store.NewCollection(d.deps.FS, d.deps.Paths, cfg.Global.Collection)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	applier := apply.NewApplier(
		d.deps.FS,
		d.deps.Installer,
		d.deps.ConfigWriter,
		d.deps.Registry,
		d.deps.ManifestPath,
		d.deps.LauncherManager,
		d.deps.Resolver,
		symlink.NewCreator(d.deps.FS),
	)
	applier.SteamManager = steam.NewDefaultManager()

	preflight, err := applier.Preflight(context.Background(), cfg, collection)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("preflight check: %v", err))
	}

	manifest, err := d.manifestStore.Load(d.deps.ManifestPath)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("loading manifest: %v", err))
	}

	controllerConfig, _ := cfg.ResolveControllerConfig()
	diffCtx := &emulators.DiffContext{
		CurrentConfigInputs: map[string]string{
			string(model.ConfigInputNintendoConfirm): string(controllerConfig.NintendoConfirm),
			string(model.ConfigInputHotkeys):         controllerConfig.Hotkeys.Fingerprint(),
			string(model.ConfigInputCollection):      collection.Root(),
			string(model.ConfigInputPreset):          cfg.Graphics.Preset,
			string(model.ConfigInputResume):          cfg.Savestate.Resume,
		},
	}

	diffComputer := emulators.NewDiffComputer(d.deps.FS, d.deps.Resolver)

	type fileDiffState struct {
		diff             ConfigFileDiff
		seenChanges      map[string]bool
		seenUserKeys     map[string]bool
		seenKyarabenKeys map[string]bool
		seenRegions      map[string]bool
	}

	diffsByPath := make(map[string]*fileDiffState)
	var pathOrder []string

	for _, patch := range preflight.Patches {
		baseline, found := manifest.GetManagedConfig(patch.Target)
		var baselinePtr *model.ManagedConfig
		if found {
			baselinePtr = &baseline
		}

		diff, err := diffComputer.ComputeDiffWithBaseline(patch, baselinePtr, diffCtx)
		if err != nil {
			continue
		}

		state, seen := diffsByPath[diff.Path]
		if !seen {
			state = &fileDiffState{
				diff: ConfigFileDiff{
					Path:      diff.Path,
					IsNewFile: diff.IsNewFile,
				},
				seenChanges:      make(map[string]bool),
				seenUserKeys:     make(map[string]bool),
				seenKyarabenKeys: make(map[string]bool),
				seenRegions:      make(map[string]bool),
			}
			diffsByPath[diff.Path] = state
			pathOrder = append(pathOrder, diff.Path)
		}

		for _, r := range patch.ManagedRegions {
			var info ManagedRegionInfo
			switch v := r.(type) {
			case model.FileRegion:
				info = ManagedRegionInfo{Type: "file"}
			case model.SectionRegion:
				info = ManagedRegionInfo{Type: "section", Section: v.Section, KeyPrefix: v.KeyPrefix}
			}
			regionKey := info.Type + ":" + info.Section + ":" + info.KeyPrefix
			if !state.seenRegions[regionKey] {
				state.seenRegions[regionKey] = true
				state.diff.ManagedRegions = append(state.diff.ManagedRegions, info)
			}
		}

		state.diff.HasChanges = state.diff.HasChanges || diff.HasChanges()
		state.diff.UserModified = state.diff.UserModified || diff.UserModified
		state.diff.KyarabenChanged = state.diff.KyarabenChanged || diff.KyarabenChanged

		for _, vu := range diff.VersionUpgrades {
			if state.seenKyarabenKeys[vu.Key] {
				continue
			}
			state.seenKyarabenKeys[vu.Key] = true
			state.diff.KyarabenUpdates = append(state.diff.KyarabenUpdates, KyarabenUpdateDetail{
				Key:      vu.Key,
				OldValue: vu.OldValue,
				NewValue: vu.NewValue,
			})
		}

		for _, uc := range diff.UserChanges {
			if state.seenUserKeys[uc.Key] {
				continue
			}
			state.seenUserKeys[uc.Key] = true
			state.diff.UserChanges = append(state.diff.UserChanges, UserChangeDetail{
				Key:          uc.Key,
				WrittenValue: uc.WrittenValue,
				CurrentValue: uc.CurrentValue,
			})
		}

		for _, c := range diff.Changes {
			changeKey := c.Section() + ":" + c.Key()
			if state.seenChanges[changeKey] {
				continue
			}
			state.seenChanges[changeKey] = true

			changeType := "add"
			switch c.Type {
			case emulators.ChangeModify:
				changeType = "modify"
			case emulators.ChangeRemove:
				changeType = "remove"
			}
			state.diff.Changes = append(state.diff.Changes, ConfigChangeDetail{
				Type:     changeType,
				Key:      c.Key(),
				Section:  c.Section(),
				OldValue: c.OldValue,
				NewValue: c.NewValue,
			})
		}
	}

	var diffs []ConfigFileDiff
	for _, path := range pathOrder {
		diffs = append(diffs, diffsByPath[path].diff)
	}

	return []Event{{
		Type: EventTypeResult,
		Data: PreflightResponse{
			Diffs:         diffs,
			FilesToBackup: preflight.FilesToBackup,
		},
	}}
}

func (d *Daemon) handleCancelApply() []Event {
	d.mu.Lock()
	cancel := d.applyCancelFunc
	d.mu.Unlock()

	if cancel != nil {
		cancel()
		return []Event{{
			Type: EventTypeCancelled,
			Data: CancelledResponse{Message: "Installation cancelled"},
		}}
	}

	return []Event{{
		Type: EventTypeResult,
		Data: CancelledResponse{Message: "No apply in progress"},
	}}
}

func (d *Daemon) handleGetSystems() []Event {
	systems := d.deps.Registry.AllSystems()
	vers, _ := versions.Get()
	detectedTarget := hardware.DetectTarget()

	result := make(GetSystemsResponse, 0, len(systems))
	for _, sys := range systems {
		emus := d.deps.Registry.GetEmulatorsForSystem(sys.ID)
		emuList := make([]EmulatorRef, 0, len(emus))
		for _, emu := range emus {
			ref := EmulatorRef{
				ID:                emu.ID,
				Name:              emu.Name,
				PackageName:       emu.Package.PackageName(),
				SupportedSettings: emu.SupportedSettings,
				SupportedHotkeys:  emu.SupportedHotkeys,
				ResumeRecommended: emu.ResumeRecommended,
			}

			if vers != nil {
				pkgName := emu.Package.PackageName()
				coreName := retroArchCoreName(emu.ID)
				if coreName != "" {
					pkgName = coreName
				}
				if spec, ok := vers.GetPackage(pkgName); ok {
					ref.DefaultVersion = spec.Default
					ref.AvailableVersions = spec.AvailableVersions()

					if entry := spec.GetDefault(); entry != nil {
						if target := entry.SelectTarget(detectedTarget.Name, detectedTarget.Arch); target != "" {
							if build := entry.Target(target); build != nil {
								if spec.IsRetroArchCore() {
									ref.CoreBytes = build.Size
								} else if build.Size > 0 {
									ref.DownloadBytes = build.Size
								}
							}
						}
					}
				}

				if coreName != "" {
					if raSpec, ok := vers.GetPackage("retroarch"); ok {
						if entry := raSpec.GetDefault(); entry != nil {
							if target := entry.SelectTarget(detectedTarget.Name, detectedTarget.Arch); target != "" {
								if build := entry.Target(target); build != nil && build.Size > 0 {
									ref.DownloadBytes = build.Size
								}
							}
						}
					}
				}
			}

			emuList = append(emuList, ref)
		}

		var defaultEmuID model.EmulatorID
		if defaultEmu, err := d.deps.Registry.GetDefaultEmulator(sys.ID); err == nil {
			defaultEmuID = defaultEmu.ID
		}

		result = append(result, SystemWithEmulators{
			ID:                sys.ID,
			Name:              sys.Name,
			Description:       sys.Description,
			Manufacturer:      sys.Manufacturer,
			Label:             sys.Label,
			DefaultEmulatorID: defaultEmuID,
			Emulators:         emuList,
			NintendoDiamond:   model.NintendoDiamondSystems[sys.ID],
		})
	}

	return []Event{{
		Type: EventTypeResult,
		Data: result,
	}}
}

func (d *Daemon) handleGetFrontends() []Event {
	frontends := d.deps.Registry.AllFrontends()
	vers, _ := versions.Get()
	detectedTarget := hardware.DetectTarget()

	result := make(GetFrontendsResponse, 0, len(frontends))
	for _, fe := range frontends {
		ref := FrontendRef{
			ID:   fe.ID,
			Name: fe.Name,
		}

		if vers != nil {
			if spec, ok := vers.GetPackage(fe.Package.PackageName()); ok {
				ref.DefaultVersion = spec.Default
				ref.AvailableVersions = spec.AvailableVersions()

				if entry := spec.GetDefault(); entry != nil {
					if target := entry.SelectTarget(detectedTarget.Name, detectedTarget.Arch); target != "" {
						if build := entry.Target(target); build != nil && build.Size > 0 {
							ref.DownloadBytes = build.Size
						}
					}
				}
			}
		}

		result = append(result, ref)
	}

	return []Event{{
		Type: EventTypeResult,
		Data: result,
	}}
}

func (d *Daemon) handleGetConfig() []Event {
	loadResult, err := d.loadConfigWithWarnings()
	if err != nil {
		return d.errorResponse(fmt.Sprintf("loading config: %v", err))
	}
	cfg := loadResult.Config

	systems := make(map[string][]model.EmulatorID)
	for sys, emulators := range cfg.Systems {
		systems[string(sys)] = emulators
	}

	emulators := make(map[string]EmulatorConfResponse)
	for emuID, emuConf := range cfg.Emulators {
		if emuConf.Version != "" || emuConf.Preset != nil || emuConf.Resume != nil {
			emulators[string(emuID)] = EmulatorConfResponse{
				Version: emuConf.Version,
				Preset:  emuConf.Preset,
				Resume:  emuConf.Resume,
			}
		}
	}

	frontends := make(map[string]FrontendConfResponse)
	for feID, feConf := range cfg.Frontends {
		frontends[string(feID)] = FrontendConfResponse{
			Enabled: feConf.Enabled,
			Version: feConf.Version,
		}
	}

	var warnings []ConfigWarning
	for _, w := range loadResult.Warnings {
		warnings = append(warnings, ConfigWarning{
			Field:   w.Field,
			Message: w.Message,
		})
	}

	hk := cfg.Controller.Hotkeys
	hotkeyResp := HotkeyConfigResponse{
		Modifier:         hk.Modifier,
		SaveState:        hk.SaveState,
		LoadState:        hk.LoadState,
		NextSlot:         hk.NextSlot,
		PrevSlot:         hk.PrevSlot,
		FastForward:      hk.FastForward,
		Rewind:           hk.Rewind,
		Pause:            hk.Pause,
		Screenshot:       hk.Screenshot,
		Quit:             hk.Quit,
		ToggleFullscreen: hk.ToggleFullscreen,
		OpenMenu:         hk.OpenMenu,
	}
	if hotkeyResp.Modifier == "" {
		hotkeyResp.Modifier = string(model.ButtonBack)
	}
	if hotkeyResp.SaveState == "" {
		hotkeyResp.SaveState = string(model.ButtonRightShoulder)
	}
	if hotkeyResp.LoadState == "" {
		hotkeyResp.LoadState = string(model.ButtonLeftShoulder)
	}
	if hotkeyResp.NextSlot == "" {
		hotkeyResp.NextSlot = string(model.ButtonDPadRight)
	}
	if hotkeyResp.PrevSlot == "" {
		hotkeyResp.PrevSlot = string(model.ButtonDPadLeft)
	}
	if hotkeyResp.FastForward == "" {
		hotkeyResp.FastForward = string(model.ButtonY)
	}
	if hotkeyResp.Rewind == "" {
		hotkeyResp.Rewind = string(model.ButtonX)
	}
	if hotkeyResp.Pause == "" {
		hotkeyResp.Pause = string(model.ButtonA)
	}
	if hotkeyResp.Screenshot == "" {
		hotkeyResp.Screenshot = string(model.ButtonB)
	}
	if hotkeyResp.Quit == "" {
		hotkeyResp.Quit = string(model.ButtonStart)
	}
	if hotkeyResp.ToggleFullscreen == "" {
		hotkeyResp.ToggleFullscreen = string(model.ButtonLeftStick)
	}
	if hotkeyResp.OpenMenu == "" {
		hotkeyResp.OpenMenu = string(model.ButtonRightStick)
	}

	detectedTarget := hardware.DetectTarget()

	return []Event{{
		Type: EventTypeResult,
		Data: ConfigResponse{
			Collection: cfg.Global.Collection,
			Headless:   cfg.Global.Headless,
			Graphics: GraphicsConfigResponse{
				Preset:         cfg.Graphics.Preset,
				Target:         cfg.GraphicsTarget(),
				DetectedTarget: detectedTarget.Name,
			},
			Savestate: SavestateConfigResponse{Resume: cfg.Savestate.Resume},
			Controller: ControllerConfigResponse{
				NintendoConfirm: cfg.Controller.NintendoConfirm,
				Hotkeys:         hotkeyResp,
			},
			Systems:   systems,
			Emulators: emulators,
			Frontends: frontends,
			Warnings:  warnings,
		},
	}}
}

func (d *Daemon) handleSetConfig(data *SetConfigRequest) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(fmt.Sprintf("loading config: %v", err))
	}

	if data != nil {
		if data.Collection != "" {
			cfg.Global.Collection = data.Collection
		}

		if data.Graphics != nil {
			if data.Graphics.Preset != "" {
				cfg.Graphics.Preset = data.Graphics.Preset
			}
			if data.Graphics.Target != "" {
				cfg.Graphics.Target = data.Graphics.Target
			}
		}

		if data.Savestate != nil {
			cfg.Savestate.Resume = data.Savestate.Resume
		}

		if data.Controller != nil {
			cfg.Controller.NintendoConfirm = data.Controller.NintendoConfirm
			if data.Controller.Hotkeys != nil {
				hk := data.Controller.Hotkeys
				if hk.Modifier != "" {
					cfg.Controller.Hotkeys.Modifier = hk.Modifier
				}
				if hk.SaveState != "" {
					cfg.Controller.Hotkeys.SaveState = hk.SaveState
				}
				if hk.LoadState != "" {
					cfg.Controller.Hotkeys.LoadState = hk.LoadState
				}
				if hk.NextSlot != "" {
					cfg.Controller.Hotkeys.NextSlot = hk.NextSlot
				}
				if hk.PrevSlot != "" {
					cfg.Controller.Hotkeys.PrevSlot = hk.PrevSlot
				}
				if hk.FastForward != "" {
					cfg.Controller.Hotkeys.FastForward = hk.FastForward
				}
				if hk.Rewind != "" {
					cfg.Controller.Hotkeys.Rewind = hk.Rewind
				}
				if hk.Pause != "" {
					cfg.Controller.Hotkeys.Pause = hk.Pause
				}
				if hk.Screenshot != "" {
					cfg.Controller.Hotkeys.Screenshot = hk.Screenshot
				}
				if hk.Quit != "" {
					cfg.Controller.Hotkeys.Quit = hk.Quit
				}
				if hk.ToggleFullscreen != "" {
					cfg.Controller.Hotkeys.ToggleFullscreen = hk.ToggleFullscreen
				}
				if hk.OpenMenu != "" {
					cfg.Controller.Hotkeys.OpenMenu = hk.OpenMenu
				}
			}
		}

		if data.Systems != nil {
			cfg.Systems = make(map[model.SystemID][]model.EmulatorID)
			for sysStr, emuStrs := range data.Systems {
				emulators := make([]model.EmulatorID, len(emuStrs))
				for i, emuStr := range emuStrs {
					emulators[i] = model.EmulatorID(emuStr)
				}
				cfg.Systems[model.SystemID(sysStr)] = emulators
			}
		}

		if data.Emulators != nil {
			if cfg.Emulators == nil {
				cfg.Emulators = make(map[model.EmulatorID]model.EmulatorConf)
			}
			for emuStr, emuConf := range data.Emulators {
				emuID := model.EmulatorID(emuStr)
				existing := cfg.Emulators[emuID]
				if emuConf.Version != nil {
					existing.Version = *emuConf.Version
				}
				if emuConf.Preset != nil {
					existing.Preset = emuConf.Preset
				}
				if emuConf.Resume != nil {
					existing.Resume = emuConf.Resume
				}
				if existing.Version == "" && existing.Preset == nil && existing.Resume == nil {
					delete(cfg.Emulators, emuID)
				} else {
					cfg.Emulators[emuID] = existing
				}
			}
		}

		if data.Frontends != nil {
			if cfg.Frontends == nil {
				cfg.Frontends = make(map[model.FrontendID]model.FrontendConfig)
			}
			for feStr, feConf := range data.Frontends {
				feID := model.FrontendID(feStr)
				existing := cfg.Frontends[feID]
				existing.Enabled = feConf.Enabled
				if feConf.Version != nil {
					existing.Version = *feConf.Version
				}
				cfg.Frontends[feID] = existing
			}
		}
	}

	path := d.deps.ConfigPath
	if path == "" {
		path, _ = d.deps.Paths.ConfigPath()
	}

	if err := d.configStore.Save(cfg, path); err != nil {
		return d.errorResponse(err.Error())
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SetConfigResponse{Success: true},
	}}
}

func (d *Daemon) handleSyncStatus() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	unit := syncpkg.NewSystemdUnit(d.deps.FS, d.deps.Paths, d.deps.Service)
	serviceInstalled := unit.IsEnabled()

	manifest, _ := d.loadManifest()
	installed := manifest.SyncthingInstall != nil && d.fileExists(manifest.SyncthingInstall.BinaryPath)

	if cfg.Sync.Enabled && manifest.SyncthingInstall != nil &&
		manifest.SyncthingInstall.ConfigSchemaVersion < syncpkg.ConfigSchemaVersion {
		log.Info("Sync config schema version changed (%d -> %d), regenerating config",
			manifest.SyncthingInstall.ConfigSchemaVersion, syncpkg.ConfigSchemaVersion)
		go d.ensureSyncthingManaged(cfg)
	}

	if !cfg.Sync.Enabled {
		return []Event{{
			Type: EventTypeResult,
			Data: SyncStatusResponse{
				Enabled:                false,
				Installed:              installed,
				ServiceInstalled:       serviceInstalled,
				GlobalDiscoveryEnabled: cfg.Sync.Syncthing.GlobalDiscoveryEnabled,
				AutostartEnabled:       cfg.Sync.Autostart,
			},
		}}
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	} else {
		log.Debug("No API key found at %s", filepath.Join(d.deps.StateDir, "syncthing", "config", ".apikey"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !client.IsRunning(ctx) {
		status := unit.Status()
		systemdManaged := status.Active == "active" || status.Active == "activating"

		log.Debug("Syncthing API not responding (systemd state=%s, port=%d, hasApiKey=%v)",
			status.Active, cfg.Sync.Syncthing.GUIPort, loadedKey != "")

		if !status.Failed && !systemdManaged {
			go d.ensureSyncthingManaged(cfg)
		}

		serviceError := status.Message
		if serviceError == "" && status.Active == "activating" {
			serviceError = "Syncthing is starting..."
		}

		return []Event{{
			Type: EventTypeResult,
			Data: SyncStatusResponse{
				Enabled:                true,
				Running:                false,
				Installed:              installed,
				ServiceInstalled:       serviceInstalled,
				GUIURL:                 fmt.Sprintf("http://127.0.0.1:%d", cfg.Sync.Syncthing.GUIPort),
				ServiceError:           serviceError,
				GlobalDiscoveryEnabled: cfg.Sync.Syncthing.GlobalDiscoveryEnabled,
				AutostartEnabled:       cfg.Sync.Autostart,
			},
		}}
	}

	syncStatus, err := client.GetStatus(ctx, nil)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if len(syncStatus.Devices) > 0 {
		go d.reconcileFolderSharing(cfg)
	}

	collection, err := store.NewCollection(d.deps.FS, d.deps.Paths, cfg.Global.Collection)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	devices := make([]SyncDevice, len(syncStatus.Devices))
	for i, dev := range syncStatus.Devices {
		devices[i] = SyncDevice{
			ID:                dev.ID,
			Name:              dev.Name,
			Connected:         dev.Connected,
			Paused:            dev.Paused,
			ConnectionType:    dev.ConnectionType,
			IsLocal:           dev.IsLocal,
			ConnectivityIssue: dev.ConnectivityIssue,
		}
		if dev.LastSeen != "" {
			devices[i].LastSeen = &dev.LastSeen
		}
		if dev.Connected {
			completion, err := client.GetDeviceCompletion(ctx, dev.ID)
			if err != nil {
				log.Debug("Failed to get completion for device %s (%s): %v", dev.Name, dev.ID[:7], err)
				continue
			}
			percent := int(completion.Completion)
			devices[i].Completion = &percent
			if percent < 100 {
				d.logCompletionDiagnostics(ctx, client, dev, completion, syncStatus.Folders)
			}
		} else if dev.ConnectivityIssue != "" {
			log.Info("Device %s connectivity issue: %s", dev.Name, dev.ConnectivityIssue)
		}
	}

	folders := make([]SyncFolder, len(syncStatus.Folders))
	for i, f := range syncStatus.Folders {
		path := f.Path
		if !filepath.IsAbs(path) {
			path = computeFolderPath(collection.Root(), f.ID)
		}
		var conflictCount int
		if conflicts, err := syncpkg.ScanForConflicts(d.deps.FS, path); err == nil {
			conflictCount = len(conflicts)
		}
		folders[i] = SyncFolder{
			ID:                 f.ID,
			Path:               path,
			Label:              syncpkg.FolderLabel(f.ID),
			State:              f.State,
			Error:              f.Error,
			Type:               f.Type,
			GlobalSize:         f.GlobalSize,
			LocalSize:          f.LocalSize,
			NeedSize:           f.NeedSize,
			ReceiveOnlyChanges: f.ReceiveOnlyChanges,
			ConflictCount:      conflictCount,
		}
	}

	var progress *SyncProgress
	if progressInfo, err := client.GetSyncProgress(ctx); err == nil && progressInfo.NeedFiles > 0 {
		progress = &SyncProgress{
			NeedFiles:   progressInfo.NeedFiles,
			NeedBytes:   progressInfo.NeedBytes,
			GlobalBytes: progressInfo.GlobalBytes,
			Percent:     progressInfo.Percent,
		}
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SyncStatusResponse{
			Enabled:                true,
			Running:                true,
			Installed:              installed,
			ServiceInstalled:       serviceInstalled,
			DeviceID:               syncStatus.DeviceID,
			GUIURL:                 syncStatus.GUIURL,
			State:                  SyncState(syncStatus.OverallState()),
			Devices:                devices,
			Folders:                folders,
			Progress:               progress,
			GlobalDiscoveryEnabled: cfg.Sync.Syncthing.GlobalDiscoveryEnabled,
			AutostartEnabled:       cfg.Sync.Autostart,
			LocalConnectivityIssue: syncStatus.LocalConnectivityIssue,
		},
	}}
}

func (d *Daemon) handleSyncRemoveDevice(data *SyncRemoveDeviceRequest) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if data == nil || data.DeviceID == "" {
		return d.errorResponse("deviceId is required")
	}

	deviceID := strings.ToUpper(strings.TrimSpace(data.DeviceID))

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var removedName string
	if devices, err := client.GetConfiguredDevices(ctx); err == nil {
		for _, dev := range devices {
			if strings.ToUpper(dev.ID) == deviceID {
				removedName = dev.Name
				break
			}
		}
	}

	if err := client.RemoveDevice(ctx, deviceID); err != nil {
		return d.errorResponse(fmt.Sprintf("failed to remove device from syncthing: %v", err))
	}

	d.removedDevicesMu.Lock()
	d.removedDevices[deviceID] = time.Now()
	d.removedDevicesMu.Unlock()
	log.Info("Device removed: %s (%s) - tracking for zombie detection", removedName, deviceID[:7])

	return []Event{{
		Type: EventTypeResult,
		Data: SyncRemoveDeviceResponse{
			Success:  true,
			DeviceID: deviceID,
			Name:     removedName,
		},
	}}
}

func (d *Daemon) handleSyncAcceptDevice(data *SyncAcceptDeviceRequest) []Event {
	if data == nil || data.DeviceID == "" {
		return d.errorResponse("deviceId is required")
	}

	if !data.Accept {
		log.Info("User rejected pending device: %s", data.DeviceID[:7])
		d.stopAutoAcceptLoop()
		return []Event{{
			Type: EventTypeResult,
			Data: map[string]bool{"rejected": true},
		}}
	}

	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("User accepted pending device: %s", data.DeviceID[:7])
	if err := client.AddDevice(ctx, data.DeviceID, ""); err != nil {
		return d.errorResponse(fmt.Sprintf("failed to add device: %v", err))
	}
	if err := client.ShareFoldersWithDevice(ctx, data.DeviceID); err != nil {
		log.Error("Failed to share folders with accepted device: %v", err)
	}

	d.stopAutoAcceptLoop()

	return []Event{{
		Type: EventTypeResult,
		Data: map[string]bool{"accepted": true},
	}}
}

func (d *Daemon) handleSyncRevertFolder(data *SyncRevertFolderRequest) []Event {
	if data == nil || data.FolderID == "" {
		return d.errorResponse("folderId is required")
	}

	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.RevertFolder(ctx, data.FolderID); err != nil {
		return d.errorResponse(fmt.Sprintf("failed to revert folder: %v", err))
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SyncRevertFolderResponse{Success: true},
	}}
}

func (d *Daemon) handleSyncLocalChanges(data *SyncLocalChangesRequest) []Event {
	if data == nil || data.FolderID == "" {
		return d.errorResponse("folderId is required")
	}

	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	changes, err := client.GetLocalChanges(ctx, data.FolderID)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("failed to get local changes: %v", err))
	}

	result := make([]SyncLocalChange, len(changes))
	for i, c := range changes {
		result[i] = SyncLocalChange{
			Action:   c.Action,
			Type:     c.Type,
			Path:     c.Path,
			Modified: c.Modified,
			Size:     c.Size,
		}
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SyncLocalChangesResponse{Changes: result},
	}}
}

func (d *Daemon) HandleSyncStartPairing(cmd Command, emit func(Event)) []Event {
	return d.handleSyncStartPairing(emit)
}

func (d *Daemon) HandleSyncJoinPeer(cmd SyncJoinPeerCommand, emit func(Event)) []Event {
	return d.handleSyncJoinPeer(&cmd.Data, emit)
}

func (d *Daemon) HandleSyncDiscoveredDevices(cmd Command, emit func(Event)) []Event {
	return d.handleSyncDiscoveredDevices()
}

func newRelayClient(relays []string) (*syncpkg.RelayClient, error) {
	if len(relays) > 0 {
		return syncpkg.NewRelayClient(relays)
	}
	return syncpkg.NewDefaultRelayClient()
}

func (d *Daemon) handleSyncStartPairing(emit func(Event)) []Event {
	log.Info("Starting pairing mode")

	d.removedDevicesMu.Lock()
	d.removedDevices = make(map[string]time.Time)
	d.removedDevicesMu.Unlock()
	cfg, err := d.loadConfig()
	if err != nil {
		log.Error("Failed to load config: %v", err)
		return d.errorResponse(err.Error())
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	deviceCtx, deviceCancel := context.WithTimeout(context.Background(), 5*time.Second)
	deviceID, err := client.GetDeviceID(deviceCtx)
	deviceCancel()
	if err != nil {
		log.Error("Failed to get device ID: %v", err)
		return d.errorResponse(fmt.Sprintf("getting device ID: %v", err))
	}

	relayClient, err := newRelayClient(cfg.Sync.Relays)
	if err != nil {
		log.Info("No relay server available, falling back to device ID only: %v", err)
		d.startPendingDeviceLoop(cfg, emit)
		return []Event{{
			Type: EventTypeResult,
			Data: SyncStartPairingResponse{DeviceID: deviceID},
		}}
	}

	createCtx, createCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer createCancel()
	relayResp, err := relayClient.CreateSession(createCtx, deviceID)
	if err != nil {
		log.Info("Failed to create relay session, falling back to device ID only: %v", err)
		d.startPendingDeviceLoop(cfg, emit)
		return []Event{{
			Type: EventTypeResult,
			Data: SyncStartPairingResponse{DeviceID: deviceID},
		}}
	}

	log.Info("Created relay session with code %s for device %s", relayResp.Code, deviceID)
	d.startRelayPollLoop(cfg, relayClient, relayResp.Code, client, emit)
	d.startPendingDeviceLoop(cfg, emit)

	return []Event{{
		Type: EventTypeResult,
		Data: SyncStartPairingResponse{DeviceID: deviceID, Code: relayResp.Code},
	}}
}

func (d *Daemon) handleSyncJoinPeer(data *SyncJoinPeerRequest, emit func(Event)) []Event {
	if data == nil || (data.Code == "" && data.DeviceID == "") {
		return d.errorResponse("pairing code or device ID is required")
	}

	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	var peerDeviceID string
	var relayCode string
	var relayClient *syncpkg.RelayClient

	if data.DeviceID != "" {
		peerDeviceID = strings.ToUpper(strings.TrimSpace(data.DeviceID))
		log.Info("handleSyncJoinPeer called with deviceId=%s", peerDeviceID)
	} else {
		code := strings.ToUpper(strings.TrimSpace(data.Code))
		if isRelayCode(code) {
			log.Info("handleSyncJoinPeer called with relay code=%s", code)
			relayCode = code
			relayClient, err = newRelayClient(cfg.Sync.Relays)
			if err != nil {
				return d.errorResponse(fmt.Sprintf("relay unavailable: %v", err))
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			session, err := relayClient.GetSession(ctx, code)
			cancel()
			if err != nil {
				return d.errorResponse(fmt.Sprintf("invalid pairing code: %v", err))
			}
			peerDeviceID = session.DeviceID
			log.Info("Resolved relay code %s to device ID %s", code, peerDeviceID)
		} else {
			peerDeviceID = code
			log.Info("handleSyncJoinPeer called with device ID=%s", peerDeviceID)
		}
	}

	d.pairingMu.Lock()
	if d.pairingActive {
		d.pairingMu.Unlock()
		return d.errorResponse("pairing already in progress")
	}
	d.pairingActive = true
	ctx, cancel := context.WithCancel(context.Background())
	d.pairingCancelFunc = cancel
	d.pairingMu.Unlock()

	log.Info("Starting joiner pairing flow for device %s", peerDeviceID)

	go func(relayCode string, relayClient *syncpkg.RelayClient) {
		defer func() {
			pairingLog.Info("Joiner pairing goroutine ended")
			d.pairingMu.Lock()
			d.pairingActive = false
			d.pairingMu.Unlock()
		}()

		client := syncpkg.NewClient(cfg.Sync)
		loadedKey := d.loadSyncAPIKey()
		if loadedKey != "" {
			client.SetAPIKey(loadedKey)
		}

		pairingLog.Info("Waiting for syncthing to be ready")
		if err := d.waitForSyncthingReady(ctx, client); err != nil {
			pairingLog.Error("Syncthing not ready: %v", err)
			emit(Event{Type: EventTypeError, Data: ErrorResponse{Error: fmt.Sprintf("syncthing not ready: %v", err)}})
			return
		}

		localDeviceID, err := client.GetDeviceID(ctx)
		if err != nil {
			pairingLog.Error("Failed to get local device ID: %v", err)
			emit(Event{Type: EventTypeError, Data: ErrorResponse{Error: fmt.Sprintf("getting local device ID: %v", err)}})
			return
		}
		pairingLog.Info("Local device ID: %s", localDeviceID[:7]+"...")

		if relayCode != "" && relayClient != nil {
			submitCtx, submitCancel := context.WithTimeout(ctx, 10*time.Second)
			if err := relayClient.SubmitResponse(submitCtx, relayCode, localDeviceID); err != nil {
				pairingLog.Error("Failed to submit response to relay: %v", err)
			} else {
				pairingLog.Info("Submitted device ID to relay for code %s", relayCode)
			}
			submitCancel()
		}

		flow := syncpkg.NewJoinerPairingFlow(syncpkg.PairingFlowConfig{
			SyncConfig: cfg.Sync,
			Instance:   d.deps.Paths.Instance,
			Client:     client,
			OnProgress: func(msg string) {
				emit(Event{
					Type: EventTypeProgress,
					Data: SyncPairingProgressEvent{Message: msg},
				})
			},
		})

		result, err := flow.Run(ctx, peerDeviceID)
		if err != nil {
			pairingLog.Error("Joiner pairing flow failed: %v", err)
			emit(Event{Type: EventTypeError, Data: ErrorResponse{Error: err.Error()}})
			return
		}

		pairingLog.Info("Joiner pairing flow succeeded, persisting config")
		d.persistSyncEnabled(cfg)

		d.dismissUnwantedPendingFolders(client)

		pairingLog.Info("Emitting success result for peer %s (%s)", result.PeerName, result.PeerDeviceID[:7]+"...")
		emit(Event{
			Type: EventTypeResult,
			Data: SyncJoinPeerResponse{
				Success:      true,
				PeerDeviceID: result.PeerDeviceID,
				PeerName:     result.PeerName,
			},
		})
	}(relayCode, relayClient)

	return nil
}

func (d *Daemon) handleSyncCancelPairing() []Event {
	d.stopAutoAcceptLoop()
	d.stopJoinerPairing()
	return []Event{{
		Type: EventTypeResult,
		Data: map[string]bool{"cancelled": true},
	}}
}

func (d *Daemon) stopJoinerPairing() {
	d.pairingMu.Lock()
	defer d.pairingMu.Unlock()
	if d.pairingCancelFunc != nil {
		log.Info("Stopping joiner pairing flow")
		d.pairingCancelFunc()
		d.pairingCancelFunc = nil
	}
	d.pairingActive = false
}

func (d *Daemon) handleSyncEnable(emit func(Event)) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if cfg.Sync.Enabled {
		return d.errorResponse("sync is already enabled")
	}

	cfg.Sync.Enabled = true

	collection, err := store.NewCollection(d.deps.FS, d.deps.Paths, cfg.Global.Collection)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	allSystems := d.syncSystems(cfg)
	allEmulators := d.syncEmulators(cfg)
	allFrontends := d.syncFrontends(cfg)

	emitProgress := func(phase, message string, percent int) {
		if emit != nil {
			emit(Event{
				Type: EventTypeProgress,
				Data: SyncEnableProgressEvent{
					Phase:   phase,
					Message: message,
					Percent: percent,
				},
			})
		}
	}

	go func() {
		emitProgress("installing", "Installing syncthing...", 0)

		setup := d.syncSetup()
		result, err := setup.Install(context.Background(), cfg.Sync, collection.Root(), allSystems, allEmulators, allFrontends, func(p packages.InstallProgress) {
			percent := 0
			if p.BytesTotal > 0 {
				percent = int(p.BytesDownloaded * 80 / p.BytesTotal)
			}
			emitProgress("installing", p.Phase, percent)
		})
		if err != nil {
			emit(Event{Type: EventTypeError, Data: ErrorResponse{Error: err.Error()}})
			return
		}

		emitProgress("configuring", "Saving configuration...", 90)

		path := d.deps.ConfigPath
		if path == "" {
			path, _ = d.deps.Paths.ConfigPath()
		}
		if err := d.configStore.Save(cfg, path); err != nil {
			emit(Event{Type: EventTypeError, Data: ErrorResponse{Error: err.Error()}})
			return
		}

		emitProgress("configuring", "Updating manifest...", 95)

		manifest, err := d.loadManifest()
		if err != nil {
			emit(Event{Type: EventTypeError, Data: ErrorResponse{Error: err.Error()}})
			return
		}
		manifest.SyncthingInstall = &model.SyncthingInstall{
			Version:             d.deps.Installer.ResolveVersion("syncthing"),
			ConfigSchemaVersion: syncpkg.ConfigSchemaVersion,
			BinaryPath:          result.SyncthingBinary,
			ConfigDir:           result.ConfigDir,
			DataDir:             result.DataDir,
			SystemdUnitPath:     result.SystemdUnitPath,
		}
		if saveErr := manifest.SaveWithBackup(d.deps.ManifestPath); saveErr != nil {
			emit(Event{Type: EventTypeError, Data: ErrorResponse{Error: saveErr.Error()}})
			return
		}

		emitProgress("complete", "Sync enabled successfully", 100)

		emit(Event{
			Type: EventTypeResult,
			Data: SyncEnableResponse{Success: true},
		})
	}()

	return nil
}

func (d *Daemon) handleSyncReset() []Event {
	d.stopAutoAcceptLoop()
	d.stopJoinerPairing()
	setup := d.syncSetup()
	var removedFiles []string

	if setup.IsEnabled() {
		removedFiles = append(removedFiles, "systemd unit: "+d.deps.Paths.DirName()+"-syncthing.service")
	}

	syncthingDir := filepath.Join(d.deps.StateDir, "syncthing")
	if d.dirExists(syncthingDir) {
		removedFiles = append(removedFiles, syncthingDir)
	}

	if err := setup.Reset(); err != nil {
		return d.errorResponse(fmt.Sprintf("resetting sync: %v", err))
	}

	cfg, err := d.loadConfig()
	if err == nil && cfg.Sync.Enabled {
		cfg.Sync.Enabled = false
		_ = d.configStore.Save(cfg, d.deps.ConfigPath)
	}

	manifest, err := d.loadManifest()
	if err == nil && manifest.SyncthingInstall != nil {
		manifest.SyncthingInstall = nil
		_ = manifest.SaveWithBackup(d.deps.ManifestPath)
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SyncResetResponse{
			Success:      true,
			RemovedFiles: removedFiles,
		},
	}}
}

func (d *Daemon) handleSyncSetSettings(data *SyncSetSettingsRequest) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if data == nil {
		return d.errorResponse("missing settings data")
	}

	unit := syncpkg.NewSystemdUnit(d.deps.FS, d.deps.Paths, d.deps.Service)

	if data.Running != nil {
		if *data.Running {
			if err := unit.Start(); err != nil {
				return d.errorResponse(fmt.Sprintf("starting syncthing: %v", err))
			}
		} else {
			if err := unit.Stop(); err != nil {
				return d.errorResponse(fmt.Sprintf("stopping syncthing: %v", err))
			}
		}
	}

	if data.AutostartEnabled != nil {
		unitName := unit.UnitName()
		if *data.AutostartEnabled {
			if err := d.deps.Service.EnableAutostart(unitName); err != nil {
				return d.errorResponse(fmt.Sprintf("enabling autostart: %v", err))
			}
		} else {
			if err := d.deps.Service.DisableAutostart(unitName); err != nil {
				return d.errorResponse(fmt.Sprintf("disabling autostart: %v", err))
			}
		}
		cfg.Sync.Autostart = *data.AutostartEnabled
	}

	changed := false
	if data.GlobalDiscoveryEnabled != nil {
		cfg.Sync.Syncthing.GlobalDiscoveryEnabled = *data.GlobalDiscoveryEnabled
		changed = true
	}

	if data.AutostartEnabled != nil {
		changed = true
	}

	if !changed {
		return []Event{{
			Type: EventTypeResult,
			Data: SyncSetSettingsResponse{Success: true},
		}}
	}

	path := d.deps.ConfigPath
	if path == "" {
		path, _ = d.deps.Paths.ConfigPath()
	}
	if err := d.configStore.Save(cfg, path); err != nil {
		return d.errorResponse(err.Error())
	}

	if cfg.Sync.Enabled && data.GlobalDiscoveryEnabled != nil {
		collection, err := store.NewCollection(d.deps.FS, d.deps.Paths, cfg.Global.Collection)
		if err != nil {
			return d.errorResponse(err.Error())
		}

		allSystems := d.syncSystems(cfg)
		allEmulators := d.syncEmulators(cfg)
		allFrontends := d.syncFrontends(cfg)
		setup := d.syncSetup()
		if err := setup.UpdateConfig(context.Background(), cfg.Sync, collection.Root(), allSystems, allEmulators, allFrontends); err != nil {
			return d.errorResponse(fmt.Sprintf("updating syncthing config: %v", err))
		}

		if unit.IsEnabled() {
			if err := d.deps.Service.Restart(unit.UnitName()); err != nil {
				log.Error("Failed to restart syncthing after settings change: %v", err)
			}
		}
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SyncSetSettingsResponse{Success: true},
	}}
}

func (d *Daemon) handleSyncDiscoveredDevices() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if !cfg.Sync.Enabled {
		return []Event{{
			Type: EventTypeResult,
			Data: SyncDiscoveredDevicesResponse{Devices: []SyncDiscoveredDevice{}},
		}}
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !client.IsRunning(ctx) {
		return []Event{{
			Type: EventTypeResult,
			Data: SyncDiscoveredDevicesResponse{Devices: []SyncDiscoveredDevice{}},
		}}
	}

	discovered, err := client.GetDiscoveredDevices(ctx)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("getting discovered devices: %v", err))
	}

	devices := make([]SyncDiscoveredDevice, len(discovered))
	for i, dev := range discovered {
		devices[i] = SyncDiscoveredDevice{
			DeviceID:  dev.DeviceID,
			Addresses: dev.Addresses,
		}
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SyncDiscoveredDevicesResponse{Devices: devices},
	}}
}

func (d *Daemon) handleSyncPending() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if !cfg.Sync.Enabled {
		return []Event{{
			Type: EventTypeResult,
			Data: SyncPendingResponse{Pending: false},
		}}
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !client.IsRunning(ctx) {
		return []Event{{
			Type: EventTypeResult,
			Data: SyncPendingResponse{Pending: false},
		}}
	}

	pending, err := client.GetPendingStatus(ctx)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("getting sync pending status: %v", err))
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SyncPendingResponse{
			Pending:    pending.TotalFiles > 0,
			TotalFiles: pending.TotalFiles,
			TotalBytes: pending.TotalBytes,
		},
	}}
}

func (d *Daemon) loadSyncAPIKey() string {
	keyPath := filepath.Join(d.deps.StateDir, "syncthing", "config", ".apikey")
	data, err := d.deps.FS.ReadFile(keyPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (d *Daemon) stopSyncthing(cfg *model.KyarabenConfig) bool {
	unit := syncpkg.NewSystemdUnit(d.deps.FS, d.deps.Paths, d.deps.Service)
	unitName := unit.UnitName()
	state := d.deps.Service.State(unitName)
	if state != "active" && state != "activating" {
		return false
	}
	log.Info("Stopping syncthing before apply")
	if err := d.deps.Service.Stop(unitName); err != nil {
		log.Debug("Error stopping syncthing: %v", err)
	}
	ports := []int{cfg.Sync.Syncthing.GUIPort, cfg.Sync.Syncthing.ListenPort}
	for _, port := range ports {
		if err := syncpkg.WaitForPortRelease(port, 5*time.Second); err != nil {
			log.Debug("Port %d not released: %v", port, err)
		}
	}
	return true
}

func (d *Daemon) syncSystems(cfg *model.KyarabenConfig) []model.SystemID {
	return cfg.EffectiveSyncSystems(d.deps.Registry)
}

func (d *Daemon) syncEmulators(cfg *model.KyarabenConfig) []folders.EmulatorInfo {
	syncInfos := cfg.EffectiveSyncEmulators(d.deps.Registry)
	result := make([]folders.EmulatorInfo, len(syncInfos))
	for i, info := range syncInfos {
		result[i] = folders.EmulatorInfo{
			ID:                 info.ID,
			UsesStatesDir:      info.UsesStatesDir,
			UsesScreenshotsDir: info.UsesScreenshotsDir,
		}
	}
	return result
}

func (d *Daemon) syncFrontends(cfg *model.KyarabenConfig) []model.FrontendID {
	return cfg.EffectiveSyncFrontends()
}

func (d *Daemon) dismissUnwantedPendingFolders(client syncpkg.SyncClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configured, err := client.GetFolderConfigs(ctx)
	if err != nil {
		log.Error("Failed to get configured folders: %v", err)
		return
	}

	configuredIDs := make(map[string]bool)
	for _, f := range configured {
		configuredIDs[f.ID] = true
	}

	pending, err := client.GetPendingFolders(ctx)
	if err != nil {
		log.Error("Failed to get pending folders: %v", err)
		return
	}

	for _, p := range pending {
		if !configuredIDs[p.ID] {
			log.Info("Dismissing pending folder %s (not configured on this device)", p.ID)
			if err := client.DismissPendingFolder(ctx, p.ID, p.OfferedBy); err != nil {
				log.Error("Failed to dismiss pending folder %s: %v", p.ID, err)
			}
		}
	}
}

// ensureSyncthingManaged sets up syncthing if it's not already being managed
// by our systemd unit. "Managed" means systemd state is "active" or "activating".
func (d *Daemon) ensureSyncthingManaged(cfg *model.KyarabenConfig) {
	unit := syncpkg.NewSystemdUnit(d.deps.FS, d.deps.Paths, d.deps.Service)
	state := d.deps.Service.State(unit.UnitName())
	if state == "active" || state == "activating" {
		log.Debug("Syncthing already managed by systemd (state=%s)", state)
		return
	}

	d.syncReconfigMu.Lock()
	if d.syncReconfigRunning {
		d.syncReconfigMu.Unlock()
		return
	}
	if !d.shouldRetrySyncSetup() {
		d.syncReconfigMu.Unlock()
		return
	}
	d.syncReconfigRunning = true
	d.syncReconfigMu.Unlock()

	defer func() {
		d.syncReconfigMu.Lock()
		d.syncReconfigRunning = false
		d.syncReconfigMu.Unlock()
	}()

	collection, err := store.NewCollection(d.deps.FS, d.deps.Paths, cfg.Global.Collection)
	if err != nil {
		log.Error("Failed to create user store for sync setup: %v", err)
		d.recordSyncSetupFailure()
		return
	}

	allSystems := d.syncSystems(cfg)
	allEmulators := d.syncEmulators(cfg)
	allFrontends := d.syncFrontends(cfg)

	setup := d.syncSetup()
	result, err := setup.Install(context.Background(), cfg.Sync, collection.Root(), allSystems, allEmulators, allFrontends, nil)
	if err != nil {
		log.Error("Syncthing setup failed: %v", err)
		d.recordSyncSetupFailure()
		return
	}

	d.resetSyncSetupBackoff()

	manifest, err := d.loadManifest()
	if err != nil {
		log.Error("Failed to load manifest: %v", err)
		return
	}

	manifest.SyncthingInstall = &model.SyncthingInstall{
		Version:             d.deps.Installer.ResolveVersion("syncthing"),
		ConfigSchemaVersion: syncpkg.ConfigSchemaVersion,
		BinaryPath:          result.SyncthingBinary,
		ConfigDir:           result.ConfigDir,
		DataDir:             result.DataDir,
		SystemdUnitPath:     result.SystemdUnitPath,
	}
	if err := manifest.SaveWithBackup(d.deps.ManifestPath); err != nil {
		log.Error("Failed to save manifest: %v", err)
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

}

func (d *Daemon) shouldRetrySyncSetup() bool {
	if d.syncReconfigFailCount == 0 {
		return true
	}

	backoffSeconds := 5 << min(d.syncReconfigFailCount-1, 3)
	if backoffSeconds > 60 {
		backoffSeconds = 60
	}

	elapsed := time.Since(d.syncReconfigLastFailure)
	return elapsed >= time.Duration(backoffSeconds)*time.Second
}

func (d *Daemon) recordSyncSetupFailure() {
	d.syncReconfigMu.Lock()
	defer d.syncReconfigMu.Unlock()
	d.syncReconfigLastFailure = time.Now()
	d.syncReconfigFailCount++
}

func (d *Daemon) resetSyncSetupBackoff() {
	d.syncReconfigMu.Lock()
	defer d.syncReconfigMu.Unlock()
	d.syncReconfigFailCount = 0
	d.syncReconfigLastFailure = time.Time{}
}

func (d *Daemon) reconcileFolderSharing(cfg *model.KyarabenConfig) {
	d.reconcileMu.Lock()
	if d.reconcileRunning {
		d.reconcileMu.Unlock()
		return
	}
	if !d.shouldRetryReconcile() {
		d.reconcileMu.Unlock()
		return
	}
	d.reconcileRunning = true
	d.reconcileMu.Unlock()

	defer func() {
		d.reconcileMu.Lock()
		d.reconcileRunning = false
		d.reconcileMu.Unlock()
	}()

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	localDeviceID, err := client.GetDeviceID(ctx)
	if err != nil {
		log.Debug("Reconcile: failed to get local device ID: %v", err)
		d.recordReconcileFailure()
		return
	}

	devices, err := client.GetConfiguredDevices(ctx)
	if err != nil {
		log.Debug("Reconcile: failed to get configured devices: %v", err)
		d.recordReconcileFailure()
		return
	}

	if len(devices) == 0 {
		d.resetReconcileBackoff()
		return
	}

	deviceIDs := make([]string, len(devices))
	for i, dev := range devices {
		deviceIDs[i] = dev.ID
	}

	folders, err := client.GetFoldersWithDevices(ctx)
	if err != nil {
		log.Debug("Reconcile: failed to get folders with devices: %v", err)
		d.recordReconcileFailure()
		return
	}

	drift := syncpkg.ComputeFolderSharingDrift(folders, deviceIDs, localDeviceID)
	if len(drift) == 0 {
		d.resetReconcileBackoff()
		return
	}

	log.Info("Reconcile: detected folder sharing drift, fixing %d folders", len(drift))
	if err := client.ReconcileFolderSharing(ctx, drift); err != nil {
		log.Error("Reconcile: failed to fix folder sharing: %v", err)
		d.recordReconcileFailure()
		return
	}

	d.resetReconcileBackoff()
	log.Info("Reconcile: folder sharing drift fixed")
}

func (d *Daemon) logCompletionDiagnostics(ctx context.Context, client syncpkg.SyncClient, dev syncpkg.DeviceStatus, completion *syncpkg.CompletionResponse, folders []syncpkg.FolderStatusSummary) {
	d.completionLogMu.Lock()
	if time.Since(d.completionLastLog) < 30*time.Second {
		d.completionLogMu.Unlock()
		return
	}
	d.completionLastLog = time.Now()
	d.completionLogMu.Unlock()

	percent := int(completion.Completion)
	log.Info("Completion diagnostics for %s (%s): aggregate=%d%% globalBytes=%d needBytes=%d globalItems=%d needItems=%d needDeletes=%d",
		dev.Name, dev.ID[:7], percent,
		completion.GlobalBytes, completion.NeedBytes,
		completion.GlobalItems, completion.NeedItems, completion.NeedDeletes)

	for _, f := range folders {
		folderCompletion, err := client.GetFolderCompletionForDevice(ctx, f.ID, dev.ID)
		if err != nil {
			log.Debug("  folder %s: error getting completion: %v", f.ID, err)
			continue
		}
		if int(folderCompletion.Completion) < 100 {
			log.Info("  folder %s: completion=%d%% globalBytes=%d needBytes=%d needItems=%d",
				f.ID, int(folderCompletion.Completion),
				folderCompletion.GlobalBytes, folderCompletion.NeedBytes, folderCompletion.NeedItems)
		}
	}
}

func (d *Daemon) shouldRetryReconcile() bool {
	if d.reconcileFailCount == 0 {
		return true
	}

	backoffSeconds := 5 << min(d.reconcileFailCount-1, 3)
	if backoffSeconds > 60 {
		backoffSeconds = 60
	}

	elapsed := time.Since(d.reconcileLastFailure)
	return elapsed >= time.Duration(backoffSeconds)*time.Second
}

func (d *Daemon) recordReconcileFailure() {
	d.reconcileMu.Lock()
	defer d.reconcileMu.Unlock()
	d.reconcileLastFailure = time.Now()
	d.reconcileFailCount++
}

func (d *Daemon) resetReconcileBackoff() {
	d.reconcileMu.Lock()
	defer d.reconcileMu.Unlock()
	d.reconcileFailCount = 0
	d.reconcileLastFailure = time.Time{}
}

func (d *Daemon) updateSyncConfig(cfg *model.KyarabenConfig, collectionPath string) error {
	allSystems := d.syncSystems(cfg)
	allEmulators := d.syncEmulators(cfg)
	allFrontends := d.syncFrontends(cfg)

	manifest, err := d.loadManifest()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	setup := d.syncSetup()
	result, installErr := setup.Install(context.Background(), cfg.Sync, collectionPath, allSystems, allEmulators, allFrontends, nil)
	if installErr != nil {
		return fmt.Errorf("updating syncthing: %w", installErr)
	}

	expectedVersion := d.deps.Installer.ResolveVersion("syncthing")
	manifest.SyncthingInstall = &model.SyncthingInstall{
		Version:             expectedVersion,
		ConfigSchemaVersion: syncpkg.ConfigSchemaVersion,
		BinaryPath:          result.SyncthingBinary,
		ConfigDir:           result.ConfigDir,
		DataDir:             result.DataDir,
		SystemdUnitPath:     result.SystemdUnitPath,
	}
	if saveErr := manifest.SaveWithBackup(d.deps.ManifestPath); saveErr != nil {
		return fmt.Errorf("saving manifest: %w", saveErr)
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if client.IsRunning(ctx) {
		if err := client.Restart(ctx); err != nil {
			log.Info("Syncthing restart requested (may take a moment)")
		}
	}

	return nil
}

func (d *Daemon) syncSetup() *syncpkg.Setup {
	clientFactory := func(config model.SyncConfig) syncpkg.SyncClient {
		return syncpkg.NewClient(config)
	}
	return syncpkg.NewSetup(d.deps.FS, d.deps.Paths, d.deps.Installer, d.deps.StateDir, d.deps.Service, clientFactory)
}

func (d *Daemon) waitForSyncthingReady(ctx context.Context, client syncpkg.SyncClient) error {
	const maxDuration = 10 * time.Second
	deadline := time.Now().Add(maxDuration)
	attempt := 0
	interval := 50 * time.Millisecond

	for time.Now().Before(deadline) {
		attempt++
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if client.IsRunning(checkCtx) {
			cancel()
			pairingLog.Info("Syncthing ready after %d attempts", attempt)
			return nil
		}
		cancel()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}

		if interval < 500*time.Millisecond {
			interval = interval * 2
		}
	}
	return fmt.Errorf("syncthing not ready after %v", maxDuration)
}

func (d *Daemon) persistSyncEnabled(cfg *model.KyarabenConfig) {
	cfg.Sync.Enabled = true

	path := d.deps.ConfigPath
	if path == "" {
		path, _ = d.deps.Paths.ConfigPath()
	}

	if err := d.configStore.Save(cfg, path); err != nil {
		log.Error("Failed to persist sync config: %v", err)
	}
}

func (d *Daemon) startPendingDeviceLoop(cfg *model.KyarabenConfig, emit func(Event)) {
	d.autoAcceptMu.Lock()
	if d.autoAcceptCancelFunc != nil {
		d.autoAcceptCancelFunc()
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.autoAcceptCancelFunc = cancel
	d.autoAcceptMu.Unlock()

	go func() {
		client := syncpkg.NewClient(cfg.Sync)
		loadedKey := d.loadSyncAPIKey()
		if loadedKey != "" {
			client.SetAPIKey(loadedKey)
		}

		seenDevices := make(map[string]bool)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pollCtx, pollCancel := context.WithTimeout(ctx, 5*time.Second)
				pending, err := client.GetPendingDevices(pollCtx)
				pollCancel()
				if err != nil {
					log.Debug("Pending device loop: error getting pending devices: %v", err)
					continue
				}

				for _, dev := range pending {
					if seenDevices[dev.DeviceID] {
						continue
					}
					seenDevices[dev.DeviceID] = true

					log.Info("Pending device found: %s (%s), requesting user confirmation", dev.Name, dev.DeviceID[:7])
					emit(Event{
						Type: EventTypePendingDevice,
						Data: SyncPendingDeviceEvent{
							DeviceID: dev.DeviceID,
							Name:     dev.Name,
						},
					})
				}
			}
		}
	}()
}

func (d *Daemon) stopAutoAcceptLoop() {
	d.autoAcceptMu.Lock()
	defer d.autoAcceptMu.Unlock()
	if d.autoAcceptCancelFunc != nil {
		d.autoAcceptCancelFunc()
		d.autoAcceptCancelFunc = nil
	}
}

func (d *Daemon) startRelayPollLoop(cfg *model.KyarabenConfig, relayClient *syncpkg.RelayClient, code string, syncClient syncpkg.SyncClient, emit func(Event)) {
	go func() {
		pairingLog.Info("Starting relay poll loop for code %s", code)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		timeout := time.After(5 * time.Minute)
		pollCount := 0

		for {
			select {
			case <-timeout:
				pairingLog.Info("Relay poll loop timed out for code %s after %d polls", code, pollCount)
				_ = relayClient.DeleteSession(context.Background(), code)
				return
			case <-ticker.C:
				pollCount++
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				resp, err := relayClient.GetResponse(ctx, code)
				cancel()
				if err != nil {
					pairingLog.Info("Relay poll %d error for code %s: %v", pollCount, code, err)
					continue
				}

				if !resp.Ready {
					pairingLog.Debug("Relay poll %d for code %s: not ready", pollCount, code)
					continue
				}

				pairingLog.Info("Relay poll %d: joiner device ID received: %s", pollCount, resp.DeviceID[:7]+"...")

				addCtx, addCancel := context.WithTimeout(context.Background(), 10*time.Second)
				pairingLog.Info("Adding joiner device to syncthing config")
				if err := syncClient.AddDeviceAutoName(addCtx, resp.DeviceID); err != nil {
					pairingLog.Error("Failed to add device from relay: %v", err)
					addCancel()
					continue
				}
				pairingLog.Info("Sharing folders with joiner device")
				if err := syncClient.ShareFoldersWithDevice(addCtx, resp.DeviceID); err != nil {
					pairingLog.Error("Failed to share folders with relay device: %v", err)
				}
				addCancel()

				pairingLog.Info("Initiator pairing via relay completed successfully for device %s", resp.DeviceID[:7])
				pairingLog.Info("Emitting progress event with deviceId: %s", resp.DeviceID)
				emit(Event{
					Type: EventTypeProgress,
					Data: SyncPairingProgressEvent{
						Message:  "Device connected via pairing code",
						DeviceID: resp.DeviceID,
					},
				})
				pairingLog.Info("Progress event emitted successfully")

				_ = relayClient.DeleteSession(context.Background(), code)
				return
			}
		}
	}()
}

func isRelayCode(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, c := range s {
		isUpper := c >= 'A' && c <= 'Z'
		isDigit := c >= '0' && c <= '9'
		if !isUpper && !isDigit {
			return false
		}
	}
	return true
}

func (d *Daemon) handleUninstallPreview() []Event {
	manifest, err := d.manifestStore.Load(d.deps.ManifestPath)
	if err != nil {
		log.Error("Failed to load manifest for uninstall preview: %v", err)
		manifest = model.NewManifest()
	}

	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}
	collection := cfg.Global.Collection

	configDir, _ := d.deps.Paths.ConfigDir()

	var desktopFiles []string
	for _, f := range manifest.DesktopFiles {
		if d.fileExists(f) {
			desktopFiles = append(desktopFiles, f)
		}
	}

	var iconFiles []string
	for _, f := range manifest.IconFiles {
		if d.fileExists(f) {
			iconFiles = append(iconFiles, f)
		}
	}

	var configFiles []string
	for _, cfg := range manifest.ManagedConfigs {
		path, err := cfg.Target.ResolveWith(d.deps.Resolver)
		if err == nil && d.fileExists(path) {
			configFiles = append(configFiles, path)
		}
	}

	var kyarabenFiles []string
	if manifest.KyarabenInstall != nil {
		for _, p := range []string{
			manifest.KyarabenInstall.AppPath,
			manifest.KyarabenInstall.CLIPath,
			manifest.KyarabenInstall.DesktopPath,
		} {
			if p != "" && d.fileExists(p) {
				kyarabenFiles = append(kyarabenFiles, p)
			}
		}
	}
	if len(kyarabenFiles) == 0 {
		homeDir, _ := os.UserHomeDir()
		dataDir, _ := paths.DataDir()
		kyarabenPaths := []string{
			filepath.Join(homeDir, ".local", "bin", "kyaraben-ui"),
			filepath.Join(homeDir, ".local", "bin", "kyaraben"),
			filepath.Join(dataDir, "applications", "kyaraben.desktop"),
		}
		for _, p := range kyarabenPaths {
			if d.fileExists(p) {
				kyarabenFiles = append(kyarabenFiles, p)
			}
		}
	}

	var syncthingFiles []string
	if manifest.SyncthingInstall != nil {
		si := manifest.SyncthingInstall
		if si.BinaryPath != "" && d.fileExists(si.BinaryPath) {
			syncthingFiles = append(syncthingFiles, si.BinaryPath)
		}
		for _, dir := range []string{si.ConfigDir, si.DataDir} {
			if dir != "" && d.dirExists(dir) {
				syncthingFiles = append(syncthingFiles, dir)
			}
		}
	}
	syncServices, _ := syncpkg.FindKyarabenSyncthingServices()
	syncthingFiles = append(syncthingFiles, syncServices...)

	var retroArchCoresDir string
	var retroArchCoreFiles []string
	if coresDir, err := d.deps.Paths.CoresDir(); err == nil && d.dirExists(coresDir) {
		retroArchCoresDir = coresDir
		entries, _ := os.ReadDir(coresDir)
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_libretro.so") {
				retroArchCoreFiles = append(retroArchCoreFiles, entry.Name())
			}
		}
	}

	return []Event{{
		Type: EventTypeResult,
		Data: UninstallPreviewResponse{
			StateDir:           d.deps.StateDir,
			StateDirExists:     d.dirExists(d.deps.StateDir),
			RetroArchCoresDir:  retroArchCoresDir,
			RetroArchCoreFiles: retroArchCoreFiles,
			DesktopFiles:       desktopFiles,
			IconFiles:          iconFiles,
			ConfigFiles:        configFiles,
			KyarabenFiles:      kyarabenFiles,
			SyncthingFiles:     syncthingFiles,
			Preserved: PreservedPaths{
				Collection: shortenPath(collection),
				ConfigDir:  shortenPath(configDir),
			},
		},
	}}
}

func (d *Daemon) handleUninstall() []Event {
	manifest, err := d.manifestStore.Load(d.deps.ManifestPath)
	if err != nil {
		log.Error("Failed to load manifest for uninstall: %v", err)
		manifest = model.NewManifest()
	}

	var removedFiles []string
	var errors []string

	if manifest.KyarabenInstall != nil {
		ki := manifest.KyarabenInstall
		for _, path := range []string{ki.AppPath, ki.CLIPath, ki.DesktopPath} {
			if path != "" && d.fileExists(path) {
				if err := os.Remove(path); err != nil {
					errors = append(errors, fmt.Sprintf("could not remove %s: %v", path, err))
				} else {
					removedFiles = append(removedFiles, path)
				}
			}
		}
	}

	for _, cfg := range manifest.ManagedConfigs {
		path, err := cfg.Target.ResolveWith(d.deps.Resolver)
		if err != nil {
			continue
		}
		if d.fileExists(path) {
			if err := os.Remove(path); err != nil {
				errors = append(errors, fmt.Sprintf("could not remove %s: %v", path, err))
			} else {
				removedFiles = append(removedFiles, path)
			}
		}
	}

	for _, f := range manifest.DesktopFiles {
		if d.fileExists(f) {
			if err := os.Remove(f); err != nil {
				errors = append(errors, fmt.Sprintf("could not remove %s: %v", f, err))
			} else {
				removedFiles = append(removedFiles, f)
			}
		}
	}

	for _, f := range manifest.IconFiles {
		if d.fileExists(f) {
			if err := os.Remove(f); err != nil {
				errors = append(errors, fmt.Sprintf("could not remove %s: %v", f, err))
			} else {
				removedFiles = append(removedFiles, f)
			}
		}
	}

	cfg, cfgErr := d.loadConfig()
	var syncPorts []int
	if cfgErr == nil {
		syncPorts = []int{cfg.Sync.Syncthing.GUIPort, cfg.Sync.Syncthing.ListenPort}
	}

	syncServices, _ := syncpkg.FindKyarabenSyncthingServices()
	for _, servicePath := range syncServices {
		if err := syncpkg.StopAndRemoveServiceWithWait(d.deps.Service, servicePath, 10*time.Second, syncPorts); err != nil {
			errors = append(errors, fmt.Sprintf("could not remove syncthing service %s: %v", servicePath, err))
		} else {
			removedFiles = append(removedFiles, servicePath)
		}
	}

	if d.dirExists(d.deps.StateDir) {
		if err := forceRemoveAll(d.deps.StateDir); err != nil {
			errors = append(errors, fmt.Sprintf("could not remove %s: %v", d.deps.StateDir, err))
		} else {
			removedFiles = append(removedFiles, d.deps.StateDir)
		}
	}

	dataDir, _ := paths.DataDir()
	iconsDir := filepath.Join(dataDir, "icons", "hicolor")
	launcher.UpdateIconCaches(iconsDir)

	return []Event{{
		Type: EventTypeResult,
		Data: UninstallResponse{
			Success:      len(errors) == 0,
			RemovedFiles: removedFiles,
			Errors:       errors,
		},
	}}
}

func forceRemoveAll(path string) error {
	if err := forceChmodRecursive(path); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func forceChmodRecursive(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}

	if !info.IsDir() {
		return os.Chmod(path, 0644)
	}

	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := forceChmodRecursive(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}

func (d *Daemon) fileExists(path string) bool {
	info, err := d.deps.FS.Stat(path)
	return err == nil && !info.IsDir()
}

func (d *Daemon) dirExists(path string) bool {
	info, err := d.deps.FS.Stat(path)
	return err == nil && info.IsDir()
}

func (d *Daemon) handleInstallKyaraben(data *InstallKyarabenRequest) []Event {
	if data == nil || data.AppImagePath == "" || data.SidecarPath == "" {
		return d.errorResponse("appImagePath and sidecarPath are required")
	}

	result, err := d.deps.LauncherManager.InstallApp(data.AppImagePath, data.SidecarPath)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	lockDir := filepath.Dir(d.deps.ManifestPath)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return d.errorResponse(fmt.Sprintf("creating state directory: %v", err))
	}
	lockPath := filepath.Join(lockDir, "apply.lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("creating lock file: %v", err))
	}
	defer func() { _ = lockFile.Close() }()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return d.errorResponse(fmt.Sprintf("acquiring lock: %v", err))
	}
	defer func() { _ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN) }()

	manifest, err := d.loadManifest()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	var shortcutGiven bool
	var shortcutPath string
	if manifest.KyarabenInstall != nil {
		shortcutGiven = manifest.KyarabenInstall.DesktopShortcutGiven
		shortcutPath = manifest.KyarabenInstall.DesktopShortcutPath
	}

	manifest.KyarabenInstall = &model.KyarabenInstall{
		AppPath:              result.AppPath,
		CLIPath:              result.CLIPath,
		DesktopPath:          result.DesktopPath,
		DesktopShortcutGiven: shortcutGiven,
		DesktopShortcutPath:  shortcutPath,
	}

	if !shortcutGiven {
		newShortcutPath, err := d.deps.LauncherManager.CreateDesktopShortcut()
		if err != nil {
			log.Debug("Failed to create desktop shortcut: %v", err)
		} else {
			manifest.KyarabenInstall.DesktopShortcutGiven = true
			manifest.KyarabenInstall.DesktopShortcutPath = newShortcutPath
			log.Info("Created one-shot desktop shortcut: %s", newShortcutPath)
		}
	}

	if saveErr := manifest.SaveWithBackup(d.deps.ManifestPath); saveErr != nil {
		return d.errorResponse(fmt.Sprintf("failed to save manifest: %v", saveErr))
	}

	log.Info("Kyaraben installed successfully to %s", result.AppPath)

	return []Event{{
		Type: EventTypeResult,
		Data: InstallKyarabenResponse{Success: true, AppPath: result.AppPath},
	}}
}

func (d *Daemon) handleInstallStatus() []Event {
	manifest, err := d.manifestStore.Load(d.deps.ManifestPath)
	if err != nil {
		log.Error("Failed to load manifest for install status: %v", err)
		manifest = model.NewManifest()
	}

	if manifest.KyarabenInstall != nil {
		ki := manifest.KyarabenInstall
		cliExists := ki.CLIPath != "" && d.fileExists(ki.CLIPath)
		desktopExists := ki.DesktopPath != "" && d.fileExists(ki.DesktopPath)
		appExists := ki.AppPath != "" && d.fileExists(ki.AppPath)

		return []Event{{
			Type: EventTypeResult,
			Data: InstallStatusResponse{
				Installed:   cliExists && desktopExists,
				AppPath:     boolToPath(appExists, ki.AppPath),
				DesktopPath: boolToPath(desktopExists, ki.DesktopPath),
				CLIPath:     boolToPath(cliExists, ki.CLIPath),
			},
		}}
	}

	status := d.deps.LauncherManager.GetInstallStatus()
	installed := status.CLIPath != "" && status.DesktopPath != ""

	return []Event{{
		Type: EventTypeResult,
		Data: InstallStatusResponse{
			Installed:   installed,
			AppPath:     status.AppPath,
			DesktopPath: status.DesktopPath,
			CLIPath:     status.CLIPath,
		},
	}}
}

func boolToPath(exists bool, path string) string {
	if exists {
		return path
	}
	return ""
}

func (d *Daemon) handleRefreshIconCaches() []Event {
	refreshed := d.deps.LauncherManager.RefreshIconCaches()
	return []Event{{
		Type: EventTypeResult,
		Data: RefreshIconCachesResponse{Refreshed: refreshed},
	}}
}

func retroArchCoreName(id model.EmulatorID) string {
	if !strings.HasPrefix(string(id), "retroarch:") {
		return ""
	}
	return strings.TrimPrefix(string(id), "retroarch:")
}

func (d *Daemon) maybeCreateDesktopShortcut() {
	manifest, err := d.loadManifest()
	if err != nil {
		log.Debug("Failed to load manifest for desktop shortcut: %v", err)
		return
	}

	if manifest.KyarabenInstall == nil {
		return
	}

	if manifest.KyarabenInstall.DesktopShortcutGiven {
		return
	}

	shortcutPath, err := d.deps.LauncherManager.CreateDesktopShortcut()
	if err != nil {
		log.Debug("Failed to create desktop shortcut: %v", err)
		return
	}

	manifest.KyarabenInstall.DesktopShortcutGiven = true
	manifest.KyarabenInstall.DesktopShortcutPath = shortcutPath
	if err := d.manifestStore.Save(manifest, d.deps.ManifestPath); err != nil {
		log.Debug("Failed to save manifest after desktop shortcut: %v", err)
		return
	}

	log.Info("Created one-shot desktop shortcut: %s", shortcutPath)
}

func (d *Daemon) handleGetStorageDevices() []Event {
	var devices []StorageDevice

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return d.errorResponse(fmt.Sprintf("getting home directory: %v", err))
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(homeDir, &stat); err == nil {
		devices = append(devices, StorageDevice{
			ID:         "internal",
			Label:      "Internal storage",
			Path:       filepath.Join(homeDir, "Emulation"),
			FreeBytes:  int64(stat.Bavail) * stat.Bsize,
			TotalBytes: int64(stat.Blocks) * stat.Bsize,
		})
	}

	user := os.Getenv("USER")
	if user == "" {
		user = filepath.Base(homeDir)
	}

	mediaPaths := []string{
		filepath.Join("/run/media", user),
		filepath.Join("/media", user),
	}

	var internalDevID uint64
	if info, err := os.Stat(homeDir); err == nil {
		if sys, ok := info.Sys().(*syscall.Stat_t); ok {
			internalDevID = sys.Dev
		}
	}

	type externalMount struct {
		path     string
		isSDCard bool
		stat     syscall.Statfs_t
	}
	var candidates []externalMount

	for _, mediaBase := range mediaPaths {
		entries, err := os.ReadDir(mediaBase)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			mountPath := filepath.Join(mediaBase, entry.Name())

			info, err := os.Stat(mountPath)
			if err != nil {
				continue
			}
			sys, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				continue
			}
			if sys.Dev == internalDevID {
				continue
			}

			var stat syscall.Statfs_t
			if err := syscall.Statfs(mountPath, &stat); err != nil {
				continue
			}

			candidates = append(candidates, externalMount{
				path:     mountPath,
				isSDCard: isSDCard(mountPath),
				stat:     stat,
			})
		}
	}

	if len(candidates) > 0 {
		best := candidates[0]
		for _, c := range candidates[1:] {
			if c.isSDCard && !best.isSDCard {
				best = c
			}
		}
		devices = append(devices, StorageDevice{
			ID:         "external",
			Label:      filepath.Base(best.path),
			Path:       filepath.Join(best.path, "Emulation"),
			FreeBytes:  int64(best.stat.Bavail) * best.stat.Bsize,
			TotalBytes: int64(best.stat.Blocks) * best.stat.Bsize,
		})
	}

	return []Event{{
		Type: EventTypeResult,
		Data: StorageDevicesResponse{Devices: devices},
	}}
}

func isSDCard(mountPath string) bool {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return false
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		if fields[1] == mountPath {
			device := filepath.Base(fields[0])
			return strings.HasPrefix(device, "mmcblk")
		}
	}
	return false
}

func (d *Daemon) handleImportScan(req *ImportScanRequest) []Event {
	if req == nil {
		return d.errorResponse("import_scan requires data")
	}

	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(fmt.Sprintf("loading config: %v", err))
	}

	collection, err := store.NewCollection(d.deps.FS, d.deps.Paths, cfg.Global.Collection)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("creating collection: %v", err))
	}

	scanner := importscanner.NewScanner(d.deps.FS, d.deps.Registry, collection)
	report, err := scanner.Scan(importscanner.ScanOptions{
		SourcePath: req.SourcePath,
		ESDEPath:   req.ESDEPath,
		Layout:     req.Layout,
	})
	if err != nil {
		return d.errorResponse(fmt.Sprintf("scanning: %v", err))
	}

	resp := convertImportReport(report, cfg)
	return []Event{{
		Type: EventTypeResult,
		Data: resp,
	}}
}

func convertImportReport(r *importscanner.ImportReport, cfg *model.KyarabenConfig) ImportScanResponse {
	resp := ImportScanResponse{
		SourcePath:   r.SourcePath,
		ESDEPath:     r.ESDEPath,
		KyarabenPath: r.KyarabenPath,
		Mode:         string(r.Mode),
		Systems:      []ImportSystemReport{},
		Summary: ImportDiffSummary{
			TotalOnlyInSource:   r.Summary.TotalOnlyInSource,
			TotalOnlyInKyaraben: r.Summary.TotalOnlyInKyaraben,
		},
	}

	for _, sys := range r.Systems {
		enabledEmulators := cfg.Systems[sys.System]
		sysReport := ImportSystemReport{
			System:     sys.System,
			SystemName: sys.SystemName,
			Enabled:    len(enabledEmulators) > 0,
			SystemData: []ImportDataComparison{},
			Emulators:  []ImportEmulatorReport{},
		}
		for _, data := range sys.SystemData {
			sysReport.SystemData = append(sysReport.SystemData, convertDataComparison(data))
		}
		for _, emu := range sys.Emulators {
			emuEnabled := false
			for _, e := range enabledEmulators {
				if e == emu.Emulator {
					emuEnabled = true
					break
				}
			}
			emuReport := ImportEmulatorReport{
				Emulator:     emu.Emulator,
				EmulatorName: emu.EmulatorName,
				Enabled:      emuEnabled,
				EmulatorData: []ImportDataComparison{},
			}
			for _, data := range emu.EmulatorData {
				emuReport.EmulatorData = append(emuReport.EmulatorData, convertDataComparison(data))
			}
			sysReport.Emulators = append(sysReport.Emulators, emuReport)
		}
		resp.Systems = append(resp.Systems, sysReport)
	}

	for _, fe := range r.Frontends {
		feReport := ImportFrontendReport{
			Frontend:     fe.Frontend,
			FrontendName: fe.FrontendName,
			FrontendData: []ImportDataComparison{},
		}
		for _, data := range fe.FrontendData {
			feReport.FrontendData = append(feReport.FrontendData, convertDataComparison(data))
		}
		resp.Frontends = append(resp.Frontends, feReport)
	}

	return resp
}

func convertDataComparison(d importscanner.DataComparison) ImportDataComparison {
	result := ImportDataComparison{
		DataType: string(d.DataType),
		Source:   convertFolderInfo(d.Source),
		Kyaraben: convertFolderInfo(d.Kyaraben),
		Diff:     convertDiffInfo(d.Diff),
		Notes:    d.Notes,
	}
	return result
}

func convertFolderInfo(f importscanner.FolderInfo) ImportFolderInfo {
	result := ImportFolderInfo{
		Path:      f.Path,
		FileCount: f.FileCount,
		TotalSize: f.TotalSize,
		Exists:    f.Exists,
		IsFlat:    f.IsFlat,
	}
	if f.Symlink != nil {
		result.Symlink = &ImportSymlinkInfo{
			Target: f.Symlink.Target,
			Intact: f.Symlink.Intact,
		}
	}
	return result
}

func convertDiffInfo(d importscanner.DiffInfo) ImportDiffInfo {
	result := ImportDiffInfo{
		SourceDelta:   d.SourceDelta,
		KyarabenDelta: d.KyarabenDelta,
	}
	for _, f := range d.OnlyInSource {
		result.OnlyInSource = append(result.OnlyInSource, ImportFileInfo{
			RelPath: f.RelPath,
			Size:    f.Size,
		})
	}
	for _, f := range d.OnlyInKyaraben {
		result.OnlyInKyaraben = append(result.OnlyInKyaraben, ImportFileInfo{
			RelPath: f.RelPath,
			Size:    f.Size,
		})
	}
	return result
}
