package apply

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
)

const nixBuildTimeout = 30 * time.Minute

type Progress struct {
	Step    string
	Message string
	Output  string // Optional streaming output line (e.g., from nix build)
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

		if !exists {
			continue
		}

		if _, managed := manifest.GetManagedConfig(patch.Target); managed {
			continue
		}

		diff, err := emulators.ComputeDiff(patch)
		if err != nil {
			return nil, fmt.Errorf("computing diff for %s: %w", path, err)
		}

		if diff.HasChanges() {
			filesToBackup = append(filesToBackup, path)
		}
	}

	return &PreflightResult{
		Patches:       allPatches,
		FilesToBackup: filesToBackup,
	}, nil
}

type Applier struct {
	NixClient       nix.NixClient
	FlakeGenerator  *nix.FlakeGenerator
	ConfigWriter    *emulators.ConfigWriter
	Registry        *registry.Registry
	ManifestPath    string
	LauncherManager *launcher.Manager
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

	genPath, err := a.FlakeGenerator.Generate(a.NixClient.GetFlakePath(), emulatorsToInstall)
	if err != nil {
		return nil, fmt.Errorf("generating flake: %w", err)
	}

	opts.OnProgress(Progress{Step: "build", Message: "Installing emulators..."})

	a.NixClient.SetOutputCallback(func(line string) {
		opts.OnProgress(Progress{Step: "build", Output: line})
	})
	defer a.NixClient.SetOutputCallback(nil)

	buildCtx, cancel := context.WithTimeout(context.Background(), nixBuildTimeout)
	defer cancel()

	flakeRef := a.FlakeGenerator.DefaultFlakeRef(string(genPath))

	var profileLink string
	if a.LauncherManager != nil {
		profileLink = a.LauncherManager.CurrentLink()
	}

	var storePath string
	if profileLink != "" {
		// Remove existing symlink before building - nix won't overwrite it
		if _, err := os.Lstat(profileLink); err == nil {
			if err := os.Remove(profileLink); err != nil {
				return nil, fmt.Errorf("removing existing profile link: %w", err)
			}
		}

		if err := a.NixClient.BuildWithLink(buildCtx, flakeRef, profileLink); err != nil {
			return nil, fmt.Errorf("building emulators: %w", err)
		}
		target, err := os.Readlink(profileLink)
		if err != nil {
			return nil, fmt.Errorf("reading profile link: %w", err)
		}

		// nix-portable virtualizes /nix/store, so the symlink target doesn't exist
		// on the real filesystem. Translate to the real store path.
		realTarget := a.NixClient.RealStorePath(target)
		if realTarget != target {
			if err := os.Remove(profileLink); err != nil {
				return nil, fmt.Errorf("removing old profile link: %w", err)
			}
			if err := os.Symlink(realTarget, profileLink); err != nil {
				return nil, fmt.Errorf("creating real profile link: %w", err)
			}
		}
		storePath = realTarget
	} else {
		var err error
		storePath, err = a.NixClient.Build(buildCtx, flakeRef)
		if err != nil {
			return nil, fmt.Errorf("building emulators: %w", err)
		}
	}

	if a.LauncherManager != nil {
		opts.OnProgress(Progress{Step: "wrappers", Message: "Generating launcher scripts..."})

		a.LauncherManager.SetNixPortableBinary(a.NixClient.GetNixPortableBinary())
		a.LauncherManager.SetNixPortableLocation(a.NixClient.GetNixPortableLocation())
		if err := a.LauncherManager.GenerateWrappers(); err != nil {
			return nil, fmt.Errorf("generating launcher wrappers: %w", err)
		}

		desktopEntries := a.buildDesktopEntries(emulatorsToInstall)
		if err := a.LauncherManager.GenerateDesktopFiles(desktopEntries); err != nil {
			return nil, fmt.Errorf("generating desktop files: %w", err)
		}
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

func (a *Applier) buildDesktopEntries(emulatorIDs []model.EmulatorID) []launcher.DesktopEntry {
	seenBinaries := make(map[string]bool)
	var entries []launcher.DesktopEntry

	for _, emuID := range emulatorIDs {
		emu, err := a.Registry.GetEmulator(emuID)
		if err != nil || emu.Launcher.Binary == "" {
			continue
		}

		if seenBinaries[emu.Launcher.Binary] {
			continue
		}
		seenBinaries[emu.Launcher.Binary] = true

		displayName := emu.Launcher.DisplayName
		if displayName == "" {
			displayName = emu.Name
		}

		if emu.Package.Source() == model.PackageSourceNixpkgs {
			entries = append(entries, launcher.NixStoreDesktop{
				BinaryName: emu.Launcher.Binary,
			})
		} else {
			entries = append(entries, launcher.GeneratedDesktop{
				BinaryName:    emu.Launcher.Binary,
				Name:          displayName,
				GenericName:   emu.Launcher.GenericName,
				CategoriesStr: strings.Join(emu.Launcher.Categories, ";"),
			})
		}
	}

	return entries
}
