package esde

import "github.com/fnune/kyaraben/internal/model"

type SystemMapping struct {
	Name     string
	FullName string
	Platform string
}

var systemMappings = map[model.SystemID]SystemMapping{
	model.SystemIDNES: {
		Name:     "nes",
		FullName: "Nintendo Entertainment System",
		Platform: "nes",
	},
	model.SystemIDSNES: {
		Name:     "snes",
		FullName: "Super Nintendo Entertainment System",
		Platform: "snes",
	},
	model.SystemIDN64: {
		Name:     "n64",
		FullName: "Nintendo 64",
		Platform: "n64",
	},
	model.SystemIDGB: {
		Name:     "gb",
		FullName: "Nintendo Game Boy",
		Platform: "gb",
	},
	model.SystemIDGBC: {
		Name:     "gbc",
		FullName: "Nintendo Game Boy Color",
		Platform: "gbc",
	},
	model.SystemIDGBA: {
		Name:     "gba",
		FullName: "Nintendo Game Boy Advance",
		Platform: "gba",
	},
	model.SystemIDNDS: {
		Name:     "nds",
		FullName: "Nintendo DS",
		Platform: "nds",
	},
	model.SystemIDN3DS: {
		Name:     "n3ds",
		FullName: "Nintendo 3DS",
		Platform: "n3ds",
	},
	model.SystemIDGameCube: {
		Name:     "gc",
		FullName: "Nintendo GameCube",
		Platform: "gc",
	},
	model.SystemIDWii: {
		Name:     "wii",
		FullName: "Nintendo Wii",
		Platform: "wii",
	},
	model.SystemIDWiiU: {
		Name:     "wiiu",
		FullName: "Nintendo Wii U",
		Platform: "wiiu",
	},
	model.SystemIDSwitch: {
		Name:     "switch",
		FullName: "Nintendo Switch",
		Platform: "switch",
	},
	model.SystemIDPSX: {
		Name:     "psx",
		FullName: "Sony PlayStation",
		Platform: "psx",
	},
	model.SystemIDPS2: {
		Name:     "ps2",
		FullName: "Sony PlayStation 2",
		Platform: "ps2",
	},
	model.SystemIDPS3: {
		Name:     "ps3",
		FullName: "Sony PlayStation 3",
		Platform: "ps3",
	},
	model.SystemIDPSP: {
		Name:     "psp",
		FullName: "Sony PlayStation Portable",
		Platform: "psp",
	},
	model.SystemIDPSVita: {
		Name:     "psvita",
		FullName: "Sony PlayStation Vita",
		Platform: "psvita",
	},
	model.SystemIDGenesis: {
		Name:     "genesis",
		FullName: "Sega Genesis",
		Platform: "genesis",
	},
	model.SystemIDSaturn: {
		Name:     "saturn",
		FullName: "Sega Saturn",
		Platform: "saturn",
	},
	model.SystemIDDreamcast: {
		Name:     "dreamcast",
		FullName: "Sega Dreamcast",
		Platform: "dreamcast",
	},
	model.SystemIDMasterSystem: {
		Name:     "mastersystem",
		FullName: "Sega Master System",
		Platform: "mastersystem",
	},
	model.SystemIDGameGear: {
		Name:     "gamegear",
		FullName: "Sega Game Gear",
		Platform: "gamegear",
	},
	model.SystemIDPCEngine: {
		Name:     "pcengine",
		FullName: "NEC PC Engine",
		Platform: "pcengine",
	},
	model.SystemIDNGP: {
		Name:     "ngp",
		FullName: "SNK Neo Geo Pocket",
		Platform: "ngp",
	},
	model.SystemIDNeoGeo: {
		Name:     "neogeo",
		FullName: "SNK Neo Geo",
		Platform: "neogeo",
	},
	model.SystemIDXbox: {
		Name:     "xbox",
		FullName: "Microsoft Xbox",
		Platform: "xbox",
	},
	model.SystemIDXbox360: {
		Name:     "xbox360",
		FullName: "Microsoft Xbox 360",
		Platform: "xbox360",
	},
	model.SystemIDAtari2600: {
		Name:     "atari2600",
		FullName: "Atari 2600",
		Platform: "atari2600",
	},
	model.SystemIDC64: {
		Name:     "c64",
		FullName: "Commodore 64",
		Platform: "c64",
	},
	model.SystemIDArcade: {
		Name:     "arcade",
		FullName: "Arcade",
		Platform: "arcade",
	},
}

func GetSystemMapping(id model.SystemID) (SystemMapping, bool) {
	m, ok := systemMappings[id]
	return m, ok
}
