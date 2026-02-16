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
	"github.com/fnune/kyaraben/internal/store"
	syncpkg "github.com/fnune/kyaraben/internal/sync"
	"github.com/fnune/kyaraben/internal/version"
	"github.com/fnune/kyaraben/internal/versions"
)

var log = logging.New("daemon")

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

	mu              sync.Mutex
	applyCancelFunc context.CancelFunc

	pairingMu         sync.Mutex
	pairingCancelFunc context.CancelFunc
	pairingActive     bool
}

func New(fs vfs.FS, p *paths.Paths, configPath, stateDir, manifestPath string, reg *registry.Registry, installer packages.Installer, configWriter *emulators.ConfigWriter, launcherManager *launcher.Manager) *Daemon {
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
	}
}

func NewDefault(p *paths.Paths, configPath, stateDir, manifestPath string, reg *registry.Registry, installer packages.Installer, configWriter *emulators.ConfigWriter, launcherManager *launcher.Manager) *Daemon {
	return New(vfs.OSFS, p, configPath, stateDir, manifestPath, reg, installer, configWriter, launcherManager)
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
		for j, cfg := range emu.ManagedConfigs {
			keys := make([]ManagedKeyInfo, len(cfg.Keys))
			for k, key := range cfg.Keys {
				keys[k] = ManagedKeyInfo{
					Key:   key.Key,
					Value: shortenPath(key.Value),
				}
			}
			managedConfigs[j] = ManagedConfigInfo{
				Path: shortenPath(cfg.Path),
				Keys: keys,
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

	preflight, err := applier.Preflight(context.Background(), cfg, userStore)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("preflight check: %v", err))
	}

	manifest, err := d.manifestStore.Load(d.manifestPath)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("loading manifest: %v", err))
	}

	type fileDiffState struct {
		diff         ConfigFileDiff
		seenChanges  map[string]bool
		seenUserKeys map[string]bool
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
				seenChanges:  make(map[string]bool),
				seenUserKeys: make(map[string]bool),
			}
			diffsByPath[diff.Path] = state
			pathOrder = append(pathOrder, diff.Path)
		}

		state.diff.HasChanges = state.diff.HasChanges || diff.HasChanges()
		state.diff.UserModified = state.diff.UserModified || diff.UserModified

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

	if !cfg.Sync.Enabled {
		return []Event{{
			Type: EventTypeResult,
			Data: SyncStatusResponse{Enabled: false},
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
		go d.ensureSyncthingRunning(cfg)

		return []Event{{
			Type: EventTypeResult,
			Data: SyncStatusResponse{
				Enabled: true,
				Mode:    string(cfg.Sync.Mode),
				Running: false,
				GUIURL:  fmt.Sprintf("http://127.0.0.1:%d", cfg.Sync.Syncthing.GUIPort),
			},
		}}
	}

	syncStatus, err := client.GetStatus(ctx)
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
	}

	folders := make([]SyncFolder, len(syncStatus.Folders))
	for i, f := range syncStatus.Folders {
		folders[i] = SyncFolder{
			ID:         f.ID,
			Path:       f.Path,
			Label:      folderLabel(f.ID),
			State:      f.State,
			GlobalSize: f.GlobalSize,
			LocalSize:  f.LocalSize,
			NeedSize:   f.NeedSize,
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
			Enabled:  true,
			Mode:     string(syncStatus.Mode),
			Running:  true,
			DeviceID: syncStatus.DeviceID,
			GUIURL:   syncStatus.GUIURL,
			State:    SyncState(syncStatus.OverallState()),
			Devices:  devices,
			Folders:  folders,
			Paused:   syncStatus.Paused,
			Progress: progress,
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

	found := -1
	for i, dev := range cfg.Sync.Devices {
		if strings.ToUpper(dev.ID) == deviceID || strings.EqualFold(dev.Name, deviceID) {
			found = i
			break
		}
	}

	if found == -1 {
		return d.errorResponse("device not found")
	}

	removed := cfg.Sync.Devices[found]
	cfg.Sync.Devices = append(cfg.Sync.Devices[:found], cfg.Sync.Devices[found+1:]...)

	path := d.configPath
	if path == "" {
		path, _ = d.paths.ConfigPath()
	}

	if err := d.configStore.Save(cfg, path); err != nil {
		return d.errorResponse(err.Error())
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SyncRemoveDeviceResponse{
			Success:  true,
			DeviceID: removed.ID,
			Name:     removed.Name,
		},
	}}
}

func (d *Daemon) HandleSyncStartPairing(cmd Command, emit func(Event)) []Event {
	return d.handleSyncStartPairing(emit)
}

func (d *Daemon) HandleSyncJoinPrimary(cmd SyncJoinPrimaryCommand, emit func(Event)) []Event {
	return d.handleSyncJoinPrimary(&cmd.Data, emit)
}

func (d *Daemon) handleSyncStartPairing(emit func(Event)) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
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

	go func() {
		defer func() {
			d.pairingMu.Lock()
			d.pairingActive = false
			d.pairingMu.Unlock()
		}()

		client := syncpkg.NewClient(cfg.Sync)
		loadedKey := d.loadSyncAPIKey()
		if loadedKey != "" {
			client.SetAPIKey(loadedKey)
		}

		flow := syncpkg.NewPrimaryPairingFlow(syncpkg.PairingFlowConfig{
			SyncConfig: cfg.Sync,
			Instance:   d.paths.Instance,
			Advertiser: syncpkg.NewMDNSAdvertiser(),
			Client:     client,
			OnProgress: func(msg string) {
				emit(Event{
					Type: EventTypeProgress,
					Data: SyncPairingProgressEvent{Message: msg},
				})
			},
		})

		result, code, err := flow.Run(ctx)
		if err != nil {
			emit(Event{Type: EventTypeError, Data: ErrorResponse{Error: err.Error()}})
			return
		}

		_ = code

		d.persistPairedDevice(cfg, result.PeerDeviceID, result.PeerName, model.SyncModePrimary)

		emit(Event{
			Type: EventTypeResult,
			Data: SyncPairingCompleteResponse{
				Success:      true,
				PeerDeviceID: result.PeerDeviceID,
				PeerName:     result.PeerName,
			},
		})
	}()

	return nil
}

