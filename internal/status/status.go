package status

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
	"github.com/fnune/kyaraben/internal/versions"
)

type SystemInfo struct {
	ID   model.SystemID
	Name string
}

type EmulatorInfo struct {
	ID             model.EmulatorID
	Name           string
	Version        string   // Installed version
	PinnedVersion  string   // User-pinned version (empty if auto)
	DefaultVersion string   // Latest default version from versions.toml
	ManagedConfigs []string // Paths to config files managed by kyaraben
}

type Result struct {
	ConfigPath           string
	UserStorePath        string
	UserStoreInitialized bool
	EnabledSystems       []SystemInfo
	InstalledEmulators   []EmulatorInfo
	LastApplied          time.Time
	MissingRequiredCount int
	HealthWarning        string // Non-empty if inconsistent state detected
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

	vers, _ := versions.Get()

	for _, emu := range manifest.InstalledEmulators {
		info := EmulatorInfo{
			ID:            emu.ID,
			Name:          string(emu.ID),
			Version:       emu.Version,
			PinnedVersion: cfg.EmulatorVersion(emu.ID),
		}
		if e, err := reg.GetEmulator(emu.ID); err == nil {
			info.Name = e.Name
		}

		if vers != nil {
			if e, err := reg.GetEmulator(emu.ID); err == nil {
				if spec, ok := vers.GetEmulator(e.Package.PackageName()); ok {
					info.DefaultVersion = spec.Default
				}
			}
		}

		for _, cfg := range manifest.GetManagedConfigsForEmulator(emu.ID) {
			if path, err := cfg.Target.Resolve(); err == nil {
				info.ManagedConfigs = append(info.ManagedConfigs, path)
			}
		}

		result.InstalledEmulators = append(result.InstalledEmulators, info)
	}

	checker := store.NewProvisionChecker(userStore)
	for sys, emulatorIDs := range cfg.Systems {
		for _, emuID := range emulatorIDs {
			emu, err := reg.GetEmulator(emuID)
			if err != nil {
				continue
			}
			results := checker.Check(emu, sys)
			if store.HasMissingRequired(results) {
				result.MissingRequiredCount++
			}
		}
	}

	// Health check: detect if artifacts exist but manifest is empty
	if warning := checkHealthInconsistency(manifest); warning != "" {
		result.HealthWarning = warning
	}

	return result, nil
}

// checkHealthInconsistency detects when installation artifacts exist but the
// manifest doesn't track them. This can happen if the manifest was corrupted
// or deleted.
func checkHealthInconsistency(manifest *model.Manifest) string {
	if len(manifest.InstalledEmulators) > 0 {
		return "" // Manifest has data, no inconsistency
	}

	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return ""
	}

	// Check if wrapper scripts exist
	binDir := filepath.Join(stateDir, "bin")
	if entries, err := os.ReadDir(binDir); err == nil && len(entries) > 0 {
		return "Installation artifacts found but manifest is empty. Your installation state may have been lost. Please report this as a bug and run 'kyaraben apply' to restore."
	}

	// Check if "current" profile symlink exists
	currentLink := filepath.Join(stateDir, "current")
	if _, err := os.Lstat(currentLink); err == nil {
		return "Installation artifacts found but manifest is empty. Your installation state may have been lost. Please report this as a bug and run 'kyaraben apply' to restore."
	}

	return ""
}
