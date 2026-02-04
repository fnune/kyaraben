import { type ChangeEvent, useState } from 'react'
import { SystemLogo } from '@/components/SystemLogo/SystemLogo'
import { formatBytes } from '@/lib/changeUtils'
import { Modal } from '@/lib/Modal'
import type { EmulatorID, EmulatorRef, ProvisionResult, System, SystemID } from '@/types/daemon'

export interface SystemRowProps {
  readonly system: System
  readonly onEnableDefault: (systemId: SystemID) => void
}

export interface EmulatorRowProps {
  readonly systemId: SystemID
  readonly systemName: string
  readonly emulator: EmulatorRef
  readonly pinnedVersion: string | null
  readonly installedVersion: string | null
  readonly enabled: boolean
  readonly emulatorSharedWith: readonly string[]
  readonly emulatorInstalledFor: readonly string[]
  readonly provisions: readonly ProvisionResult[]
  readonly userStore: string
  readonly onToggle: (systemId: SystemID, emulatorId: EmulatorID, enabled: boolean) => void
  readonly onVersionChange: (emulatorId: EmulatorID, version: string | null) => void
}

function ProvisionsBadges({
  provisions,
  onClick,
}: {
  readonly provisions: readonly ProvisionResult[]
  readonly onClick: (e: React.MouseEvent) => void
}) {
  if (provisions.length === 0) return null

  return (
    <button
      type="button"
      onClick={onClick}
      className="flex flex-nowrap gap-1 py-1 overflow-hidden min-[720px]:flex-wrap min-[720px]:overflow-visible cursor-pointer hover:opacity-80 transition-opacity"
    >
      {provisions.map((p) => {
        const isFound = p.status === 'found'
        const bgColor = isFound
          ? 'bg-green-100 border-green-200'
          : p.required
            ? 'bg-red-50 border-red-200'
            : 'bg-amber-50 border-amber-200'
        const textColor = isFound
          ? 'text-green-700'
          : p.required
            ? 'text-red-700'
            : 'text-amber-700'
        const descColor = isFound
          ? 'text-green-600'
          : p.required
            ? 'text-red-600'
            : 'text-amber-600'

        return (
          <span
            key={p.filename}
            className={`inline-flex flex-col px-1.5 py-0.5 rounded border ${bgColor}`}
          >
            <span className={`text-[10px] font-medium font-mono leading-tight ${textColor}`}>
              {p.filename}
            </span>
            <span className={`text-[9px] leading-tight ${descColor}`}>{p.description}</span>
          </span>
        )
      })}
    </button>
  )
}

type ActionType =
  | 'will-install'
  | 'will-update'
  | 'will-uninstall'
  | 'already-installed'
  | 'shared-uninstall'
  | null

function getAction(
  enabled: boolean,
  installedVersion: string | null,
  effectiveVersion: string | null,
  emulatorSharedWith: readonly string[],
  emulatorInstalledFor: readonly string[],
): ActionType {
  if (!enabled) {
    if (!installedVersion) return null
    if (emulatorSharedWith.length > 0) return 'shared-uninstall'
    return 'will-uninstall'
  }

  if (!installedVersion) {
    if (emulatorInstalledFor.length > 0) return 'already-installed'
    return 'will-install'
  }

  if (effectiveVersion && installedVersion !== effectiveVersion) return 'will-update'
  return null
}

function ActionLabel({
  action,
  installedVersion,
  emulatorSharedWith,
  emulatorInstalledFor,
}: {
  readonly action: ActionType
  readonly installedVersion: string | null
  readonly emulatorSharedWith: readonly string[]
  readonly emulatorInstalledFor: readonly string[]
}) {
  if (!action) return null

  const formatSystemList = (systems: readonly string[]) => {
    if (systems.length === 1) return systems[0]
    if (systems.length === 2) return `${systems[0]} and ${systems[1]}`
    return `${systems.slice(0, -1).join(', ')}, and ${systems[systems.length - 1]}`
  }

  const config: Record<NonNullable<ActionType>, { text: string; color: string }> = {
    'will-install': { text: 'Will install', color: 'text-blue-600' },
    'will-update': {
      text: installedVersion ? `Will update from ${installedVersion}` : 'Will update',
      color: 'text-amber-600',
    },
    'will-uninstall': { text: 'Will uninstall', color: 'text-red-600' },
    'already-installed': {
      text: `Already installed for ${formatSystemList(emulatorInstalledFor)}`,
      color: 'text-green-600',
    },
    'shared-uninstall': {
      text: `In use by ${formatSystemList(emulatorSharedWith)}`,
      color: 'text-gray-500',
    },
  }

  const { text, color } = config[action]

  return <span className={`text-sm ${color}`}>{text}</span>
}

function VersionSelector({
  emulator,
  pinnedVersion,
  onChange,
  disabled,
}: {
  readonly emulator: EmulatorRef
  readonly pinnedVersion: string | null
  readonly onChange: (version: string | null) => void
  readonly disabled: boolean
}) {
  if (!emulator.availableVersions || emulator.availableVersions.length === 0) {
    return <span className="text-sm text-gray-500">{emulator.defaultVersion}</span>
  }

  const handleChange = (e: ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value
    onChange(value === '' ? null : value)
  }

  return (
    <select
      value={pinnedVersion ?? ''}
      onChange={handleChange}
      disabled={disabled}
      onClick={(e) => e.stopPropagation()}
      className="text-sm bg-transparent border-none text-gray-500 focus:outline-none focus:ring-0 cursor-pointer disabled:cursor-default disabled:opacity-50 tabular-nums"
    >
      <option value="">{emulator.defaultVersion}</option>
      {emulator.availableVersions.map((v) => (
        <option key={v} value={v}>
          {v} 📌
        </option>
      ))}
    </select>
  )
}

