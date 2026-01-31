import type { ChangeEvent } from 'react'
import type { EmulatorRef } from '@/types/daemon'

export interface VersionSelectorProps {
  readonly emulator: EmulatorRef
  readonly pinnedVersion: string | null
  readonly installedVersion: string | null
  readonly onChange: (version: string | null) => void
  readonly disabled?: boolean
}

export function VersionSelector({
  emulator,
  pinnedVersion,
  installedVersion,
  onChange,
  disabled,
}: VersionSelectorProps) {
  if (!emulator.availableVersions || emulator.availableVersions.length === 0) {
    return null
  }

  const handleChange = (e: ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value
    onChange(value === '' ? null : value)
  }

  const currentValue = pinnedVersion ?? ''
  const effectiveVersion = pinnedVersion ?? emulator.defaultVersion
  const willUpdate = installedVersion && effectiveVersion !== installedVersion
  const notInstalled = !installedVersion

  // Determine status message
  let statusMessage: string | null = null
  let statusColor = 'text-gray-500'

  if (notInstalled) {
    statusMessage = `will install ${effectiveVersion}`
    statusColor = 'text-blue-600'
  } else if (willUpdate) {
    statusMessage = `${installedVersion} → ${effectiveVersion} on apply`
    statusColor = 'text-amber-600'
  }

  return (
    <div className="flex flex-col items-end gap-0.5">
      <select
        value={currentValue}
        onChange={handleChange}
        disabled={disabled}
        className="text-xs bg-gray-50 border border-gray-200 rounded px-1.5 py-0.5 text-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:opacity-50 cursor-pointer"
        title={pinnedVersion ? `Pinned to ${pinnedVersion}` : `Using default (${emulator.defaultVersion})`}
      >
        <option value="">{emulator.defaultVersion}</option>
        {emulator.availableVersions.map((version) => (
          <option key={version} value={version}>
            {version} (pinned)
          </option>
        ))}
      </select>
      {statusMessage && <span className={`text-[10px] ${statusColor}`}>{statusMessage}</span>}
    </div>
  )
}
