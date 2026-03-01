import bsnes from '@/assets/emulators/bsnes.png'
import cemu from '@/assets/emulators/cemu.png'
import citra from '@/assets/emulators/citra.png'
import dolphin from '@/assets/emulators/dolphin.png'
import duckstation from '@/assets/emulators/duckstation.png'
import eden from '@/assets/emulators/eden.png'
import fbneo from '@/assets/emulators/fbneo.png'
import flycast from '@/assets/emulators/flycast.png'
import genesisPlusGx from '@/assets/emulators/genesis_plus_gx.png'
import mednafen from '@/assets/emulators/mednafen.svg'
import melonds from '@/assets/emulators/melonds.png'
import mesen from '@/assets/emulators/mesen.png'
import mgba from '@/assets/emulators/mgba.png'
import mupen64plus from '@/assets/emulators/mupen64plus.png'
import pcsx2 from '@/assets/emulators/pcsx2.png'
import ppsspp from '@/assets/emulators/ppsspp.png'
import retroarch from '@/assets/emulators/retroarch.svg'
import rpcs3 from '@/assets/emulators/rpcs3.svg'
import snes9x from '@/assets/emulators/snes9x.svg'
import stella from '@/assets/emulators/stella.png'
import vice from '@/assets/emulators/vice.svg'
import vita3k from '@/assets/emulators/vita3k.png'
import xemu from '@/assets/emulators/xemu.png'
import xeniaEdge from '@/assets/emulators/xenia-edge.png'
import type { EmulatorID } from '@/types/daemon'

const EMULATOR_LOGOS: Partial<Record<EmulatorID, string>> = {
  cemu,
  dolphin,
  duckstation,
  eden,
  flycast,
  pcsx2,
  ppsspp,
  retroarch,
  rpcs3,
  vita3k,
  'retroarch:snes9x': snes9x,
  'retroarch:bsnes': bsnes,
  'retroarch:mesen': mesen,
  'retroarch:genesis_plus_gx': genesisPlusGx,
  'retroarch:mupen64plus_next': mupen64plus,
  'retroarch:mednafen_saturn': mednafen,
  'retroarch:mednafen_pce_fast': mednafen,
  'retroarch:mednafen_ngp': mednafen,
  'retroarch:mgba': mgba,
  'retroarch:melondsds': melonds,
  'retroarch:citra': citra,
  'retroarch:fbneo': fbneo,
  'retroarch:stella': stella,
  'retroarch:vice_x64sc': vice,
  xemu,
  'xenia-edge': xeniaEdge,
}

export function getEmulatorLogo(emulatorId: EmulatorID): string | undefined {
  return EMULATOR_LOGOS[emulatorId]
}

export interface EmulatorLogoProps {
  readonly emulatorId: EmulatorID
  readonly emulatorName?: string
  readonly className?: string
}

export function EmulatorLogo({ emulatorId, emulatorName, className = '' }: EmulatorLogoProps) {
  const logo = EMULATOR_LOGOS[emulatorId]

  if (!logo) {
    return null
  }

  return <img src={logo} alt={emulatorName ?? emulatorId} className={className} />
}
