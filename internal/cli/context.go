package cli

import (
	"fmt"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
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

func (c *Context) NewRegistry() *registry.Registry   { return registry.NewDefault() }
func (c *Context) NewNixClient() (*nix.Client, error) { return nix.NewClient() }

func (c *Context) NewUserStore(cfg *model.KyarabenConfig) (*store.UserStore, error) {
	path, err := cfg.ExpandUserStore()
	if err != nil {
		return nil, err
	}
	return store.NewUserStore(path), nil
}
