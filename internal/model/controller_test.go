package model

import (
	"testing"
)

func TestParseHotkeyBinding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []SDLButton
		wantErr bool
	}{
		{
			name:  "single button",
			input: "Back",
			want:  []SDLButton{ButtonBack},
		},
		{
			name:  "two-button chord",
			input: "Back+RightShoulder",
			want:  []SDLButton{ButtonBack, ButtonRightShoulder},
		},
		{
			name:  "three-button chord",
			input: "Back+LeftShoulder+A",
			want:  []SDLButton{ButtonBack, ButtonLeftShoulder, ButtonA},
		},
		{
			name:  "whitespace around buttons is trimmed",
			input: " Back + A ",
			want:  []SDLButton{ButtonBack, ButtonA},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "unknown button",
			input:   "Back+Turbo",
			wantErr: true,
		},
		{
			name:    "too many components",
			input:   "A+B+X+Y",
			wantErr: true,
		},
		{
			name:  "all face buttons as chord",
			input: "A+B+X",
			want:  []SDLButton{ButtonA, ButtonB, ButtonX},
		},
		{
			name:  "dpad button",
			input: "DPadUp",
			want:  []SDLButton{ButtonDPadUp},
		},
		{
			name:  "trigger buttons",
			input: "LeftTrigger+RightTrigger",
			want:  []SDLButton{ButtonLeftTrigger, ButtonRightTrigger},
		},
		{
			name:  "stick buttons",
			input: "LeftStick+RightStick",
			want:  []SDLButton{ButtonLeftStick, ButtonRightStick},
		},
		{
			name:    "case sensitive",
			input:   "back",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseHotkeyBinding(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseHotkeyBinding(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if len(got.Buttons) != len(tt.want) {
				t.Fatalf("got %d buttons, want %d", len(got.Buttons), len(tt.want))
			}
			for i := range tt.want {
				if got.Buttons[i] != tt.want[i] {
					t.Errorf("button[%d] = %q, want %q", i, got.Buttons[i], tt.want[i])
				}
			}
		})
	}
}

func TestHotkeyBindingString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		binding HotkeyBinding
		want    string
	}{
		{
			name:    "single button",
			binding: HotkeyBinding{Buttons: []SDLButton{ButtonBack}},
			want:    "Back",
		},
		{
			name:    "two-button chord",
			binding: HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonRightShoulder}},
			want:    "Back+RightShoulder",
		},
		{
			name:    "three-button chord",
			binding: HotkeyBinding{Buttons: []SDLButton{ButtonBack, ButtonA, ButtonB}},
			want:    "Back+A+B",
		},
		{
			name:    "empty binding",
			binding: HotkeyBinding{},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.binding.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHotkeyBindingRoundTrip(t *testing.T) {
	t.Parallel()

	inputs := []string{
		"Back",
		"Back+RightShoulder",
		"Back+LeftShoulder+A",
		"Start+LeftStick",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseHotkeyBinding(input)
			if err != nil {
				t.Fatalf("ParseHotkeyBinding(%q) error = %v", input, err)
			}
			got := parsed.String()
			if got != input {
				t.Errorf("roundtrip: String() = %q, want %q", got, input)
			}
		})
	}
}

