import azahar from '@/assets/emulators/azahar.png'
import cemu from '@/assets/emulators/cemu.png'
import dolphin from '@/assets/emulators/dolphin.png'
import duckstation from '@/assets/emulators/duckstation.png'
import eden from '@/assets/emulators/eden.png'
import flycast from '@/assets/emulators/flycast.png'
import melonds from '@/assets/emulators/melonds.png'
import mgba from '@/assets/emulators/mgba.png'
import pcsx2 from '@/assets/emulators/pcsx2.png'
import ppsspp from '@/assets/emulators/ppsspp.png'
import retroarch from '@/assets/emulators/retroarch.svg'
import rpcs3 from '@/assets/emulators/rpcs3.svg'
import vita3k from '@/assets/emulators/vita3k.png'
import type { EmulatorID } from '@/types/daemon'

const EMULATOR_LOGOS: Partial<Record<EmulatorID, string>> = {
  azahar,
  cemu,
  dolphin,
  duckstation,
  eden,
  flycast,
  melonds,
  mgba,
  pcsx2,
  ppsspp,
  retroarch,
  rpcs3,
  vita3k,
  'retroarch:bsnes': retroarch,
  'retroarch:mesen': retroarch,
  'retroarch:genesis_plus_gx': retroarch,
  'retroarch:mupen64plus_next': retroarch,
  'retroarch:mednafen_saturn': retroarch,
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
