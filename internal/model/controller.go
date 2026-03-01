// Controller configuration types and hotkey binding parser.
// Hotkey defaults inspired by:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
package model

import (
	"fmt"
	"strings"
)

// NintendoConfirmButton specifies where the confirm button should be for
// Nintendo systems with diamond-layout face buttons (SNES, GameCube, Switch, etc.).
// This does not affect N64 (unique layout) or non-Nintendo systems.
type NintendoConfirmButton string

const (
	// NintendoConfirmSouth places confirm at the south button position.
	// Use this for consistent muscle memory across all systems.
	NintendoConfirmSouth NintendoConfirmButton = "south"

	// NintendoConfirmEast places confirm at the east button position,
	// matching the original Nintendo controller feel where A is at east.
	// This is the default.
	NintendoConfirmEast NintendoConfirmButton = "east"
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
	NintendoConfirm NintendoConfirmButton
	Hotkeys         HotkeyConfig
}

// DefaultHotkeys returns the default hotkey configuration.
func DefaultHotkeys() HotkeyConfig {
	return HotkeyConfig{
		SaveState:        HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonRightShoulder}},
		LoadState:        HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonLeftShoulder}},
		NextSlot:         HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonDPadRight}},
		PrevSlot:         HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonDPadLeft}},
		FastForward:      HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonY}},
		Rewind:           HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonX}},
		Pause:            HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonA}},
		Screenshot:       HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonB}},
		Quit:             HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonStart}},
		ToggleFullscreen: HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonLeftStick}},
		OpenMenu:         HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonRightStick}},
	}
}

func ValidateNintendoConfirmButton(s string) (NintendoConfirmButton, error) {
	switch NintendoConfirmButton(s) {
	case NintendoConfirmSouth, NintendoConfirmEast:
		return NintendoConfirmButton(s), nil
	default:
		return "", fmt.Errorf("unknown nintendo confirm button %q (valid: %q, %q)", s, NintendoConfirmSouth, NintendoConfirmEast)
	}
}

// SteamDeckGUID is the virtual gamepad GUID that Steam Input presents for all
// controllers connected through Steam. On Steam Deck in game mode, every
// controller (Xbox, PlayStation, etc.) appears with this GUID.
const SteamDeckGUID = "03000000de280000ff11000001000000"

// nintendoDiamondSystems lists Nintendo systems with diamond-layout face buttons
// where the NintendoConfirmButton setting applies. N64 is excluded because its
// controller has a unique layout where A/B aren't in a standard diamond.
var nintendoDiamondSystems = map[SystemID]bool{
	SystemIDNES:      true,
	SystemIDSNES:     true,
	SystemIDGB:       true,
	SystemIDGBC:      true,
	SystemIDGBA:      true,
	SystemIDNDS:      true,
	SystemIDN3DS:     true,
	SystemIDGameCube: true,
	SystemIDWii:      true,
	SystemIDWiiU:     true,
	SystemIDSwitch:   true,
}

// FaceButtonMapping holds the SDL buttons for each physical position.
// South/East/West/North refer to positions on the user's controller.
type FaceButtonMapping struct {
	South SDLButton
	East  SDLButton
	West  SDLButton
	North SDLButton
}

// FaceButtons returns which SDL button is at each physical position for the
// given system. For Nintendo diamond-layout systems, buttons are swapped when
// NintendoConfirmSouth is set so that physical south triggers Nintendo A.
// For all other systems (including N64), positional mapping is used.
func (cc *ControllerConfig) FaceButtons(sys SystemID) FaceButtonMapping {
	if nintendoDiamondSystems[sys] && cc.NintendoConfirm == NintendoConfirmSouth {
		return FaceButtonMapping{
			South: ButtonB, // Nintendo A (originally east) now at south
			East:  ButtonA, // Nintendo B (originally south) now at east
			West:  ButtonY,
			North: ButtonX,
		}
	}
	// Default: standard positional mapping (A=south, B=east, X=west, Y=north)
	return FaceButtonMapping{
		South: ButtonA,
		East:  ButtonB,
		West:  ButtonX,
		North: ButtonY,
	}
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

// SDLIndex returns the SDL button index for the given button.
// This returns the raw index without any layout transformation,
// suitable for hotkeys which are defined in terms of physical buttons.
func (cc *ControllerConfig) SDLIndex(btn SDLButton) int {
	return SDLButtonIndex[btn]
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
