package importscanner

import (
	"path/filepath"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
)

type SourceLayout interface {
	Name() string
	Detect(fs vfs.FS, root string) bool
	Classify(relPath string) *Classification
	ExpectedPaths(dataType DataType) []string
}

type Classification struct {
	DataType  DataType
	System    *model.SystemID
	Emulator  *model.EmulatorID
	Ambiguous bool
	Note      string
}

var systemAliases = map[string]model.SystemID{
	"gc":             model.SystemIDGameCube,
	"gamecube":       model.SystemIDGameCube,
	"ngc":            model.SystemIDGameCube,
	"wii":            model.SystemIDWii,
	"wiiu":           model.SystemIDWiiU,
	"wii_u":          model.SystemIDWiiU,
	"switch":         model.SystemIDSwitch,
	"ns":             model.SystemIDSwitch,
	"nes":            model.SystemIDNES,
	"famicom":        model.SystemIDNES,
	"fc":             model.SystemIDNES,
	"snes":           model.SystemIDSNES,
	"superfamicom":   model.SystemIDSNES,
	"sfc":            model.SystemIDSNES,
	"n64":            model.SystemIDN64,
	"nintendo64":     model.SystemIDN64,
	"gb":             model.SystemIDGB,
	"gameboy":        model.SystemIDGB,
	"gbc":            model.SystemIDGBC,
	"gameboycolor":   model.SystemIDGBC,
	"gba":            model.SystemIDGBA,
	"gameboyadvance": model.SystemIDGBA,
	"nds":            model.SystemIDNDS,
	"ds":             model.SystemIDNDS,
	"nintendods":     model.SystemIDNDS,
	"3ds":            model.SystemIDN3DS,
	"n3ds":           model.SystemIDN3DS,
	"nintendo3ds":    model.SystemIDN3DS,
	"psx":            model.SystemIDPSX,
	"ps1":            model.SystemIDPSX,
	"playstation":    model.SystemIDPSX,
	"ps2":            model.SystemIDPS2,
	"playstation2":   model.SystemIDPS2,
	"ps3":            model.SystemIDPS3,
	"playstation3":   model.SystemIDPS3,
	"psp":            model.SystemIDPSP,
	"psvita":         model.SystemIDPSVita,
	"vita":           model.SystemIDPSVita,
	"genesis":        model.SystemIDGenesis,
	"megadrive":      model.SystemIDGenesis,
	"md":             model.SystemIDGenesis,
	"mastersystem":   model.SystemIDMasterSystem,
	"sms":            model.SystemIDMasterSystem,
	"gamegear":       model.SystemIDGameGear,
	"gg":             model.SystemIDGameGear,
	"saturn":         model.SystemIDSaturn,
	"dreamcast":      model.SystemIDDreamcast,
	"dc":             model.SystemIDDreamcast,
	"pcengine":       model.SystemIDPCEngine,
	"pce":            model.SystemIDPCEngine,
	"tg16":           model.SystemIDPCEngine,
	"turbografx16":   model.SystemIDPCEngine,
	"ngp":            model.SystemIDNGP,
	"ngpc":           model.SystemIDNGP,
	"neogeopocket":   model.SystemIDNGP,
	"xbox":           model.SystemIDXbox,
	"xbox360":        model.SystemIDXbox360,
	"arcade":         model.SystemIDArcade,
	"mame":           model.SystemIDArcade,
	"neogeo":         model.SystemIDNeoGeo,
	"atari2600":      model.SystemIDAtari2600,
	"2600":           model.SystemIDAtari2600,
	"c64":            model.SystemIDC64,
	"commodore64":    model.SystemIDC64,
}

var emulatorAliases = map[string]model.EmulatorID{
	"dolphin":           model.EmulatorIDDolphin,
	"dolphin-emu":       model.EmulatorIDDolphin,
	"duckstation":       model.EmulatorIDDuckStation,
	"pcsx2":             model.EmulatorIDPCSX2,
	"rpcs3":             model.EmulatorIDRPCS3,
	"vita3k":            model.EmulatorIDVita3K,
	"ppsspp":            model.EmulatorIDPPSSPP,
	"flycast":           model.EmulatorIDFlycast,
	"cemu":              model.EmulatorIDCemu,
	"eden":              model.EmulatorIDEden,
	"xemu":              model.EmulatorIDXemu,
	"xenia":             model.EmulatorIDXeniaEdge,
	"xenia-edge":        model.EmulatorIDXeniaEdge,
	"retroarch":         model.EmulatorIDRetroArch,
	"bsnes":             model.EmulatorIDRetroArchBsnes,
	"mesen":             model.EmulatorIDRetroArchMesen,
	"genesis_plus_gx":   model.EmulatorIDRetroArchGenesisPlusGX,
	"mupen64plus":       model.EmulatorIDRetroArchMupen64Plus,
	"mupen64plus_next":  model.EmulatorIDRetroArchMupen64Plus,
	"mednafen_saturn":   model.EmulatorIDRetroArchBeetleSaturn,
	"beetle_saturn":     model.EmulatorIDRetroArchBeetleSaturn,
	"mednafen_pce":      model.EmulatorIDRetroArchBeetlePCE,
	"mednafen_pce_fast": model.EmulatorIDRetroArchBeetlePCE,
	"beetle_pce":        model.EmulatorIDRetroArchBeetlePCE,
	"mednafen_ngp":      model.EmulatorIDRetroArchBeetleNGP,
	"beetle_ngp":        model.EmulatorIDRetroArchBeetleNGP,
	"mgba":              model.EmulatorIDRetroArchMGBA,
	"melonds":           model.EmulatorIDRetroArchMelonDS,
	"citra":             model.EmulatorIDRetroArchCitra,
	"fbneo":             model.EmulatorIDRetroArchFBNeo,
	"stella":            model.EmulatorIDRetroArchStella,
	"vice":              model.EmulatorIDRetroArchVICE,
	"vice_x64sc":        model.EmulatorIDRetroArchVICE,
}

