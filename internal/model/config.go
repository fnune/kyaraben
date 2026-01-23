package model

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// KyarabenConfig represents the user's kyaraben configuration.
type KyarabenConfig struct {
	Global  GlobalConfig            `toml:"global"`
	Sync    SyncConfig              `toml:"sync"`
	Systems map[SystemID]SystemConf `toml:"systems"`
}

// GlobalConfig holds global settings.
type GlobalConfig struct {
	UserStore string `toml:"user_store"` // Path to emulation directory
}

// SystemConf holds per-system configuration.
type SystemConf struct {
	Emulator EmulatorID `toml:"emulator"`
}

// DefaultConfigPath returns the default path to the kyaraben config file.
func DefaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config directory: %w", err)
	}
	return filepath.Join(configDir, "kyaraben", "config.toml"), nil
}

// DefaultUserStore returns the default path to the user's emulation directory.
func DefaultUserStore() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, "Emulation"), nil
}

// LoadConfig loads the kyaraben configuration from a file.
func LoadConfig(path string) (*KyarabenConfig, error) {
	var cfg KyarabenConfig
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}
	return &cfg, nil
}

// SaveConfig writes the kyaraben configuration to a file.
func SaveConfig(cfg *KyarabenConfig, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}

	encoder := toml.NewEncoder(f)
	encodeErr := encoder.Encode(cfg)
	closeErr := f.Close()

	if encodeErr != nil {
		return fmt.Errorf("encoding config: %w", encodeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing config file: %w", closeErr)
	}
	return nil
}

// ExpandUserStore expands ~ in the user store path.
func (c *KyarabenConfig) ExpandUserStore() (string, error) {
	path := c.Global.UserStore
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expanding home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}
	return path, nil
}

// EnabledSystems returns a list of enabled system IDs.
func (c *KyarabenConfig) EnabledSystems() []SystemID {
	systems := make([]SystemID, 0, len(c.Systems))
	for id := range c.Systems {
		systems = append(systems, id)
	}
	return systems
}

// NewDefaultConfig creates a new config with default values.
func NewDefaultConfig() (*KyarabenConfig, error) {
	userStore, err := DefaultUserStore()
	if err != nil {
		return nil, err
	}
	return &KyarabenConfig{
		Global: GlobalConfig{
			UserStore: userStore,
		},
		Sync:    DefaultSyncConfig(),
		Systems: make(map[SystemID]SystemConf),
	}, nil
}
