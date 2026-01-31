package registry

import (
	"github.com/fnune/kyaraben/internal/emulators/duckstation"
	"github.com/fnune/kyaraben/internal/emulators/e2etestemu"
	"github.com/fnune/kyaraben/internal/emulators/eden"
	"github.com/fnune/kyaraben/internal/emulators/mgba"
	"github.com/fnune/kyaraben/internal/emulators/pcsx2"
	"github.com/fnune/kyaraben/internal/emulators/ppsspp"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbsnes"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmelonds"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmgba"
	"github.com/fnune/kyaraben/internal/emulators/retroarchppsspp"
	"github.com/fnune/kyaraben/internal/emulators/tic80emu"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/systems/e2etest"
	"github.com/fnune/kyaraben/internal/systems/gba"
	"github.com/fnune/kyaraben/internal/systems/nds"
	"github.com/fnune/kyaraben/internal/systems/ps2"
	"github.com/fnune/kyaraben/internal/systems/psp"
	"github.com/fnune/kyaraben/internal/systems/psx"
	"github.com/fnune/kyaraben/internal/systems/snes"
	switchsys "github.com/fnune/kyaraben/internal/systems/switch"
	"github.com/fnune/kyaraben/internal/systems/tic80"
)

func NewDefault() *Registry {
	return New(
		[]model.SystemDefinition{
			snes.Definition{},
			psx.Definition{},
			ps2.Definition{},
			tic80.Definition{},
			gba.Definition{},
			nds.Definition{},
			psp.Definition{},
			switchsys.Definition{},
			e2etest.Definition{},
		},
		[]model.EmulatorDefinition{
			retroarchbsnes.Definition{},
			retroarchmgba.Definition{},
			retroarchmelonds.Definition{},
			retroarchppsspp.Definition{},
			duckstation.Definition{},
			pcsx2.Definition{},
			ppsspp.Definition{},
			mgba.Definition{},
			tic80emu.Definition{},
			eden.Definition{},
			e2etestemu.Definition{},
		},
	)
}
