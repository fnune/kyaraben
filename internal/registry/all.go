package registry

import (
	"github.com/fnune/kyaraben/internal/emulators/duckstation"
	"github.com/fnune/kyaraben/internal/emulators/e2etestemu"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbsnes"
	"github.com/fnune/kyaraben/internal/emulators/tic80emu"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/systems/e2etest"
	"github.com/fnune/kyaraben/internal/systems/psx"
	"github.com/fnune/kyaraben/internal/systems/snes"
	"github.com/fnune/kyaraben/internal/systems/tic80"
)

func NewDefault() *Registry {
	return New(
		[]model.SystemDefinition{
			snes.Definition{},
			psx.Definition{},
			tic80.Definition{},
			e2etest.Definition{},
		},
		[]model.EmulatorDefinition{
			retroarchbsnes.Definition{},
			duckstation.Definition{},
			tic80emu.Definition{},
			e2etestemu.Definition{},
		},
	)
}