func NormalizeSystem(name string) *model.SystemID {
	normalized := strings.ToLower(strings.ReplaceAll(name, "-", ""))
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, " ", "")
	if sys, ok := systemAliases[normalized]; ok {
		return &sys
	}
	sysID := model.SystemID(normalized)
	return &sysID
}

func NormalizeEmulator(name string) *model.EmulatorID {
	normalized := strings.ToLower(name)
	if emu, ok := emulatorAliases[normalized]; ok {
		return &emu
	}
	emuID := model.EmulatorID(normalized)
	return &emuID
}

type GenericLayout struct{}

func (g *GenericLayout) Name() string { return "Generic" }

func (g *GenericLayout) Detect(fs vfs.FS, root string) bool {
	return true
}

func (g *GenericLayout) Classify(relPath string) *Classification {
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(parts) < 2 {
		return nil
	}

	topDir := strings.ToLower(parts[0])
	subDir := parts[1]

	switch topDir {
	case "roms":
		sys := NormalizeSystem(subDir)
		return &Classification{
			DataType: DataTypeROMs,
			System:   sys,
		}
	case "bios":
		sys := NormalizeSystem(subDir)
		return &Classification{
			DataType: DataTypeBIOS,
			System:   sys,
		}
	case "saves":
		sys := NormalizeSystem(subDir)
		return &Classification{
			DataType: DataTypeSaves,
			System:   sys,
		}
	case "states":
		emu := NormalizeEmulator(subDir)
		return &Classification{
			DataType: DataTypeStates,
			Emulator: emu,
		}
	case "screenshots":
		emu := NormalizeEmulator(subDir)
		return &Classification{
			DataType: DataTypeScreenshots,
			Emulator: emu,
		}
	}

	return nil
}

func (g *GenericLayout) ExpectedPaths(dataType DataType) []string {
	switch dataType {
	case DataTypeROMs:
		return []string{"roms"}
	case DataTypeBIOS:
		return []string{"bios"}
	case DataTypeSaves:
		return []string{"saves"}
	case DataTypeStates:
		return []string{"states"}
	case DataTypeScreenshots:
		return []string{"screenshots"}
	default:
		return nil
	}
}

type EmuDeckLayout struct{}

func (e *EmuDeckLayout) Name() string { return "EmuDeck" }

func (e *EmuDeckLayout) Detect(fs vfs.FS, root string) bool {
	_, err := fs.Stat(filepath.Join(root, "roms"))
	return err == nil
}

func (e *EmuDeckLayout) Classify(relPath string) *Classification {
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(parts) < 1 {
		return nil
	}

	topDir := strings.ToLower(parts[0])

	switch topDir {
	case "roms":
		if len(parts) < 2 {
			return nil
		}
		sys := NormalizeSystem(parts[1])
		return &Classification{
			DataType: DataTypeROMs,
			System:   sys,
		}
	case "bios":
		if len(parts) >= 2 {
			sys := NormalizeSystem(parts[1])
			return &Classification{
				DataType: DataTypeBIOS,
				System:   sys,
			}
		}
		return &Classification{
			DataType:  DataTypeBIOS,
			Ambiguous: true,
			Note:      "EmuDeck uses flat BIOS folder",
		}
	case "saves":
		if len(parts) < 2 {
			return nil
		}
		sys := NormalizeSystem(parts[1])
		return &Classification{
			DataType: DataTypeSaves,
			System:   sys,
		}
	}

	return nil
}

func (e *EmuDeckLayout) ExpectedPaths(dataType DataType) []string {
	switch dataType {
	case DataTypeROMs:
		return []string{"roms"}
	case DataTypeBIOS:
		return []string{"bios"}
	case DataTypeSaves:
		return []string{"saves"}
	default:
		return nil
	}
}

var AvailableLayouts = map[string]SourceLayout{
	"emudeck": &EmuDeckLayout{},
	"generic": &GenericLayout{},
}

func GetLayout(name string) SourceLayout {
	if layout, ok := AvailableLayouts[name]; ok {
		return layout
	}
	return &EmuDeckLayout{}
}

func LayoutNames() []string {
	return []string{"emudeck", "generic"}
}
