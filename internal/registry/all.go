package registry

import (
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
	"github.com/fnune/kyaraben/internal/systems/wii"
	"github.com/fnune/kyaraben/internal/systems/wiiu"
)

func NewDefault() *Registry {
	return New(
		[]model.SystemDefinition{
			snes.Definition{},
			psx.Definition{},
			ps2.Definition{},
			ps3.Definition{},
			psvita.Definition{},
			dreamcast.Definition{},
			gba.Definition{},
			nds.Definition{},
			psp.Definition{},
			gamecube.Definition{},
			wii.Definition{},
			wiiu.Definition{},
			n3ds.Definition{},
			switchsys.Definition{},
			e2etest.Definition{},
		},
		[]model.EmulatorDefinition{
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
			eden.Definition{},
			e2etestemu.Definition{},
		},
	)
}
