package emulators

import (
	"fmt"
	"os"

	"github.com/fnune/kyaraben/internal/model"
)

type Registry struct {
	systems   map[model.SystemID]model.System
	emulators map[model.EmulatorID]model.Emulator
}

func NewRegistry() *Registry {
	r := &Registry{
		systems:   make(map[model.SystemID]model.System),
		emulators: make(map[model.EmulatorID]model.Emulator),
	}
	r.registerSystems()
	r.registerEmulators()
	return r
}

func (r *Registry) registerSystems() {
	r.systems[model.SystemSNES] = model.System{
		ID:          model.SystemSNES,
		Name:        "Super Nintendo",
		Description: "16-bit home console by Nintendo (1990)",
	}
	r.systems[model.SystemPSX] = model.System{
		ID:          model.SystemPSX,
		Name:        "PlayStation",
		Description: "32-bit home console by Sony (1994)",
	}
	r.systems[model.SystemTIC80] = model.System{
		ID:          model.SystemTIC80,
		Name:        "TIC-80",
		Description: "Fantasy console for making and playing tiny games",
	}
	r.systems[model.SystemE2ETest] = model.System{
		ID:          model.SystemE2ETest,
		Name:        "E2E Test",
		Description: "Hidden system for CI testing (uses hello from nixpkgs)",
		Hidden:      true,
	}
}

func (r *Registry) registerEmulators() {
	r.emulators[model.EmulatorRetroArchBsnes] = model.Emulator{
		ID:         model.EmulatorRetroArchBsnes,
		Name:       "RetroArch (bsnes)",
		Systems:    []model.SystemID{model.SystemSNES},
		Source:     model.PackageSourceNixpkgs,
		NixAttr:    "retroarch-bsnes",
		Provisions: []model.Provision{},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		ConfigPaths: []string{
			"~/.config/retroarch/retroarch.cfg",
		},
	}

	r.emulators[model.EmulatorDuckStation] = model.Emulator{
		ID:      model.EmulatorDuckStation,
		Name:    "DuckStation",
		Systems: []model.SystemID{model.SystemPSX},
		Source:  model.PackageSourceNixpkgs,
		NixAttr: "duckstation",
		Provisions: []model.Provision{
			{
				ID:          "psx-bios-usa",
				Kind:        model.ProvisionBIOS,
				Filename:    "scph5501.bin",
				Description: "PlayStation BIOS (USA)",
				Required:    true,
				MD5Hash:     "490f666e1afb15b7362b406ed1cea246",
			},
			{
				ID:          "psx-bios-japan",
				Kind:        model.ProvisionBIOS,
				Filename:    "scph5500.bin",
				Description: "PlayStation BIOS (Japan)",
				Required:    false,
				MD5Hash:     "8dd7d5296a650fac7319bce665a6a53c",
			},
			{
				ID:          "psx-bios-europe",
				Kind:        model.ProvisionBIOS,
				Filename:    "scph5502.bin",
				Description: "PlayStation BIOS (Europe)",
				Required:    false,
				MD5Hash:     "32736f17079d0b2b7024407c39bd3050",
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		ConfigPaths: []string{
			"~/.config/duckstation/settings.ini",
		},
	}

	r.emulators[model.EmulatorTIC80] = model.Emulator{
		ID:         model.EmulatorTIC80,
		Name:       "TIC-80",
		Systems:    []model.SystemID{model.SystemTIC80},
		Source:     model.PackageSourceNixpkgs,
		NixAttr:    "tic-80",
		Provisions: []model.Provision{},
		StateKinds: []model.StateKind{
			model.StateSaves,
		},
		ConfigPaths: []string{},
	}
	r.emulators[model.EmulatorE2ETest] = model.Emulator{
		ID:          model.EmulatorE2ETest,
		Name:        "E2E Test",
		Systems:     []model.SystemID{model.SystemE2ETest},
		Source:      model.PackageSourceNixpkgs,
		NixAttr:     "hello",
		Provisions:  []model.Provision{},
		StateKinds:  []model.StateKind{},
		ConfigPaths: []string{},
	}
}

func (r *Registry) GetSystem(id model.SystemID) (model.System, error) {
	sys, ok := r.systems[id]
	if !ok {
		return model.System{}, fmt.Errorf("unknown system: %s", id)
	}
	return sys, nil
}

func (r *Registry) GetEmulator(id model.EmulatorID) (model.Emulator, error) {
	emu, ok := r.emulators[id]
	if !ok {
		return model.Emulator{}, fmt.Errorf("unknown emulator: %s", id)
	}
	return emu, nil
}

func (r *Registry) GetEmulatorsForSystem(sys model.SystemID) []model.Emulator {
	var result []model.Emulator
	for _, emu := range r.emulators {
		if emu.SupportsSystem(sys) {
			result = append(result, emu)
		}
	}
	return result
}

func (r *Registry) GetDefaultEmulator(sys model.SystemID) (model.Emulator, error) {
	defaults := map[model.SystemID]model.EmulatorID{
		model.SystemSNES:    model.EmulatorRetroArchBsnes,
		model.SystemPSX:     model.EmulatorDuckStation,
		model.SystemTIC80:   model.EmulatorTIC80,
		model.SystemE2ETest: model.EmulatorE2ETest,
	}

	emuID, ok := defaults[sys]
	if !ok {
		return model.Emulator{}, fmt.Errorf("no default emulator for system: %s", sys)
	}
	return r.GetEmulator(emuID)
}

func (r *Registry) AllSystems() []model.System {
	showHidden := os.Getenv("KYARABEN_SHOW_HIDDEN") == "1"
	result := make([]model.System, 0, len(r.systems))
	for _, sys := range r.systems {
		if !sys.Hidden || showHidden {
			result = append(result, sys)
		}
	}
	return result
}

func (r *Registry) AllEmulators() []model.Emulator {
	result := make([]model.Emulator, 0, len(r.emulators))
	for _, emu := range r.emulators {
		result = append(result, emu)
	}
	return result
}

func (r *Registry) GetConfigGenerator(emuID model.EmulatorID) ConfigGenerator {
	return GetConfigGenerator(emuID)
}
