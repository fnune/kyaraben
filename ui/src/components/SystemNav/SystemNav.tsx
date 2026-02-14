import { EmulatorLogo } from '@/components/EmulatorLogo/EmulatorLogo'
import {
  OPTIONAL_PROVISION_BG,
  OPTIONAL_PROVISION_BORDER,
  OPTIONAL_PROVISION_COLOR,
  OPTIONAL_PROVISION_ICON,
  PROVISION_FOUND_BG,
  PROVISION_FOUND_BORDER,
  PROVISION_FOUND_COLOR,
  PROVISION_FOUND_ICON,
  PROVISION_MISSING_BG,
  PROVISION_MISSING_BORDER,
  PROVISION_MISSING_COLOR,
  PROVISION_MISSING_ICON,
} from '@/lib/provisionStatus'
import type { EmulatorID } from '@/types/daemon'

export type ProvisionStatus = 'ok' | 'optional-missing' | 'required-missing' | null

export interface SystemNavItem {
  readonly id: EmulatorID
  readonly name: string
  readonly systemName: string
  readonly installed: boolean
  readonly provisionStatus: ProvisionStatus
}

export interface SystemNavProps {
  readonly emulators: readonly SystemNavItem[]
  readonly onEmulatorClick: (id: EmulatorID) => void
}

function ProvisionBadge({ status }: { status: ProvisionStatus }) {
  if (!status) return null

  const config = {
    ok: {
      icon: PROVISION_FOUND_ICON,
      color: PROVISION_FOUND_COLOR,
      border: PROVISION_FOUND_BORDER,
      bg: PROVISION_FOUND_BG,
    },
    'optional-missing': {
      icon: OPTIONAL_PROVISION_ICON,
      color: OPTIONAL_PROVISION_COLOR,
      border: OPTIONAL_PROVISION_BORDER,
      bg: OPTIONAL_PROVISION_BG,
    },
    'required-missing': {
      icon: PROVISION_MISSING_ICON,
      color: PROVISION_MISSING_COLOR,
      border: PROVISION_MISSING_BORDER,
      bg: PROVISION_MISSING_BG,
    },
  }[status]

  return (
    <span
      className={`absolute -top-1 -right-1 w-3 h-3 rounded-full border flex items-center justify-center text-[8px] font-medium ${config.color} ${config.border} ${config.bg}`}
    >
      {config.icon}
    </span>
  )
}

export function SystemNav({ emulators, onEmulatorClick }: SystemNavProps) {
  if (emulators.length === 0) {
    return null
  }

  return (
    <div className="flex flex-wrap gap-0.5">
      {emulators.map((emulator) => (
        <button
          key={emulator.id}
          type="button"
          onClick={() => onEmulatorClick(emulator.id)}
          title={`${emulator.name} (${emulator.systemName})`}
          className="flex flex-col items-center gap-0.5 p-1 rounded-sm transition-all w-10 hover:bg-surface-raised"
        >
          <div className="relative">
            <EmulatorLogo
              emulatorId={emulator.id}
              emulatorName={emulator.name}
              className="w-5 h-5 rounded-sm"
            />
            {emulator.installed && <ProvisionBadge status={emulator.provisionStatus} />}
          </div>
          <span className="text-[9px] text-on-surface-dim leading-tight w-full text-center truncate">
            {emulator.name}
          </span>
        </button>
      ))}
    </div>
  )
}
