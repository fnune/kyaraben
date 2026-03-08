package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type ServiceConfig struct {
	Autostart bool `toml:"autostart"`
}

type Config struct {
	Service ServiceConfig `toml:"service"`

	ROMs        map[string]string `toml:"roms"`
	Saves       map[string]string `toml:"saves"`
	Screenshots map[string]string `toml:"screenshots"`
}

func DefaultConfig() Config {
	return Config{
		Service: ServiceConfig{
			Autostart: false,
		},
		ROMs:        defaultROMs(),
		Saves:       defaultSaves(),
		Screenshots: defaultScreenshots(),
	}
}

func defaultROMs() map[string]string {
	systems := []string{
		"nes", "snes", "n64", "gb", "gbc", "gba", "nds",
		"gamecube", "wii",
		"psx", "ps2", "psp",
		"megadrive", "mastersystem", "gamegear", "saturn", "dreamcast",
		"pcengine", "ngp", "neogeo",
		"atari2600", "c64",
	}

	roms := make(map[string]string, len(systems))
	for _, sys := range systems {
		roms[sys] = filepath.Join("roms", sys)
	}
	return roms
}

func defaultSaves() map[string]string {
	systems := []string{
		"nes", "snes", "n64", "gb", "gbc", "gba", "nds",
		"gamecube", "wii",
		"psx", "ps2", "psp",
		"megadrive", "mastersystem", "gamegear", "saturn", "dreamcast",
		"pcengine", "ngp", "neogeo",
		"atari2600", "c64",
	}

	saves := make(map[string]string, len(systems))
	for _, sys := range systems {
		saves[sys] = filepath.Join("saves", sys)
	}
	return saves
}

func defaultScreenshots() map[string]string {
	return map[string]string{
		"retroarch": "screenshots",
	}
}

type ConfigStore struct {
	path string
}

func NewConfigStore(dataDir string) *ConfigStore {
	return &ConfigStore{
		path: filepath.Join(dataDir, "config.toml"),
	}
}

func (s *ConfigStore) Load(defaults Config) (*Config, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &defaults, nil
		}
		return nil, err
	}

	cfg := defaults
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (s *ConfigStore) Save(cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	f, err := os.Create(s.path)
	if err != nil {
		return err
	}

	encodeErr := toml.NewEncoder(f).Encode(cfg)
	closeErr := f.Close()
	if encodeErr != nil {
		return encodeErr
	}
	return closeErr
}
