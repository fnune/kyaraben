package model

import "strings"

// EmulatorID uniquely identifies an emulator implementation.
// Format: "emulator" or "emulator:core" for RetroArch cores.
//
// Constants use the full type name as prefix (EmulatorID*) because tygo's
// enum_style: union requires the prefix to match the type name exactly.
type EmulatorID string

// IsRetroArchCore returns true if this is a RetroArch core (not the base retroarch).
func (id EmulatorID) IsRetroArchCore() bool {
	return strings.HasPrefix(string(id), "retroarch:") && id != EmulatorIDRetroArch
}

// RetroArchCoreName returns the core name (e.g., "bsnes" from "retroarch:bsnes"),
// or empty string if this is not a RetroArch core.
func (id EmulatorID) RetroArchCoreName() string {
	if !id.IsRetroArchCore() {
		return ""
	}
	return strings.TrimPrefix(string(id), "retroarch:")
}

const (
	// EmulatorIDRetroArch is the base RetroArch installation shared by all cores.
	// Used for shared data like the cores directory.
	EmulatorIDRetroArch              EmulatorID = "retroarch"
	EmulatorIDRetroArchSnes9x        EmulatorID = "retroarch:snes9x"
	EmulatorIDRetroArchBsnes         EmulatorID = "retroarch:bsnes"
	EmulatorIDRetroArchMesen         EmulatorID = "retroarch:mesen"
	EmulatorIDRetroArchGenesisPlusGX EmulatorID = "retroarch:genesis_plus_gx"
	EmulatorIDRetroArchMupen64Plus   EmulatorID = "retroarch:mupen64plus_next"
	EmulatorIDRetroArchBeetleSaturn  EmulatorID = "retroarch:mednafen_saturn"
	EmulatorIDRetroArchBeetlePCE     EmulatorID = "retroarch:mednafen_pce_fast"
	EmulatorIDRetroArchBeetleNGP     EmulatorID = "retroarch:mednafen_ngp"
	EmulatorIDRetroArchMGBA          EmulatorID = "retroarch:mgba"
	EmulatorIDRetroArchMelonDS       EmulatorID = "retroarch:melondsds"
	EmulatorIDRetroArchCitra         EmulatorID = "retroarch:citra"
	EmulatorIDRetroArchAzahar        EmulatorID = "retroarch:azahar"
	EmulatorIDRetroArchFBNeo         EmulatorID = "retroarch:fbneo"
	EmulatorIDRetroArchStella        EmulatorID = "retroarch:stella"
	EmulatorIDRetroArchVICE          EmulatorID = "retroarch:vice_x64sc"
	EmulatorIDDuckStation            EmulatorID = "duckstation"
	EmulatorIDPCSX2                  EmulatorID = "pcsx2"
	EmulatorIDRPCS3                  EmulatorID = "rpcs3"
	EmulatorIDVita3K                 EmulatorID = "vita3k"
	EmulatorIDPPSSPP                 EmulatorID = "ppsspp"
	EmulatorIDFlycast                EmulatorID = "flycast"
	EmulatorIDCemu                   EmulatorID = "cemu"
	EmulatorIDDolphin                EmulatorID = "dolphin"
	EmulatorIDEden                   EmulatorID = "eden"
	EmulatorIDXemu                   EmulatorID = "xemu"
	EmulatorIDXeniaEdge              EmulatorID = "xenia-edge"
)
