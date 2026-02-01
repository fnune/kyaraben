import { useState } from 'react'
import { SystemLogo } from '@/components/SystemLogo/SystemLogo'
import { Modal } from '@/lib/Modal'
import type { DoctorResponse, EmulatorID, ProvisionResult, System, SystemID } from '@/types/daemon'

export interface EmulatorListProps {
  readonly systems: readonly System[]
  readonly enabledEmulators: ReadonlySet<EmulatorID>
  readonly emulatorVersions: ReadonlyMap<EmulatorID, string | null>
  readonly installedVersions: ReadonlyMap<EmulatorID, string>
  readonly provisions: DoctorResponse
  readonly userStore: string
  readonly onToggle: (emulatorId: EmulatorID, enabled: boolean) => void
  readonly onVersionChange: (emulatorId: EmulatorID, version: string | null) => void
}

interface EmulatorWithSystems {
  id: EmulatorID
  name: string
  systems: { id: SystemID; name: string; label: string }[]
  defaultVersion: string | undefined
  availableVersions: string[] | undefined
}

function buildEmulatorList(systems: readonly System[]): EmulatorWithSystems[] {
  const emulatorMap = new Map<EmulatorID, EmulatorWithSystems>()

  for (const system of systems) {
    for (const emu of system.emulators) {
      const existing = emulatorMap.get(emu.id)
      if (existing) {
        existing.systems.push({ id: system.id, name: system.name, label: system.label })
      } else {
        emulatorMap.set(emu.id, {
          id: emu.id,
          name: emu.name,
          systems: [{ id: system.id, name: system.name, label: system.label }],
          defaultVersion: emu.defaultVersion,
          availableVersions: emu.availableVersions,
        })
      }
    }
  }

  return Array.from(emulatorMap.values()).sort((a, b) => a.name.localeCompare(b.name))
}

function SystemBadge({ id, name }: { readonly id: SystemID; readonly name: string }) {
  return (
    <div className="w-8 h-6 flex items-center justify-center" title={name}>
      <SystemLogo systemId={id} systemName={name} className="!w-8" />
    </div>
  )
}

function ProvisionsBadges({
  provisions,
  onClick,
}: {
  readonly provisions: readonly ProvisionResult[]
  readonly onClick: () => void
}) {
  if (provisions.length === 0) return null

  const missing = provisions.filter((p) => p.status !== 'found')
  const hasMissingRequired = missing.some((p) => p.required)

  if (missing.length === 0) {
    return (
      <button
        type="button"
        onClick={onClick}
        className="px-1.5 py-0.5 text-xs bg-green-100 text-green-700 rounded hover:bg-green-200 transition-colors"
      >
        {provisions.length} file{provisions.length > 1 ? 's' : ''} ready
      </button>
    )
  }

  const bgColor = hasMissingRequired
    ? 'bg-red-100 hover:bg-red-200'
    : 'bg-amber-100 hover:bg-amber-200'
  const textColor = hasMissingRequired ? 'text-red-700' : 'text-amber-700'

  return (
    <button
      type="button"
      onClick={onClick}
      className={`px-1.5 py-0.5 text-xs ${bgColor} ${textColor} rounded transition-colors`}
    >
      {missing.length} file{missing.length > 1 ? 's' : ''} needed
    </button>
  )
}

function ProvisionsDialog({
  open,
  onClose,
  emulatorName,
  provisions,
  userStore,
  systemId,
}: {
  readonly open: boolean
  readonly onClose: () => void
  readonly emulatorName: string
  readonly provisions: readonly ProvisionResult[]
  readonly userStore: string
  readonly systemId: SystemID
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
    <Modal open={open} onClose={onClose} title={`${emulatorName} files`}>
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
                <span className="text-red-500 text-lg">⚠</span>
              ) : (
                <span className="text-amber-500 text-lg">○</span>
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
              <span className="text-green-500 text-lg">✓</span>
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
    return <span className="text-xs text-gray-400 tabular-nums">{defaultVersion}</span>
  }

  return (
    <select
      value={pinnedVersion ?? ''}
      onChange={(e) => onChange(e.target.value === '' ? null : e.target.value)}
      disabled={disabled}
      className="text-xs bg-transparent border-none text-gray-500 focus:outline-none focus:ring-0 cursor-pointer disabled:cursor-default disabled:opacity-50 tabular-nums"
    >
      <option value="">{defaultVersion}</option>
      {availableVersions.map((v) => (
        <option key={v} value={v}>
          {v} 📌
        </option>
      ))}
    </select>
  )
}

