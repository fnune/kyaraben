import { CHANGE_CONFIG, type ChangeType } from '@/lib/changeUtils'

export interface ChangeNotchProps {
  readonly type: NonNullable<ChangeType>
}

export function ChangeNotch({ type }: ChangeNotchProps) {
  const config = CHANGE_CONFIG[type]

  return (
    <div
      className={`
        absolute -bottom-0.5 left-1/2 -translate-x-1/2 z-10
        ${config.bgColor} text-white text-xs font-medium
        px-2.5 py-0.5 rounded-t-lg whitespace-nowrap
      `}
    >
      {config.icon} {config.label}
    </div>
  )
}
