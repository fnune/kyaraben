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

	mu               sync.Mutex
	applyCancelFunc  context.CancelFunc
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

// HandleWithEmit processes a command. If emit is provided, progress events
// are sent immediately via emit rather than being batched.
func (d *Daemon) HandleWithEmit(cmd Command, emit func(Event)) []Event {
	switch cmd.Type {
	case CmdStatus:
		return d.handleStatus()
	case CmdDoctor:
		return d.handleDoctor()
	case CmdApply:
		return d.handleApply(cmd.Data, emit)
	case CmdCancelApply:
		return d.handleCancelApply()
	case CmdGetSystems:
		return d.handleGetSystems()
	case CmdGetConfig:
		return d.handleGetConfig()
	case CmdSetConfig:
		return d.handleSetConfig(cmd.Data)
	case CmdSyncStatus:
		return d.handleSyncStatus()
	case CmdSyncAddDevice:
		return d.handleSyncAddDevice(cmd.Data)
	case CmdSyncRemoveDevice:
		return d.handleSyncRemoveDevice(cmd.Data)
	case CmdUninstallPreview:
		return d.handleUninstallPreview()
	default:
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": fmt.Sprintf("unknown command: %s", cmd.Type)},
		}}
	}
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
	return model.LoadConfig(path)
}

func (d *Daemon) loadConfigOrDefault() (*model.KyarabenConfig, error) {
	cfg, err := d.loadConfig()
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
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	configPath := d.configPath
	if configPath == "" {
		configPath, _ = model.DefaultConfigPath()
	}

	userStore, err := store.NewUserStore(cfg.Global.UserStore)
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	result, err := status.Get(context.Background(), cfg, configPath, d.reg, userStore, manifestPath)
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	systems := make([]string, len(result.EnabledSystems))
	for i, sys := range result.EnabledSystems {
		systems[i] = string(sys.ID)
	}

	installedEmulators := make([]map[string]string, len(result.InstalledEmulators))
	for i, emu := range result.InstalledEmulators {
		installedEmulators[i] = map[string]string{
			"id":      string(emu.ID),
			"version": emu.Version,
		}
	}

	return []Event{{
		Type: EventResult,
		Data: map[string]interface{}{
			"userStore":          result.UserStorePath,
			"enabledSystems":     systems,
			"installedEmulators": installedEmulators,
			"lastApplied":        result.LastApplied.Format(time.RFC3339),
		},
	}}
}

func (d *Daemon) handleDoctor() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	userStore, err := store.NewUserStore(cfg.Global.UserStore)
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	result, err := doctor.Run(context.Background(), cfg, d.reg, userStore)
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	systems := make(map[string][]map[string]interface{})
	for _, sys := range result.Systems {
		provisions := make([]map[string]interface{}, len(sys.Provisions))
		for i, prov := range sys.Provisions {
			provisions[i] = map[string]interface{}{
				"filename":    prov.Filename,
				"description": prov.Description,
				"required":    prov.Required,
				"status":      string(prov.Status),
				"foundPath":   prov.FoundPath,
			}
		}
		systems[string(sys.SystemID)] = provisions
	}

	return []Event{{
		Type: EventResult,
		Data: systems,
	}}
}

func (d *Daemon) handleApply(_ map[string]interface{}, emit func(Event)) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	userStore, err := store.NewUserStore(cfg.Global.UserStore)
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
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
			data := map[string]interface{}{
				"step":    p.Step,
				"message": p.Message,
			}
			if p.Output != "" {
				data["output"] = p.Output
			}
			if p.Speed != "" {
				data["speed"] = p.Speed
			}
			event := Event{
				Type: EventProgress,
				Data: data,
			}
			if emit != nil {
				emit(event)
			}
		},
	}

	result, err := applier.Apply(ctx, cfg, userStore, opts)
	if ctx.Err() != nil {
		return []Event{{
			Type: EventCancelled,
			Data: map[string]string{"message": "Installation cancelled"},
		}}
	}
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	return []Event{{
		Type: EventResult,
		Data: map[string]interface{}{
			"success":   true,
			"storePath": result.StorePath,
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
			Type: EventCancelled,
			Data: map[string]string{"message": "Installation cancelled"},
		}}
	}

	return []Event{{
		Type: EventResult,
		Data: map[string]interface{}{"cancelled": false, "message": "No apply in progress"},
	}}
}

func (d *Daemon) handleGetSystems() []Event {
	systems := d.reg.AllSystems()
	vers, _ := versions.Get()

	result := make([]map[string]interface{}, 0, len(systems))
	for _, sys := range systems {
		emus := d.reg.GetEmulatorsForSystem(sys.ID)
		emuList := make([]map[string]interface{}, 0, len(emus))
		for _, emu := range emus {
			emuData := map[string]interface{}{
				"id":   string(emu.ID),
				"name": emu.Name,
			}

			// Add version info if available
			if vers != nil {
				// Use the binary name as the key in versions.toml
				emuName := string(emu.ID)
				if spec, ok := vers.GetEmulator(emuName); ok {
					emuData["defaultVersion"] = spec.Default
					availableVersions := spec.AvailableVersions()
					sort.Strings(availableVersions)
					emuData["availableVersions"] = availableVersions
				}
			}

			emuList = append(emuList, emuData)
		}

		result = append(result, map[string]interface{}{
			"id":          string(sys.ID),
			"name":        sys.Name,
			"description": sys.Description,
			"emulators":   emuList,
		})
	}

	return []Event{{
		Type: EventResult,
		Data: result,
	}}
}

