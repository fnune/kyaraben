// Controller configuration types and hotkey binding parser.
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

func ValidateLayoutID(s string) (LayoutID, error) {
	switch LayoutID(s) {
	case LayoutStandard, LayoutNintendo:
		return LayoutID(s), nil
	default:
		return "", fmt.Errorf("unknown controller layout %q (valid: standard, nintendo)", s)
	}
}

// SteamDeckGUID is the virtual gamepad GUID that Steam Input presents for all
// controllers connected through Steam. On Steam Deck in game mode, every
// controller (Xbox, PlayStation, etc.) appears with this GUID.
const SteamDeckGUID = "03000000de280000ff11000001000000"

// FaceButtons returns the four face buttons (south, east, west, north) adjusted
// for the configured layout. Standard layout: A=south, B=east, X=west, Y=north.
// Nintendo layout swaps A/B and X/Y so that the physical button positions match
// the expected console behavior.
func (cc *ControllerConfig) FaceButtons() (south, east, west, north SDLButton) {
	if cc.Layout == LayoutNintendo {
		return ButtonB, ButtonA, ButtonY, ButtonX
	}
	return ButtonA, ButtonB, ButtonX, ButtonY
}

// SDLButtonIndex returns the standard SDL GameController button index.
// These indices are consistent across all controllers in SDL GameControllerDB.
var SDLButtonIndex = map[SDLButton]int{
	ButtonA:             0,
	ButtonB:             1,
	ButtonX:             2,
	ButtonY:             3,
	ButtonBack:          4,
	ButtonGuide:         5,
	ButtonStart:         6,
	ButtonLeftStick:     7,
	ButtonRightStick:    8,
	ButtonLeftShoulder:  9,
	ButtonRightShoulder: 10,
	ButtonDPadUp:        11,
	ButtonDPadDown:      12,
	ButtonDPadLeft:      13,
	ButtonDPadRight:     14,
}

// SDLAxisIndex maps SDL axis names to their standard indices.
const (
	AxisLeftX        = 0
	AxisLeftY        = 1
	AxisLeftTrigger  = 2
	AxisRightX       = 3
	AxisRightY       = 4
	AxisRightTrigger = 5
)
