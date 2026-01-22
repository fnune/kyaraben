package cli

import (
	"fmt"

	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
	"github.com/fnune/kyaraben/internal/store"
)

// Context holds shared state for CLI commands.
type Context struct {
	ConfigPath string
}

// LoadConfig loads the kyaraben configuration.
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

// GetConfigPath returns the config path, resolving the default if needed.
func (c *Context) GetConfigPath() (string, error) {
	if c.ConfigPath != "" {
		return c.ConfigPath, nil
	}
	return model.DefaultConfigPath()
}

// NewRegistry creates a new emulator registry.
func (c *Context) NewRegistry() *emulators.Registry {
	return emulators.NewRegistry()
}

// NewNixClient creates a new Nix client.
func (c *Context) NewNixClient() (*nix.Client, error) {
	return nix.NewClient()
}

// NewUserStore creates a UserStore from the config.
func (c *Context) NewUserStore(cfg *model.KyarabenConfig) (*store.UserStore, error) {
	path, err := cfg.ExpandUserStore()
	if err != nil {
		return nil, err
	}
	return store.NewUserStore(path), nil
}