function ProvisionsDialog({
  open,
  onClose,
  systemId,
  systemName,
  provisions,
  userStore,
}: {
  readonly open: boolean
  readonly onClose: () => void
  readonly systemId: SystemID
  readonly systemName: string
  readonly provisions: readonly ProvisionResult[]
  readonly userStore: string
}) {
  const missing = provisions.filter((p) => p.status !== 'found')
  const found = provisions.filter((p) => p.status === 'found')
  const biosPath = userStore ? `${userStore}/bios/${systemId}` : null

  const handleOpenFolder = () => {
    if (biosPath) {
      window.electron.invoke('open_path', biosPath)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={`${systemName} files`}>
      <div className="space-y-4">
        {biosPath && (
          <div className="bg-gray-50 rounded-lg p-3">
            <p className="text-sm text-gray-600 mb-2">Place files in:</p>
            <div className="flex items-center gap-2">
              <code className="text-sm bg-gray-100 px-2 py-1 rounded flex-1 truncate">
                {biosPath}
              </code>
              <button
                type="button"
                onClick={handleOpenFolder}
                className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors flex-shrink-0"
              >
                Open folder
              </button>
            </div>
          </div>
        )}

        <ul className="space-y-3">
          {missing.map((p) => (
            <li key={p.filename} className="flex items-start gap-3">
              {p.required ? (
                <svg
                  className="w-6 h-6 text-red-500 flex-shrink-0"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                  role="img"
                  aria-label={`${p.filename} is missing and required`}
                >
                  <path d="M12 2L1 21h22L12 2zm0 4l7.53 13H4.47L12 6zm-1 5v4h2v-4h-2zm0 6v2h2v-2h-2z" />
                </svg>
              ) : (
                <svg
                  className="w-6 h-6 text-amber-500 flex-shrink-0"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                  role="img"
                  aria-label={`${p.filename} is missing but optional`}
                >
                  <circle cx="12" cy="12" r="10" />
                </svg>
              )}
              <div>
                <code className="text-sm font-medium text-gray-900 bg-gray-100 px-1.5 py-0.5 rounded">
                  {p.filename}
                </code>
                {p.required ? (
                  <span className="text-xs text-red-600 ml-2">required</span>
                ) : (
                  <span className="text-xs text-amber-600 ml-2">optional</span>
                )}
                <p className="text-sm text-gray-500 mt-1">{p.description}</p>
              </div>
            </li>
          ))}
          {found.map((p) => (
            <li key={p.filename} className="flex items-start gap-3">
              <svg
                className="w-6 h-6 text-green-500 flex-shrink-0"
                viewBox="0 0 24 24"
                fill="currentColor"
                role="img"
                aria-label={`${p.filename} found`}
              >
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
              </svg>
              <div>
                <code className="text-sm text-gray-600 bg-gray-100 px-1.5 py-0.5 rounded">
                  {p.filename}
                </code>
                <p className="text-sm text-gray-500 mt-1">{p.description}</p>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </Modal>
  )
}

export function EmulatorRow({
  systemId,
  systemName,
  emulator,
  pinnedVersion,
  installedVersion,
  enabled,
  emulatorSharedWith,
  emulatorInstalledFor,
  provisions,
  userStore,
  onToggle,
  onVersionChange,
}: EmulatorRowProps) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const effectiveVersion = pinnedVersion ?? emulator.defaultVersion ?? null
  const action = getAction(
    enabled,
    installedVersion,
    effectiveVersion,
    emulatorSharedWith,
    emulatorInstalledFor,
  )

  const statusColor = (() => {
    if (!enabled) return 'border-gray-300'
    if (action === 'will-install') return 'border-blue-500'
    if (action === 'will-update') return 'border-amber-500'
    if (installedVersion) return 'border-green-500'
    return 'border-gray-300'
  })()

  return (
    <>
      <label
        className={`flex flex-wrap items-center gap-x-3 gap-y-1 min-h-10 py-2 px-3 bg-gray-50/50 border-l-4 ${statusColor} hover:bg-gray-100/50 transition-colors cursor-pointer`}
      >
        <input
          type="checkbox"
          checked={enabled}
          onChange={(e) => onToggle(systemId, emulator.id, e.target.checked)}
          className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500 cursor-pointer"
        />
        <span className="text-sm text-gray-700">{emulator.name}</span>
        <ProvisionsBadges
          provisions={provisions}
          onClick={(e) => {
            e.preventDefault()
            setDialogOpen(true)
          }}
        />
        <span className="flex-1" />
        {emulator.downloadBytes && (
          <span className="text-xs text-gray-400">{formatBytes(emulator.downloadBytes)}</span>
        )}
        <ActionLabel
          action={action}
          installedVersion={installedVersion}
          emulatorSharedWith={emulatorSharedWith}
          emulatorInstalledFor={emulatorInstalledFor}
        />
        <VersionSelector
          emulator={emulator}
          pinnedVersion={pinnedVersion}
          onChange={(v) => onVersionChange(emulator.id, v)}
          disabled={!enabled}
        />
      </label>

      <ProvisionsDialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        systemId={systemId}
        systemName={systemName}
        provisions={provisions}
        userStore={userStore}
      />
    </>
  )
}

export function SystemRow({ system, onEnableDefault }: SystemRowProps) {
  return (
    <button
      type="button"
      onClick={() => onEnableDefault(system.id)}
      className="w-full min-h-12 flex items-center justify-center py-2 px-3 border-l-4 border-gray-300 hover:bg-gray-50 transition-colors"
    >
      <SystemLogo systemId={system.id} systemName={system.name} />
    </button>
  )
}
