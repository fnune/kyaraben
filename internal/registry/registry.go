package registry

import (
	"fmt"
	"os"

	"github.com/fnune/kyaraben/internal/model"
)

type Registry struct {
	systems          map[model.SystemID]model.System
	emulators        map[model.EmulatorID]emulatorEntry
	defaultEmulators map[model.SystemID]model.EmulatorID
}

type emulatorEntry struct {
	model.Emulator
	configGen model.ConfigGenerator
}

func New(systems []model.SystemDefinition, emulators []model.EmulatorDefinition) *Registry {
	r := &Registry{
		systems:          make(map[model.SystemID]model.System),
		emulators:        make(map[model.EmulatorID]emulatorEntry),
		defaultEmulators: make(map[model.SystemID]model.EmulatorID),
	}

	for _, def := range emulators {
		emu := def.Emulator()
		r.emulators[emu.ID] = emulatorEntry{
			Emulator:  emu,
			configGen: def.ConfigGenerator(),
		}
	}

	for _, def := range systems {
		sys := def.System()
		r.systems[sys.ID] = sys
		r.defaultEmulators[sys.ID] = def.DefaultEmulatorID()
	}

	return r
}

func (r *Registry) GetSystem(id model.SystemID) (model.System, error) {
	sys, ok := r.systems[id]
	if !ok {
		return model.System{}, fmt.Errorf("unknown system: %s", id)
	}
	return sys, nil
}

func (r *Registry) GetEmulator(id model.EmulatorID) (model.Emulator, error) {
	entry, ok := r.emulators[id]
	if !ok {
		return model.Emulator{}, fmt.Errorf("unknown emulator: %s", id)
	}
	return entry.Emulator, nil
}

func (r *Registry) GetEmulatorsForSystem(sys model.SystemID) []model.Emulator {
	var result []model.Emulator
	for _, entry := range r.emulators {
		if entry.SupportsSystem(sys) {
			result = append(result, entry.Emulator)
		}
	}
	return result
}

func (r *Registry) GetDefaultEmulator(sys model.SystemID) (model.Emulator, error) {
	emuID, ok := r.defaultEmulators[sys]
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
	for _, entry := range r.emulators {
		result = append(result, entry.Emulator)
	}
	return result
}

func (r *Registry) GetConfigGenerator(emuID model.EmulatorID) model.ConfigGenerator {
	entry, ok := r.emulators[emuID]
	if !ok {
		return nil
	}
	return entry.configGen
}
