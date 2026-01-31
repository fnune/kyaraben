package registry

import (
	"testing"

	"github.com/fnune/kyaraben/internal/emulators/azahar"
	"github.com/fnune/kyaraben/internal/emulators/cemu"
	"github.com/fnune/kyaraben/internal/emulators/dolphin"
	"github.com/fnune/kyaraben/internal/emulators/duckstation"
	"github.com/fnune/kyaraben/internal/emulators/e2etestemu"
	"github.com/fnune/kyaraben/internal/emulators/eden"
	"github.com/fnune/kyaraben/internal/emulators/flycast"
	"github.com/fnune/kyaraben/internal/emulators/melonds"
	"github.com/fnune/kyaraben/internal/emulators/mgba"
	"github.com/fnune/kyaraben/internal/emulators/pcsx2"
	"github.com/fnune/kyaraben/internal/emulators/ppsspp"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbsnes"
	"github.com/fnune/kyaraben/internal/emulators/rpcs3"
	"github.com/fnune/kyaraben/internal/emulators/tic80emu"
	"github.com/fnune/kyaraben/internal/emulators/vita3k"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/systems/dreamcast"
	"github.com/fnune/kyaraben/internal/systems/e2etest"
	"github.com/fnune/kyaraben/internal/systems/gamecube"
	"github.com/fnune/kyaraben/internal/systems/gba"
	"github.com/fnune/kyaraben/internal/systems/nds"
	n3ds "github.com/fnune/kyaraben/internal/systems/nintendo3ds"
	"github.com/fnune/kyaraben/internal/systems/ps2"
	"github.com/fnune/kyaraben/internal/systems/ps3"
	"github.com/fnune/kyaraben/internal/systems/psp"
	"github.com/fnune/kyaraben/internal/systems/psvita"
	"github.com/fnune/kyaraben/internal/systems/psx"
	"github.com/fnune/kyaraben/internal/systems/snes"
	switchsys "github.com/fnune/kyaraben/internal/systems/switch"
	"github.com/fnune/kyaraben/internal/systems/tic80"
	"github.com/fnune/kyaraben/internal/systems/wii"
	"github.com/fnune/kyaraben/internal/systems/wiiu"
)

