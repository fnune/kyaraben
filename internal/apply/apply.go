package apply

import (
	"context"
	"fmt"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
	"github.com/fnune/kyaraben/internal/store"
)

type Progress struct {
	Step    string
	Message string
}

type Result struct {
	StorePath string
	Patches   []model.ConfigPatch
}

type Options struct {
	DryRun   bool
	ShowDiff bool
	OnProgress func(Progress)
}

func Apply(cfg *model.KyarabenConfig, opts Options) (*Result, error) {
	if opts.OnProgress == nil {
		opts.OnProgress = func(Progress) {}
	}

	opts.OnProgress(Progress{Step: "start", Message: "Starting apply..."})

	userStorePath, err := cfg.ExpandUserStore()
	if err != nil {
		return nil, fmt.Errorf("expanding user store path: %w", err)
	}
	userStore := store.NewUserStore(userStorePath)

	nixClient, err := nix.NewClient()
	if err != nil {
		return nil, fmt.Errorf("creating nix client: %w", err)
	}

	if !opts.DryRun && !nixClient.IsAvailable() {
		return nil, fmt.Errorf("nix is not available")
	}

	emulatorsToInstall := make([]model.EmulatorID, 0, len(cfg.Systems))
	allPatches := make([]model.ConfigPatch, 0)

	for sys, sysConf := range cfg.Systems {
		emulatorsToInstall = append(emulatorsToInstall, sysConf.Emulator)

		gen := emulators.GetConfigGenerator(sysConf.Emulator)
		if gen == nil {
			continue
		}

		patches, err := gen.Generate(userStore, []model.SystemID{sys})
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

	if err := nixClient.EnsureFlakeDir(); err != nil {
		return nil, fmt.Errorf("creating flake directory: %w", err)
	}

	flakeGen := nix.NewFlakeGenerator()
	if err := flakeGen.Generate(nixClient.FlakePath, emulatorsToInstall); err != nil {
		return nil, fmt.Errorf("generating flake: %w", err)
	}

	opts.OnProgress(Progress{Step: "build", Message: "Building emulators..."})

	buildCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	flakeRef := flakeGen.DefaultFlakeRef(nixClient.FlakePath)
	storePath, err := nixClient.Build(buildCtx, flakeRef)
	if err != nil {
		return nil, fmt.Errorf("building emulators: %w", err)
	}

	opts.OnProgress(Progress{Step: "configs", Message: "Applying emulator configurations..."})

	configWriter := emulators.NewConfigWriter()
	for _, patch := range allPatches {
		if err := configWriter.Apply(patch); err != nil {
			return nil, fmt.Errorf("applying config %s: %w", patch.Config.Path, err)
		}
	}

	opts.OnProgress(Progress{Step: "manifest", Message: "Updating manifest..."})

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return nil, fmt.Errorf("getting manifest path: %w", err)
	}

	manifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	manifest.LastApplied = time.Now()

	for _, emuID := range emulatorsToInstall {
		manifest.AddEmulator(model.InstalledEmulator{
			ID:        emuID,
			Version:   "latest",
			StorePath: storePath,
			Installed: time.Now(),
		})
	}

	for _, patch := range allPatches {
		manifest.AddManagedConfig(model.ManagedConfig{
			Path:         patch.Config.Path,
			LastModified: time.Now(),
			EmulatorID:   patch.Config.EmulatorID,
		})
	}

	if err := manifest.Save(manifestPath); err != nil {
		return nil, fmt.Errorf("saving manifest: %w", err)
	}

	return &Result{
		StorePath: storePath,
		Patches:   allPatches,
	}, nil
}

func ComputeDiffs(patches []model.ConfigPatch) ([]*emulators.ConfigDiff, error) {
	diffs := make([]*emulators.ConfigDiff, 0, len(patches))
	for _, patch := range patches {
		diff, err := emulators.ComputeDiff(patch)
		if err != nil {
			return nil, fmt.Errorf("computing diff for %s: %w", patch.Config.Path, err)
		}
		diffs = append(diffs, diff)
	}
	return diffs, nil
}
