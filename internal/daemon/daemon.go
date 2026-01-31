package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/apply"
	"github.com/fnune/kyaraben/internal/doctor"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/status"
	"github.com/fnune/kyaraben/internal/store"
	"github.com/fnune/kyaraben/internal/sync"
)

type Daemon struct {
	configPath      string
	reg             *registry.Registry
	nixClient       nix.NixClient
	flakeGenerator  *nix.FlakeGenerator
	configWriter    *emulators.ConfigWriter
	launcherManager *launcher.Manager
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

// Handle processes a command and returns all events at once.
// For streaming events during long operations, use HandleWithEmit.
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

func (d *Daemon) handleStatus() []Event {
	cfg, err := d.loadConfig()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = model.NewDefaultConfig()
		} else {
			return []Event{{
				Type: EventError,
				Data: map[string]string{"error": fmt.Sprintf("loading config: %v", err)},
			}}
		}
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

	result, err := status.Get(cfg, configPath, d.reg, userStore, manifestPath)
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
		if errors.Is(err, os.ErrNotExist) {
			cfg = model.NewDefaultConfig()
		} else {
			return []Event{{
				Type: EventError,
				Data: map[string]string{"error": fmt.Sprintf("loading config: %v", err)},
			}}
		}
	}

	userStore, err := store.NewUserStore(cfg.Global.UserStore)
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	result, err := doctor.Run(cfg, d.reg, userStore)
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
			event := Event{
				Type: EventProgress,
				Data: data,
			}
			if emit != nil {
				emit(event)
			}
		},
	}

	result, err := applier.Apply(cfg, userStore, opts)
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

func (d *Daemon) handleGetSystems() []Event {
	systems := d.reg.AllSystems()

	result := make([]map[string]interface{}, 0, len(systems))
	for _, sys := range systems {
		emulators := d.reg.GetEmulatorsForSystem(sys.ID)
		emuList := make([]map[string]string, 0, len(emulators))
		for _, emu := range emulators {
			emuList = append(emuList, map[string]string{
				"id":   string(emu.ID),
				"name": emu.Name,
			})
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
	cfg, err := d.loadConfig()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = model.NewDefaultConfig()
		} else {
			return []Event{{
				Type: EventError,
				Data: map[string]string{"error": fmt.Sprintf("loading config: %v", err)},
			}}
		}
	}

	systems := make(map[string]string)
	for sys, sysConf := range cfg.Systems {
		systems[string(sys)] = string(sysConf.Emulator)
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
	cfg, err := d.loadConfig()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = model.NewDefaultConfig()
		} else {
			return []Event{{
				Type: EventError,
				Data: map[string]string{"error": fmt.Sprintf("loading config: %v", err)},
			}}
		}
	}

	if userStore, ok := data["userStore"].(string); ok {
		cfg.Global.UserStore = userStore
	}

	if systems, ok := data["systems"].(map[string]interface{}); ok {
		cfg.Systems = make(map[model.SystemID]model.SystemConf)
		for sysStr, emuVal := range systems {
			if emuStr, ok := emuVal.(string); ok {
				cfg.Systems[model.SystemID(sysStr)] = model.SystemConf{
					Emulator: model.EmulatorID(emuStr),
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
		if errors.Is(err, os.ErrNotExist) {
			cfg = model.NewDefaultConfig()
		} else {
			return []Event{{
				Type: EventError,
				Data: map[string]string{"error": fmt.Sprintf("loading config: %v", err)},
			}}
		}
	}

	if !cfg.Sync.Enabled {
		return []Event{{
			Type: EventResult,
			Data: map[string]interface{}{
				"enabled": false,
			},
		}}
	}

	client := sync.NewClient(cfg.Sync)
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
