import { useState } from 'react'
import { ChangeNotch } from '@/components/ChangeNotch/ChangeNotch'
import { getEmulatorLogo } from '@/components/EmulatorLogo/EmulatorLogo'
import { PathsModal } from '@/components/PathsModal/PathsModal'
import { CHANGE_CONFIG, formatBytes, getChangeType } from '@/lib/changeUtils'
import { CopyIcon, FolderIcon, PlayIcon } from '@/lib/icons'
import { useToast } from '@/lib/ToastContext'
import { ToggleSwitch } from '@/lib/ToggleSwitch'
import type { EmulatorPaths, EmulatorRef, ManagedConfigInfo, ProvisionResult } from '@/types/daemon'

export interface EmulatorSubcardProps {
  readonly emulator: EmulatorRef
  readonly enabled: boolean
  readonly pinnedVersion: string | null
  readonly installedVersion: string | null
  readonly provisions: readonly ProvisionResult[]
  readonly managedConfigs?: readonly ManagedConfigInfo[]
  readonly paths?: EmulatorPaths
  readonly execLine?: string
  readonly sharedPackage?: boolean
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
  emulatorName,
  disabled,
  onOpenFolder,
  onCopy,
  onLaunch,
}: {
  readonly provision: ProvisionResult
  readonly emulatorName: string
  readonly disabled: boolean
  readonly onOpenFolder: (path: string) => void
  readonly onCopy: (text: string) => void
  readonly onLaunch?: () => void
}) {
  const isReady = provision.status === 'found'
  const isOptional = !provision.required
  const statusColor = isReady ? 'text-emerald-400' : isOptional ? 'text-amber-400' : 'text-red-400'
  const kindLabel = KIND_LABELS[provision.kind] ?? provision.kind
  const expectedPath = provision.expectedPath ?? ''

  const handleOpenFolder = () => {
    if (!disabled && expectedPath) onOpenFolder(expectedPath)
  }

  const handleCopy = () => {
    if (!disabled) {
      navigator.clipboard.writeText(provision.filename)
      onCopy(provision.filename)
    }
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

  const actionButton = provision.importViaUI ? (
    onLaunch && !disabled ? (
      <button
        type="button"
        onClick={onLaunch}
        className="ml-auto flex items-center gap-1.5 text-blue-400 hover:text-blue-300 hover:underline transition-colors shrink-0"
      >
        <span className="hidden md:inline">Import in {emulatorName}</span>
        <PlayIcon />
      </button>
    ) : (
      <span className="ml-auto text-gray-500 text-xs shrink-0">
        <span className="hidden md:inline">Import in {emulatorName} after install</span>
      </span>
    )
  ) : (
    <button
      type="button"
      onClick={handleOpenFolder}
      disabled={disabled}
      className={`ml-auto flex items-center gap-1.5 text-blue-400 transition-colors shrink-0 ${disabled ? 'opacity-50 cursor-not-allowed' : 'hover:text-blue-300 hover:underline'}`}
      aria-label={`Open ${expectedPath}`}
    >
      <span className="hidden md:inline">Place in {expectedPath}</span>
      <FolderIcon />
    </button>
  )

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
          disabled={disabled}
          className={`text-gray-600 transition-colors ${disabled ? 'cursor-not-allowed' : 'hover:text-white'}`}
          aria-label={`Copy ${provision.filename}`}
        >
          <CopyIcon />
        </button>
      </span>
      {actionButton}
    </div>
  )
}

export function EmulatorSubcard({
  emulator,
  enabled,
  pinnedVersion,
  installedVersion,
  provisions,
  managedConfigs,
  paths,
  execLine,
  sharedPackage,
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

  const handleLaunch = () => {
    if (onLaunch) {
      onLaunch()
      showToast(`Launching ${emulator.name}`)
    }
  }

  const handleOpenFolder = (path: string) => {
    window.electron.invoke('open_path', path)
    showToast(`Opening ${path}`)
  }

  const handleCopy = (text: string) => {
    showToast(`Copied ${text}`)
  }

  const logo = getEmulatorLogo(emulator.id)

  return (
    <div className={`rounded-lg overflow-hidden relative ${cardClasses}`}>
      {changeType && <ChangeNotch type={changeType} />}

      <div className={`flex items-center gap-4 p-3 ${!enabled ? 'opacity-60' : ''}`}>
        {logo && (
          <div className="hidden min-[720px]:flex items-center justify-center w-10 h-10 shrink-0">
            <img src={logo} alt="" className="w-full h-full object-contain" />
          </div>
        )}
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
                    <button
                      type="button"
                      onClick={handleLaunch}
                      disabled={!enabled}
                      className={enabled ? 'hover:text-white' : 'cursor-not-allowed'}
                    >
                      Launch
                    </button>
                    <span className="text-gray-600">·</span>
                  </>
                )}
                {paths && (
                  <button
                    type="button"
                    onClick={() => setPathsOpen(true)}
                    disabled={!enabled}
                    className={enabled ? 'hover:text-white' : 'cursor-not-allowed'}
                  >
                    Paths
                  </button>
                )}
              </>
            ) : (
              (emulator.downloadBytes || emulator.coreBytes) && (
                <span className="text-gray-500">
                  {emulator.downloadBytes ? formatBytes(emulator.downloadBytes) : ''}
                  {sharedPackage && emulator.downloadBytes && (
                    <span className="text-blue-400 ml-1">(shared)</span>
                  )}
                  {emulator.coreBytes && (
                    <span className="ml-1">
                      {emulator.downloadBytes ? '+ ' : ''}
                      {formatBytes(emulator.coreBytes)} core
                    </span>
                  )}
                </span>
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
              emulatorName={emulator.name}
              disabled={!enabled}
              onOpenFolder={handleOpenFolder}
              onCopy={handleCopy}
              {...(execLine && onLaunch && { onLaunch })}
            />
          ))}
        </div>
      )}

      {paths && (
        <PathsModal
          open={pathsOpen}
          onClose={() => setPathsOpen(false)}
          emulatorName={emulator.name}
          paths={paths}
          {...(managedConfigs && { managedConfigs })}
        />
      )}
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
        outline-2 outline-offset-1 focus:outline-solid focus:outline-blue-400
        ${isPinned ? 'ring-2 ring-amber-500' : 'border border-gray-600'}
        ${disabled ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'}
        tabular-nums
      `}
    >
      <option value="">{defaultVersion} (auto)</option>
      {availableVersions.map((v) => (
        <option key={v} value={v}>
          {v} (pin)
        </option>
      ))}
    </select>
  )
}
