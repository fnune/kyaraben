package model

// EmulatorID uniquely identifies an emulator implementation.
// Format: "emulator" or "emulator:core" for RetroArch cores.
//
// Constants use the full type name as prefix (EmulatorID*) because tygo's
// enum_style: union requires the prefix to match the type name exactly.
type EmulatorID string

const (
	// EmulatorIDRetroArch is the base RetroArch installation shared by all cores.
	// Used for shared data like the cores directory.
	EmulatorIDRetroArch              EmulatorID = "retroarch"
	EmulatorIDRetroArchBsnes         EmulatorID = "retroarch:bsnes"
	EmulatorIDRetroArchMesen         EmulatorID = "retroarch:mesen"
	EmulatorIDRetroArchGenesisPlusGX EmulatorID = "retroarch:genesis_plus_gx"
	EmulatorIDRetroArchMupen64Plus   EmulatorID = "retroarch:mupen64plus_next"
	EmulatorIDRetroArchBeetleSaturn  EmulatorID = "retroarch:mednafen_saturn"
	EmulatorIDRetroArchMGBA          EmulatorID = "retroarch:mgba"
	EmulatorIDRetroArchMelonDS       EmulatorID = "retroarch:melonds"
	EmulatorIDDuckStation            EmulatorID = "duckstation"
	EmulatorIDPCSX2                  EmulatorID = "pcsx2"
	EmulatorIDRPCS3                  EmulatorID = "rpcs3"
	EmulatorIDVita3K                 EmulatorID = "vita3k"
	EmulatorIDPPSSPP                 EmulatorID = "ppsspp"
	EmulatorIDFlycast                EmulatorID = "flycast"
	EmulatorIDCemu                   EmulatorID = "cemu"
	EmulatorIDAzahar                 EmulatorID = "azahar"
	EmulatorIDDolphin                EmulatorID = "dolphin"
	EmulatorIDEden                   EmulatorID = "eden"
)
