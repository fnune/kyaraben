package registry

import (
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
	"github.com/fnune/kyaraben/internal/emulators/retroarchfbneo"
	"github.com/fnune/kyaraben/internal/emulators/retroarchgenesisplusgx"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmelonds"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmesen"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmgba"
	"github.com/fnune/kyaraben/internal/emulators/retroarchmupen64plus"
	"github.com/fnune/kyaraben/internal/emulators/retroarchstella"
	"github.com/fnune/kyaraben/internal/emulators/retroarchvice"
	"github.com/fnune/kyaraben/internal/emulators/rpcs3"
	"github.com/fnune/kyaraben/internal/emulators/vita3k"
	"github.com/fnune/kyaraben/internal/emulators/xemu"
	"github.com/fnune/kyaraben/internal/emulators/xeniaedge"
	"github.com/fnune/kyaraben/internal/frontends/esde"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/systems/arcade"
	"github.com/fnune/kyaraben/internal/systems/atari2600"
	"github.com/fnune/kyaraben/internal/systems/c64"
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
	"github.com/fnune/kyaraben/internal/systems/neogeo"
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
	"github.com/fnune/kyaraben/internal/systems/xbox"
	"github.com/fnune/kyaraben/internal/systems/xbox360"
)

func NewDefault() *Registry {
	return New(
		[]model.SystemDefinition{
			// Nintendo
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
			// Sony
			psx.Definition{},
			ps2.Definition{},
			ps3.Definition{},
			psp.Definition{},
			psvita.Definition{},
			// Sega
			genesis.Definition{},
			mastersystem.Definition{},
			gamegear.Definition{},
			saturn.Definition{},
			dreamcast.Definition{},
			// NEC
			pcengine.Definition{},
			// SNK
			ngp.Definition{},
			neogeo.Definition{},
			// Microsoft
			xbox.Definition{},
			xbox360.Definition{},
			// Atari
			atari2600.Definition{},
			// Commodore
			c64.Definition{},
			// Arcade
			arcade.Definition{},
		},
		[]model.EmulatorDefinition{
			retroarchbsnes.Definition{},
			retroarchmesen.Definition{},
			retroarchgenesisplusgx.Definition{},
			retroarchmupen64plus.Definition{},
			retroarchbeetlesaturn.Definition{},
			retroarchbeetlepce.Definition{},
			retroarchbeetlengp.Definition{},
			retroarchmgba.Definition{},
			retroarchmelonds.Definition{},
			retroarchfbneo.Definition{},
			retroarchstella.Definition{},
			retroarchvice.Definition{},
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
			xemu.Definition{},
			xeniaedge.Definition{},
		},
		[]model.FrontendDefinition{
			esde.Definition{},
		},
	)
}
