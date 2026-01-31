import type { System } from '@/types/daemon'
import type { Manufacturer } from '@/types/model.gen'

export interface SystemIconProps {
  readonly system: System
  readonly size?: 'small' | 'medium' | 'large'
  readonly className?: string
}

const MANUFACTURER_BG: Readonly<Record<Manufacturer, string>> = {
  Nintendo: 'bg-nintendo',
  Sony: 'bg-sony',
  Sega: 'bg-sega',
  Other: 'bg-gray-500',
}

const SIZE_CLASSES: Readonly<Record<NonNullable<SystemIconProps['size']>, string>> = {
  small: 'w-8 h-8 text-xs',
  medium: 'w-12 h-12 text-sm',
  large: 'w-16 h-16 text-base',
}

export function SystemIcon({ system, size = 'medium', className = '' }: SystemIconProps) {
  const bgClass = MANUFACTURER_BG[system.manufacturer]
  const sizeClass = SIZE_CLASSES[size]

  return (
    <div
      className={`${bgClass} ${sizeClass} rounded-lg flex items-center justify-center text-white font-bold ${className}`}
      role="img"
      aria-label={`${system.id} icon`}
    >
      <span>{system.label}</span>
    </div>
  )
}
