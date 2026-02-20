import dreamcast from '@/assets/systems/dreamcast.svg'
import gamegear from '@/assets/systems/gamegear.svg'
import gb from '@/assets/systems/gb.svg'
import gba from '@/assets/systems/gba.svg'
import gbc from '@/assets/systems/gbc.svg'
import gc from '@/assets/systems/gc.svg'
import genesis from '@/assets/systems/genesis.svg'
import mastersystem from '@/assets/systems/mastersystem.svg'
import n3ds from '@/assets/systems/n3ds.svg'
import n64 from '@/assets/systems/n64.svg'
import nds from '@/assets/systems/nds.svg'
import nes from '@/assets/systems/nes.svg'
import ngp from '@/assets/systems/ngp.svg'
import pcengine from '@/assets/systems/pcengine.svg'
import ps2 from '@/assets/systems/ps2.svg'
import ps3 from '@/assets/systems/ps3.svg'
import psp from '@/assets/systems/psp.svg'
import psvita from '@/assets/systems/psvita.svg'
import psx from '@/assets/systems/psx.svg'
import saturn from '@/assets/systems/saturn.svg'
import snes from '@/assets/systems/snes.svg'
import nswitch from '@/assets/systems/switch.svg'
import wii from '@/assets/systems/wii.svg'
import wiiu from '@/assets/systems/wiiu.svg'
import type { SystemID } from '@/types/daemon'

export const SYSTEM_LOGOS: Record<SystemID, string> = {
  // Nintendo
  nes,
  snes,
  n64,
  gb,
  gbc,
  gba,
  nds,
  n3ds,
  gamecube: gc,
  wii,
  wiiu,
  switch: nswitch,
  // Sony
  psx,
  ps2,
  ps3,
  psp,
  psvita,
  // Sega
  genesis,
  mastersystem,
  gamegear,
  saturn,
  dreamcast,
  // NEC
  pcengine,
  // SNK
  ngp,
}

export interface SystemLogoProps {
  readonly systemId: SystemID
  readonly systemName?: string
  readonly className?: string
}

export function SystemLogo({ systemId, systemName, className = '' }: SystemLogoProps) {
  const logo = SYSTEM_LOGOS[systemId]

  return (
    <div className={`w-16 h-10 flex items-center justify-center ${className}`}>
      <img
        src={logo}
        alt={systemName ?? systemId}
        title={systemName}
        className="max-w-full max-h-full object-contain"
        style={{ filter: 'var(--t-logo-filter)' }}
      />
    </div>
  )
}