type ActionType = 'will-install' | 'will-update' | 'will-uninstall' | 'installed' | null

function getAction(
  enabled: boolean,
  installedVersion: string | null,
  effectiveVersion: string | null,
): ActionType {
  if (!enabled) {
    return installedVersion ? 'will-uninstall' : null
  }
  if (!installedVersion) {
    return 'will-install'
  }
  if (effectiveVersion && installedVersion !== effectiveVersion) {
    return 'will-update'
  }
  return 'installed'
}

function ActionLabel({
  action,
  installedVersion,
}: {
  action: ActionType
  installedVersion: string | null
}) {
  if (!action) return null

  const config: Record<NonNullable<ActionType>, { text: string; color: string }> = {
    'will-install': { text: 'Will install', color: 'text-blue-600' },
    'will-update': { text: `Update from ${installedVersion}`, color: 'text-amber-600' },
    'will-uninstall': { text: 'Will uninstall', color: 'text-red-600' },
    installed: { text: 'Installed', color: 'text-green-600' },
  }

  const { text, color } = config[action]
  return <span className={`text-xs ${color}`}>{text}</span>
}

function EmulatorRow({
  emulator,
  enabled,
  pinnedVersion,
  installedVersion,
  provisions,
  userStore,
  onToggle,
  onVersionChange,
}: {
  readonly emulator: EmulatorWithSystems
  readonly enabled: boolean
  readonly pinnedVersion: string | null
  readonly installedVersion: string | null
  readonly provisions: readonly ProvisionResult[]
  readonly userStore: string
  readonly onToggle: (enabled: boolean) => void
  readonly onVersionChange: (version: string | null) => void
}) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const effectiveVersion = pinnedVersion ?? emulator.defaultVersion ?? null
  const action = getAction(enabled, installedVersion, effectiveVersion)

  const statusColor = (() => {
    if (action === 'will-install') return 'border-blue-400'
    if (action === 'will-update') return 'border-amber-400'
    if (action === 'will-uninstall') return 'border-red-400'
    if (installedVersion) return 'border-green-400'
    return 'border-gray-200'
  })()

  return (
    <div>
      <label
        className={`flex items-center gap-3 py-2 px-3 bg-white border-l-4 ${statusColor} hover:bg-gray-50 transition-colors cursor-pointer`}
      >
        <input
          type="checkbox"
          checked={enabled}
          onChange={(e) => onToggle(e.target.checked)}
          className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500 cursor-pointer flex-shrink-0"
        />

        <span className="font-medium text-sm text-gray-900 min-w-[120px]">{emulator.name}</span>

        <div className="flex flex-wrap gap-1">
          {emulator.systems.map((sys) => (
            <SystemBadge key={sys.id} id={sys.id} name={sys.name} />
          ))}
        </div>

        <div className="flex-1" />

        <ProvisionsBadges provisions={provisions} onClick={() => setDialogOpen(true)} />

        <ActionLabel action={action} installedVersion={installedVersion} />

        <VersionSelector
          defaultVersion={emulator.defaultVersion}
          availableVersions={emulator.availableVersions}
          pinnedVersion={pinnedVersion}
          onChange={onVersionChange}
          disabled={!enabled}
        />
      </label>

      {provisions.length > 0 && (
        <ProvisionsDialog
          open={dialogOpen}
          onClose={() => setDialogOpen(false)}
          emulatorName={emulator.name}
          provisions={provisions}
          userStore={userStore}
          systemId={emulator.systems[0]?.id ?? ('' as SystemID)}
        />
      )}
    </div>
  )
}

export function EmulatorList({
  systems,
  enabledEmulators,
  emulatorVersions,
  installedVersions,
  provisions,
  userStore,
  onToggle,
  onVersionChange,
}: EmulatorListProps) {
  const emulators = buildEmulatorList(systems)

  return (
    <div className="bg-white rounded-lg border border-gray-200 divide-y divide-gray-100">
      {emulators.map((emulator) => (
        <EmulatorRow
          key={emulator.id}
          emulator={emulator}
          enabled={enabledEmulators.has(emulator.id)}
          pinnedVersion={emulatorVersions.get(emulator.id) ?? null}
          installedVersion={installedVersions.get(emulator.id) ?? null}
          provisions={provisions[emulator.id] ?? []}
          userStore={userStore}
          onToggle={(enabled) => onToggle(emulator.id, enabled)}
          onVersionChange={(version) => onVersionChange(emulator.id, version)}
        />
      ))}
    </div>
  )
}
