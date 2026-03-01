package status

import (
	"context"
	"path/filepath"
	"time"

	"github.com/twpayne/go-vfs/v5"

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

type ManagedRegionInfo struct {
	Type      string // "file" or "section"
	Section   string
	KeyPrefix string
}

type ManagedConfigInfo struct {
	Path           string
	ManagedRegions []ManagedRegionInfo
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

type Getter struct {
	fs            vfs.FS
	paths         *paths.Paths
	manifestStore *model.ManifestStore
	resolver      model.BaseDirResolver
}

func NewGetter(fs vfs.FS, p *paths.Paths, resolver model.BaseDirResolver) *Getter {
	return &Getter{
		fs:            fs,
		paths:         p,
		manifestStore: model.NewManifestStore(fs),
		resolver:      resolver,
	}
}

func NewDefaultGetter() *Getter {
	return NewGetter(vfs.OSFS, paths.DefaultPaths(), model.NewDefaultResolver())
}

func (g *Getter) Get(ctx context.Context, cfg *model.KyarabenConfig, configPath string, reg *registry.Registry, userStore *store.UserStore, manifestPath string) (*Result, error) {
	manifest, err := g.manifestStore.Load(manifestPath)
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
				if spec, ok := vers.GetPackage(e.Package.PackageName()); ok {
					info.DefaultVersion = spec.Default
				}
			}
		}

		for _, mc := range manifest.GetManagedConfigsForEmulator(emu.ID) {
			path, err := mc.Target.ResolveWith(g.resolver)
			if err != nil {
				continue
			}
			configInfo := ManagedConfigInfo{Path: path}
			for _, r := range mc.ManagedRegions {
				switch v := r.(type) {
				case model.FileRegion:
					configInfo.ManagedRegions = append(configInfo.ManagedRegions, ManagedRegionInfo{Type: "file"})
				case model.SectionRegion:
					configInfo.ManagedRegions = append(configInfo.ManagedRegions, ManagedRegionInfo{
						Type:      "section",
						Section:   v.Section,
						KeyPrefix: v.KeyPrefix,
					})
				}
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

	result.HealthWarning = g.detectOrphanedArtifacts(manifest)

	return result, nil
}

func (g *Getter) detectOrphanedArtifacts(manifest *model.Manifest) string {
	if len(manifest.InstalledEmulators) > 0 {
		return ""
	}

	stateDir, err := g.paths.StateDir()
	if err != nil {
		return ""
	}

	binDir := filepath.Join(stateDir, "bin")
	if entries, err := g.fs.ReadDir(binDir); err == nil && len(entries) > 0 {
		return "orphaned_artifacts"
	}

	currentLink := filepath.Join(stateDir, "current")
	if _, err := g.fs.Lstat(currentLink); err == nil {
		return "orphaned_artifacts"
	}

	return ""
}
