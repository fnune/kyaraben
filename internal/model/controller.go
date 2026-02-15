// Controller configuration types and hotkey binding parser.
// GUID mapping data compiled from:
// - SDL GameControllerDB (https://github.com/mdqinc/SDL_GameControllerDB)
// - Linux xpad driver (https://github.com/torvalds/linux/blob/master/drivers/input/joystick/xpad.c)
// Hotkey defaults inspired by:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
package model

import (
	"fmt"
	"strings"
)

type LayoutID string

const (
	LayoutStandard LayoutID = "standard"
	LayoutNintendo LayoutID = "nintendo"
)

type ProfileID string

const (
	ProfileSteamDeck   ProfileID = "steam-deck"
	ProfileXbox        ProfileID = "xbox"
	ProfilePlayStation ProfileID = "playstation"
	ProfileNintendo    ProfileID = "nintendo"
)

// SDLButton represents a standard SDL GameController button name.
type SDLButton string

const (
	ButtonA             SDLButton = "A"
	ButtonB             SDLButton = "B"
	ButtonX             SDLButton = "X"
	ButtonY             SDLButton = "Y"
	ButtonBack          SDLButton = "Back"
	ButtonStart         SDLButton = "Start"
	ButtonGuide         SDLButton = "Guide"
	ButtonLeftShoulder  SDLButton = "LeftShoulder"
	ButtonRightShoulder SDLButton = "RightShoulder"
	ButtonLeftTrigger   SDLButton = "LeftTrigger"
	ButtonRightTrigger  SDLButton = "RightTrigger"
	ButtonLeftStick     SDLButton = "LeftStick"
	ButtonRightStick    SDLButton = "RightStick"
	ButtonDPadUp        SDLButton = "DPadUp"
	ButtonDPadDown      SDLButton = "DPadDown"
	ButtonDPadLeft      SDLButton = "DPadLeft"
	ButtonDPadRight     SDLButton = "DPadRight"
)

var validButtons = map[SDLButton]bool{
	ButtonA: true, ButtonB: true, ButtonX: true, ButtonY: true,
	ButtonBack: true, ButtonStart: true, ButtonGuide: true,
	ButtonLeftShoulder: true, ButtonRightShoulder: true,
	ButtonLeftTrigger: true, ButtonRightTrigger: true,
	ButtonLeftStick: true, ButtonRightStick: true,
	ButtonDPadUp: true, ButtonDPadDown: true,
	ButtonDPadLeft: true, ButtonDPadRight: true,
}

// HotkeyBinding represents a validated controller hotkey chord.
type HotkeyBinding struct {
	Buttons []SDLButton
}

func (h HotkeyBinding) String() string {
	parts := make([]string, len(h.Buttons))
	for i, b := range h.Buttons {
		parts[i] = string(b)
	}
	return strings.Join(parts, "+")
}

// ParseHotkeyBinding parses and validates a hotkey string like "Back+RightShoulder".
func ParseHotkeyBinding(s string) (HotkeyBinding, error) {
	if s == "" {
		return HotkeyBinding{}, fmt.Errorf("empty hotkey binding")
	}

	parts := strings.Split(s, "+")
	if len(parts) > 3 {
		return HotkeyBinding{}, fmt.Errorf("hotkey %q has %d components (max 3)", s, len(parts))
	}

	buttons := make([]SDLButton, len(parts))
	for i, p := range parts {
		b := SDLButton(strings.TrimSpace(p))
		if !validButtons[b] {
			return HotkeyBinding{}, fmt.Errorf("unknown button %q in hotkey %q", p, s)
		}
		buttons[i] = b
	}

	return HotkeyBinding{Buttons: buttons}, nil
}

// HotkeyConfig holds all configurable hotkey bindings.
type HotkeyConfig struct {
	SaveState        HotkeyBinding
	LoadState        HotkeyBinding
	NextSlot         HotkeyBinding
	PrevSlot         HotkeyBinding
	FastForward      HotkeyBinding
	Rewind           HotkeyBinding
	Pause            HotkeyBinding
	Screenshot       HotkeyBinding
	Quit             HotkeyBinding
	ToggleFullscreen HotkeyBinding
	OpenMenu         HotkeyBinding
}

