package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fnune/kyaraben/internal/apply"
	"github.com/fnune/kyaraben/internal/doctor"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/hardware"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/status"
	"github.com/fnune/kyaraben/internal/store"
	syncpkg "github.com/fnune/kyaraben/internal/sync"
	"github.com/fnune/kyaraben/internal/versions"
)

type Daemon struct {
	configPath      string
	reg             *registry.Registry
	nixClient       nix.NixClient
	flakeGenerator  *nix.FlakeGenerator
	configWriter    *emulators.ConfigWriter
	launcherManager *launcher.Manager

	mu              sync.Mutex
	applyCancelFunc context.CancelFunc
}

func New(configPath string, reg *registry.Registry, nixClient nix.NixClient, flakeGenerator *nix.FlakeGenerator, configWriter *emulators.ConfigWriter, launcherManager *launcher.Manager) *Daemon {
	return &Daemon{
		configPath:      configPath,
		reg:             reg,
		nixClient:       nixClient,
		flakeGenerator:  flakeGenerator,
		configWriter:    configWriter,
		launcherManager: launcherManager,
	}
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
	case CommandTypeGetConfig:
		return d.handleGetConfig()
	case CommandTypeSetConfig:
		return d.handleSetConfig(nil)
	case CommandTypeSyncStatus:
		return d.handleSyncStatus()
	case CommandTypeSyncAddDevice:
		return d.handleSyncAddDevice(nil)
	case CommandTypeSyncRemoveDevice:
		return d.handleSyncRemoveDevice(nil)
	case CommandTypeUninstallPreview:
		return d.handleUninstallPreview()
	default:
		return d.errorResponse(fmt.Sprintf("unknown command: %s", cmd.Type))
	}
}

func (d *Daemon) HandleSetConfig(cmd SetConfigCommand, emit func(Event)) []Event {
	return d.handleSetConfig(&cmd.Data)
}

func (d *Daemon) HandleSyncAddDevice(cmd SyncAddDeviceCommand, emit func(Event)) []Event {
	return d.handleSyncAddDevice(&cmd.Data)
}

func (d *Daemon) HandleSyncRemoveDevice(cmd SyncRemoveDeviceCommand, emit func(Event)) []Event {
	return d.handleSyncRemoveDevice(&cmd.Data)
}

func (d *Daemon) errorResponse(msg string) []Event {
	return []Event{{
		Type: EventTypeError,
		Data: ErrorResponse{Error: msg},
	}}
}

func (d *Daemon) loadConfig() (*model.KyarabenConfig, error) {
	path := d.configPath
	if path == "" {
		var err error
		path, err = model.DefaultConfigPath()
		if err != nil {
			return nil, err
		}
	}
	cfg, err := model.LoadConfig(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return model.NewDefaultConfig(), nil
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
		configPath, _ = model.DefaultConfigPath()
	}

	userStore, err := store.NewUserStore(cfg.Global.UserStore)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	result, err := status.Get(context.Background(), cfg, configPath, d.reg, userStore, manifestPath)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	systems := make([]model.SystemID, len(result.EnabledSystems))
	for i, sys := range result.EnabledSystems {
		systems[i] = sys.ID
	}

	installedEmulators := make([]InstalledEmulator, len(result.InstalledEmulators))
	for i, emu := range result.InstalledEmulators {
		installed := InstalledEmulator{
			ID:             emu.ID,
			Version:        emu.Version,
			ManagedConfigs: emu.ManagedConfigs,
		}
		if e, err := d.reg.GetEmulator(emu.ID); err == nil && e.Launcher.Binary != "" {
			installed.ExecLine = fmt.Sprintf("%s/%s", d.launcherManager.BinDir(), e.Launcher.Binary)
		}
		installedEmulators[i] = installed
	}

	return []Event{{
		Type: EventTypeResult,
		Data: StatusResponse{
			UserStore:          result.UserStorePath,
			EnabledSystems:     systems,
			InstalledEmulators: installedEmulators,
			LastApplied:        result.LastApplied.Format(time.RFC3339),
		},
	}}
}

