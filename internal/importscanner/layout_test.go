package importscanner

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestNormalizeSystem(t *testing.T) {
	tests := []struct {
		input    string
		expected model.SystemID
	}{
		{"gc", model.SystemIDGameCube},
		{"gamecube", model.SystemIDGameCube},
		{"GameCube", model.SystemIDGameCube},
		{"ngc", model.SystemIDGameCube},
		{"wii", model.SystemIDWii},
		{"Wii", model.SystemIDWii},
		{"ps1", model.SystemIDPSX},
		{"psx", model.SystemIDPSX},
		{"playstation", model.SystemIDPSX},
		{"snes", model.SystemIDSNES},
		{"superfamicom", model.SystemIDSNES},
		{"sfc", model.SystemIDSNES},
		{"gba", model.SystemIDGBA},
		{"gameboyadvance", model.SystemIDGBA},
		{"nds", model.SystemIDNDS},
		{"ds", model.SystemIDNDS},
		{"n64", model.SystemIDN64},
		{"genesis", model.SystemIDGenesis},
		{"megadrive", model.SystemIDGenesis},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeSystem(tt.input)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if *result != tt.expected {
				t.Errorf("NormalizeSystem(%q) = %s, want %s", tt.input, *result, tt.expected)
			}
		})
	}
}

func TestNormalizeEmulator(t *testing.T) {
	tests := []struct {
		input    string
		expected model.EmulatorID
	}{
		{"dolphin", model.EmulatorIDDolphin},
		{"dolphin-emu", model.EmulatorIDDolphin},
		{"duckstation", model.EmulatorIDDuckStation},
		{"pcsx2", model.EmulatorIDPCSX2},
		{"retroarch", model.EmulatorIDRetroArch},
		{"mgba", model.EmulatorIDRetroArchMGBA},
		{"bsnes", model.EmulatorIDRetroArchBsnes},
		{"ppsspp", model.EmulatorIDPPSSPP},
		{"flycast", model.EmulatorIDFlycast},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeEmulator(tt.input)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if *result != tt.expected {
				t.Errorf("NormalizeEmulator(%q) = %s, want %s", tt.input, *result, tt.expected)
			}
		})
	}
}

func TestGenericLayoutClassify(t *testing.T) {
	layout := &GenericLayout{}

	tests := []struct {
		path     string
		dataType DataType
		isSystem bool
	}{
		{"roms/gamecube/game.iso", DataTypeROMs, true},
		{"bios/psx/scph1001.bin", DataTypeBIOS, true},
		{"saves/snes/game.srm", DataTypeSaves, true},
		{"states/dolphin/state.sav", DataTypeStates, false},
		{"screenshots/dolphin/screen.png", DataTypeScreenshots, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := layout.Classify(tt.path)
			if result == nil {
				t.Fatal("expected classification, got nil")
			}
			if result.DataType != tt.dataType {
				t.Errorf("DataType = %s, want %s", result.DataType, tt.dataType)
			}
			if tt.isSystem {
				if result.System == nil {
					t.Error("expected System to be set")
				}
				if result.Emulator != nil {
					t.Error("expected Emulator to be nil for system data")
				}
			} else {
				if result.Emulator == nil {
					t.Error("expected Emulator to be set")
				}
			}
		})
	}
}

func TestGenericLayoutClassifyUnknown(t *testing.T) {
	layout := &GenericLayout{}

	tests := []string{
		"unknown/file.txt",
		"single.txt",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			result := layout.Classify(path)
			if result != nil {
				t.Errorf("expected nil for unknown path %q, got %+v", path, result)
			}
		})
	}
}
