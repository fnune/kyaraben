package collections

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/sync"
)

func Dir(p *paths.Paths) (string, error) {
	stateDir, err := p.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, "collections"), nil
}

func Port(p *paths.Paths) int {
	return DefaultPort + p.InstancePortOffset()
}

func CommandPort(p *paths.Paths) int {
	return DefaultCommandPort + p.InstancePortOffset()
}

func NewDefaultService(fs vfs.FS, p *paths.Paths, resolver model.BaseDirResolver, cfg *model.KyarabenConfig, store model.StoreReader, getEmulator EmulatorLookup) (*Service, error) {
	dir, err := Dir(p)
	if err != nil {
		return nil, err
	}
	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return nil, err
	}
	lm, err := launcher.NewManager(fs, p, resolver)
	if err != nil {
		return nil, err
	}
	return NewService(ServiceDeps{
		FS:            fs,
		Config:        cfg,
		Store:         store,
		GetEmulator:   getEmulator,
		Source:        NewSource(fs, NewHTTPFetcher(), filepath.Join(dir, "cache")),
		Overrides:     NewOverrideStore(fs, filepath.Join(dir, "owned.json")),
		UserConfigDir: configDir,
		CommandPort:   CommandPort(p),
		BinDir:        lm.BinDir(),
		CoresDir:      lm.CoresDir(),
	}), nil
}

func NewDefaultUnits(fs vfs.FS, p *paths.Paths) (*Units, error) {
	configDir, err := paths.ConfigDir()
	if err != nil {
		return nil, err
	}
	return NewUnits(fs, sync.NewDefaultServiceManager(), filepath.Join(configDir, "systemd", "user"), p.DirName()+"-collections", p.Instance), nil
}

func launcherBinary(p *paths.Paths) (string, error) {
	if installed, err := p.CLIInstallPath(); err == nil {
		if _, err := os.Stat(installed); err == nil {
			return installed, nil
		}
	}
	exe, err := os.Executable() // may be a temporary AppImage mount, hence the installed CLI first
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

type SetupReport struct {
	Results  map[string][]FileResult
	Warnings []string
}

func SetUp(ctx context.Context, svc *Service, units *Units, p *paths.Paths, configured bool) SetupReport {
	report := SetupReport{Results: map[string][]FileResult{}}
	if !configured {
		if units.Installed() {
			if err := units.Remove(); err != nil {
				report.Warnings = append(report.Warnings, fmt.Sprintf("removing collections launcher service: %v", err))
			}
		}
		return report
	}

	binary, err := launcherBinary(p)
	if err == nil {
		err = units.Ensure(binary, Port(p))
	}
	if err != nil {
		report.Warnings = append(report.Warnings, fmt.Sprintf("setting up collections launcher service: %v", err))
	}

	for _, id := range svc.IDs() {
		results, err := svc.Sync(ctx, id, false)
		if err != nil {
			report.Warnings = append(report.Warnings, fmt.Sprintf("collection %s: %v", id, err))
			continue
		}
		report.Results[id] = results
	}
	return report
}

func Uninstall(fs vfs.FS, units *Units, p *paths.Paths) (removed []string, errs []error) {
	if units.Installed() {
		if err := units.Remove(); err != nil {
			errs = append(errs, fmt.Errorf("removing collections launcher units: %w", err))
		} else {
			removed = append(removed, units.socketPath(), units.servicePath())
		}
	}
	dir, err := Dir(p)
	if err != nil {
		return removed, append(errs, err)
	}
	results, err := NewOverrideStore(fs, filepath.Join(dir, "owned.json")).RemoveAll()
	if err != nil {
		errs = append(errs, err)
	}
	for _, r := range results {
		if r.Outcome == FileRemoved {
			removed = append(removed, r.Path)
		}
	}
	return removed, errs
}
