import { ChangeNotch } from '@/components/ChangeNotch/ChangeNotch'
import { getFrontendLogo } from '@/components/FrontendLogo/FrontendLogo'
import { CHANGE_CONFIG, formatBytes, getChangeType } from '@/lib/changeUtils'
import { Select } from '@/lib/Select'
import { ToggleSwitch } from '@/lib/ToggleSwitch'
import type { FrontendRef } from '@/types/daemon.gen'

export interface FrontendCardProps {
  readonly frontend: FrontendRef
  readonly enabled: boolean
  readonly pinnedVersion: string | null
  readonly installedVersion: string | null
  readonly onToggle: (enabled: boolean) => void
  readonly onVersionChange: (version: string | null) => void
}

export function FrontendCard({
  frontend,
  enabled,
  pinnedVersion,
  installedVersion,
  onToggle,
  onVersionChange,
}: FrontendCardProps) {
  const effectiveVersion = pinnedVersion ?? frontend.defaultVersion ?? null
  const changeType = getChangeType(
    enabled,
    installedVersion,
    effectiveVersion,
    frontend.availableVersions,
  )

  const cardClasses = (() => {
    if (changeType) {
      const config = CHANGE_CONFIG[changeType]
      return `ring-1 ${config.ringColor} bg-surface-alt`
    }
    return 'bg-surface-alt'
  })()

  const logo = getFrontendLogo(frontend.id)
  const borderClasses = installedVersion ? 'border-t-2 border-t-accent' : ''

  return (
    <div className={`rounded-element overflow-hidden relative ${cardClasses} ${borderClasses}`}>
      {changeType && <ChangeNotch type={changeType} />}

      <div className="flex items-center gap-4 p-3">
        {logo && (
          <div className="hidden min-[720px]:flex items-center justify-center w-10 h-10 shrink-0">
            <img src={logo} alt="" className="w-full h-full object-contain" />
          </div>
        )}
        <div className="flex-1 space-y-0.5">
          <div className="flex items-center gap-2">
            <span className="text-on-surface font-medium text-sm">{frontend.name}</span>
            <div className="ml-auto flex items-center gap-3">
              <VersionSelector
                defaultVersion={frontend.defaultVersion}
                availableVersions={frontend.availableVersions}
                pinnedVersion={pinnedVersion}
                onChange={onVersionChange}
                disabled={!enabled}
              />
              <ToggleSwitch enabled={enabled} onChange={onToggle} />
            </div>
          </div>
          <div className="flex items-center gap-2 text-xs text-on-surface-muted">
            {installedVersion ? (
              <span className="text-on-surface-dim">Installed: {installedVersion}</span>
            ) : (
              frontend.downloadBytes && (
                <span className="text-on-surface-dim">{formatBytes(frontend.downloadBytes)}</span>
              )
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function VersionSelector({
  defaultVersion,
  availableVersions,
  pinnedVersion,
  onChange,
  disabled,
}: {
  readonly defaultVersion: string | undefined
  readonly availableVersions: string[] | undefined
  readonly pinnedVersion: string | null
  readonly onChange: (version: string | null) => void
  readonly disabled: boolean
}) {
  if (!availableVersions || availableVersions.length === 0) {
    return (
      <span className="text-xs text-on-surface-muted tabular-nums font-mono">{defaultVersion}</span>
    )
  }

  const isPinned = pinnedVersion !== null
  const options = [
    { value: '', label: `${defaultVersion} (auto)` },
    ...availableVersions.map((v) => ({ value: v, label: `${v} (pin)` })),
  ]

  return (
    <Select
      value={pinnedVersion ?? ''}
      options={options}
      onChange={(v) => onChange(v === '' ? null : v)}
      disabled={disabled}
      className={isPinned ? '[&>button]:ring-2 [&>button]:ring-status-warning' : ''}
    />
  )
}
