package config

import (
	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/guestapp"
	"github.com/fnune/kyaraben/internal/model"
)

type Config = guestapp.Config
type ServiceConfig = guestapp.ServiceConfig
type SyncConfig = guestapp.SyncConfig
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
				string(model.SystemIDNES):          "Saves/FC",
				string(model.SystemIDSNES):         "Saves/SFC",
				string(model.SystemIDGB):           "Saves/GB",
				string(model.SystemIDGBC):          "Saves/GBC",
				string(model.SystemIDGBA):          "Saves/GBA",
				string(model.SystemIDPSX):          "Saves/PS",
				string(model.SystemIDGenesis):      "Saves/MD",
				string(model.SystemIDGameGear):     "Saves/GG",
				string(model.SystemIDMasterSystem): "Saves/SMS",
				string(model.SystemIDPCEngine):     "Saves/PCE",
				string(model.SystemIDNGP):          "Saves/NGP",
				string(model.SystemIDAtari2600):    "Saves/A2600",
				string(model.SystemIDC64):          "Saves/C64",
				string(model.SystemIDArcade):       "Saves/FBN",
			},
			ROMs: map[string]string{
				string(model.SystemIDNES):          "Roms/Nintendo (FC)",
				string(model.SystemIDSNES):         "Roms/Super Nintendo (SFC)",
				string(model.SystemIDGB):           "Roms/Game Boy (GB)",
				string(model.SystemIDGBC):          "Roms/Game Boy Color (GBC)",
				string(model.SystemIDGBA):          "Roms/Game Boy Advance (GBA)",
				string(model.SystemIDPSX):          "Roms/PlayStation (PS)",
				string(model.SystemIDGenesis):      "Roms/Mega Drive (MD)",
				string(model.SystemIDGameGear):     "Roms/Game Gear (GG)",
				string(model.SystemIDMasterSystem): "Roms/Master System (SMS)",
				string(model.SystemIDPCEngine):     "Roms/PC Engine (PCE)",
				string(model.SystemIDNGP):          "Roms/Neo Geo Pocket (NGP)",
				string(model.SystemIDAtari2600):    "Roms/Atari 2600 (A2600)",
				string(model.SystemIDC64):          "Roms/Commodore 64 (C64)",
				string(model.SystemIDArcade):       "Roms/Arcade (FBN)",
			},
			BIOS: map[string]string{
				string(model.SystemIDGBA): "Bios/GBA",
				string(model.SystemIDPSX): "Bios/PS",
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
