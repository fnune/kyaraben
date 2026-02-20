package registry

import (
	"testing"

	"github.com/fnune/kyaraben/internal/emulators/cemu"
	"github.com/fnune/kyaraben/internal/emulators/dolphin"
	"github.com/fnune/kyaraben/internal/emulators/duckstation"
	"github.com/fnune/kyaraben/internal/emulators/eden"
	"github.com/fnune/kyaraben/internal/emulators/flycast"
	"github.com/fnune/kyaraben/internal/emulators/pcsx2"
	"github.com/fnune/kyaraben/internal/emulators/ppsspp"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbeetlengp"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbeetlepce"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbeetlesaturn"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbsnes"
	"github.com/fnune/kyaraben/internal/emulators/retroarchcitra"
	"github.com/fnune/kyaraben/internal/emulators/retroarchgenesisplusgx"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmelonds"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmesen"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmgba"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmupen64plus"
	"github.com/fnune/kyaraben/internal/emulators/rpcs3"
	"github.com/fnune/kyaraben/internal/emulators/vita3k"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/systems/dreamcast"
	"github.com/fnune/kyaraben/internal/systems/gamecube"
	"github.com/fnune/kyaraben/internal/systems/gamegear"
	"github.com/fnune/kyaraben/internal/systems/gb"
	"github.com/fnune/kyaraben/internal/systems/gba"
	"github.com/fnune/kyaraben/internal/systems/gbc"
	"github.com/fnune/kyaraben/internal/systems/genesis"
	"github.com/fnune/kyaraben/internal/systems/mastersystem"
	"github.com/fnune/kyaraben/internal/systems/n64"
	"github.com/fnune/kyaraben/internal/systems/nds"
	"github.com/fnune/kyaraben/internal/systems/nes"
	"github.com/fnune/kyaraben/internal/systems/ngp"
	n3ds "github.com/fnune/kyaraben/internal/systems/nintendo3ds"
	"github.com/fnune/kyaraben/internal/systems/pcengine"
	"github.com/fnune/kyaraben/internal/systems/ps2"
	"github.com/fnune/kyaraben/internal/systems/ps3"
	"github.com/fnune/kyaraben/internal/systems/psp"
	"github.com/fnune/kyaraben/internal/systems/psvita"
	"github.com/fnune/kyaraben/internal/systems/psx"
	"github.com/fnune/kyaraben/internal/systems/saturn"
	"github.com/fnune/kyaraben/internal/systems/snes"
	switchsys "github.com/fnune/kyaraben/internal/systems/switch"
	"github.com/fnune/kyaraben/internal/systems/wii"
	"github.com/fnune/kyaraben/internal/systems/wiiu"
)

