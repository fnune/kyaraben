package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Saves       map[string]string `toml:"saves"`
	ROMs        map[string]string `toml:"roms"`
	BIOS        map[string]string `toml:"bios"`
	Screenshots map[string]string `toml:"screenshots"`
	Service     ServiceConfig     `toml:"service"`
}

type ServiceConfig struct {
	Enabled     bool `toml:"enabled"`
	StartOnBoot bool `toml:"start_on_boot"`
}

func DefaultConfig() Config {
	return Config{
		Service: ServiceConfig{
			Enabled:     true,
			StartOnBoot: true,
		},
		Saves: map[string]string{
			"nes":          "Saves/FC",
			"snes":         "Saves/SFC",
			"gb":           "Saves/GB",
			"gbc":          "Saves/GBC",
			"gba":          "Saves/GBA",
			"psx":          "Saves/PS",
			"genesis":      "Saves/MD",
			"gamegear":     "Saves/GG",
			"mastersystem": "Saves/SMS",
			"pcengine":     "Saves/PCE",
			"ngp":          "Saves/NGP",
			"atari2600":    "Saves/A2600",
			"c64":          "Saves/C64",
			"arcade":       "Saves/FBN",
		},
		ROMs: map[string]string{
			"nes":          "Roms/Nintendo (FC)",
			"snes":         "Roms/Super Nintendo (SFC)",
			"gb":           "Roms/Game Boy (GB)",
			"gbc":          "Roms/Game Boy Color (GBC)",
			"gba":          "Roms/Game Boy Advance (GBA)",
			"psx":          "Roms/PlayStation (PS)",
			"genesis":      "Roms/Mega Drive (MD)",
			"gamegear":     "Roms/Game Gear (GG)",
			"mastersystem": "Roms/Master System (SMS)",
			"pcengine":     "Roms/PC Engine (PCE)",
			"ngp":          "Roms/Neo Geo Pocket (NGP)",
			"atari2600":    "Roms/Atari 2600 (A2600)",
			"c64":          "Roms/Commodore 64 (C64)",
			"arcade":       "Roms/Arcade (FBN)",
		},
		BIOS: map[string]string{
			"gba": "Bios/GBA",
			"psx": "Bios/PS",
		},
		Screenshots: map[string]string{
			"retroarch":   "Screenshots",
			"duckstation": "Screenshots",
		},
	}
}

func Load(dataDir string) (*Config, error) {
	cfg := DefaultConfig()

	if dataDir == "" {
		return &cfg, nil
	}

	configPath := filepath.Join(dataDir, "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := cfg.save(dataDir); err != nil {
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

func (c *Config) Save(dataDir string) error {
	return c.save(dataDir)
}

func (c *Config) save(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(dataDir, "config.toml")
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return toml.NewEncoder(f).Encode(c)
}