func TestAllDefinitions(t *testing.T) {
	systemDefs := []model.SystemDefinition{
		snes.Definition{},
		psx.Definition{},
		ps2.Definition{},
		ps3.Definition{},
		psvita.Definition{},
		dreamcast.Definition{},
		tic80.Definition{},
		gba.Definition{},
		nds.Definition{},
		psp.Definition{},
		gamecube.Definition{},
		wii.Definition{},
		wiiu.Definition{},
		n3ds.Definition{},
		switchsys.Definition{},
		e2etest.Definition{},
	}

	emulatorDefs := []model.EmulatorDefinition{
		retroarchbsnes.Definition{},
		duckstation.Definition{},
		pcsx2.Definition{},
		rpcs3.Definition{},
		vita3k.Definition{},
		ppsspp.Definition{},
		mgba.Definition{},
		melonds.Definition{},
		flycast.Definition{},
		cemu.Definition{},
		azahar.Definition{},
		dolphin.Definition{},
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

	// Test all registered systems can be retrieved
	for _, sys := range reg.AllSystems() {
		t.Run(string(sys.ID), func(t *testing.T) {
			got, err := reg.GetSystem(sys.ID)
			if err != nil {
				t.Errorf("GetSystem(%s) error = %v", sys.ID, err)
				return
			}
			if got.ID != sys.ID {
				t.Errorf("GetSystem(%s) returned wrong system: got %s", sys.ID, got.ID)
			}
		})
	}

	// Test unknown system returns error
	t.Run("unknown", func(t *testing.T) {
		_, err := reg.GetSystem(model.SystemID("unknown"))
		if err == nil {
			t.Error("GetSystem(unknown) should return error")
		}
	})
}

func TestRegistryGetEmulator(t *testing.T) {
	reg := NewDefault()

	tests := []struct {
		id      model.EmulatorID
		wantErr bool
	}{
		{model.EmulatorRetroArchBsnes, false},
		{model.EmulatorDuckStation, false},
		{model.EmulatorPCSX2, false},
		{model.EmulatorRPCS3, false},
		{model.EmulatorVita3K, false},
		{model.EmulatorPPSSPP, false},
		{model.EmulatorMGBA, false},
		{model.EmulatorMelonDS, false},
		{model.EmulatorFlycast, false},
		{model.EmulatorCemu, false},
		{model.EmulatorAzahar, false},
		{model.EmulatorDolphin, false},
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
		wantLen int
		wantAny []model.EmulatorID
	}{
		{model.SystemSNES, 1, []model.EmulatorID{model.EmulatorRetroArchBsnes}},
		{model.SystemPSX, 1, []model.EmulatorID{model.EmulatorDuckStation}},
		{model.SystemPS2, 1, []model.EmulatorID{model.EmulatorPCSX2}},
		{model.SystemPS3, 1, []model.EmulatorID{model.EmulatorRPCS3}},
		{model.SystemPSVita, 1, []model.EmulatorID{model.EmulatorVita3K}},
		{model.SystemDreamcast, 1, []model.EmulatorID{model.EmulatorFlycast}},
		{model.SystemTIC80, 1, []model.EmulatorID{model.EmulatorTIC80}},
		{model.SystemGBA, 1, []model.EmulatorID{model.EmulatorMGBA}},
		{model.SystemNDS, 1, []model.EmulatorID{model.EmulatorMelonDS}},
		{model.SystemPSP, 1, []model.EmulatorID{model.EmulatorPPSSPP}},
		{model.SystemGameCube, 1, []model.EmulatorID{model.EmulatorDolphin}},
		{model.SystemWii, 1, []model.EmulatorID{model.EmulatorDolphin}},
		{model.SystemWiiU, 1, []model.EmulatorID{model.EmulatorCemu}},
		{model.System3DS, 1, []model.EmulatorID{model.EmulatorAzahar}},
		{model.SystemSwitch, 1, []model.EmulatorID{model.EmulatorEden}},
	}

	for _, tt := range tests {
		t.Run(string(tt.system), func(t *testing.T) {
			emulators := reg.GetEmulatorsForSystem(tt.system)

			if len(emulators) != tt.wantLen {
				t.Errorf("Got %d emulators, want %d", len(emulators), tt.wantLen)
				return
			}

			found := make(map[model.EmulatorID]bool)
			for _, emu := range emulators {
				found[emu.ID] = true
			}
			for _, wantID := range tt.wantAny {
				if !found[wantID] {
					t.Errorf("Expected emulator %s not found", wantID)
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
		{model.SystemPS2, model.EmulatorPCSX2, false},
		{model.SystemPS3, model.EmulatorRPCS3, false},
		{model.SystemPSVita, model.EmulatorVita3K, false},
		{model.SystemDreamcast, model.EmulatorFlycast, false},
		{model.SystemTIC80, model.EmulatorTIC80, false},
		{model.SystemGBA, model.EmulatorMGBA, false},
		{model.SystemNDS, model.EmulatorMelonDS, false},
		{model.SystemPSP, model.EmulatorPPSSPP, false},
		{model.SystemGameCube, model.EmulatorDolphin, false},
		{model.SystemWii, model.EmulatorDolphin, false},
		{model.SystemWiiU, model.EmulatorCemu, false},
		{model.System3DS, model.EmulatorAzahar, false},
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
	if len(systems) < 15 {
		t.Errorf("Expected at least 15 systems, got %d", len(systems))
	}

	expected := []model.SystemID{
		model.SystemSNES,
		model.SystemPSX,
		model.SystemPS2,
		model.SystemPS3,
		model.SystemPSVita,
		model.SystemDreamcast,
		model.SystemTIC80,
		model.SystemGBA,
		model.SystemNDS,
		model.SystemPSP,
		model.SystemGameCube,
		model.SystemWii,
		model.SystemWiiU,
		model.System3DS,
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
		{model.EmulatorDuckStation, true},
		{model.EmulatorPCSX2, true},
		{model.EmulatorRPCS3, true},
		{model.EmulatorVita3K, true},
		{model.EmulatorPPSSPP, true},
		{model.EmulatorMGBA, true},
		{model.EmulatorMelonDS, true},
		{model.EmulatorFlycast, true},
		{model.EmulatorCemu, true},
		{model.EmulatorAzahar, true},
		{model.EmulatorDolphin, true},
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
