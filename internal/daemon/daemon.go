package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/apply"
	"github.com/fnune/kyaraben/internal/doctor"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/emulators/symlink"
	"github.com/fnune/kyaraben/internal/hardware"
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

type Daemon struct {
	fs              vfs.FS
	paths           *paths.Paths
	configStore     *model.ConfigStore
	manifestStore   *model.ManifestStore
	configPath      string
	stateDir        string
	manifestPath    string
	reg             *registry.Registry
	installer       packages.Installer
	configWriter    *emulators.ConfigWriter
	launcherManager *launcher.Manager
	service         syncpkg.ServiceManager

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
}

func New(fs vfs.FS, p *paths.Paths, configPath, stateDir, manifestPath string, reg *registry.Registry, installer packages.Installer, configWriter *emulators.ConfigWriter, launcherManager *launcher.Manager, service syncpkg.ServiceManager) *Daemon {
	return &Daemon{
		fs:              fs,
		paths:           p,
		configStore:     model.NewConfigStore(fs),
		manifestStore:   model.NewManifestStore(fs),
		configPath:      configPath,
		stateDir:        stateDir,
		manifestPath:    manifestPath,
		reg:             reg,
		installer:       installer,
		configWriter:    configWriter,
		launcherManager: launcherManager,
		service:         service,
	}
}

func NewDefault(p *paths.Paths, configPath, stateDir, manifestPath string, reg *registry.Registry, installer packages.Installer, configWriter *emulators.ConfigWriter, launcherManager *launcher.Manager) *Daemon {
	return New(vfs.OSFS, p, configPath, stateDir, manifestPath, reg, installer, configWriter, launcherManager, syncpkg.NewDefaultServiceManager())
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

func folderLabel(id string) string {
	id = strings.TrimPrefix(id, "kyaraben-")
	parts := strings.SplitN(id, "-", 2)
	if len(parts) == 2 {
		return fmt.Sprintf("%s (%s)", parts[1], parts[0])
	}
	return id
}

func computeFolderPath(userStoreRoot, folderID string) string {
	id := strings.TrimPrefix(folderID, "kyaraben-")
	parts := strings.SplitN(id, "-", 2)
	if len(parts) == 2 {
		return filepath.Join(userStoreRoot, parts[0], parts[1])
	}
	return filepath.Join(userStoreRoot, id)
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
		return d.handleSyncEnable(nil, emit)
	case CommandTypeSyncRevertFolder:
		return d.handleSyncRevertFolder(nil)
	case CommandTypeSyncLocalChanges:
		return d.handleSyncLocalChanges(nil)
	case CommandTypeSyncReset:
		return d.handleSyncReset()
	case CommandTypeSyncDiscoveredDevices:
		return d.handleSyncDiscoveredDevices()
	case CommandTypeGetStorageDevices:
		return d.handleGetStorageDevices()
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

func (d *Daemon) HandleInstallKyaraben(cmd InstallKyarabenCommand, emit func(Event)) []Event {
	return d.handleInstallKyaraben(&cmd.Data)
}

func (d *Daemon) HandleSyncEnable(cmd SyncEnableCommand, emit func(Event)) []Event {
	return d.handleSyncEnable(&cmd.Data, emit)
}

func (d *Daemon) HandleSyncRevertFolder(cmd SyncRevertFolderCommand, emit func(Event)) []Event {
	return d.handleSyncRevertFolder(&cmd.Data)
}

func (d *Daemon) HandleSyncLocalChanges(cmd SyncLocalChangesCommand, emit func(Event)) []Event {
	return d.handleSyncLocalChanges(&cmd.Data)
}

func (d *Daemon) errorResponse(msg string) []Event {
	return []Event{{
		Type: EventTypeError,
		Data: ErrorResponse{Error: msg},
	}}
}

func (d *Daemon) loadManifest() (*model.Manifest, error) {
	manifest, err := d.manifestStore.Load(d.manifestPath)
	if err != nil {
		return nil, fmt.Errorf("manifest data appears corrupted: %w. Please report this as a bug and run 'kyaraben apply' to restore your configuration", err)
	}
	return manifest, nil
}

func (d *Daemon) loadConfig() (*model.KyarabenConfig, error) {
	path := d.configPath
	if path == "" {
		var err error
		path, err = d.paths.ConfigPath()
		if err != nil {
			return nil, err
		}
	}
	cfg, err := d.configStore.Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := model.NewDefaultConfig()
			if d.paths.Instance != "" {
				offset := d.paths.InstancePortOffset()
				cfg.Sync.Syncthing.ListenPort = 22100 + offset
				cfg.Sync.Syncthing.DiscoveryPort = 21127 + offset
				cfg.Sync.Syncthing.GUIPort = 8484 + offset
				cfg.Global.UserStore = "~/Emulation-" + d.paths.Instance
			}
			return cfg, nil
		}
		return nil, err
	}
	return cfg, nil
}

