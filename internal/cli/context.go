package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/packages"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/status"
	"github.com/fnune/kyaraben/internal/store"
)

type Context struct {
	FS          vfs.FS
	Paths       *paths.Paths
	Resolver    model.BaseDirResolver
	ConfigStore *model.ConfigStore
	ConfigPath  string
}

func NewContext(fs vfs.FS, p *paths.Paths, resolver model.BaseDirResolver, configPath string) *Context {
	return &Context{
		FS:          fs,
		Paths:       p,
		Resolver:    resolver,
		ConfigStore: model.NewConfigStore(fs),
		ConfigPath:  configPath,
	}
}

func NewDefaultContext(instance, configPath string) *Context {
	return NewContext(vfs.OSFS, paths.NewPaths(instance), model.NewDefaultResolver(), configPath)
}

func (c *Context) GetPaths() *paths.Paths {
	return c.Paths
}

func (c *Context) LoadConfig() (*model.KyarabenConfig, error) {
	result, err := c.LoadConfigWithWarnings()
	if err != nil {
		return nil, err
	}
	return result.Config, nil
}

func (c *Context) LoadConfigWithWarnings() (*model.LoadResult, error) {
	path, err := c.GetConfigPath()
	if err != nil {
		return nil, err
	}

	reg := c.NewRegistry()
	validators := &model.ConfigValidators{
		GetEmulator: reg.GetEmulator,
		GetSystem:   reg.GetSystem,
		GetFrontend: reg.GetFrontend,
	}

	result, err := c.ConfigStore.LoadWithWarnings(path, validators)
	if err != nil {
		return nil, fmt.Errorf("loading config from %s: %w", path, err)
	}
	return result, nil
}

func (c *Context) SaveConfig(cfg *model.KyarabenConfig, path string) error {
	return c.ConfigStore.Save(cfg, path)
}

func (c *Context) GetConfigPath() (string, error) {
	if c.ConfigPath != "" {
		return c.ConfigPath, nil
	}
	return c.Paths.ConfigPath()
}

func (c *Context) NewRegistry() *registry.Registry { return registry.NewDefault() }

func (c *Context) NewInstaller() (packages.Installer, error) {
	stateDir, err := c.stateDir()
	if err != nil {
		return nil, fmt.Errorf("getting state directory: %w", err)
	}
	if useFakeInstaller() {
		packagesDir := filepath.Join(stateDir, "packages")
		return packages.NewFakeInstaller(c.FS, packagesDir), nil
	}
	downloader := packages.NewHTTPDownloader()
	extractor := packages.NewDefaultExtractor()
	return packages.NewDefaultPackageInstaller(stateDir, downloader, extractor), nil
}

func (c *Context) NewCollection(cfg *model.KyarabenConfig) (*store.Collection, error) {
	return store.NewCollection(c.FS, c.Paths, cfg.Global.Collection)
}

func (c *Context) NewStatusGetter() *status.Getter {
	return status.NewGetter(c.FS, c.Paths, c.Resolver)
}

func (c *Context) stateDir() (string, error) {
	return c.Paths.StateDir()
}

func useFakeInstaller() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("KYARABEN_E2E_FAKE_INSTALLER")))
	return value == "1" || value == "true" || value == "yes"
}
