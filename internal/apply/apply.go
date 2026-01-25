package apply

import (
	"context"
	"fmt"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
)

const nixBuildTimeout = 30 * time.Minute

type Progress struct {
	Step    string
	Message string
}

type Result struct {
	StorePath string
	Patches   []model.ConfigPatch
	Backups   []BackupInfo
}

type BackupInfo struct {
	OriginalPath string
	BackupPath   string
}

type Options struct {
	DryRun        bool
	ShowDiff      bool
	CreateBackups bool
	OnProgress    func(Progress)
}

type PreflightResult struct {
	Patches       []model.ConfigPatch
	FilesToBackup []string
}

func (a *Applier) Preflight(cfg *model.KyarabenConfig, userStore *store.UserStore) (*PreflightResult, error) {
	allPatches := make([]model.ConfigPatch, 0)

	for _, sysConf := range cfg.Systems {
		gen := a.Registry.GetConfigGenerator(sysConf.Emulator)
		if gen == nil {
			continue
		}

		patches, err := gen.Generate(userStore)
		if err != nil {
			return nil, fmt.Errorf("generating config for %s: %w", sysConf.Emulator, err)
		}
		allPatches = append(allPatches, patches...)
	}

	manifest, err := model.LoadManifest(a.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	var filesToBackup []string
	for _, patch := range allPatches {
		path, exists, err := a.ConfigWriter.NeedsBackup(patch)
		if err != nil {
			return nil, fmt.Errorf("checking config file: %w", err)
		}

		if exists {
			if _, managed := manifest.GetManagedConfig(patch.Target); !managed {
				filesToBackup = append(filesToBackup, path)
			}
		}
	}

	return &PreflightResult{
		Patches:       allPatches,
		FilesToBackup: filesToBackup,
	}, nil
}

type Applier struct {
	NixClient      nix.NixClient
	FlakeGenerator *nix.FlakeGenerator
	ConfigWriter   *emulators.ConfigWriter
	Registry       *registry.Registry
	ManifestPath   string
}

func (a *Applier) Apply(cfg *model.KyarabenConfig, userStore *store.UserStore, opts Options) (*Result, error) {
	if opts.OnProgress == nil {
		opts.OnProgress = func(Progress) {}
	}

	opts.OnProgress(Progress{Step: "start", Message: "Starting apply..."})

	if !opts.DryRun && !a.NixClient.IsAvailable() {
		return nil, fmt.Errorf("nix is not available (nix-portable not found)")
	}

	emulatorsToInstall := make([]model.EmulatorID, 0, len(cfg.Systems))
	allPatches := make([]model.ConfigPatch, 0)

	for _, sysConf := range cfg.Systems {
		emulatorsToInstall = append(emulatorsToInstall, sysConf.Emulator)

		gen := a.Registry.GetConfigGenerator(sysConf.Emulator)
		if gen == nil {
			continue
		}

		patches, err := gen.Generate(userStore)
		if err != nil {
			return nil, fmt.Errorf("generating config for %s: %w", sysConf.Emulator, err)
		}
		allPatches = append(allPatches, patches...)
	}

	if opts.DryRun {
		return &Result{Patches: allPatches}, nil
	}

	opts.OnProgress(Progress{Step: "directories", Message: "Creating directory structure..."})

	if err := userStore.Initialize(); err != nil {
		return nil, fmt.Errorf("initializing user store: %w", err)
	}

	for sys := range cfg.Systems {
		if err := userStore.InitializeSystem(sys); err != nil {
			return nil, fmt.Errorf("initializing system %s: %w", sys, err)
		}
	}

	opts.OnProgress(Progress{Step: "flake", Message: "Generating Nix flake..."})

	if err := a.NixClient.EnsureFlakeDir(); err != nil {
		return nil, fmt.Errorf("creating flake directory: %w", err)
	}

	if err := a.FlakeGenerator.Generate(a.NixClient.GetFlakePath(), emulatorsToInstall); err != nil {
		return nil, fmt.Errorf("generating flake: %w", err)
	}

	opts.OnProgress(Progress{Step: "build", Message: "Building emulators..."})

	buildCtx, cancel := context.WithTimeout(context.Background(), nixBuildTimeout)
	defer cancel()

	flakeRef := a.FlakeGenerator.DefaultFlakeRef(a.NixClient.GetFlakePath())
	storePath, err := a.NixClient.Build(buildCtx, flakeRef)
	if err != nil {
		return nil, fmt.Errorf("building emulators: %w", err)
	}

	opts.OnProgress(Progress{Step: "configs", Message: "Applying emulator configurations..."})

	manifest, err := model.LoadManifest(a.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	configResults := make([]emulators.ApplyResult, len(allPatches))
	var backups []BackupInfo
	for i, patch := range allPatches {
		createBackup := false
		if opts.CreateBackups {
			_, exists, err := a.ConfigWriter.NeedsBackup(patch)
			if err != nil {
				return nil, fmt.Errorf("checking config file: %w", err)
			}
			if exists {
				if _, managed := manifest.GetManagedConfig(patch.Target); !managed {
					createBackup = true
				}
			}
		}

		result, err := a.ConfigWriter.ApplyWithOptions(patch, emulators.ApplyOptions{
			CreateBackup: createBackup,
		})
		if err != nil {
			return nil, fmt.Errorf("applying config: %w", err)
		}
		configResults[i] = result

		if result.BackupPath != "" {
			backups = append(backups, BackupInfo{
				OriginalPath: result.Path,
				BackupPath:   result.BackupPath,
			})
		}
	}

	opts.OnProgress(Progress{Step: "manifest", Message: "Updating manifest..."})

	manifest.LastApplied = time.Now()

	for _, emuID := range emulatorsToInstall {
		manifest.AddEmulator(model.InstalledEmulator{
			ID:        emuID,
			Version:   "latest",
			StorePath: storePath,
			Installed: time.Now(),
		})
	}

	for i, patch := range allPatches {
		managedKeys := make([]model.ManagedKey, len(patch.Entries))
		for j, entry := range patch.Entries {
			managedKeys[j] = model.ManagedKey(entry)
		}

		manifest.AddManagedConfig(model.ManagedConfig{
			Target:       patch.Target,
			BaselineHash: configResults[i].BaselineHash,
			LastModified: time.Now(),
			ManagedKeys:  managedKeys,
		})
	}

	if err := manifest.Save(a.ManifestPath); err != nil {
		return nil, fmt.Errorf("saving manifest: %w", err)
	}

	return &Result{
		StorePath: storePath,
		Patches:   allPatches,
		Backups:   backups,
	}, nil
}

func ComputeDiffs(patches []model.ConfigPatch) ([]*emulators.ConfigDiff, error) {
	diffs := make([]*emulators.ConfigDiff, 0, len(patches))
	for _, patch := range patches {
		diff, err := emulators.ComputeDiff(patch)
		if err != nil {
			return nil, fmt.Errorf("computing diff: %w", err)
		}
		diffs = append(diffs, diff)
	}
	return diffs, nil
}
