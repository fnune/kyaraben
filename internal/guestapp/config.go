package guestapp

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/twpayne/go-vfs/v5"
)

type ServiceConfig struct {
	Autostart  bool `toml:"autostart"`
	SyncStates bool `toml:"sync_states"`
}

type PathMappings struct {
	Saves       map[string]string `toml:"saves"`
	ROMs        map[string]string `toml:"roms"`
	BIOS        map[string]string `toml:"bios"`
	Screenshots map[string]string `toml:"screenshots"`
	States      map[string]string `toml:"states"`
}

type Config struct {
	PathMappings
	Service ServiceConfig `toml:"service"`
}

type ConfigStore struct {
	fs      vfs.FS
	dataDir string
}

func NewConfigStore(fs vfs.FS, dataDir string) *ConfigStore {
	return &ConfigStore{fs: fs, dataDir: dataDir}
}

func NewDefaultConfigStore(dataDir string) *ConfigStore {
	return NewConfigStore(vfs.OSFS, dataDir)
}

func (s *ConfigStore) Load(defaults Config) (*Config, error) {
	cfg := defaults

	if s.dataDir == "" {
		return &cfg, nil
	}

	configPath := filepath.Join(s.dataDir, "config.toml")
	data, err := s.fs.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := s.save(&cfg); err != nil {
				return &cfg, err
			}
			return &cfg, nil
		}
		return &cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}

func (s *ConfigStore) Save(c *Config) error {
	return s.save(c)
}

func (s *ConfigStore) save(c *Config) error {
	if err := vfs.MkdirAll(s.fs, s.dataDir, 0755); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(c); err != nil {
		return err
	}

	configPath := filepath.Join(s.dataDir, "config.toml")
	return s.fs.WriteFile(configPath, buf.Bytes(), 0644)
}
