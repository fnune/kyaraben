package status

import (
	"context"
	"time"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
)

type SystemInfo struct {
	ID   model.SystemID
	Name string
}

type EmulatorInfo struct {
	ID      model.EmulatorID
	Name    string
	Version string
}

type Result struct {
	ConfigPath           string
	UserStorePath        string
	UserStoreInitialized bool
	EnabledSystems       []SystemInfo
	InstalledEmulators   []EmulatorInfo
	LastApplied          time.Time
	MissingRequiredCount int
}

func Get(ctx context.Context, cfg *model.KyarabenConfig, configPath string, reg *registry.Registry, userStore *store.UserStore, manifestPath string) (*Result, error) {
	manifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	result := &Result{
		ConfigPath:           configPath,
		UserStorePath:        userStore.Root(),
		UserStoreInitialized: userStore.IsInitialized(),
		LastApplied:          manifest.LastApplied,
	}

	for _, sysID := range cfg.EnabledSystems() {
		info := SystemInfo{ID: sysID, Name: string(sysID)}
		if sys, err := reg.GetSystem(sysID); err == nil {
			info.Name = sys.Name
		}
		result.EnabledSystems = append(result.EnabledSystems, info)
	}

	for _, emu := range manifest.InstalledEmulators {
		info := EmulatorInfo{
			ID:      emu.ID,
			Name:    string(emu.ID),
			Version: emu.Version,
		}
		if e, err := reg.GetEmulator(emu.ID); err == nil {
			info.Name = e.Name
		}
		result.InstalledEmulators = append(result.InstalledEmulators, info)
	}

	checker := store.NewProvisionChecker(userStore)
	for sys, sysConf := range cfg.Systems {
		emu, err := reg.GetEmulator(sysConf.Emulator)
		if err != nil {
			continue
		}
		results := checker.Check(emu, sys)
		if store.HasMissingRequired(results) {
			result.MissingRequiredCount++
		}
	}

	return result, nil
}
