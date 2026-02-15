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
	"github.com/fnune/kyaraben/internal/store"
)

type Context struct {
	ConfigPath string
}

func (c *Context) LoadConfig() (*model.KyarabenConfig, error) {
	path := c.ConfigPath
	if path == "" {
		var err error
		path, err = model.DefaultConfigPath()
		if err != nil {
			return nil, err
		}
	}

	cfg, err := model.LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("loading config from %s: %w", path, err)
	}
	return cfg, nil
}

func (c *Context) GetConfigPath() (string, error) {
	if c.ConfigPath != "" {
		return c.ConfigPath, nil
	}
	return model.DefaultConfigPath()
}

func (c *Context) NewRegistry() *registry.Registry { return registry.NewDefault() }

func (c *Context) NewInstaller() (packages.Installer, error) {
	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return nil, fmt.Errorf("getting state directory: %w", err)
	}
	if useFakeInstaller() {
		packagesDir := filepath.Join(stateDir, "packages")
		return packages.NewFakeInstaller(vfs.OSFS, packagesDir), nil
	}
	downloader := packages.NewHTTPDownloader()
	extractor := packages.NewDefaultExtractor()
	return packages.NewDefaultPackageInstaller(stateDir, downloader, extractor), nil
}

func (c *Context) NewUserStore(cfg *model.KyarabenConfig) (*store.UserStore, error) {
	return store.NewDefaultUserStore(cfg.Global.UserStore)
}

func (c *Context) stateDir() (string, error) {
	return paths.KyarabenStateDir()
}

func useFakeInstaller() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("KYARABEN_E2E_FAKE_INSTALLER")))
	return value == "1" || value == "true" || value == "yes"
}
