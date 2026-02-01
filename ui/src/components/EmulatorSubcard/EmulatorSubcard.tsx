import { useState } from 'react'
import { ChangeNotch } from '@/components/ChangeNotch/ChangeNotch'
import { PathsModal } from '@/components/PathsModal/PathsModal'
import { CHANGE_CONFIG, formatBytes, getChangeType } from '@/lib/changeUtils'
import { useToast } from '@/lib/ToastContext'
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
  readonly execLine?: string
  readonly onToggle: (enabled: boolean) => void
  readonly onVersionChange: (version: string | null) => void
  readonly onLaunch?: () => void
}

const KIND_LABELS: Record<string, string> = {
  bios: 'BIOS',
  keys: 'keys',
  firmware: 'firmware',
}

function ProvisionItem({
  provision,
  provisionPath,
  onOpenFolder,
  onCopy,
}: {
  readonly provision: ProvisionResult
  readonly provisionPath: string
  readonly onOpenFolder: (path: string) => void
  readonly onCopy: (text: string) => void
}) {
  const isReady = provision.status === 'found'
  const isOptional = !provision.required
  const statusColor = isReady ? 'text-emerald-400' : isOptional ? 'text-amber-400' : 'text-red-400'
  const kindLabel = KIND_LABELS[provision.kind] ?? provision.kind

  const handleOpenFolder = () => {
    onOpenFolder(provisionPath)
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(provision.filename)
    onCopy(provision.filename)
  }

  if (isReady) {
    return (
      <div className="flex items-center text-xs px-3 py-1.5">
        <span className={statusColor}>✓</span>
        <span className="ml-2 text-gray-400">
          {kindLabel} ({provision.description})
        </span>
      </div>
    )
  }

  return (
    <div className="flex items-center text-xs px-3 py-1.5 gap-2">
      <span className={statusColor}>✗</span>
      <span className="text-gray-400">
        <span className="hidden md:inline">Missing {isOptional ? 'optional' : 'required'} </span>
        {kindLabel} ({provision.description})
      </span>
      <span className="inline-flex items-center gap-1">
        <code className="text-gray-500">{provision.filename}</code>
        <button
          type="button"
          onClick={handleCopy}
          className="text-gray-600 hover:text-white transition-colors"
          aria-label={`Copy ${provision.filename}`}
        >
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
            />
          </svg>
        </button>
      </span>
      <button
        type="button"
        onClick={handleOpenFolder}
        className="ml-auto flex items-center gap-1.5 text-blue-400 hover:text-blue-300 hover:underline transition-colors shrink-0"
      >
        <span className="hidden md:inline">Place in {provisionPath}</span>
        <svg
          className="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-label={`Open ${provisionPath}`}
          role="img"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z"
          />
        </svg>
      </button>
    </div>
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
  execLine,
  onToggle,
  onVersionChange,
  onLaunch,
}: EmulatorSubcardProps) {
  const [pathsOpen, setPathsOpen] = useState(false)
  const { showToast } = useToast()

  const effectiveVersion = pinnedVersion ?? emulator.defaultVersion ?? null
  const changeType = getChangeType(
    enabled,
    installedVersion,
    effectiveVersion,
    emulator.availableVersions,
  )

  const cardClasses = (() => {
    if (changeType) {
      const config = CHANGE_CONFIG[changeType]
      return `ring-1 ${config.ringColor} bg-gray-800`
    }
    return enabled ? 'bg-gray-800' : 'bg-gray-800/50'
  })()

  const biosPath = userStore ? `${userStore}/bios/${systemId}` : ''

  const handleLaunch = () => {
    if (onLaunch) {
      onLaunch()
    }
  }

  const handleOpenFolder = (path: string) => {
    window.electron.invoke('open_path', path)
    showToast(`Opening ${path}`)
  }

  const handleCopy = (text: string) => {
    showToast(`Copied ${text}`)
  }

  return (
    <div className={`rounded-lg overflow-hidden relative ${cardClasses}`}>
      {changeType && <ChangeNotch type={changeType} />}

      <div className={`flex items-center gap-4 p-3 ${!enabled ? 'opacity-60' : ''}`}>
        <div className="flex-1 space-y-1">
          <div className="flex items-center gap-2">
            <span className="text-white font-medium text-sm">{emulator.name}</span>
            <div className="ml-auto flex items-center gap-1.5">
              <VersionSelector
                defaultVersion={emulator.defaultVersion}
                availableVersions={emulator.availableVersions}
                pinnedVersion={pinnedVersion}
                onChange={onVersionChange}
                disabled={!enabled}
              />
            </div>
          </div>
          <div className="flex items-center gap-2 text-xs text-gray-400">
            {installedVersion ? (
              <>
                {execLine && (
                  <>
                    <button type="button" onClick={handleLaunch} className="hover:text-white">
                      Launch
                    </button>
                    <span className="text-gray-600">·</span>
                  </>
                )}
                <button
                  type="button"
                  onClick={() => setPathsOpen(true)}
                  className="hover:text-white"
                >
                  Paths
                </button>
              </>
            ) : (
              emulator.downloadBytes && (
                <span className="text-gray-500">{formatBytes(emulator.downloadBytes)}</span>
              )
            )}
          </div>
        </div>

        <ToggleSwitch enabled={enabled} onChange={onToggle} />
      </div>

      {provisions.length > 0 && (
        <div className={`border-t border-gray-700/50 ${!enabled ? 'opacity-60' : ''}`}>
          {provisions.map((p) => (
            <ProvisionItem
              key={p.filename}
              provision={p}
              provisionPath={biosPath}
              onOpenFolder={handleOpenFolder}
              onCopy={handleCopy}
            />
          ))}
        </div>
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

  const isPinned = pinnedVersion !== null

  return (
    <select
      value={pinnedVersion ?? ''}
      onChange={(e) => onChange(e.target.value === '' ? null : e.target.value)}
      disabled={disabled}
      className={`
        bg-gray-700 rounded px-2 py-1 text-xs text-gray-200
        outline-none focus:ring-1 focus:ring-blue-400
        ${isPinned ? 'ring-1 ring-amber-500' : 'border border-gray-600'}
        ${disabled ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'}
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
