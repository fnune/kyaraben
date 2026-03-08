import { VERSION_DEFAULT } from '@shared/ui'
import { Select } from '@/lib/Select'

export interface VersionSelectProps {
  readonly defaultVersion: string | undefined
  readonly availableVersions: string[] | undefined
  readonly selectedVersion: string
  readonly onChange: (version: string) => void
  readonly disabled?: boolean
  readonly size?: 'sm'
}

export function VersionSelect({
  defaultVersion,
  availableVersions,
  selectedVersion,
  onChange,
  disabled,
  size,
}: VersionSelectProps) {
  if (!availableVersions || availableVersions.length === 0) {
    return (
      <span className="text-xs text-on-surface-muted tabular-nums font-mono">{defaultVersion}</span>
    )
  }

  const isPinned = selectedVersion !== VERSION_DEFAULT
  const options = [
    { value: VERSION_DEFAULT, label: `${defaultVersion} (auto)` },
    ...availableVersions.map((v) => ({ value: v, label: `${v} (pin)` })),
  ]

  return (
    <Select
      value={selectedVersion}
      options={options}
      onChange={onChange}
      disabled={disabled ?? false}
      {...(size && { size })}
      className={isPinned ? '[&>button]:ring-2 [&>button]:ring-status-warning' : ''}
    />
  )
}
