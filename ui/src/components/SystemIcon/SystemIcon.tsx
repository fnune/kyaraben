import type { SystemID } from '@/types/daemon'
import { type Manufacturer, SYSTEM_MANUFACTURERS } from '@/types/ui'

export interface SystemIconProps {
  readonly systemId: SystemID
  readonly size?: 'small' | 'medium' | 'large'
  readonly className?: string
}

const SYSTEM_LABELS: Readonly<Record<SystemID, string>> = {
  snes: 'SNES',
  gba: 'GBA',
  nds: 'NDS',
  switch: 'NSW',
  psx: 'PSX',
  psp: 'PSP',
  tic80: 'TIC',
  'e2e-test': 'TST',
}

const MANUFACTURER_BG: Readonly<Record<Manufacturer, string>> = {
  Nintendo: 'bg-nintendo',
  Sony: 'bg-sony',
  Other: 'bg-gray-500',
}

const SIZE_CLASSES: Readonly<Record<NonNullable<SystemIconProps['size']>, string>> = {
  small: 'w-8 h-8 text-xs',
  medium: 'w-12 h-12 text-sm',
  large: 'w-16 h-16 text-base',
}

export function SystemIcon({ systemId, size = 'medium', className = '' }: SystemIconProps) {
  const manufacturer = SYSTEM_MANUFACTURERS[systemId]
  const label = SYSTEM_LABELS[systemId]
  const bgClass = MANUFACTURER_BG[manufacturer]
  const sizeClass = SIZE_CLASSES[size]

  return (
    <div
      className={`${bgClass} ${sizeClass} rounded-lg flex items-center justify-center text-white font-bold ${className}`}
      role="img"
      aria-label={`${systemId} icon`}
    >
      <span>{label}</span>
    </div>
  )
}
