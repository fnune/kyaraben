package config

import (
	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/guestapp"
)

type Config = guestapp.Config
type ServiceConfig = guestapp.ServiceConfig
type PathMappings = guestapp.PathMappings
type ConfigStore = guestapp.ConfigStore

func NewConfigStore(fs vfs.FS, dataDir string) *ConfigStore {
	return guestapp.NewConfigStore(fs, dataDir)
}

func NewDefaultConfigStore(dataDir string) *ConfigStore {
	return guestapp.NewDefaultConfigStore(dataDir)
}

func DefaultConfig() Config {
	return Config{
		Service: ServiceConfig{
			Autostart: true,
		},
		PathMappings: PathMappings{
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
				"retroarch": "Screenshots",
			},
			States: map[string]string{
				"retroarch:fceumm":            ".userdata/shared/FC-fceumm",
				"retroarch:snes9x":            ".userdata/shared/SFC-snes9x",
				"retroarch:gambatte":          ".userdata/shared/GB-gambatte",
				"retroarch:mgba":              ".userdata/shared/GBA-mgba",
				"retroarch:gpsp":              ".userdata/shared/GBA-gpsp",
				"retroarch:pcsx_rearmed":      ".userdata/shared/PS-pcsx_rearmed",
				"retroarch:picodrive":         ".userdata/shared/MD-picodrive",
				"retroarch:fbneo":             ".userdata/shared/FBN-fbneo",
				"retroarch:mednafen_pce_fast": ".userdata/shared/PCE-mednafen_pce_fast",
				"retroarch:mednafen_ngp":      ".userdata/shared/NGP-race",
				"retroarch:stella":            ".userdata/shared/A2600-stella2014",
			},
		},
	}
}

func Load(dataDir string) (*Config, error) {
	return NewDefaultConfigStore(dataDir).Load(DefaultConfig())
}
