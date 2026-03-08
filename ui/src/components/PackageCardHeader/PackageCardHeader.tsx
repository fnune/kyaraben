import { VERSION_DEFAULT } from '@shared/ui'
import type { ReactNode } from 'react'
import { ChangeNotch } from '@/components/ChangeNotch/ChangeNotch'
import { CHANGE_CONFIG, getChangeType } from '@/lib/changeUtils'
import { ToggleSwitch } from '@/lib/ToggleSwitch'
import { VersionSelect } from '@/lib/VersionSelect'

export type ChangeType = ReturnType<typeof getChangeType>

export interface PackageCardProps {
  readonly changeType: ChangeType
  readonly installed: boolean
  readonly enabled: boolean
  readonly children: ReactNode
}

export function PackageCard({ changeType, installed, enabled, children }: PackageCardProps) {
  const cardClasses = (() => {
    if (changeType) {
      const config = CHANGE_CONFIG[changeType]
      return `ring-1 ${config.ringColor} bg-surface-alt`
    }
    return enabled ? 'bg-surface-alt' : 'bg-surface-alt/50'
  })()

  const borderClasses = installed ? 'border-t-2 border-t-accent' : ''

  return (
    <div className={`rounded-element overflow-hidden relative ${cardClasses} ${borderClasses}`}>
      {changeType && <ChangeNotch type={changeType} />}
      {children}
    </div>
  )
}

export interface PackageCardHeaderProps {
  readonly name: string
  readonly logo?: ReactNode
  readonly defaultVersion: string | undefined
  readonly availableVersions: string[] | undefined
  readonly selectedVersion: string
  readonly enabled: boolean
  readonly readOnly?: boolean
  readonly onToggle: (enabled: boolean) => void
  readonly onVersionChange: (version: string) => void
  readonly secondaryContent?: ReactNode
  readonly nameAction?: ReactNode
}

export function PackageCardHeader({
  name,
  logo,
  defaultVersion,
  availableVersions,
  selectedVersion,
  enabled,
  readOnly,
  onToggle,
  onVersionChange,
  secondaryContent,
  nameAction,
}: PackageCardHeaderProps) {
  return (
    <div className="flex items-center gap-4 p-3">
      {logo && (
        <div className="hidden min-[720px]:flex items-center justify-center w-10 h-10 shrink-0">
          {logo}
        </div>
      )}
      <div className="flex-1 space-y-0.5">
        <div className="flex flex-col gap-2 min-[400px]:flex-row min-[400px]:items-center">
          <span className="text-on-surface font-medium text-sm flex items-center gap-1.5">
            {name}
            {nameAction}
          </span>
          <div className="flex items-center gap-3 min-[400px]:ml-auto">
            <VersionSelect
              defaultVersion={defaultVersion}
              availableVersions={availableVersions}
              selectedVersion={selectedVersion}
              onChange={onVersionChange}
              disabled={!enabled || !!readOnly}
              size="sm"
            />
            <ToggleSwitch enabled={enabled} onChange={onToggle} disabled={!!readOnly} />
          </div>
        </div>
        {secondaryContent}
      </div>
    </div>
  )
}

export function useChangeType(
  enabled: boolean,
  installedVersion: string | null,
  selectedVersion: string,
  defaultVersion: string | undefined,
  availableVersions: string[] | undefined,
): ChangeType {
  const effectiveVersion =
    selectedVersion === VERSION_DEFAULT ? (defaultVersion ?? null) : selectedVersion
  return getChangeType(enabled, installedVersion, effectiveVersion, availableVersions)
}
