package daemon

import (
	"context"
	"fmt"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
	"github.com/fnune/kyaraben/internal/store"
)

// Daemon handles JSON protocol commands from the UI.
type Daemon struct {
	configPath string
	registry   *emulators.Registry
}

// New creates a new daemon instance.
func New(configPath string) *Daemon {
	return &Daemon{
		configPath: configPath,
		registry:   emulators.NewRegistry(),
	}
}

// Handle processes a command and returns events.
func (d *Daemon) Handle(cmd Command) []Event {
	switch cmd.Type {
	case CmdStatus:
		return d.handleStatus()
	case CmdDoctor:
		return d.handleDoctor()
	case CmdApply:
		return d.handleApply(cmd.Data)
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
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
	}

	manifestPath, _ := model.DefaultManifestPath()
	manifest, _ := model.LoadManifest(manifestPath)

	userStorePath, _ := cfg.ExpandUserStore()

	systems := make([]string, 0, len(cfg.Systems))
	for sys := range cfg.Systems {
		systems = append(systems, string(sys))
	}

	installedEmulators := make([]map[string]string, 0, len(manifest.InstalledEmulators))
	for _, emu := range manifest.InstalledEmulators {
		installedEmulators = append(installedEmulators, map[string]string{
			"id":      string(emu.ID),
			"version": emu.Version,
		})
	}

	return []Event{{
		Type: EventResult,
		Data: map[string]interface{}{
			"userStore":          userStorePath,
			"enabledSystems":     systems,
			"installedEmulators": installedEmulators,
			"lastApplied":        manifest.LastApplied.Format(time.RFC3339),
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

	userStorePath, _ := cfg.ExpandUserStore()
	userStore := store.NewUserStore(userStorePath)
	checker := store.NewProvisionChecker(userStore)

	results := make(map[string][]map[string]interface{})

	for sys, sysConf := range cfg.Systems {
		emu, err := d.registry.GetEmulator(sysConf.Emulator)
		if err != nil {
			continue
		}

		provResults := checker.Check(emu, sys)
		systemResults := make([]map[string]interface{}, 0, len(provResults))

		for _, r := range provResults {
			systemResults = append(systemResults, map[string]interface{}{
				"filename":    r.Provision.Filename,
				"description": r.Provision.Description,
				"required":    r.Provision.Required,
				"status":      string(r.Status),
				"foundPath":   r.FoundPath,
			})
		}

		results[string(sys)] = systemResults
	}

	return []Event{{
		Type: EventResult,
		Data: results,
	}}
}

func (d *Daemon) handleApply(data map[string]interface{}) []Event {
	events := []Event{}

	// Progress: starting
	events = append(events, Event{
		Type: EventProgress,
		Data: map[string]interface{}{
			"step":    "start",
			"message": "Starting apply...",
		},
	})

	cfg, err := d.loadConfig()
	if err != nil {
		return append(events, Event{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		})
	}

	userStorePath, _ := cfg.ExpandUserStore()
	userStore := store.NewUserStore(userStorePath)

	// Initialize directories
	events = append(events, Event{
		Type: EventProgress,
		Data: map[string]interface{}{
			"step":    "directories",
			"message": "Creating directory structure...",
		},
	})

	if err := userStore.Initialize(); err != nil {
		return append(events, Event{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		})
	}

	for sys := range cfg.Systems {
		if err := userStore.InitializeSystem(sys); err != nil {
			return append(events, Event{
				Type: EventError,
				Data: map[string]string{"error": err.Error()},
			})
		}
	}

	// Generate flake
	events = append(events, Event{
		Type: EventProgress,
		Data: map[string]interface{}{
			"step":    "flake",
			"message": "Generating Nix flake...",
		},
	})

	nixClient, err := nix.NewClient()
	if err != nil {
		return append(events, Event{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		})
	}

	emulatorsToInstall := make([]model.EmulatorID, 0, len(cfg.Systems))
	for _, sysConf := range cfg.Systems {
		emulatorsToInstall = append(emulatorsToInstall, sysConf.Emulator)
	}

	flakeGen := nix.NewFlakeGenerator()
	if err := nixClient.EnsureFlakeDir(); err != nil {
		return append(events, Event{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		})
	}

	if err := flakeGen.Generate(nixClient.FlakePath, emulatorsToInstall); err != nil {
		return append(events, Event{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		})
	}

	// Build emulators
	events = append(events, Event{
		Type: EventProgress,
		Data: map[string]interface{}{
			"step":    "build",
			"message": "Building emulators...",
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	flakeRef := flakeGen.DefaultFlakeRef(nixClient.FlakePath)
	storePath, err := nixClient.Build(ctx, flakeRef)
	if err != nil {
		return append(events, Event{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		})
	}

	// Generate configs
	events = append(events, Event{
		Type: EventProgress,
		Data: map[string]interface{}{
			"step":    "configs",
			"message": "Generating emulator configurations...",
		},
	})

	configWriter := emulators.NewConfigWriter()
	for sys, sysConf := range cfg.Systems {
		gen := emulators.GetConfigGenerator(sysConf.Emulator)
		if gen == nil {
			continue
		}

		patches, err := gen.Generate(userStore, []model.SystemID{sys})
		if err != nil {
			return append(events, Event{
				Type: EventError,
				Data: map[string]string{"error": err.Error()},
			})
		}

		for _, patch := range patches {
			if err := configWriter.Apply(patch); err != nil {
				return append(events, Event{
					Type: EventError,
					Data: map[string]string{"error": err.Error()},
				})
			}
		}
	}

	// Update manifest
	manifestPath, _ := model.DefaultManifestPath()
	manifest, _ := model.LoadManifest(manifestPath)
	manifest.LastApplied = time.Now()

	for _, emuID := range emulatorsToInstall {
		manifest.AddEmulator(model.InstalledEmulator{
			ID:        emuID,
			Version:   "latest",
			StorePath: storePath,
			Installed: time.Now(),
		})
	}

	if err := manifest.Save(manifestPath); err != nil {
		return append(events, Event{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		})
	}

	// Done
	events = append(events, Event{
		Type: EventResult,
		Data: map[string]interface{}{
			"success":   true,
			"storePath": storePath,
		},
	})

	return events
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
		return []Event{{
			Type: EventError,
			Data: map[string]string{"error": err.Error()},
		}}
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
