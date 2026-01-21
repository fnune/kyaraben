package emulators

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestRegistryGetSystem(t *testing.T) {
	reg := NewRegistry()

	tests := []struct {
		id      model.SystemID
		wantErr bool
	}{
		{model.SystemSNES, false},
		{model.SystemPSX, false},
		{model.SystemTIC80, false},
		{model.SystemID("unknown"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			sys, err := reg.GetSystem(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSystem(%s) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}
			if !tt.wantErr && sys.ID != tt.id {
				t.Errorf("GetSystem(%s) returned wrong system: got %s", tt.id, sys.ID)
			}
		})
	}
}

func TestRegistryGetEmulator(t *testing.T) {
	reg := NewRegistry()

	tests := []struct {
		id      model.EmulatorID
		wantErr bool
	}{
		{model.EmulatorRetroArchBsnes, false},
		{model.EmulatorDuckStation, false},
		{model.EmulatorTIC80, false},
		{model.EmulatorID("unknown"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			emu, err := reg.GetEmulator(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEmulator(%s) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}
			if !tt.wantErr && emu.ID != tt.id {
				t.Errorf("GetEmulator(%s) returned wrong emulator: got %s", tt.id, emu.ID)
			}
		})
	}
}

func TestRegistryGetEmulatorsForSystem(t *testing.T) {
	reg := NewRegistry()

	tests := []struct {
		system  model.SystemID
		wantIDs []model.EmulatorID
	}{
		{model.SystemSNES, []model.EmulatorID{model.EmulatorRetroArchBsnes}},
		{model.SystemPSX, []model.EmulatorID{model.EmulatorDuckStation}},
		{model.SystemTIC80, []model.EmulatorID{model.EmulatorTIC80}},
	}

	for _, tt := range tests {
		t.Run(string(tt.system), func(t *testing.T) {
			emulators := reg.GetEmulatorsForSystem(tt.system)

			if len(emulators) != len(tt.wantIDs) {
				t.Errorf("Got %d emulators, want %d", len(emulators), len(tt.wantIDs))
				return
			}

			for i, emu := range emulators {
				if emu.ID != tt.wantIDs[i] {
					t.Errorf("Emulator %d: got %s, want %s", i, emu.ID, tt.wantIDs[i])
				}
			}
		})
	}
}

func TestRegistryGetDefaultEmulator(t *testing.T) {
	reg := NewRegistry()

	tests := []struct {
		system  model.SystemID
		wantID  model.EmulatorID
		wantErr bool
	}{
		{model.SystemSNES, model.EmulatorRetroArchBsnes, false},
		{model.SystemPSX, model.EmulatorDuckStation, false},
		{model.SystemTIC80, model.EmulatorTIC80, false},
		{model.SystemID("unknown"), "", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.system), func(t *testing.T) {
			emu, err := reg.GetDefaultEmulator(tt.system)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDefaultEmulator(%s) error = %v, wantErr %v", tt.system, err, tt.wantErr)
				return
			}
			if !tt.wantErr && emu.ID != tt.wantID {
				t.Errorf("GetDefaultEmulator(%s) = %s, want %s", tt.system, emu.ID, tt.wantID)
			}
		})
	}
}

func TestEmulatorSupportsSystem(t *testing.T) {
	reg := NewRegistry()

	emu, _ := reg.GetEmulator(model.EmulatorRetroArchBsnes)

	if !emu.SupportsSystem(model.SystemSNES) {
		t.Error("RetroArch bsnes should support SNES")
	}

	if emu.SupportsSystem(model.SystemPSX) {
		t.Error("RetroArch bsnes should not support PSX")
	}
}

func TestAllSystems(t *testing.T) {
	reg := NewRegistry()

	systems := reg.AllSystems()
	if len(systems) < 3 {
		t.Errorf("Expected at least 3 systems, got %d", len(systems))
	}

	// Check for known systems
	foundSNES := false
	foundPSX := false
	foundTIC80 := false

	for _, sys := range systems {
		switch sys.ID {
		case model.SystemSNES:
			foundSNES = true
		case model.SystemPSX:
			foundPSX = true
		case model.SystemTIC80:
			foundTIC80 = true
		}
	}

	if !foundSNES {
		t.Error("SNES not found in AllSystems")
	}
	if !foundPSX {
		t.Error("PSX not found in AllSystems")
	}
	if !foundTIC80 {
		t.Error("TIC80 not found in AllSystems")
	}
}
