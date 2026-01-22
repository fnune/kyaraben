package daemon

import (
	"fmt"
	"time"

	"github.com/fnune/kyaraben/internal/apply"
	"github.com/fnune/kyaraben/internal/doctor"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
	"github.com/fnune/kyaraben/internal/status"
	"github.com/fnune/kyaraben/internal/store"
)

type Daemon struct {
	configPath     string
	registry       *emulators.Registry
	nixClient      nix.NixClient
	flakeGenerator *nix.FlakeGenerator
	configWriter   *emulators.ConfigWriter
}

func New(configPath string, registry *emulators.Registry, nixClient nix.NixClient, flakeGenerator *nix.FlakeGenerator, configWriter *emulators.ConfigWriter) *Daemon {
	return &Daemon{
		configPath:     configPath,
		registry:       registry,
		nixClient:      nixClient,
		flakeGenerator: flakeGenerator,
		configWriter:   configWriter,
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
		// Use default config if none exists
		cfg, err = model.NewDefaultConfig()
		if err != nil {
			return []Event{{
				Type: EventError,
				Data: map[string]string{"error": err.Error()},
			}}
		}
	}

	configPath := d.configPath
	if configPath == "" {
		configPath, _ = model.DefaultConfigPath()
	}

	userStorePath, err := cfg.ExpandUserStore()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	userStore := store.NewUserStore(userStorePath)

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	result, err := status.Get(cfg, configPath, d.registry, userStore, manifestPath)
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
		// Use default config if none exists
		cfg, err = model.NewDefaultConfig()
		if err != nil {
			return []Event{{
				Type: EventError,
				Data: map[string]string{"error": err.Error()},
			}}
		}
	}

	userStorePath, err := cfg.ExpandUserStore()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	userStore := store.NewUserStore(userStorePath)

	result, err := doctor.Run(cfg, d.registry, userStore)
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

	userStorePath, err := cfg.ExpandUserStore()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	userStore := store.NewUserStore(userStorePath)

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	applier := &apply.Applier{
		NixClient:      d.nixClient,
		FlakeGenerator: d.flakeGenerator,
		ConfigWriter:   d.configWriter,
		Registry:       d.registry,
		ManifestPath:   manifestPath,
	}

	opts := apply.Options{
		OnProgress: func(p apply.Progress) {
			event := Event{
				Type: EventProgress,
				Data: map[string]interface{}{
					"step":    p.Step,
					"message": p.Message,
				},
			}
			// Stream immediately if emit is provided, otherwise events are lost
			// (we don't batch progress events anymore)
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
	systems := d.registry.AllSystems()

	result := make([]map[string]interface{}, 0, len(systems))
	for _, sys := range systems {
		emulators := d.registry.GetEmulatorsForSystem(sys.ID)
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
		cfg, err = model.NewDefaultConfig()
		if err != nil {
			return []Event{{
				Type: EventError,
				Data: map[string]string{"error": err.Error()},
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
		// Create new config if it doesn't exist
		cfg, err = model.NewDefaultConfig()
		if err != nil {
			return []Event{{
				Type: EventError,
				Data: map[string]string{"error": err.Error()},
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
