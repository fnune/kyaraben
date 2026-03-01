package registry

import (
	"fmt"
	"slices"

	"github.com/fnune/kyaraben/internal/model"
)

type Registry struct {
	systems          map[model.SystemID]model.System
	emulators        map[model.EmulatorID]emulatorEntry
	defaultEmulators map[model.SystemID]model.EmulatorID
	frontends        map[model.FrontendID]frontendEntry
}

type emulatorEntry struct {
	model.Emulator
	configGen model.ConfigGenerator
}

type frontendEntry struct {
	model.Frontend
	configGen  model.FrontendConfigGenerator
	definition model.FrontendDefinition
}

func New(systems []model.SystemDefinition, emulators []model.EmulatorDefinition, frontends []model.FrontendDefinition) *Registry {
	r := &Registry{
		systems:          make(map[model.SystemID]model.System),
		emulators:        make(map[model.EmulatorID]emulatorEntry),
		defaultEmulators: make(map[model.SystemID]model.EmulatorID),
		frontends:        make(map[model.FrontendID]frontendEntry),
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

	for _, def := range frontends {
		fe := def.Frontend()
		r.frontends[fe.ID] = frontendEntry{
			Frontend:   fe,
			configGen:  def.ConfigGenerator(),
			definition: def,
		}
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
	defaultEmuID := r.defaultEmulators[sys]
	slices.SortFunc(result, func(a, b model.Emulator) int {
		aIsDefault := a.ID == defaultEmuID
		bIsDefault := b.ID == defaultEmuID
		if aIsDefault && !bIsDefault {
			return -1
		}
		if bIsDefault && !aIsDefault {
			return 1
		}
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
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
	result := make([]model.System, 0, len(r.systems))
	for _, sys := range r.systems {
		result = append(result, sys)
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

func (r *Registry) GetFrontend(id model.FrontendID) (model.Frontend, error) {
	entry, ok := r.frontends[id]
	if !ok {
		return model.Frontend{}, fmt.Errorf("unknown frontend: %s", id)
	}
	return entry.Frontend, nil
}

func (r *Registry) GetFrontendConfigGenerator(id model.FrontendID) model.FrontendConfigGenerator {
	entry, ok := r.frontends[id]
	if !ok {
		return nil
	}
	return entry.configGen
}

func (r *Registry) GetFrontendDefinition(id model.FrontendID) model.FrontendDefinition {
	entry, ok := r.frontends[id]
	if !ok {
		return nil
	}
	return entry.definition
}

func (r *Registry) AllFrontends() []model.Frontend {
	result := make([]model.Frontend, 0, len(r.frontends))
	for _, entry := range r.frontends {
		result = append(result, entry.Frontend)
	}
	return result
}