func (d *Daemon) handleSyncJoinPrimary(data *SyncJoinPrimaryRequest, emit func(Event)) []Event {
	if data == nil || data.Code == "" {
		return d.errorResponse("code is required")
	}

	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
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

	go func() {
		defer func() {
			d.pairingMu.Lock()
			d.pairingActive = false
			d.pairingMu.Unlock()
		}()

		client := syncpkg.NewClient(cfg.Sync)
		loadedKey := d.loadSyncAPIKey()
		if loadedKey != "" {
			client.SetAPIKey(loadedKey)
		}

		flow := syncpkg.NewSecondaryPairingFlow(syncpkg.PairingFlowConfig{
			SyncConfig: cfg.Sync,
			Instance:   d.paths.Instance,
			Browser:    syncpkg.NewMDNSBrowser(),
			Client:     client,
			OnProgress: func(msg string) {
				emit(Event{
					Type: EventTypeProgress,
					Data: SyncPairingProgressEvent{Message: msg},
				})
			},
		})

		result, err := flow.Run(ctx, data.Code)
		if err != nil {
			emit(Event{Type: EventTypeError, Data: ErrorResponse{Error: err.Error()}})
			return
		}

		d.persistPairedDevice(cfg, result.PeerDeviceID, result.PeerName, model.SyncModeSecondary)

		emit(Event{
			Type: EventTypeResult,
			Data: SyncJoinPrimaryResponse{
				Success:      true,
				PeerDeviceID: result.PeerDeviceID,
				PeerName:     result.PeerName,
			},
		})
	}()

	return nil
}

