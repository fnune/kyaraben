import type { ChangeEvent } from 'react'
import { SystemIcon } from '@/components/SystemIcon/SystemIcon'
import { VersionSelector } from '@/components/VersionSelector/VersionSelector'
import type { EmulatorID, ProvisionResult, System, SystemID } from '@/types/daemon'

export interface SystemCardProps {
  readonly system: System
  readonly selectedEmulator: EmulatorID | null
  readonly pinnedVersion: string | null
  readonly installedVersion: string | null
  readonly provisions: readonly ProvisionResult[]
  readonly enabled: boolean
  readonly onToggle: (systemId: SystemID, enabled: boolean) => void
  readonly onVersionChange: (systemId: SystemID, version: string | null) => void
}

function ProvisionBadge({ provision }: { readonly provision: ProvisionResult }) {
  const isOk = provision.status === 'found'
  const isOptional = !provision.required

  const badgeClasses = isOk
    ? 'bg-green-100 text-green-800'
    : isOptional
      ? 'bg-yellow-100 text-yellow-800'
      : 'bg-red-100 text-red-800'

  const statusText = isOk ? 'OK' : isOptional ? 'optional' : 'missing'

  return (
    <span
      className={`${badgeClasses} px-2 py-0.5 rounded text-xs font-medium`}
      title={provision.description}
    >
      {provision.filename} ({statusText})
    </span>
  )
}

export function SystemCard({
  system,
  selectedEmulator,
  pinnedVersion,
  installedVersion,
  provisions,
  enabled,
  onToggle,
  onVersionChange,
}: SystemCardProps) {
  const emulator = system.emulators.find((e) => e.id === selectedEmulator) ?? system.emulators[0]
  const hasRequiredMissing = provisions.some((p) => p.required && p.status !== 'found')
  const hasProvisions = provisions.length > 0
  const hasVersions = emulator?.availableVersions && emulator.availableVersions.length > 0

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    onToggle(system.id, e.target.checked)
  }

  const handleVersionChange = (version: string | null) => {
    onVersionChange(system.id, version)
  }

  return (
    <article className="border border-gray-200 rounded-lg p-4 hover:border-gray-300 transition-colors">
      <label className="flex items-center gap-3 cursor-pointer">
        <input
          type="checkbox"
          checked={enabled}
          onChange={handleChange}
          className="w-5 h-5 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
        />
        <SystemIcon system={system} size="medium" />
        <div className="flex flex-col flex-1">
          <h3 className="font-semibold text-gray-900">{system.name}</h3>
          <div className="flex items-center justify-between gap-2">
            {emulator && <span className="text-sm text-gray-500">{emulator.name}</span>}
            {hasVersions && enabled && emulator && (
              <VersionSelector
                emulator={emulator}
                pinnedVersion={pinnedVersion}
                installedVersion={installedVersion}
                onChange={handleVersionChange}
              />
            )}
          </div>
        </div>
      </label>

      {hasProvisions && (
        <div className="mt-3 pl-8">
          {hasRequiredMissing && (
            <span className="text-amber-600 text-sm font-medium block mb-2">Requires files</span>
          )}
          <div className="flex flex-wrap gap-1">
            {provisions.map((p) => (
              <ProvisionBadge key={p.filename} provision={p} />
            ))}
          </div>
        </div>
      )}
    </article>
  )
}