func TestAllDefinitions(t *testing.T) {
	systemDefs := []model.SystemDefinition{
		nes.Definition{},
		snes.Definition{},
		n64.Definition{},
		gb.Definition{},
		gbc.Definition{},
		gba.Definition{},
		nds.Definition{},
		n3ds.Definition{},
		gamecube.Definition{},
		wii.Definition{},
		wiiu.Definition{},
		switchsys.Definition{},
		psx.Definition{},
		ps2.Definition{},
		ps3.Definition{},
		psp.Definition{},
		psvita.Definition{},
		genesis.Definition{},
		mastersystem.Definition{},
		gamegear.Definition{},
		saturn.Definition{},
		dreamcast.Definition{},
		pcengine.Definition{},
		ngp.Definition{},
	}

	emulatorDefs := []model.EmulatorDefinition{
		retroarchbsnes.Definition{},
		retroarchmesen.Definition{},
		retroarchgenesisplusgx.Definition{},
		retroarchmupen64plus.Definition{},
		retroarchbeetlesaturn.Definition{},
		retroarchbeetlepce.Definition{},
		retroarchbeetlengp.Definition{},
		retroarchmgba.Definition{},
		retroarchmelonds.Definition{},
		duckstation.Definition{},
		pcsx2.Definition{},
		rpcs3.Definition{},
		vita3k.Definition{},
		ppsspp.Definition{},
		flycast.Definition{},
		cemu.Definition{},
		retroarchcitra.Definition{},
		dolphin.Definition{},
		eden.Definition{},
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
		if emu.Launcher.RomCommand == nil {
			t.Errorf("emulator %q: Launcher.RomCommand is nil", emu.ID)
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
		{model.EmulatorIDRetroArchBsnes, false},
		{model.EmulatorIDRetroArchMesen, false},
		{model.EmulatorIDRetroArchGenesisPlusGX, false},
		{model.EmulatorIDRetroArchMupen64Plus, false},
		{model.EmulatorIDRetroArchBeetleSaturn, false},
		{model.EmulatorIDDuckStation, false},
		{model.EmulatorIDPCSX2, false},
		{model.EmulatorIDRPCS3, false},
		{model.EmulatorIDVita3K, false},
		{model.EmulatorIDPPSSPP, false},
		{model.EmulatorIDRetroArchMGBA, false},
		{model.EmulatorIDRetroArchMelonDS, false},
		{model.EmulatorIDFlycast, false},
		{model.EmulatorIDCemu, false},
		{model.EmulatorIDRetroArchCitra, false},
		{model.EmulatorIDDolphin, false},
		{model.EmulatorIDEden, false},
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
		{model.SystemIDNES, 1, []model.EmulatorID{model.EmulatorIDRetroArchMesen}},
		{model.SystemIDSNES, 1, []model.EmulatorID{model.EmulatorIDRetroArchBsnes}},
		{model.SystemIDN64, 1, []model.EmulatorID{model.EmulatorIDRetroArchMupen64Plus}},
		{model.SystemIDGB, 1, []model.EmulatorID{model.EmulatorIDRetroArchMGBA}},
		{model.SystemIDGBC, 1, []model.EmulatorID{model.EmulatorIDRetroArchMGBA}},
		{model.SystemIDGBA, 1, []model.EmulatorID{model.EmulatorIDRetroArchMGBA}},
		{model.SystemIDNDS, 1, []model.EmulatorID{model.EmulatorIDRetroArchMelonDS}},
		{model.SystemIDN3DS, 1, []model.EmulatorID{model.EmulatorIDRetroArchCitra}},
		{model.SystemIDGameCube, 1, []model.EmulatorID{model.EmulatorIDDolphin}},
		{model.SystemIDWii, 1, []model.EmulatorID{model.EmulatorIDDolphin}},
		{model.SystemIDWiiU, 1, []model.EmulatorID{model.EmulatorIDCemu}},
		{model.SystemIDSwitch, 1, []model.EmulatorID{model.EmulatorIDEden}},
		{model.SystemIDPSX, 1, []model.EmulatorID{model.EmulatorIDDuckStation}},
		{model.SystemIDPS2, 1, []model.EmulatorID{model.EmulatorIDPCSX2}},
		{model.SystemIDPS3, 1, []model.EmulatorID{model.EmulatorIDRPCS3}},
		{model.SystemIDPSP, 1, []model.EmulatorID{model.EmulatorIDPPSSPP}},
		{model.SystemIDPSVita, 1, []model.EmulatorID{model.EmulatorIDVita3K}},
		{model.SystemIDGenesis, 1, []model.EmulatorID{model.EmulatorIDRetroArchGenesisPlusGX}},
		{model.SystemIDSaturn, 1, []model.EmulatorID{model.EmulatorIDRetroArchBeetleSaturn}},
		{model.SystemIDDreamcast, 1, []model.EmulatorID{model.EmulatorIDFlycast}},
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
		{model.SystemIDNES, model.EmulatorIDRetroArchMesen, false},
		{model.SystemIDSNES, model.EmulatorIDRetroArchBsnes, false},
		{model.SystemIDN64, model.EmulatorIDRetroArchMupen64Plus, false},
		{model.SystemIDGB, model.EmulatorIDRetroArchMGBA, false},
		{model.SystemIDGBC, model.EmulatorIDRetroArchMGBA, false},
		{model.SystemIDGBA, model.EmulatorIDRetroArchMGBA, false},
		{model.SystemIDNDS, model.EmulatorIDRetroArchMelonDS, false},
		{model.SystemIDN3DS, model.EmulatorIDRetroArchCitra, false},
		{model.SystemIDGameCube, model.EmulatorIDDolphin, false},
		{model.SystemIDWii, model.EmulatorIDDolphin, false},
		{model.SystemIDWiiU, model.EmulatorIDCemu, false},
		{model.SystemIDSwitch, model.EmulatorIDEden, false},
		{model.SystemIDPSX, model.EmulatorIDDuckStation, false},
		{model.SystemIDPS2, model.EmulatorIDPCSX2, false},
		{model.SystemIDPS3, model.EmulatorIDRPCS3, false},
		{model.SystemIDPSP, model.EmulatorIDPPSSPP, false},
		{model.SystemIDPSVita, model.EmulatorIDVita3K, false},
		{model.SystemIDGenesis, model.EmulatorIDRetroArchGenesisPlusGX, false},
		{model.SystemIDSaturn, model.EmulatorIDRetroArchBeetleSaturn, false},
		{model.SystemIDDreamcast, model.EmulatorIDFlycast, false},
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

	emu, _ := reg.GetEmulator(model.EmulatorIDRetroArchBsnes)

	if !emu.SupportsSystem(model.SystemIDSNES) {
		t.Error("RetroArch bsnes should support SNES")
	}

	if emu.SupportsSystem(model.SystemIDPSX) {
		t.Error("RetroArch bsnes should not support PSX")
	}
}

func TestAllSystems(t *testing.T) {
	reg := NewDefault()

	systems := reg.AllSystems()
	if len(systems) < 20 {
		t.Errorf("Expected at least 20 systems, got %d", len(systems))
	}

	expected := []model.SystemID{
		model.SystemIDNES,
		model.SystemIDSNES,
		model.SystemIDN64,
		model.SystemIDGB,
		model.SystemIDGBC,
		model.SystemIDGBA,
		model.SystemIDNDS,
		model.SystemIDN3DS,
		model.SystemIDGameCube,
		model.SystemIDWii,
		model.SystemIDWiiU,
		model.SystemIDSwitch,
		model.SystemIDPSX,
		model.SystemIDPS2,
		model.SystemIDPS3,
		model.SystemIDPSP,
		model.SystemIDPSVita,
		model.SystemIDGenesis,
		model.SystemIDSaturn,
		model.SystemIDDreamcast,
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
		{model.EmulatorIDRetroArchBsnes, true},
		{model.EmulatorIDRetroArchMesen, true},
		{model.EmulatorIDRetroArchGenesisPlusGX, true},
		{model.EmulatorIDRetroArchMupen64Plus, true},
		{model.EmulatorIDRetroArchBeetleSaturn, true},
		{model.EmulatorIDDuckStation, true},
		{model.EmulatorIDPCSX2, true},
		{model.EmulatorIDRPCS3, true},
		{model.EmulatorIDVita3K, true},
		{model.EmulatorIDPPSSPP, true},
		{model.EmulatorIDRetroArchMGBA, true},
		{model.EmulatorIDRetroArchMelonDS, true},
		{model.EmulatorIDFlycast, true},
		{model.EmulatorIDCemu, true},
		{model.EmulatorIDRetroArchCitra, true},
		{model.EmulatorIDDolphin, true},
		{model.EmulatorIDEden, true},
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