func (d *Daemon) handleSyncCancelPairing() []Event {
	d.pairingMu.Lock()
	defer d.pairingMu.Unlock()
	if d.pairingCancelFunc != nil {
		d.pairingCancelFunc()
		d.pairingCancelFunc = nil
	}
	d.pairingActive = false
	return []Event{{
		Type: EventTypeResult,
		Data: map[string]bool{"cancelled": true},
	}}
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

	allSystems := make([]model.SystemID, 0, len(cfg.Systems))
	for sysID := range cfg.Systems {
		allSystems = append(allSystems, sysID)
	}

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

		setup := syncpkg.NewSetup(d.fs, d.paths, d.installer, d.stateDir)
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
			Version:         d.installer.ResolveVersion("syncthing"),
			BinaryPath:      result.SyncthingBinary,
			ConfigDir:       result.ConfigDir,
			DataDir:         result.DataDir,
			SystemdUnitPath: result.SystemdUnitPath,
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

func (d *Daemon) ensureSyncthingRunning(cfg *model.KyarabenConfig) {
	userStore, err := store.NewUserStore(d.fs, d.paths, cfg.Global.UserStore)
	if err != nil {
		log.Error("Failed to create user store for sync setup: %v", err)
		return
	}

	allSystems := make([]model.SystemID, 0, len(cfg.Systems))
	for sysID := range cfg.Systems {
		allSystems = append(allSystems, sysID)
	}

	setup := syncpkg.NewSetup(d.fs, d.paths, d.installer, d.stateDir)
	result, err := setup.Install(context.Background(), cfg.Sync, userStore.Root(), allSystems, nil)
	if err != nil {
		log.Error("Failed to ensure syncthing running: %v", err)
		return
	}

	manifest, err := d.loadManifest()
	if err != nil {
		log.Error("Failed to load manifest: %v", err)
		return
	}

	manifest.SyncthingInstall = &model.SyncthingInstall{
		Version:         d.installer.ResolveVersion("syncthing"),
		BinaryPath:      result.SyncthingBinary,
		ConfigDir:       result.ConfigDir,
		DataDir:         result.DataDir,
		SystemdUnitPath: result.SystemdUnitPath,
	}
	if err := manifest.SaveWithBackup(d.manifestPath); err != nil {
		log.Error("Failed to save manifest: %v", err)
	}
}

func (d *Daemon) updateSyncConfig(cfg *model.KyarabenConfig, userStorePath string) error {
	allSystems := make([]model.SystemID, 0, len(cfg.Systems))
	for sysID := range cfg.Systems {
		allSystems = append(allSystems, sysID)
	}

	manifest, err := d.loadManifest()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	setup := syncpkg.NewSetup(d.fs, d.paths, d.installer, d.stateDir)
	result, installErr := setup.Install(context.Background(), cfg.Sync, userStorePath, allSystems, nil)
	if installErr != nil {
		return fmt.Errorf("updating syncthing: %w", installErr)
	}

	expectedVersion := d.installer.ResolveVersion("syncthing")
	manifest.SyncthingInstall = &model.SyncthingInstall{
		Version:         expectedVersion,
		BinaryPath:      result.SyncthingBinary,
		ConfigDir:       result.ConfigDir,
		DataDir:         result.DataDir,
		SystemdUnitPath: result.SystemdUnitPath,
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
	}

	return nil
}

func (d *Daemon) persistPairedDevice(cfg *model.KyarabenConfig, peerDeviceID, peerName string, mode model.SyncMode) {
	for _, dev := range cfg.Sync.Devices {
		if dev.ID == peerDeviceID {
			return
		}
	}

	cfg.Sync.Devices = append(cfg.Sync.Devices, model.SyncDevice{
		ID:   peerDeviceID,
		Name: peerName,
	})
	cfg.Sync.Enabled = true
	cfg.Sync.Mode = mode

	path := d.configPath
	if path == "" {
		path, _ = d.paths.ConfigPath()
	}

	if err := d.configStore.Save(cfg, path); err != nil {
		log.Error("Failed to persist paired device: %v", err)
	}
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
		for _, p := range []string{si.BinaryPath, si.SystemdUnitPath} {
			if p != "" && d.fileExists(p) {
				syncthingFiles = append(syncthingFiles, p)
			}
		}
		for _, dir := range []string{si.ConfigDir, si.DataDir} {
			if dir != "" && d.dirExists(dir) {
				syncthingFiles = append(syncthingFiles, dir)
			}
		}
	}

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

	syncSetup := syncpkg.NewSetup(d.fs, d.paths, d.installer, d.stateDir)
	if syncSetup.IsEnabled() {
		if err := syncSetup.Disable(); err != nil {
			errors = append(errors, fmt.Sprintf("could not disable syncthing service: %v", err))
		}
	}
	if manifest.SyncthingInstall != nil {
		si := manifest.SyncthingInstall
		if si.SystemdUnitPath != "" && d.fileExists(si.SystemdUnitPath) {
			if err := os.Remove(si.SystemdUnitPath); err != nil {
				errors = append(errors, fmt.Sprintf("could not remove %s: %v", si.SystemdUnitPath, err))
			} else {
				removedFiles = append(removedFiles, si.SystemdUnitPath)
			}
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