func (d *Daemon) handleStatus() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	configPath := d.configPath
	if configPath == "" {
		configPath, _ = d.paths.ConfigPath()
	}

	userStore, err := store.NewUserStore(d.fs, d.paths, cfg.Global.UserStore)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	result, err := status.NewGetter(d.fs, d.paths).Get(context.Background(), cfg, configPath, d.reg, userStore, d.manifestPath)
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
		if e, err := d.reg.GetEmulator(emu.ID); err == nil && e.Launcher.Binary != "" {
			installed.ExecLine = fmt.Sprintf("%s/%s", d.launcherManager.BinDir(), e.Launcher.Binary)
		}

		installed.Paths = make(map[string]EmulatorPaths)
		emuDef, _ := d.reg.GetEmulator(emu.ID)
		for sysID, emuIDs := range cfg.Systems {
			for _, emuID := range emuIDs {
				if emuID == emu.ID {
					paths := EmulatorPaths{
						Roms: shortenPath(userStore.SystemRomsDir(sysID)),
					}
					if emuDef.PathUsage.UsesBiosDir {
						paths.Bios = shortenPath(userStore.SystemBiosDir(sysID))
					}
					if emuDef.PathUsage.UsesSavesDir {
						paths.Saves = shortenPath(userStore.SystemSavesDir(sysID))
					}
					if emuDef.PathUsage.UsesStatesDir {
						paths.Savestates = shortenPath(userStore.EmulatorStatesDir(emu.ID))
					}
					if emuDef.PathUsage.UsesScreenshotsDir {
						paths.Screenshots = shortenPath(userStore.EmulatorScreenshotsDir(emu.ID))
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
		installedFrontends[i] = InstalledFrontend{
			ID:      fe.ID,
			Version: fe.Version,
		}
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

	return []Event{{
		Type: EventTypeResult,
		Data: StatusResponse{
			UserStore:               result.UserStorePath,
			EnabledSystems:          systems,
			InstalledEmulators:      installedEmulators,
			InstalledFrontends:      installedFrontends,
			Symlinks:                symlinks,
			LastApplied:             result.LastApplied.Format(time.RFC3339),
			HealthWarning:           result.HealthWarning,
			KyarabenVersion:         version.Get(),
			ManifestKyarabenVersion: manifestVersion,
		},
	}}
}

func (d *Daemon) handleDoctor() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	userStore, err := store.NewUserStore(d.fs, d.paths, cfg.Global.UserStore)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	result, err := doctor.Run(context.Background(), cfg, d.reg, userStore)
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

	versionOverrides, err := cfg.BuildVersionOverrides(d.reg.GetEmulator, d.reg.GetFrontend)
	if err != nil {
		return d.errorResponse(err.Error())
	}
	d.installer.SetVersionOverrides(versionOverrides)

	userStore, err := store.NewUserStore(d.fs, d.paths, cfg.Global.UserStore)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	// Acquire exclusive lock to prevent concurrent Apply operations
	lockDir := filepath.Dir(d.manifestPath)
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

	applier := apply.NewApplier(
		d.fs,
		d.installer,
		d.configWriter,
		d.reg,
		d.manifestPath,
		d.launcherManager,
		model.OSBaseDirResolver{},
		symlink.NewCreator(d.fs),
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
				},
			}
			if emit != nil {
				emit(event)
			}
		},
	}

	_, err = applier.Apply(ctx, cfg, userStore, opts)
	if ctx.Err() != nil {
		return []Event{{
			Type: EventTypeCancelled,
			Data: CancelledResponse{Message: "Installation cancelled"},
		}}
	}
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if _, err := d.launcherManager.InstallKyaraben("", ""); err != nil {
		log.Debug("Failed to install Kyaraben to PATH: %v", err)
	}

	if cfg.Sync.Enabled {
		if err := d.updateSyncConfig(cfg, userStore.Root()); err != nil {
			log.Info("Failed to update sync config: %v", err)
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

	userStore, err := store.NewUserStore(d.fs, d.paths, cfg.Global.UserStore)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	applier := apply.NewApplier(
		d.fs,
		d.installer,
		d.configWriter,
		d.reg,
		d.manifestPath,
		d.launcherManager,
		model.OSBaseDirResolver{},
		symlink.NewCreator(d.fs),
	)
	applier.SteamManager = steam.NewDefaultManager()

	preflight, err := applier.Preflight(context.Background(), cfg, userStore)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("preflight check: %v", err))
	}

	manifest, err := d.manifestStore.Load(d.manifestPath)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("loading manifest: %v", err))
	}

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

		diff, err := emulators.ComputeDiffWithBaseline(patch, baselinePtr)
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

		for _, ku := range diff.KyarabenUpdates {
			key := ku.Path[len(ku.Path)-1]
			if state.seenKyarabenKeys[key] {
				continue
			}
			state.seenKyarabenKeys[key] = true
			state.diff.KyarabenUpdates = append(state.diff.KyarabenUpdates, KyarabenUpdateDetail{
				Key:      key,
				OldValue: ku.OldValue,
				NewValue: ku.NewValue,
			})
		}

		for _, uc := range diff.UserChanges {
			key := uc.Path[len(uc.Path)-1]
			if state.seenUserKeys[key] {
				continue
			}
			state.seenUserKeys[key] = true
			state.diff.UserChanges = append(state.diff.UserChanges, UserChangeDetail{
				Key:           key,
				BaselineValue: uc.BaselineValue,
				CurrentValue:  uc.CurrentValue,
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
	systems := d.reg.AllSystems()
	vers, _ := versions.Get()
	detectedTarget := hardware.DetectTarget().Name

	result := make(GetSystemsResponse, 0, len(systems))
	for _, sys := range systems {
		emus := d.reg.GetEmulatorsForSystem(sys.ID)
		emuList := make([]EmulatorRef, 0, len(emus))
		for _, emu := range emus {
			ref := EmulatorRef{
				ID:          emu.ID,
				Name:        emu.Name,
				PackageName: emu.Package.PackageName(),
			}

			if vers != nil {
				if spec, ok := vers.GetPackage(emu.Package.PackageName()); ok {
					ref.DefaultVersion = spec.Default
					availableVersions := spec.AvailableVersions()
					sort.Strings(availableVersions)
					ref.AvailableVersions = availableVersions

					if entry := spec.GetDefault(); entry != nil {
						if target := entry.SelectTarget(detectedTarget); target != "" {
							if build := entry.Target(target); build != nil && build.Size > 0 {
								ref.DownloadBytes = build.Size
							}
						}
					}
				}

				if coreName := retroArchCoreName(emu.ID); coreName != "" {
					ref.CoreBytes = vers.GetCoreSize(coreName)
				}
			}

			emuList = append(emuList, ref)
		}

		var defaultEmuID model.EmulatorID
		if defaultEmu, err := d.reg.GetDefaultEmulator(sys.ID); err == nil {
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
		})
	}

	return []Event{{
		Type: EventTypeResult,
		Data: result,
	}}
}

func (d *Daemon) handleGetFrontends() []Event {
	frontends := d.reg.AllFrontends()
	vers, _ := versions.Get()
	detectedTarget := hardware.DetectTarget().Name

	result := make(GetFrontendsResponse, 0, len(frontends))
	for _, fe := range frontends {
		ref := FrontendRef{
			ID:   fe.ID,
			Name: fe.Name,
		}

		if vers != nil {
			if spec, ok := vers.GetPackage(fe.Package.PackageName()); ok {
				ref.DefaultVersion = spec.Default
				availableVersions := spec.AvailableVersions()
				sort.Strings(availableVersions)
				ref.AvailableVersions = availableVersions

				if entry := spec.GetDefault(); entry != nil {
					if target := entry.SelectTarget(detectedTarget); target != "" {
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
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(fmt.Sprintf("loading config: %v", err))
	}

	systems := make(map[string][]model.EmulatorID)
	for sys, emulators := range cfg.Systems {
		systems[string(sys)] = emulators
	}

	emulators := make(map[string]EmulatorConfResponse)
	for emuID, emuConf := range cfg.Emulators {
		if emuConf.Version != "" {
			emulators[string(emuID)] = EmulatorConfResponse{
				Version: emuConf.Version,
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

	return []Event{{
		Type: EventTypeResult,
		Data: ConfigResponse{
			UserStore: cfg.Global.UserStore,
			Systems:   systems,
			Emulators: emulators,
			Frontends: frontends,
		},
	}}
}

func (d *Daemon) handleSetConfig(data *SetConfigRequest) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(fmt.Sprintf("loading config: %v", err))
	}

	if data != nil {
		if data.UserStore != "" {
			cfg.Global.UserStore = data.UserStore
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
			cfg.Emulators = make(map[model.EmulatorID]model.EmulatorConf)
			for emuStr, emuConf := range data.Emulators {
				cfg.Emulators[model.EmulatorID(emuStr)] = model.EmulatorConf{
					Version: emuConf.Version,
				}
			}
		}

		if data.Frontends != nil {
			cfg.Frontends = make(map[model.FrontendID]model.FrontendConfig)
			for feStr, feConf := range data.Frontends {
				cfg.Frontends[model.FrontendID(feStr)] = model.FrontendConfig{
					Enabled: feConf.Enabled,
					Version: feConf.Version,
				}
			}
		}
	}

	path := d.configPath
	if path == "" {
		path, _ = d.paths.ConfigPath()
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

	unit := syncpkg.NewSystemdUnit(d.fs, d.paths, d.service)
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
				Enabled:          false,
				Installed:        installed,
				ServiceInstalled: serviceInstalled,
			},
		}}
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	} else {
		log.Debug("No API key found at %s", filepath.Join(d.stateDir, "syncthing", "config", ".apikey"))
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
				Enabled:          true,
				Mode:             string(cfg.Sync.Mode),
				Running:          false,
				Installed:        installed,
				ServiceInstalled: serviceInstalled,
				GUIURL:           fmt.Sprintf("http://127.0.0.1:%d", cfg.Sync.Syncthing.GUIPort),
				ServiceError:     serviceError,
			},
		}}
	}

	syncStatus, err := client.GetStatus(ctx)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	userStore, err := store.NewUserStore(d.fs, d.paths, cfg.Global.UserStore)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	devices := make([]SyncDevice, len(syncStatus.Devices))
	for i, dev := range syncStatus.Devices {
		devices[i] = SyncDevice{
			ID:        dev.ID,
			Name:      dev.Name,
			Connected: dev.Connected,
			Paused:    dev.Paused,
		}
		if syncStatus.Mode == model.SyncModePrimary && dev.Connected {
			if completion, err := client.GetDeviceCompletion(ctx, dev.ID); err == nil {
				percent := int(completion.Completion)
				devices[i].Completion = &percent
			}
		}
	}

	folders := make([]SyncFolder, len(syncStatus.Folders))
	for i, f := range syncStatus.Folders {
		path := f.Path
		if !filepath.IsAbs(path) {
			path = computeFolderPath(userStore.Root(), f.ID)
		}
		folders[i] = SyncFolder{
			ID:                 f.ID,
			Path:               path,
			Label:              folderLabel(f.ID),
			State:              f.State,
			Type:               f.Type,
			GlobalSize:         f.GlobalSize,
			LocalSize:          f.LocalSize,
			NeedSize:           f.NeedSize,
			ReceiveOnlyChanges: f.ReceiveOnlyChanges,
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
			Enabled:          true,
			Mode:             string(syncStatus.Mode),
			Running:          true,
			Installed:        installed,
			ServiceInstalled: serviceInstalled,
			DeviceID:         syncStatus.DeviceID,
			GUIURL:           syncStatus.GUIURL,
			State:            SyncState(syncStatus.OverallState()),
			Devices:          devices,
			Folders:          folders,
			Progress:         progress,
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

	return []Event{{
		Type: EventTypeResult,
		Data: SyncRemoveDeviceResponse{
			Success:  true,
			DeviceID: deviceID,
			Name:     removedName,
		},
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

func (d *Daemon) HandleSyncJoinPrimary(cmd SyncJoinPrimaryCommand, emit func(Event)) []Event {
	return d.handleSyncJoinPrimary(&cmd.Data, emit)
}

func (d *Daemon) HandleSyncDiscoveredDevices(cmd Command, emit func(Event)) []Event {
	return d.handleSyncDiscoveredDevices()
}

func newRelayClient(overrideURL string) (*syncpkg.RelayClient, error) {
	if overrideURL != "" {
		return syncpkg.NewRelayClient([]string{overrideURL})
	}
	return syncpkg.NewDefaultRelayClient()
}

func (d *Daemon) handleSyncStartPairing(emit func(Event)) []Event {
	log.Info("Starting pairing mode")
	cfg, err := d.loadConfig()
	if err != nil {
		log.Error("Failed to load config: %v", err)
		return d.errorResponse(err.Error())
	}

	if cfg.Sync.Mode != model.SyncModePrimary {
		log.Error("Pairing rejected: not primary mode")
		return d.errorResponse("pairing can only be started on primary device")
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deviceID, err := client.GetDeviceID(ctx)
	if err != nil {
		log.Error("Failed to get device ID: %v", err)
		return d.errorResponse(fmt.Sprintf("getting device ID: %v", err))
	}

	relayClient, err := newRelayClient(cfg.Sync.RelayURL)
	if err != nil {
		log.Info("No relay server available, falling back to device ID only: %v", err)
		d.startAutoAcceptLoop(cfg)
		return []Event{{
			Type: EventTypeResult,
			Data: SyncStartPairingResponse{DeviceID: deviceID},
		}}
	}

	relayResp, err := relayClient.CreateSession(ctx, deviceID)
	if err != nil {
		log.Info("Failed to create relay session, falling back to device ID only: %v", err)
		d.startAutoAcceptLoop(cfg)
		return []Event{{
			Type: EventTypeResult,
			Data: SyncStartPairingResponse{DeviceID: deviceID},
		}}
	}

	log.Info("Created relay session with code %s for device %s", relayResp.Code, deviceID)
	d.startAutoAcceptLoop(cfg)
	d.startRelayPollLoop(cfg, relayClient, relayResp.Code, client, emit)

	return []Event{{
		Type: EventTypeResult,
		Data: SyncStartPairingResponse{DeviceID: deviceID, Code: relayResp.Code},
	}}
}

func (d *Daemon) handleSyncJoinPrimary(data *SyncJoinPrimaryRequest, emit func(Event)) []Event {
	if data == nil || (data.Code == "" && data.DeviceID == "") {
		return d.errorResponse("pairing code or device ID is required")
	}

	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	var primaryDeviceID string
	var relayCode string
	var relayClient *syncpkg.RelayClient

	if data.DeviceID != "" {
		primaryDeviceID = strings.ToUpper(strings.TrimSpace(data.DeviceID))
		log.Info("handleSyncJoinPrimary called with deviceId=%s", primaryDeviceID)
	} else {
		code := strings.ToUpper(strings.TrimSpace(data.Code))
		if isRelayCode(code) {
			log.Info("handleSyncJoinPrimary called with relay code=%s", code)
			relayCode = code
			relayClient, err = newRelayClient(cfg.Sync.RelayURL)
			if err != nil {
				return d.errorResponse(fmt.Sprintf("relay unavailable: %v", err))
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			session, err := relayClient.GetSession(ctx, code)
			cancel()
			if err != nil {
				return d.errorResponse(fmt.Sprintf("invalid pairing code: %v", err))
			}
			primaryDeviceID = session.DeviceID
			log.Info("Resolved relay code %s to device ID %s", code, primaryDeviceID)
		} else {
			primaryDeviceID = code
			log.Info("handleSyncJoinPrimary called with device ID=%s", primaryDeviceID)
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

	log.Info("Starting secondary pairing flow for device %s", primaryDeviceID)

	go func(relayCode string, relayClient *syncpkg.RelayClient) {
		defer func() {
			pairingLog.Info("Secondary pairing goroutine ended")
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

		flow := syncpkg.NewSecondaryPairingFlow(syncpkg.PairingFlowConfig{
			SyncConfig: cfg.Sync,
			Instance:   d.paths.Instance,
			Client:     client,
			OnProgress: func(msg string) {
				emit(Event{
					Type: EventTypeProgress,
					Data: SyncPairingProgressEvent{Message: msg},
				})
			},
		})

		result, err := flow.Run(ctx, primaryDeviceID)
		if err != nil {
			pairingLog.Error("Secondary pairing flow failed: %v", err)
			emit(Event{Type: EventTypeError, Data: ErrorResponse{Error: err.Error()}})
			return
		}

		pairingLog.Info("Secondary pairing flow succeeded, persisting config")
		d.persistSyncEnabled(cfg, model.SyncModeSecondary)

		d.dismissUnwantedPendingFolders(client)

		pairingLog.Info("Emitting success result for peer %s (%s)", result.PeerName, result.PeerDeviceID[:7]+"...")
		emit(Event{
			Type: EventTypeResult,
			Data: SyncJoinPrimaryResponse{
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
	d.stopSecondaryPairing()
	return []Event{{
		Type: EventTypeResult,
		Data: map[string]bool{"cancelled": true},
	}}
}

func (d *Daemon) stopSecondaryPairing() {
	d.pairingMu.Lock()
	defer d.pairingMu.Unlock()
	if d.pairingCancelFunc != nil {
		log.Info("Stopping secondary pairing flow")
		d.pairingCancelFunc()
		d.pairingCancelFunc = nil
	}
	d.pairingActive = false
}

func (d *Daemon) handleSyncEnable(data *SyncEnableRequest, emit func(Event)) []Event {
	if data == nil || data.Mode == "" {
		return d.errorResponse("mode is required")
	}

	mode := model.SyncMode(data.Mode)
	if mode != model.SyncModePrimary && mode != model.SyncModeSecondary {
		return d.errorResponse("mode must be 'primary' or 'secondary'")
	}

	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if cfg.Sync.Enabled {
		return d.errorResponse("sync is already enabled")
	}

	cfg.Sync.Enabled = true
	cfg.Sync.Mode = mode

	userStore, err := store.NewUserStore(d.fs, d.paths, cfg.Global.UserStore)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	allSystems := d.syncSystems(cfg)

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

		setup := syncpkg.NewSetup(d.fs, d.paths, d.installer, d.stateDir, d.service)
		result, err := setup.Install(context.Background(), cfg.Sync, userStore.Root(), allSystems, func(p packages.InstallProgress) {
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

		path := d.configPath
		if path == "" {
			path, _ = d.paths.ConfigPath()
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
			Version:             d.installer.ResolveVersion("syncthing"),
			ConfigSchemaVersion: syncpkg.ConfigSchemaVersion,
			BinaryPath:          result.SyncthingBinary,
			ConfigDir:           result.ConfigDir,
			DataDir:             result.DataDir,
			SystemdUnitPath:     result.SystemdUnitPath,
		}
		if saveErr := manifest.SaveWithBackup(d.manifestPath); saveErr != nil {
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
	d.stopSecondaryPairing()
	setup := syncpkg.NewSetup(d.fs, d.paths, d.installer, d.stateDir, d.service)
	var removedFiles []string

	if setup.IsEnabled() {
		removedFiles = append(removedFiles, "systemd unit: "+d.paths.DirName()+"-syncthing.service")
	}

	syncthingDir := filepath.Join(d.stateDir, "syncthing")
	if d.dirExists(syncthingDir) {
		removedFiles = append(removedFiles, syncthingDir)
	}

	if err := setup.Reset(); err != nil {
		return d.errorResponse(fmt.Sprintf("resetting sync: %v", err))
	}

	cfg, err := d.loadConfig()
	if err == nil && cfg.Sync.Enabled {
		cfg.Sync.Enabled = false
		_ = d.configStore.Save(cfg, d.configPath)
	}

	manifest, err := d.loadManifest()
	if err == nil && manifest.SyncthingInstall != nil {
		manifest.SyncthingInstall = nil
		_ = manifest.SaveWithBackup(d.manifestPath)
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SyncResetResponse{
			Success:      true,
			RemovedFiles: removedFiles,
		},
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
	keyPath := filepath.Join(d.stateDir, "syncthing", "config", ".apikey")
	data, err := d.fs.ReadFile(keyPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (d *Daemon) syncSystems(cfg *model.KyarabenConfig) []model.SystemID {
	if cfg.Sync.Mode == model.SyncModePrimary {
		var systems []model.SystemID
		for _, sys := range d.reg.AllSystems() {
			systems = append(systems, sys.ID)
		}
		return systems
	}
	var systems []model.SystemID
	for sysID := range cfg.Systems {
		systems = append(systems, sysID)
	}
	return systems
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
	unit := syncpkg.NewSystemdUnit(d.fs, d.paths, d.service)
	state := d.service.State(unit.UnitName())
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

	userStore, err := store.NewUserStore(d.fs, d.paths, cfg.Global.UserStore)
	if err != nil {
		log.Error("Failed to create user store for sync setup: %v", err)
		d.recordSyncSetupFailure()
		return
	}

	allSystems := d.syncSystems(cfg)

	setup := syncpkg.NewSetup(d.fs, d.paths, d.installer, d.stateDir, d.service)
	result, err := setup.Install(context.Background(), cfg.Sync, userStore.Root(), allSystems, nil)
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
		Version:             d.installer.ResolveVersion("syncthing"),
		ConfigSchemaVersion: syncpkg.ConfigSchemaVersion,
		BinaryPath:          result.SyncthingBinary,
		ConfigDir:           result.ConfigDir,
		DataDir:             result.DataDir,
		SystemdUnitPath:     result.SystemdUnitPath,
	}
	if err := manifest.SaveWithBackup(d.manifestPath); err != nil {
		log.Error("Failed to save manifest: %v", err)
	}

	client := syncpkg.NewClient(cfg.Sync)
	loadedKey := d.loadSyncAPIKey()
	if loadedKey != "" {
		client.SetAPIKey(loadedKey)
	}

	if cfg.Sync.Mode == model.SyncModeSecondary {
		d.dismissUnwantedPendingFolders(client)
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

func (d *Daemon) updateSyncConfig(cfg *model.KyarabenConfig, userStorePath string) error {
	allSystems := d.syncSystems(cfg)

	manifest, err := d.loadManifest()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	setup := syncpkg.NewSetup(d.fs, d.paths, d.installer, d.stateDir, d.service)
	result, installErr := setup.Install(context.Background(), cfg.Sync, userStorePath, allSystems, nil)
	if installErr != nil {
		return fmt.Errorf("updating syncthing: %w", installErr)
	}

	expectedVersion := d.installer.ResolveVersion("syncthing")
	manifest.SyncthingInstall = &model.SyncthingInstall{
		Version:             expectedVersion,
		ConfigSchemaVersion: syncpkg.ConfigSchemaVersion,
		BinaryPath:          result.SyncthingBinary,
		ConfigDir:           result.ConfigDir,
		DataDir:             result.DataDir,
		SystemdUnitPath:     result.SystemdUnitPath,
	}
	if saveErr := manifest.SaveWithBackup(d.manifestPath); saveErr != nil {
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

		if cfg.Sync.Mode == model.SyncModeSecondary {
			d.dismissUnwantedPendingFolders(client)
		}
	}

	return nil
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

func (d *Daemon) persistSyncEnabled(cfg *model.KyarabenConfig, mode model.SyncMode) {
	cfg.Sync.Enabled = true
	cfg.Sync.Mode = mode

	path := d.configPath
	if path == "" {
		path, _ = d.paths.ConfigPath()
	}

	if err := d.configStore.Save(cfg, path); err != nil {
		log.Error("Failed to persist sync config: %v", err)
	}
}

func (d *Daemon) startAutoAcceptLoop(cfg *model.KyarabenConfig) {
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
					log.Debug("Auto-accept: error getting pending devices: %v", err)
					continue
				}

				if len(pending) > 0 {
					log.Info("Auto-accept: found %d pending device(s)", len(pending))
				}

				for _, dev := range pending {
					if seenDevices[dev.DeviceID] {
						continue
					}
					seenDevices[dev.DeviceID] = true

					log.Info("Auto-accepting pending device: %s (ID: %s)", dev.Name, dev.DeviceID[:7])
					addCtx, addCancel := context.WithTimeout(ctx, 10*time.Second)
					if err := client.AddDevice(addCtx, dev.DeviceID, dev.Name); err != nil {
						log.Error("Failed to add device %s: %v", dev.Name, err)
						addCancel()
						continue
					}
					if err := client.ShareFoldersWithDevice(addCtx, dev.DeviceID); err != nil {
						log.Error("Failed to share folders with %s: %v", dev.Name, err)
					}
					addCancel()

					log.Info("Auto-accept: pairing complete, stopping loop")
					d.stopAutoAcceptLoop()
					return
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

				pairingLog.Info("Relay poll %d: secondary device ID received: %s", pollCount, resp.DeviceID[:7]+"...")

				addCtx, addCancel := context.WithTimeout(context.Background(), 10*time.Second)
				pairingLog.Info("Adding secondary device to syncthing config")
				if err := syncClient.AddDeviceAutoName(addCtx, resp.DeviceID); err != nil {
					pairingLog.Error("Failed to add device from relay: %v", err)
					addCancel()
					continue
				}
				pairingLog.Info("Sharing folders with secondary device")
				if err := syncClient.ShareFoldersWithDevice(addCtx, resp.DeviceID); err != nil {
					pairingLog.Error("Failed to share folders with relay device: %v", err)
				}
				addCancel()

				pairingLog.Info("Primary pairing via relay completed successfully")
				emit(Event{
					Type: EventTypeProgress,
					Data: SyncPairingProgressEvent{Message: "Device connected via pairing code"},
				})

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
	manifest, err := d.manifestStore.Load(d.manifestPath)
	if err != nil {
		log.Error("Failed to load manifest for uninstall preview: %v", err)
		manifest = model.NewManifest()
	}

	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}
	userStore := cfg.Global.UserStore

	configDir, _ := d.paths.ConfigDir()

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
		path, err := cfg.Target.Resolve()
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
		kyarabenPaths := []string{
			filepath.Join(homeDir, ".local", "bin", "kyaraben-ui"),
			filepath.Join(homeDir, ".local", "bin", "kyaraben"),
			filepath.Join(homeDir, ".local", "share", "applications", "kyaraben.desktop"),
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
	if coresDir, err := d.paths.CoresDir(); err == nil && d.dirExists(coresDir) {
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
			StateDir:           d.stateDir,
			StateDirExists:     d.dirExists(d.stateDir),
			RetroArchCoresDir:  retroArchCoresDir,
			RetroArchCoreFiles: retroArchCoreFiles,
			DesktopFiles:       desktopFiles,
			IconFiles:          iconFiles,
			ConfigFiles:        configFiles,
			KyarabenFiles:      kyarabenFiles,
			SyncthingFiles:     syncthingFiles,
			Preserved: PreservedPaths{
				UserStore: shortenPath(userStore),
				ConfigDir: shortenPath(configDir),
			},
		},
	}}
}

func (d *Daemon) handleUninstall() []Event {
	manifest, err := d.manifestStore.Load(d.manifestPath)
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
		path, err := cfg.Target.Resolve()
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
		if err := syncpkg.StopAndRemoveServiceWithWait(d.service, servicePath, 10*time.Second, syncPorts); err != nil {
			errors = append(errors, fmt.Sprintf("could not remove syncthing service %s: %v", servicePath, err))
		} else {
			removedFiles = append(removedFiles, servicePath)
		}
	}

	if d.dirExists(d.stateDir) {
		if err := forceRemoveAll(d.stateDir); err != nil {
			errors = append(errors, fmt.Sprintf("could not remove %s: %v", d.stateDir, err))
		} else {
			removedFiles = append(removedFiles, d.stateDir)
		}
	}

	homeDir, _ := os.UserHomeDir()
	iconsDir := filepath.Join(homeDir, ".local", "share", "icons", "hicolor")
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
	info, err := d.fs.Stat(path)
	return err == nil && !info.IsDir()
}

func (d *Daemon) dirExists(path string) bool {
	info, err := d.fs.Stat(path)
	return err == nil && info.IsDir()
}

func (d *Daemon) handleInstallKyaraben(data *InstallKyarabenRequest) []Event {
	appImagePath := ""
	sidecarPath := ""
	if data != nil {
		appImagePath = data.AppImagePath
		sidecarPath = data.SidecarPath
	}

	result, err := d.launcherManager.InstallKyaraben(appImagePath, sidecarPath)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	lockDir := filepath.Dir(d.manifestPath)
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
	manifest.KyarabenInstall = &model.KyarabenInstall{
		AppPath:     result.AppPath,
		CLIPath:     result.CLIPath,
		DesktopPath: result.DesktopPath,
	}
	if saveErr := manifest.SaveWithBackup(d.manifestPath); saveErr != nil {
		return d.errorResponse(fmt.Sprintf("failed to save manifest: %v", saveErr))
	}

	return []Event{{
		Type: EventTypeResult,
		Data: InstallKyarabenResponse{Success: true},
	}}
}

func (d *Daemon) handleInstallStatus() []Event {
	manifest, err := d.manifestStore.Load(d.manifestPath)
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

	status := d.launcherManager.GetInstallStatus()
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
	refreshed := d.launcherManager.RefreshIconCaches()
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

			devices = append(devices, StorageDevice{
				ID:         "sdcard",
				Label:      entry.Name(),
				Path:       filepath.Join(mountPath, "Emulation"),
				FreeBytes:  int64(stat.Bavail) * stat.Bsize,
				TotalBytes: int64(stat.Blocks) * stat.Bsize,
			})
			break
		}
		if len(devices) > 1 {
			break
		}
	}

	return []Event{{
		Type: EventTypeResult,
		Data: StorageDevicesResponse{Devices: devices},
	}}
}