func TestValidateNintendoConfirmButton(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    NintendoConfirmButton
		wantErr bool
	}{
		{name: "south", input: "south", want: NintendoConfirmSouth},
		{name: "east", input: "east", want: NintendoConfirmEast},
		{name: "empty string", input: "", wantErr: true},
		{name: "unknown value", input: "north", wantErr: true},
		{name: "case sensitive", input: "South", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ValidateNintendoConfirmButton(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateNintendoConfirmButton(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFaceButtonsNonNintendoSystem(t *testing.T) {
	t.Parallel()

	cc := &ControllerConfig{NintendoConfirm: NintendoConfirmEast}
	// Non-Nintendo system should not be affected by NintendoConfirm setting
	south, east, west, north := cc.FaceButtons(SystemIDPSX)

	if south != ButtonA {
		t.Errorf("PSX south = %q, want %q", south, ButtonA)
	}
	if east != ButtonB {
		t.Errorf("PSX east = %q, want %q", east, ButtonB)
	}
	if west != ButtonX {
		t.Errorf("PSX west = %q, want %q", west, ButtonX)
	}
	if north != ButtonY {
		t.Errorf("PSX north = %q, want %q", north, ButtonY)
	}
}

func TestFaceButtonsNintendoSystemConfirmSouth(t *testing.T) {
	t.Parallel()

	cc := &ControllerConfig{NintendoConfirm: NintendoConfirmSouth}
	south, east, west, north := cc.FaceButtons(SystemIDSNES)

	if south != ButtonA {
		t.Errorf("SNES confirm-south south = %q, want %q", south, ButtonA)
	}
	if east != ButtonB {
		t.Errorf("SNES confirm-south east = %q, want %q", east, ButtonB)
	}
	if west != ButtonX {
		t.Errorf("SNES confirm-south west = %q, want %q", west, ButtonX)
	}
	if north != ButtonY {
		t.Errorf("SNES confirm-south north = %q, want %q", north, ButtonY)
	}
}

func TestFaceButtonsNintendoSystemConfirmEast(t *testing.T) {
	t.Parallel()

	cc := &ControllerConfig{NintendoConfirm: NintendoConfirmEast}
	south, east, west, north := cc.FaceButtons(SystemIDSNES)

	if south != ButtonB {
		t.Errorf("SNES confirm-east south = %q, want %q", south, ButtonB)
	}
	if east != ButtonA {
		t.Errorf("SNES confirm-east east = %q, want %q", east, ButtonA)
	}
	if west != ButtonY {
		t.Errorf("SNES confirm-east west = %q, want %q", west, ButtonY)
	}
	if north != ButtonX {
		t.Errorf("SNES confirm-east north = %q, want %q", north, ButtonX)
	}
}

func TestFaceButtonsN64NotAffected(t *testing.T) {
	t.Parallel()

	cc := &ControllerConfig{NintendoConfirm: NintendoConfirmEast}
	// N64 should not be affected even with NintendoConfirmEast
	south, east, west, north := cc.FaceButtons(SystemIDN64)

	if south != ButtonA {
		t.Errorf("N64 south = %q, want %q", south, ButtonA)
	}
	if east != ButtonB {
		t.Errorf("N64 east = %q, want %q", east, ButtonB)
	}
	if west != ButtonX {
		t.Errorf("N64 west = %q, want %q", west, ButtonX)
	}
	if north != ButtonY {
		t.Errorf("N64 north = %q, want %q", north, ButtonY)
	}
}

func TestSDLIndex(t *testing.T) {
	t.Parallel()

	cc := &ControllerConfig{NintendoConfirm: NintendoConfirmSouth}

	tests := []struct {
		btn  SDLButton
		want int
	}{
		{ButtonA, 0},
		{ButtonB, 1},
		{ButtonX, 2},
		{ButtonY, 3},
		{ButtonBack, 4},
		{ButtonStart, 6},
		{ButtonLeftShoulder, 9},
	}

	for _, tt := range tests {
		got := cc.SDLIndex(tt.btn)
		if got != tt.want {
			t.Errorf("SDLIndex(%q) = %d, want %d", tt.btn, got, tt.want)
		}
	}
}

func TestDefaultHotkeysAllPopulated(t *testing.T) {
	t.Parallel()

	hk := DefaultHotkeys()

	bindings := []struct {
		name    string
		binding HotkeyBinding
	}{
		{"SaveState", hk.SaveState},
		{"LoadState", hk.LoadState},
		{"NextSlot", hk.NextSlot},
		{"PrevSlot", hk.PrevSlot},
		{"FastForward", hk.FastForward},
		{"Rewind", hk.Rewind},
		{"Pause", hk.Pause},
		{"Screenshot", hk.Screenshot},
		{"Quit", hk.Quit},
		{"ToggleFullscreen", hk.ToggleFullscreen},
		{"OpenMenu", hk.OpenMenu},
	}

	for _, b := range bindings {
		if len(b.binding.Buttons) == 0 {
			t.Errorf("DefaultHotkeys().%s has no buttons", b.name)
		}
		if b.binding.String() == "" {
			t.Errorf("DefaultHotkeys().%s.String() is empty", b.name)
		}
	}
}

func TestDefaultHotkeysStringValues(t *testing.T) {
	t.Parallel()

	hk := DefaultHotkeys()

	expected := map[string]string{
		"SaveState":        "Back+RightShoulder",
		"LoadState":        "Back+LeftShoulder",
		"NextSlot":         "Back+DPadRight",
		"PrevSlot":         "Back+DPadLeft",
		"FastForward":      "Back+Y",
		"Rewind":           "Back+X",
		"Pause":            "Back+A",
		"Screenshot":       "Back+B",
		"Quit":             "Back+Start",
		"ToggleFullscreen": "Back+LeftStick",
		"OpenMenu":         "Back+RightStick",
	}

	actual := map[string]string{
		"SaveState":        hk.SaveState.String(),
		"LoadState":        hk.LoadState.String(),
		"NextSlot":         hk.NextSlot.String(),
		"PrevSlot":         hk.PrevSlot.String(),
		"FastForward":      hk.FastForward.String(),
		"Rewind":           hk.Rewind.String(),
		"Pause":            hk.Pause.String(),
		"Screenshot":       hk.Screenshot.String(),
		"Quit":             hk.Quit.String(),
		"ToggleFullscreen": hk.ToggleFullscreen.String(),
		"OpenMenu":         hk.OpenMenu.String(),
	}

	for name, want := range expected {
		got := actual[name]
		if got != want {
			t.Errorf("DefaultHotkeys().%s.String() = %q, want %q", name, got, want)
		}
	}
}

func TestSDLButtonIndexCoversAllStandardButtons(t *testing.T) {
	t.Parallel()

	expectedButtons := []SDLButton{
		ButtonA, ButtonB, ButtonX, ButtonY,
		ButtonBack, ButtonGuide, ButtonStart,
		ButtonLeftStick, ButtonRightStick,
		ButtonLeftShoulder, ButtonRightShoulder,
		ButtonDPadUp, ButtonDPadDown, ButtonDPadLeft, ButtonDPadRight,
	}

	for _, btn := range expectedButtons {
		if _, ok := SDLButtonIndex[btn]; !ok {
			t.Errorf("SDLButtonIndex missing entry for %q", btn)
		}
	}

	if len(SDLButtonIndex) != len(expectedButtons) {
		t.Errorf("SDLButtonIndex has %d entries, expected %d", len(SDLButtonIndex), len(expectedButtons))
	}
}

func TestResolveControllerConfigDefaults(t *testing.T) {
	t.Parallel()

	cfg := &KyarabenConfig{}
	cc, err := cfg.ResolveControllerConfig()
	if err != nil {
		t.Fatalf("ResolveControllerConfig() error = %v", err)
	}

	if cc.NintendoConfirm != NintendoConfirmEast {
		t.Errorf("default NintendoConfirm = %q, want %q", cc.NintendoConfirm, NintendoConfirmEast)
	}

	if cc.Hotkeys.SaveState.String() != "Back+RightShoulder" {
		t.Errorf("default SaveState = %q, want %q", cc.Hotkeys.SaveState.String(), "Back+RightShoulder")
	}
}

func TestResolveControllerConfigWithNintendoConfirm(t *testing.T) {
	t.Parallel()

	cfg := &KyarabenConfig{
		Controller: ControllerTomlConfig{NintendoConfirm: "east"},
	}
	cc, err := cfg.ResolveControllerConfig()
	if err != nil {
		t.Fatalf("ResolveControllerConfig() error = %v", err)
	}

	if cc.NintendoConfirm != NintendoConfirmEast {
		t.Errorf("NintendoConfirm = %q, want %q", cc.NintendoConfirm, NintendoConfirmEast)
	}
}

func TestResolveControllerConfigInvalidNintendoConfirm(t *testing.T) {
	t.Parallel()

	cfg := &KyarabenConfig{
		Controller: ControllerTomlConfig{NintendoConfirm: "north"},
	}
	_, err := cfg.ResolveControllerConfig()
	if err == nil {
		t.Error("expected error for invalid NintendoConfirm, got nil")
	}
}

func TestResolveControllerConfigWithHotkeys(t *testing.T) {
	t.Parallel()

	cfg := &KyarabenConfig{
		Controller: ControllerTomlConfig{
			Hotkeys: HotkeyTomlConfig{
				Modifier:  "Start",
				SaveState: "A",
				Quit:      "B",
			},
		},
	}
	cc, err := cfg.ResolveControllerConfig()
	if err != nil {
		t.Fatalf("ResolveControllerConfig() error = %v", err)
	}

	if cc.Hotkeys.SaveState.String() != "Start+A" {
		t.Errorf("SaveState = %q, want %q", cc.Hotkeys.SaveState.String(), "Start+A")
	}
	if cc.Hotkeys.Quit.String() != "Start+B" {
		t.Errorf("Quit = %q, want %q", cc.Hotkeys.Quit.String(), "Start+B")
	}

	// Unset hotkeys should keep defaults.
	if cc.Hotkeys.LoadState.String() != "Back+LeftShoulder" {
		t.Errorf("LoadState = %q, want default %q", cc.Hotkeys.LoadState.String(), "Back+LeftShoulder")
	}
}

func TestResolveControllerConfigInvalidHotkey(t *testing.T) {
	t.Parallel()

	cfg := &KyarabenConfig{
		Controller: ControllerTomlConfig{
			Hotkeys: HotkeyTomlConfig{
				SaveState: "InvalidButton",
			},
		},
	}
	_, err := cfg.ResolveControllerConfig()
	if err == nil {
		t.Error("expected error for invalid hotkey, got nil")
	}
}

func TestResolveControllerConfigInvalidModifier(t *testing.T) {
	t.Parallel()

	cfg := &KyarabenConfig{
		Controller: ControllerTomlConfig{
			Hotkeys: HotkeyTomlConfig{
				Modifier: "InvalidButton",
			},
		},
	}
	_, err := cfg.ResolveControllerConfig()
	if err == nil {
		t.Error("expected error for invalid modifier, got nil")
	}
}