func (d *Daemon) handleGetConfig() []Event {
	cfg, err := d.loadConfigOrDefault()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": fmt.Sprintf("loading config: %v", err)},
		}}
	}

	systems := make(map[string]map[string]interface{})
	for sys, sysConf := range cfg.Systems {
		entry := map[string]interface{}{
			"emulator": string(sysConf.EmulatorID()),
		}
		if version := sysConf.EmulatorVersion(); version != "" {
			entry["pinnedVersion"] = version
		}
		systems[string(sys)] = entry
	}

	return []Event{{
		Type: EventResult,
		Data: map[string]interface{}{
			"userStore": cfg.Global.UserStore,
			"systems":   systems,
		},
	}}
}

func (d *Daemon) handleSetConfig(data map[string]interface{}) []Event {
	cfg, err := d.loadConfigOrDefault()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": fmt.Sprintf("loading config: %v", err)},
		}}
	}

	if userStore, ok := data["userStore"].(string); ok {
		cfg.Global.UserStore = userStore
	}

	if systems, ok := data["systems"].(map[string]interface{}); ok {
		cfg.Systems = make(map[model.SystemID]model.SystemConf)
		for sysStr, emuVal := range systems {
			if emuStr, ok := emuVal.(string); ok {
				cfg.Systems[model.SystemID(sysStr)] = model.SystemConf{
					Emulator: emuStr,
				}
			}
		}
	}

	path := d.configPath
	if path == "" {
		path, _ = model.DefaultConfigPath()
	}

	if err := model.SaveConfig(cfg, path); err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	return []Event{{
		Type: EventResult,
		Data: map[string]interface{}{
			"success": true,
		},
	}}
}

func (d *Daemon) handleSyncStatus() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	if !cfg.Sync.Enabled {
		return []Event{{
			Type: EventResult,
			Data: map[string]interface{}{
				"enabled": false,
			},
		}}
	}

	client := syncpkg.NewClient(cfg.Sync)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !client.IsRunning(ctx) {
		return []Event{{
			Type: EventResult,
			Data: map[string]interface{}{
				"enabled": true,
				"mode":    string(cfg.Sync.Mode),
				"running": false,
				"guiURL":  fmt.Sprintf("http://127.0.0.1:%d", cfg.Sync.Syncthing.GUIPort),
			},
		}}
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	devices := make([]map[string]interface{}, len(status.Devices))
	for i, dev := range status.Devices {
		devices[i] = map[string]interface{}{
			"id":        dev.ID,
			"name":      dev.Name,
			"connected": dev.Connected,
		}
	}

	return []Event{{
		Type: EventResult,
		Data: map[string]interface{}{
			"enabled":  true,
			"mode":     string(status.Mode),
			"running":  true,
			"deviceId": status.DeviceID,
			"guiURL":   status.GUIURL,
			"state":    string(status.OverallState()),
			"devices":  devices,
		},
	}}
}

func (d *Daemon) handleSyncAddDevice(data map[string]interface{}) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	if !cfg.Sync.Enabled {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": "sync is not enabled"},
		}}
	}

	deviceID, ok := data["deviceId"].(string)
	if !ok || deviceID == "" {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": "deviceId is required"},
		}}
	}

	name, _ := data["name"].(string)
	if name == "" {
		name = fmt.Sprintf("device-%d", len(cfg.Sync.Devices)+1)
	}

	deviceID = strings.ToUpper(strings.TrimSpace(deviceID))

	for _, existing := range cfg.Sync.Devices {
		if existing.ID == deviceID {
			return []Event{{
				Type: EventError,
				Data: map[string]string{"error": "device already added"},
			}}
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
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	return []Event{{
		Type: EventResult,
		Data: map[string]interface{}{
			"success":  true,
			"deviceId": deviceID,
			"name":     name,
		},
	}}
}

func (d *Daemon) handleSyncRemoveDevice(data map[string]interface{}) []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	deviceID, ok := data["deviceId"].(string)
	if !ok || deviceID == "" {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": "deviceId is required"},
		}}
	}

	deviceID = strings.ToUpper(strings.TrimSpace(deviceID))

	found := -1
	for i, dev := range cfg.Sync.Devices {
		if strings.ToUpper(dev.ID) == deviceID || strings.EqualFold(dev.Name, deviceID) {
			found = i
			break
		}
	}

	if found == -1 {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": "device not found"},
		}}
	}

	removed := cfg.Sync.Devices[found]
	cfg.Sync.Devices = append(cfg.Sync.Devices[:found], cfg.Sync.Devices[found+1:]...)

	path := d.configPath
	if path == "" {
		path, _ = model.DefaultConfigPath()
	}

	if err := model.SaveConfig(cfg, path); err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	return []Event{{
		Type: EventResult,
		Data: map[string]interface{}{
			"success":  true,
			"deviceId": removed.ID,
			"name":     removed.Name,
		},
	}}
}

func (d *Daemon) handleUninstallPreview() []Event {
	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	manifest, _ := model.LoadManifest(manifestPath)

	cfg, _ := d.loadConfigOrDefault()
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
		Type: EventResult,
		Data: map[string]interface{}{
			"stateDir":     stateDir,
			"stateDirExists": dirExists(stateDir),
			"desktopFiles": desktopFiles,
			"iconFiles":    iconFiles,
			"configFiles":  configFiles,
			"preserved": map[string]string{
				"userStore": userStore,
				"configDir": configDir,
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
