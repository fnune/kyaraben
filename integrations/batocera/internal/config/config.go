package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/twpayne/go-vfs/v5"
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

var kyarabenToBatoceraSystem = map[string]string{
	"genesis": "megadrive",
}

func batoceraSystemName(kyarabenID string) string {
	if mapped, ok := kyarabenToBatoceraSystem[kyarabenID]; ok {
		return mapped
	}
	return kyarabenID
}

func defaultROMs() map[string]string {
	systems := []string{
		"nes", "snes", "n64", "gb", "gbc", "gba", "nds",
		"gamecube", "wii",
		"psx", "ps2", "psp",
		"genesis", "mastersystem", "gamegear", "saturn", "dreamcast",
		"pcengine", "ngp", "neogeo",
		"atari2600", "c64",
	}

	roms := make(map[string]string, len(systems))
	for _, sys := range systems {
		roms[sys] = filepath.Join("roms", batoceraSystemName(sys))
	}
	return roms
}

func defaultSaves() map[string]string {
	systems := []string{
		"nes", "snes", "n64", "gb", "gbc", "gba", "nds",
		"gamecube", "wii",
		"psx", "ps2", "psp",
		"genesis", "mastersystem", "gamegear", "saturn", "dreamcast",
		"pcengine", "ngp", "neogeo",
		"atari2600", "c64",
	}

	saves := make(map[string]string, len(systems))
	for _, sys := range systems {
		saves[sys] = filepath.Join("saves", batoceraSystemName(sys))
	}
	return saves
}

func defaultScreenshots() map[string]string {
	return map[string]string{
		"retroarch": "screenshots",
	}
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
	configPath := filepath.Join(s.dataDir, "config.toml")
	data, err := s.fs.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &defaults, nil
		}
		return nil, err
	}

	cfg := defaults
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (s *ConfigStore) Save(cfg *Config) error {
	if err := vfs.MkdirAll(s.fs, s.dataDir, 0755); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}

	configPath := filepath.Join(s.dataDir, "config.toml")
	return s.fs.WriteFile(configPath, buf.Bytes(), 0644)
}
