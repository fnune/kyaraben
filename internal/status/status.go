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

type ManagedConfigInfo struct {
	Path string
	Keys []ManagedKeyInfo
}

type ManagedKeyInfo struct {
	Key   string
	Value string
}

type SymlinkInfo struct {
	Source     string
	Target     string
	EmulatorID model.EmulatorID
}

type EmulatorInfo struct {
	ID             model.EmulatorID
	Name           string
	Version        string              // Installed version
	PinnedVersion  string              // User-pinned version (empty if auto)
	DefaultVersion string              // Latest default version from versions.toml
	ManagedConfigs []ManagedConfigInfo // Config files managed by kyaraben with their keys
}

type FrontendInfo struct {
	ID      model.FrontendID
	Name    string
	Version string
}

type Result struct {
	ConfigPath           string
	UserStorePath        string
	UserStoreInitialized bool
	EnabledSystems       []SystemInfo
	InstalledEmulators   []EmulatorInfo
	InstalledFrontends   []FrontendInfo
	Symlinks             []SymlinkInfo
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
			path, err := cfg.Target.Resolve()
			if err != nil {
				continue
			}
			configInfo := ManagedConfigInfo{Path: path}
			for _, key := range cfg.ManagedKeys {
				configInfo.Keys = append(configInfo.Keys, ManagedKeyInfo{
					Key:   key.Path[len(key.Path)-1],
					Value: key.Value,
				})
			}
			info.ManagedConfigs = append(info.ManagedConfigs, configInfo)
		}

		result.InstalledEmulators = append(result.InstalledEmulators, info)
	}

	for _, fe := range manifest.InstalledFrontends {
		info := FrontendInfo{
			ID:      fe.ID,
			Name:    string(fe.ID),
			Version: fe.Version,
		}
		if f, err := reg.GetFrontend(fe.ID); err == nil {
			info.Name = f.Name
		}
		result.InstalledFrontends = append(result.InstalledFrontends, info)
	}

	for _, s := range manifest.Symlinks {
		result.Symlinks = append(result.Symlinks, SymlinkInfo{
			Source:     s.Source,
			Target:     s.Target,
			EmulatorID: s.EmulatorID,
		})
	}

	checker := store.NewProvisionChecker(userStore)
	for sys, emulatorIDs := range cfg.Systems {
		for _, emuID := range emulatorIDs {
			emu, err := reg.GetEmulator(emuID)
			if err != nil {
				continue
			}
			results := checker.Check(emu, sys)
			if store.HasUnsatisfiedRequired(results) {
				result.MissingRequiredCount++
			}
		}
	}

	result.HealthWarning = detectOrphanedArtifacts(manifest)

	return result, nil
}

func detectOrphanedArtifacts(manifest *model.Manifest) string {
	if len(manifest.InstalledEmulators) > 0 {
		return ""
	}

	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return ""
	}

	binDir := filepath.Join(stateDir, "bin")
	if entries, err := os.ReadDir(binDir); err == nil && len(entries) > 0 {
		return "orphaned_artifacts"
	}

	currentLink := filepath.Join(stateDir, "current")
	if _, err := os.Lstat(currentLink); err == nil {
		return "orphaned_artifacts"
	}

	return ""
}
