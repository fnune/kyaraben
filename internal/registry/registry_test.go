package registry

import (
	"testing"

	"github.com/fnune/kyaraben/internal/emulators/duckstation"
	"github.com/fnune/kyaraben/internal/emulators/e2etestemu"
	"github.com/fnune/kyaraben/internal/emulators/eden"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbsnes"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmelonds"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmgba"
	"github.com/fnune/kyaraben/internal/emulators/retroarchppsspp"
	"github.com/fnune/kyaraben/internal/emulators/tic80emu"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/systems/e2etest"
	"github.com/fnune/kyaraben/internal/systems/gba"
	"github.com/fnune/kyaraben/internal/systems/nds"
	"github.com/fnune/kyaraben/internal/systems/psp"
	"github.com/fnune/kyaraben/internal/systems/psx"
	"github.com/fnune/kyaraben/internal/systems/snes"
	switchsys "github.com/fnune/kyaraben/internal/systems/switch"
	"github.com/fnune/kyaraben/internal/systems/tic80"
)

func TestAllDefinitions(t *testing.T) {
	systemDefs := []model.SystemDefinition{
		snes.Definition{},
		psx.Definition{},
		tic80.Definition{},
		gba.Definition{},
		nds.Definition{},
		psp.Definition{},
		switchsys.Definition{},
		e2etest.Definition{},
	}

	emulatorDefs := []model.EmulatorDefinition{
		retroarchbsnes.Definition{},
		retroarchmgba.Definition{},
		retroarchmelonds.Definition{},
		retroarchppsspp.Definition{},
		duckstation.Definition{},
		tic80emu.Definition{},
		eden.Definition{},
		e2etestemu.Definition{},
	}

	systems := make(map[model.SystemID]model.System)
	emulators := make(map[model.EmulatorID]model.Emulator)

	for _, def := range systemDefs {
		sys := def.System()
		systems[sys.ID] = sys

		if sys.Name == "" {
			t.Errorf("system %q: Name is empty", sys.ID)
		}
		if sys.Description == "" {
			t.Errorf("system %q: Description is empty", sys.ID)
		}
	}

	for _, def := range emulatorDefs {
		emu := def.Emulator()
		emulators[emu.ID] = emu

		if emu.Name == "" {
			t.Errorf("emulator %q: Name is empty", emu.ID)
		}
		if len(emu.Systems) == 0 {
			t.Errorf("emulator %q: Systems is empty", emu.ID)
		}
		if def.ConfigGenerator() == nil {
			t.Errorf("emulator %q: ConfigGenerator is nil", emu.ID)
		}
	}

	for _, def := range systemDefs {
		sys := def.System()
		defaultEmuID := def.DefaultEmulatorID()

		emu, ok := emulators[defaultEmuID]
		if !ok {
			t.Errorf("system %q: default emulator %q not registered", sys.ID, defaultEmuID)
			continue
		}

		if !emu.SupportsSystem(sys.ID) {
			t.Errorf("system %q: default emulator %q does not support it", sys.ID, defaultEmuID)
		}
	}

	for _, def := range emulatorDefs {
		emu := def.Emulator()
		for _, sysID := range emu.Systems {
			if _, ok := systems[sysID]; !ok {
				t.Errorf("emulator %q: references unknown system %q", emu.ID, sysID)
			}
		}
	}
}

func TestRegistryGetSystem(t *testing.T) {
	reg := NewDefault()

	tests := []struct {
		id      model.SystemID
		wantErr bool
	}{
		{model.SystemSNES, false},
		{model.SystemPSX, false},
		{model.SystemTIC80, false},
		{model.SystemGBA, false},
		{model.SystemNDS, false},
		{model.SystemPSP, false},
		{model.SystemSwitch, false},
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
	reg := NewDefault()

	tests := []struct {
		id      model.EmulatorID
		wantErr bool
	}{
		{model.EmulatorRetroArchBsnes, false},
		{model.EmulatorRetroArchMGBA, false},
		{model.EmulatorRetroArchMelonDS, false},
		{model.EmulatorRetroArchPPSSPP, false},
		{model.EmulatorDuckStation, false},
		{model.EmulatorTIC80, false},
		{model.EmulatorEden, false},
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
	reg := NewDefault()

	tests := []struct {
		system  model.SystemID
		wantIDs []model.EmulatorID
	}{
		{model.SystemSNES, []model.EmulatorID{model.EmulatorRetroArchBsnes}},
		{model.SystemPSX, []model.EmulatorID{model.EmulatorDuckStation}},
		{model.SystemTIC80, []model.EmulatorID{model.EmulatorTIC80}},
		{model.SystemGBA, []model.EmulatorID{model.EmulatorRetroArchMGBA}},
		{model.SystemNDS, []model.EmulatorID{model.EmulatorRetroArchMelonDS}},
		{model.SystemPSP, []model.EmulatorID{model.EmulatorRetroArchPPSSPP}},
		{model.SystemSwitch, []model.EmulatorID{model.EmulatorEden}},
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
	reg := NewDefault()

	tests := []struct {
		system  model.SystemID
		wantID  model.EmulatorID
		wantErr bool
	}{
		{model.SystemSNES, model.EmulatorRetroArchBsnes, false},
		{model.SystemPSX, model.EmulatorDuckStation, false},
		{model.SystemTIC80, model.EmulatorTIC80, false},
		{model.SystemGBA, model.EmulatorRetroArchMGBA, false},
		{model.SystemNDS, model.EmulatorRetroArchMelonDS, false},
		{model.SystemPSP, model.EmulatorRetroArchPPSSPP, false},
		{model.SystemSwitch, model.EmulatorEden, false},
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
	reg := NewDefault()

	emu, _ := reg.GetEmulator(model.EmulatorRetroArchBsnes)

	if !emu.SupportsSystem(model.SystemSNES) {
		t.Error("RetroArch bsnes should support SNES")
	}

	if emu.SupportsSystem(model.SystemPSX) {
		t.Error("RetroArch bsnes should not support PSX")
	}
}

func TestAllSystems(t *testing.T) {
	reg := NewDefault()

	systems := reg.AllSystems()
	if len(systems) < 7 {
		t.Errorf("Expected at least 7 systems, got %d", len(systems))
	}

	expected := []model.SystemID{
		model.SystemSNES,
		model.SystemPSX,
		model.SystemTIC80,
		model.SystemGBA,
		model.SystemNDS,
		model.SystemPSP,
		model.SystemSwitch,
	}

	found := make(map[model.SystemID]bool)
	for _, sys := range systems {
		found[sys.ID] = true
	}

	for _, id := range expected {
		if !found[id] {
			t.Errorf("%s not found in AllSystems", id)
		}
	}
}

func TestGetConfigGenerator(t *testing.T) {
	reg := NewDefault()

	tests := []struct {
		emuID    model.EmulatorID
		expected bool
	}{
		{model.EmulatorRetroArchBsnes, true},
		{model.EmulatorRetroArchMGBA, true},
		{model.EmulatorRetroArchMelonDS, true},
		{model.EmulatorRetroArchPPSSPP, true},
		{model.EmulatorDuckStation, true},
		{model.EmulatorTIC80, true},
		{model.EmulatorEden, true},
		{model.EmulatorID("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.emuID), func(t *testing.T) {
			gen := reg.GetConfigGenerator(tt.emuID)
			if tt.expected && gen == nil {
				t.Errorf("expected generator for %s, got nil", tt.emuID)
			}
			if !tt.expected && gen != nil {
				t.Errorf("expected nil for %s, got generator", tt.emuID)
			}
		})
	}
}