// ControllerConfig holds the resolved controller configuration passed to generators.
type ControllerConfig struct {
	Layout  LayoutID
	Hotkeys HotkeyConfig
	GUIDs   map[string]ProfileID
}

// BuiltinGUIDs maps common controller GUIDs to profiles.
// Users can override these via [controller.guids] in config.toml.
var BuiltinGUIDs = map[string]ProfileID{
	// Steam Deck
	"03000000de280000ff11000001000000": ProfileSteamDeck,

	// Xbox 360
	"030000005e0400008e02000010010000": ProfileXbox,
	"030000005e0400009102000007010000": ProfileXbox,
	// Xbox One
	"030000005e040000ea02000001030000": ProfileXbox,
	// Xbox One S
	"030000005e040000d102000001010000": ProfileXbox,
	// Xbox Series X/S
	"030000005e040000130b000011050000": ProfileXbox,

	// DualShock 3
	"030000004c0500006802000011010000": ProfilePlayStation,
	// DualShock 4 v1
	"030000004c050000c405000011010000": ProfilePlayStation,
	// DualShock 4 v2
	"030000004c050000cc09000011010000": ProfilePlayStation,
	// DualSense
	"030000004c050000e60c000011010000": ProfilePlayStation,

	// Switch Pro Controller
	"030000007e0500000920000011010000": ProfileNintendo,
	// 8BitDo SN30 Pro
	"03000000c82d00000161000000010000": ProfileNintendo,

	// Handheld devices that use their own VID/PID but behave as Xbox controllers
	// ROG Ally X
	"0300000005b000004c1b000000000000": ProfileXbox,
	// Lenovo Legion Go
	"030000007eef00008261000000000000": ProfileXbox,
	// Lenovo Legion Go S
	"03000000861a000010e3000000000000": ProfileXbox,
	// MSI Claw A1M
	"03000000b00d00000119000000000000": ProfileXbox,
	// OneXPlayer
	"030000006325000058d0000000000000": ProfileXbox,
}

// MergeGUIDs returns a new map with user overrides applied on top of built-in GUIDs.
func MergeGUIDs(user map[string]ProfileID) map[string]ProfileID {
	merged := make(map[string]ProfileID, len(BuiltinGUIDs)+len(user))
	for k, v := range BuiltinGUIDs {
		merged[k] = v
	}
	for k, v := range user {
		merged[k] = v
	}
	return merged
}

// DefaultHotkeys returns the default hotkey configuration.
func DefaultHotkeys() HotkeyConfig {
	return HotkeyConfig{
		SaveState:        HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonRightShoulder}},
		LoadState:        HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonLeftShoulder}},
		NextSlot:         HotkeyBinding{Buttons: []SDLButton{ButtonStart, ButtonRightShoulder}},
		PrevSlot:         HotkeyBinding{Buttons: []SDLButton{ButtonStart, ButtonLeftShoulder}},
		FastForward:      HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonRightTrigger}},
		Rewind:           HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonLeftTrigger}},
		Pause:            HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonA}},
		Screenshot:       HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonB}},
		Quit:             HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonStart}},
		ToggleFullscreen: HotkeyBinding{Buttons: []SDLButton{ButtonStart, ButtonLeftStick}},
		OpenMenu:         HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonRightStick}},
	}
}

func ValidateProfileID(s string) (ProfileID, error) {
	switch ProfileID(s) {
	case ProfileSteamDeck, ProfileXbox, ProfilePlayStation, ProfileNintendo:
		return ProfileID(s), nil
	default:
		return "", fmt.Errorf("unknown controller profile %q (valid: steam-deck, xbox, playstation, nintendo)", s)
	}
}

func ValidateLayoutID(s string) (LayoutID, error) {
	switch LayoutID(s) {
	case LayoutStandard, LayoutNintendo:
		return LayoutID(s), nil
	default:
		return "", fmt.Errorf("unknown controller layout %q (valid: standard, nintendo)", s)
	}
}
