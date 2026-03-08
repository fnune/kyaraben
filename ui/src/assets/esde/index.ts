import type { SystemID } from '@shared/daemon'
import arcadeLogo from './logos/arcade.svg'
import atari2600Logo from './logos/atari2600.svg'
import c64Logo from './logos/c64.svg'
import dreamcastLogo from './logos/dreamcast.svg'
import gamecubeLogo from './logos/gamecube.svg'
import gamegearLogo from './logos/gamegear.svg'
import gbLogo from './logos/gb.svg'
import gbaLogo from './logos/gba.svg'
import gbcLogo from './logos/gbc.svg'
import genesisLogo from './logos/genesis.svg'
import mastersystemLogo from './logos/mastersystem.svg'
import n3dsLogo from './logos/n3ds.svg'
import n64Logo from './logos/n64.svg'
import ndsLogo from './logos/nds.svg'
import neogeoLogo from './logos/neogeo.svg'
import nesLogo from './logos/nes.svg'
import ngpLogo from './logos/ngp.svg'
import pcengineLogo from './logos/pcengine.svg'
import ps2Logo from './logos/ps2.svg'
import ps3Logo from './logos/ps3.svg'
import pspLogo from './logos/psp.svg'
import psvitaLogo from './logos/psvita.svg'
import psxLogo from './logos/psx.svg'
import saturnLogo from './logos/saturn.svg'
import snesLogo from './logos/snes.svg'
import switchLogo from './logos/switch.svg'
import wiiLogo from './logos/wii.svg'
import wiiuLogo from './logos/wiiu.svg'
import xboxLogo from './logos/xbox.svg'
import xbox360Logo from './logos/xbox360.svg'

export const ESDE_LOGOS: Record<SystemID, string> = {
  nes: nesLogo,
  snes: snesLogo,
  n64: n64Logo,
  gb: gbLogo,
  gbc: gbcLogo,
  gba: gbaLogo,
  nds: ndsLogo,
  n3ds: n3dsLogo,
  gamecube: gamecubeLogo,
  wii: wiiLogo,
  wiiu: wiiuLogo,
  switch: switchLogo,
  psx: psxLogo,
  ps2: ps2Logo,
  ps3: ps3Logo,
  psp: pspLogo,
  psvita: psvitaLogo,
  genesis: genesisLogo,
  mastersystem: mastersystemLogo,
  gamegear: gamegearLogo,
  saturn: saturnLogo,
  dreamcast: dreamcastLogo,
  pcengine: pcengineLogo,
  ngp: ngpLogo,
  neogeo: neogeoLogo,
  xbox: xboxLogo,
  xbox360: xbox360Logo,
  atari2600: atari2600Logo,
  c64: c64Logo,
  arcade: arcadeLogo,
}
