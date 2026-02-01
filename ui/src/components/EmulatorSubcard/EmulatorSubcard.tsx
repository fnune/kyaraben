import { useState } from 'react'
import { ChangeNotch } from '@/components/ChangeNotch/ChangeNotch'
import { PathsModal } from '@/components/PathsModal/PathsModal'
import { CHANGE_CONFIG, getChangeType } from '@/lib/changeUtils'
import { Modal } from '@/lib/Modal'
import { ToggleSwitch } from '@/lib/ToggleSwitch'
import type { EmulatorRef, ProvisionResult, SystemID } from '@/types/daemon'

export interface EmulatorSubcardProps {
  readonly emulator: EmulatorRef
  readonly systemId: SystemID
  readonly enabled: boolean
  readonly pinnedVersion: string | null
  readonly installedVersion: string | null
  readonly provisions: readonly ProvisionResult[]
  readonly userStore: string
  readonly onToggle: (enabled: boolean) => void
  readonly onVersionChange: (version: string | null) => void
}

function ProvisionItem({
  provision,
  onClick,
}: {
  readonly provision: ProvisionResult
  readonly onClick: () => void
}) {
  const isReady = provision.status === 'found'
  return (
    <button
      type="button"
      onClick={onClick}
      className={`
        w-full text-left px-3 py-1.5 text-xs flex items-center gap-2
        hover:bg-gray-100 transition-colors
        ${isReady ? 'text-emerald-600' : 'text-red-500'}
      `}
    >
      <span>{isReady ? '✓' : '✗'}</span>
      <span>{provision.description}</span>
    </button>
  )
}

function ProvisionDialog({
  open,
  onClose,
  emulatorName,
  provision,
  biosPath,
}: {
  readonly open: boolean
  readonly onClose: () => void
  readonly emulatorName: string
  readonly provision: ProvisionResult
  readonly biosPath: string
}) {
  const isReady = provision.status === 'found'

  const handleOpenFolder = () => {
    window.electron.invoke('open_path', biosPath)
  }

  return (
    <Modal open={open} onClose={onClose} title={provision.description}>
      <p className="text-xs text-gray-500 mb-2">{emulatorName}</p>
      <div
        className={`
          rounded-lg p-4 border
          ${isReady ? 'bg-emerald-50 border-emerald-200' : 'bg-gray-50 border-gray-200'}
        `}
      >
        {isReady ? (
          <p className="text-emerald-700 text-sm">Configured correctly</p>
        ) : (
          <div>
            <p className="text-gray-600 text-sm mb-2">Place files in:</p>
            <code className="block bg-white border border-gray-200 px-2 py-1.5 rounded text-xs text-gray-600">
              {biosPath}
            </code>
            <button
              type="button"
              onClick={handleOpenFolder}
              className="mt-3 bg-gray-200 hover:bg-gray-300 text-gray-700 px-3 py-1.5 rounded text-xs transition-colors"
            >
              Open folder
            </button>
          </div>
        )}
      </div>
    </Modal>
  )
}

export function EmulatorSubcard({
  emulator,
  systemId,
  enabled,
  pinnedVersion,
  installedVersion,
  provisions,
  userStore,
  onToggle,
  onVersionChange,
}: EmulatorSubcardProps) {
  const [selectedProvision, setSelectedProvision] = useState<ProvisionResult | null>(null)
  const [pathsOpen, setPathsOpen] = useState(false)

  const effectiveVersion = pinnedVersion ?? emulator.defaultVersion ?? null
  const changeType = getChangeType(
    enabled,
    installedVersion,
    effectiveVersion,
    emulator.availableVersions,
  )
  const isPinned = pinnedVersion !== null

  const cardClasses = (() => {
    if (changeType) {
      const config = CHANGE_CONFIG[changeType]
      return `ring-1 ${config.ringColor} bg-blue-50/30`
    }
    return enabled ? 'bg-gray-50' : 'bg-white'
  })()

  const biosPath = userStore ? `${userStore}/bios/${systemId}` : ''

  return (
    <div className={`rounded-lg overflow-hidden relative ${cardClasses}`}>
      {changeType && <ChangeNotch type={changeType} />}

      <div className={`flex items-center gap-4 p-3 ${!enabled ? 'opacity-50' : ''}`}>
        <div className="flex-1 space-y-1">
          <div className="flex items-center gap-2">
            <span className="text-gray-900 font-medium text-sm">{emulator.name}</span>
            <div className="ml-auto flex items-center gap-1.5">
              {isPinned && <span className="text-amber-500 text-xs">📌</span>}
              <VersionSelector
                defaultVersion={emulator.defaultVersion}
                availableVersions={emulator.availableVersions}
                pinnedVersion={pinnedVersion}
                onChange={onVersionChange}
                disabled={!enabled}
                isPinned={isPinned}
              />
            </div>
          </div>
          <div className="flex items-center gap-2 text-xs text-gray-500">
            {installedVersion && (
              <>
                <button type="button" className="hover:text-gray-700">
                  Launch
                </button>
                <span className="text-gray-300">·</span>
              </>
            )}
            <button
              type="button"
              onClick={() => setPathsOpen(true)}
              className="hover:text-gray-700"
            >
              Paths
            </button>
            {emulator.downloadSize && (
              <>
                <span className="text-gray-300">·</span>
                <span className="text-gray-400">{emulator.downloadSize}</span>
              </>
            )}
          </div>
        </div>

        <ToggleSwitch enabled={enabled} onChange={onToggle} />
      </div>

      {provisions.length > 0 && (
        <div className={`border-t border-gray-200 ${!enabled ? 'opacity-50' : ''}`}>
          {provisions.map((p) => (
            <ProvisionItem key={p.filename} provision={p} onClick={() => setSelectedProvision(p)} />
          ))}
        </div>
      )}

      {selectedProvision && (
        <ProvisionDialog
          open={true}
          onClose={() => setSelectedProvision(null)}
          emulatorName={emulator.name}
          provision={selectedProvision}
          biosPath={biosPath}
        />
      )}

      <PathsModal
        open={pathsOpen}
        onClose={() => setPathsOpen(false)}
        emulatorName={emulator.name}
        emulatorId={emulator.id}
        userStore={userStore}
      />
    </div>
  )
}

function VersionSelector({
  defaultVersion,
  availableVersions,
  pinnedVersion,
  onChange,
  disabled,
  isPinned,
}: {
  readonly defaultVersion: string | undefined
  readonly availableVersions: string[] | undefined
  readonly pinnedVersion: string | null
  readonly onChange: (version: string | null) => void
  readonly disabled: boolean
  readonly isPinned: boolean
}) {
  if (!availableVersions || availableVersions.length === 0) {
    return <span className="text-xs text-gray-400 tabular-nums">{defaultVersion}</span>
  }

  return (
    <select
      value={pinnedVersion ?? ''}
      onChange={(e) => onChange(e.target.value === '' ? null : e.target.value)}
      disabled={disabled}
      className={`
        bg-transparent border border-gray-200 rounded px-2 py-1 text-xs
        outline-none focus:border-blue-400
        ${disabled ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'}
        ${isPinned ? 'text-amber-600' : 'text-gray-600'}
        tabular-nums
      `}
    >
      <option value="">{defaultVersion}</option>
      {availableVersions.map((v) => (
        <option key={v} value={v}>
          {v}
        </option>
      ))}
    </select>
  )
}
