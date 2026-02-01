import dreamcast from '@/assets/systems/dreamcast.svg'
import gba from '@/assets/systems/gba.svg'
import gc from '@/assets/systems/gc.svg'
import n3ds from '@/assets/systems/n3ds.svg'
import nds from '@/assets/systems/nds.svg'
import ps2 from '@/assets/systems/ps2.svg'
import ps3 from '@/assets/systems/ps3.svg'
import psp from '@/assets/systems/psp.svg'
import psvita from '@/assets/systems/psvita.svg'
import psx from '@/assets/systems/psx.svg'
import snes from '@/assets/systems/snes.svg'
import nswitch from '@/assets/systems/switch.svg'
import wii from '@/assets/systems/wii.svg'
import wiiu from '@/assets/systems/wiiu.svg'
import type { SystemID } from '@/types/daemon'

const SYSTEM_LOGOS: Record<SystemID, string> = {
  snes,
  psx,
  ps2,
  ps3,
  psvita,
  psp,
  gba,
  nds,
  '3ds': n3ds,
  gamecube: gc,
  wii,
  wiiu,
  switch: nswitch,
  dreamcast,
}

export interface SystemLogoProps {
  readonly systemId: SystemID
  readonly systemName?: string
  readonly className?: string
}

export function SystemLogo({ systemId, systemName, className = '' }: SystemLogoProps) {
  const logo = SYSTEM_LOGOS[systemId]

  return (
    <div className={`w-20 flex items-center justify-center ${className}`}>
      <img
        src={logo}
        alt={systemName ?? systemId}
        title={systemName}
        className="w-full h-auto"
        style={{
          filter:
            'brightness(0) saturate(100%) invert(40%) sepia(0%) saturate(0%) hue-rotate(0deg)',
        }}
      />
    </div>
  )
}