func (d *Daemon) handleDoctor() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	userStore, err := store.NewUserStore(cfg.Global.UserStore)
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
				Filename:    prov.Filename,
				Kind:        string(prov.Kind),
				Description: prov.Description,
				Required:    prov.Required,
				Status:      string(prov.Status),
				FoundPath:   prov.FoundPath,
				ImportViaUI: prov.ImportViaUI,
			}
		}
		response[string(sys.EmulatorID)] = provisions
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

	versionOverrides, err := cfg.BuildVersionOverrides(d.reg.GetEmulator)
	if err != nil {
		return d.errorResponse(err.Error())
	}
	d.flakeGenerator.SetVersionOverrides(versionOverrides)

	userStore, err := store.NewUserStore(cfg.Global.UserStore)
	if err != nil {
		return d.errorResponse(err.Error())
	}

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.mu.Lock()
	d.applyCancelFunc = cancel
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		d.applyCancelFunc = nil
		d.mu.Unlock()
	}()

	applier := &apply.Applier{
		NixClient:       d.nixClient,
		FlakeGenerator:  d.flakeGenerator,
		ConfigWriter:    d.configWriter,
		Registry:        d.reg,
		ManifestPath:    manifestPath,
		LauncherManager: d.launcherManager,
	}

	opts := apply.Options{
		OnProgress: func(p apply.Progress) {
			event := Event{
				Type: EventTypeProgress,
				Data: ProgressEvent{
					Step:    p.Step,
					Message: p.Message,
					Output:  p.Output,
					Speed:   p.Speed,
				},
			}
			if emit != nil {
				emit(event)
			}
		},
	}

	result, err := applier.Apply(ctx, cfg, userStore, opts)
	if ctx.Err() != nil {
		return []Event{{
			Type: EventTypeCancelled,
			Data: CancelledResponse{Message: "Installation cancelled"},
		}}
	}
	if err != nil {
		return d.errorResponse(err.Error())
	}

	return []Event{{
		Type: EventTypeResult,
		Data: ApplyResult{
			Success:   true,
			StorePath: result.StorePath,
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
	currentArch := hardware.DetectTarget().Arch

	result := make(GetSystemsResponse, 0, len(systems))
	for _, sys := range systems {
		emus := d.reg.GetEmulatorsForSystem(sys.ID)
		emuList := make([]EmulatorRef, 0, len(emus))
		for _, emu := range emus {
			ref := EmulatorRef{
				ID:   emu.ID,
				Name: emu.Name,
			}

			if vers != nil {
				if spec, ok := vers.GetEmulator(emu.Package.PackageName()); ok {
					ref.DefaultVersion = spec.Default
					availableVersions := spec.AvailableVersions()
					sort.Strings(availableVersions)
					ref.AvailableVersions = availableVersions

					if entry := spec.GetDefault(); entry != nil {
						if target := entry.DefaultTargetForArch(currentArch); target != "" {
							if build := entry.Target(target); build != nil && build.Size > 0 {
								ref.DownloadBytes = build.Size
							}
						}
					}
				}
			}

			emuList = append(emuList, ref)
		}

		result = append(result, SystemWithEmulators{
			ID:           sys.ID,
			Name:         sys.Name,
			Description:  sys.Description,
			Manufacturer: sys.Manufacturer,
			Label:        sys.Label,
			Emulators:    emuList,
		})
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

	return []Event{{
		Type: EventTypeResult,
		Data: ConfigResponse{
			UserStore: cfg.Global.UserStore,
			Systems:   systems,
			Emulators: emulators,
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
	}

	path := d.configPath
	if path == "" {
		path, _ = model.DefaultConfigPath()
	}

	if err := model.SaveConfig(cfg, path); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !client.IsRunning(ctx) {
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
		},
	}}
}

func (d *Daemon) handleSyncAddDevice(data *SyncAddDeviceRequest) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	if !cfg.Sync.Enabled {
		return d.errorResponse("sync is not enabled")
	}

	if data == nil || data.DeviceID == "" {
		return d.errorResponse("deviceId is required")
	}

	deviceID := strings.ToUpper(strings.TrimSpace(data.DeviceID))
	name := data.Name
	if name == "" {
		name = fmt.Sprintf("device-%d", len(cfg.Sync.Devices)+1)
	}

	for _, existing := range cfg.Sync.Devices {
		if existing.ID == deviceID {
			return d.errorResponse("device already added")
		}
	}

	cfg.Sync.Devices = append(cfg.Sync.Devices, model.SyncDevice{
		ID:   deviceID,
		Name: name,
	})

	path := d.configPath
	if path == "" {
		path, _ = model.DefaultConfigPath()
	}

	if err := model.SaveConfig(cfg, path); err != nil {
		return d.errorResponse(err.Error())
	}

	return []Event{{
		Type: EventTypeResult,
		Data: SyncAddDeviceResponse{
			Success:  true,
			DeviceID: deviceID,
			Name:     name,
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
		path, _ = model.DefaultConfigPath()
	}

	if err := model.SaveConfig(cfg, path); err != nil {
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

func (d *Daemon) handleUninstallPreview() []Event {
	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return d.errorResponse(err.Error())
	}

	manifest, _ := model.LoadManifest(manifestPath)

	cfg, _ := d.loadConfig()
	userStore := cfg.Global.UserStore

	configDir, _ := paths.KyarabenConfigDir()

	var desktopFiles []string
	for _, f := range manifest.DesktopFiles {
		if fileExists(f) {
			desktopFiles = append(desktopFiles, f)
		}
	}

	var iconFiles []string
	for _, f := range manifest.IconFiles {
		if fileExists(f) {
			iconFiles = append(iconFiles, f)
		}
	}

	var configFiles []string
	for _, cfg := range manifest.ManagedConfigs {
		path, err := cfg.Target.Resolve()
		if err == nil && fileExists(path) {
			configFiles = append(configFiles, path)
		}
	}

	return []Event{{
		Type: EventTypeResult,
		Data: UninstallPreviewResponse{
			StateDir:       stateDir,
			StateDirExists: dirExists(stateDir),
			DesktopFiles:   desktopFiles,
			IconFiles:      iconFiles,
			ConfigFiles:    configFiles,
			Preserved: PreservedPaths{
				UserStore: userStore,
				ConfigDir: configDir,
			},
		},
	}}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
