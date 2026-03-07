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
}

func DefaultConfig() Config {
	return Config{
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

func Load(userdataPath, platform string) (Config, error) {
	cfg := DefaultConfig()

	if userdataPath == "" || platform == "" {
		return cfg, nil
	}

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
