package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	TagOverrides map[string]string `toml:"tag_overrides"`
}

func DefaultConfig() Config {
	return Config{
		TagOverrides: make(map[string]string),
	}
}

func Load(userdataPath, platform string) (Config, error) {
	cfg := DefaultConfig()

	configPath := filepath.Join(userdataPath, platform, "kyaraben", "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (c *Config) Save(userdataPath, platform string) error {
	configDir := filepath.Join(userdataPath, platform, "kyaraben")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.toml")
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return toml.NewEncoder(f).Encode(c)
}
