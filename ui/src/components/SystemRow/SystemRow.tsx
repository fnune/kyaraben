import { type ChangeEvent, useState } from 'react'
import { SystemLogo } from '@/components/SystemLogo/SystemLogo'
import { Modal } from '@/lib/Modal'
import type { EmulatorID, EmulatorRef, ProvisionResult, System, SystemID } from '@/types/daemon'

export interface SystemRowProps {
  readonly system: System
  readonly selectedEmulator: EmulatorID | null
  readonly pinnedVersion: string | null
  readonly installedVersion: string | null
  readonly provisions: readonly ProvisionResult[]
  readonly enabled: boolean
  readonly userStore: string
  readonly emulatorSharedWith: readonly string[]
  readonly emulatorInstalledFor: readonly string[]
  readonly onToggle: (systemId: SystemID, enabled: boolean) => void
  readonly onEmulatorChange: (systemId: SystemID, emulatorId: EmulatorID) => void
  readonly onVersionChange: (systemId: SystemID, version: string | null) => void
}

function ProvisionsBadges({
  provisions,
  onClick,
}: {
  readonly provisions: readonly ProvisionResult[]
  readonly onClick: () => void
}) {
  if (provisions.length === 0) return <div className="ml-1" />

  return (
    <button
      type="button"
      onClick={onClick}
      className="flex flex-nowrap gap-1 py-1 ml-1 overflow-hidden min-[720px]:flex-wrap min-[720px]:overflow-visible cursor-pointer hover:opacity-80 transition-opacity"
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

function EmulatorSelector({
  emulators,
  selected,
  onChange,
  disabled,
}: {
  readonly emulators: readonly EmulatorRef[]
  readonly selected: EmulatorID | null
  readonly onChange: (id: EmulatorID) => void
  readonly disabled: boolean
}) {
  if (emulators.length <= 1) {
    return <span className="text-sm text-gray-600">{emulators[0]?.name ?? 'Unknown'}</span>
  }

  const handleChange = (e: ChangeEvent<HTMLSelectElement>) => {
    onChange(e.target.value as EmulatorID)
  }

  return (
    <select
      value={selected ?? emulators[0]?.id ?? ''}
      onChange={handleChange}
      disabled={disabled}
      onClick={(e) => e.stopPropagation()}
      className="text-sm bg-transparent border-none text-gray-600 focus:outline-none focus:ring-0 cursor-pointer disabled:cursor-default disabled:opacity-50"
    >
      {emulators.map((emu) => (
        <option key={emu.id} value={emu.id}>
          {emu.name}
        </option>
      ))}
    </select>
  )
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

export function SystemRow({
  system,
  selectedEmulator,
  pinnedVersion,
  installedVersion,
  provisions,
  enabled,
  userStore,
  emulatorSharedWith,
  emulatorInstalledFor,
  onToggle,
  onEmulatorChange,
  onVersionChange,
}: SystemRowProps) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const emulator = system.emulators.find((e) => e.id === selectedEmulator) ?? system.emulators[0]
  const effectiveVersion = pinnedVersion ?? emulator?.defaultVersion ?? null
  const action = getAction(
    enabled,
    installedVersion,
    effectiveVersion,
    emulatorSharedWith,
    emulatorInstalledFor,
  )

  const hasMissingRequired = provisions.some((p) => p.required && p.status !== 'found')
  const hasMissingOptional = provisions.some((p) => !p.required && p.status !== 'found')

  const getStatusColor = () => {
    if (action === 'will-uninstall') return 'bg-gray-300'
    if (hasMissingRequired) return 'bg-red-500'
    if (hasMissingOptional) return 'bg-amber-500'
    if (action === 'will-install') return 'bg-blue-500'
    if (action === 'will-update') return 'bg-amber-500'
    if (enabled && installedVersion) return 'bg-green-500'
    return 'bg-gray-300'
  }
  const statusColor = getStatusColor()

  const handleCheckboxChange = (e: ChangeEvent<HTMLInputElement>) => {
    onToggle(system.id, e.target.checked)
  }

  const handleEmulatorChange = (emulatorId: EmulatorID) => {
    onEmulatorChange(system.id, emulatorId)
  }

  const handleVersionChange = (version: string | null) => {
    onVersionChange(system.id, version)
  }

  return (
    <div className="border-b border-gray-100 last:border-b-0 relative">
      <div className={`absolute left-0 top-0 bottom-0 w-1 ${statusColor}`} />

      {/* Mobile layout */}
      <div className="min-[720px]:hidden flex flex-col gap-2 pr-3 py-2 hover:bg-gray-50 transition-colors">
        <div className="flex items-center gap-3">
          <label className="flex items-center justify-center pl-5 pr-3 self-stretch cursor-pointer bg-gray-100/50 border-r border-gray-200">
            <input
              type="checkbox"
              checked={enabled}
              onChange={handleCheckboxChange}
              className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500 cursor-pointer"
            />
          </label>
          <SystemLogo systemId={system.id} systemName={system.name} className="flex-shrink-0" />
          <div className="flex-1 min-w-0" />
          <div className="flex items-center gap-1">
            <EmulatorSelector
              emulators={system.emulators}
              selected={selectedEmulator}
              onChange={handleEmulatorChange}
              disabled={!enabled}
            />
            {emulator && (
              <VersionSelector
                emulator={emulator}
                pinnedVersion={pinnedVersion}
                onChange={handleVersionChange}
                disabled={!enabled}
              />
            )}
          </div>
        </div>
        <div className="flex items-center gap-2 pl-14">
          <ActionLabel
            action={action}
            installedVersion={installedVersion}
            emulatorSharedWith={emulatorSharedWith}
            emulatorInstalledFor={emulatorInstalledFor}
          />
          <ProvisionsBadges provisions={provisions} onClick={() => setDialogOpen(true)} />
        </div>
      </div>

      {/* Desktop layout */}
      <div className="hidden min-[720px]:flex min-h-12 hover:bg-gray-50 transition-colors">
        <label className="w-12 flex items-center justify-center pl-1 cursor-pointer bg-gray-100/50 border-r border-gray-200">
          <input
            type="checkbox"
            checked={enabled}
            onChange={handleCheckboxChange}
            className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500 cursor-pointer"
          />
        </label>

        <div className="flex-1 flex items-center gap-x-3 py-2 pl-3 pr-3">
          <SystemLogo systemId={system.id} systemName={system.name} className="flex-shrink-0" />

          <ProvisionsBadges provisions={provisions} onClick={() => setDialogOpen(true)} />

          <div className="flex-1" />

          <div className="flex items-center gap-1">
            <ActionLabel
              action={action}
              installedVersion={installedVersion}
              emulatorSharedWith={emulatorSharedWith}
              emulatorInstalledFor={emulatorInstalledFor}
            />
            <EmulatorSelector
              emulators={system.emulators}
              selected={selectedEmulator}
              onChange={handleEmulatorChange}
              disabled={!enabled}
            />
            {emulator && (
              <VersionSelector
                emulator={emulator}
                pinnedVersion={pinnedVersion}
                onChange={handleVersionChange}
                disabled={!enabled}
              />
            )}
          </div>
        </div>
      </div>

      <ProvisionsDialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        systemId={system.id}
        systemName={system.name}
        provisions={provisions}
        userStore={userStore}
      />
    </div>
  )
}
