import { useState } from 'react'
import { ChangeNotch } from '@/components/ChangeNotch/ChangeNotch'
import { getEmulatorLogo } from '@/components/EmulatorLogo/EmulatorLogo'
import { PathsModal } from '@/components/PathsModal/PathsModal'
import {
  getKindLabel,
  ProvisionActionInline,
  ProvisionsModal,
} from '@/components/ProvisionsModal/ProvisionsModal'
import { CHANGE_CONFIG, formatBytes, getChangeType } from '@/lib/changeUtils'
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

function ProvisionsSummary({
  provisions,
  disabled,
  onOpenFolder,
  onClick,
  onLaunch,
}: {
  readonly provisions: readonly ProvisionResult[]
  readonly disabled: boolean
  readonly onOpenFolder: (path: string) => void
  readonly onClick: () => void
  readonly onLaunch?: () => void
}) {
  const found = provisions.filter((p) => p.status === 'found')
  const missing = provisions.filter((p) => p.status !== 'found')
  const unsatisfiedRequired = missing.filter((p) => p.groupRequired && !p.groupSatisfied)

  const firstFound = found[0]
  const firstUnsatisfied = unsatisfiedRequired[0]

  if (missing.length === 0) {
    return (
      <button
        type="button"
        onClick={onClick}
        className="flex items-center text-xs px-3 py-1.5 w-full h-full hover:bg-surface-raised/50 transition-colors"
      >
        <span className="text-emerald-400">✓</span>
        <span className="ml-2 text-on-surface-muted">
          {provisions.length} file{provisions.length > 1 ? 's' : ''} ready
        </span>
      </button>
    )
  }

  if (firstUnsatisfied) {
    const kindLabel = getKindLabel(firstUnsatisfied.kind)
    const label = firstUnsatisfied.description
      ? `${kindLabel} (${firstUnsatisfied.description})`
      : kindLabel
    const statusLabel = firstUnsatisfied.groupSize > 1 ? 'at least one required' : 'required'

    return (
      <button
        type="button"
        onClick={onClick}
        className="flex items-center text-xs px-3 py-1.5 w-full hover:bg-surface-raised/50 transition-colors text-left"
      >
        <span className="text-red-400">✗</span>
        <span className="text-on-surface-muted truncate ml-2">
          {label}
          <span className="hidden md:inline text-on-surface-dim">, {statusLabel}</span>
        </span>
        {unsatisfiedRequired.length > 1 && (
          <span className="text-on-surface-dim shrink-0 ml-2">
            +{unsatisfiedRequired.length - 1}
          </span>
        )}
        <ProvisionActionInline
          provision={firstUnsatisfied}
          disabled={disabled}
          onOpenFolder={onOpenFolder}
          {...(onLaunch && { onLaunch })}
        />
      </button>
    )
  }

  if (firstFound) {
    const kindLabel = getKindLabel(firstFound.kind)
    const label = firstFound.description ? `${kindLabel} (${firstFound.description})` : kindLabel
    const optionalCount = missing.filter((p) => !p.groupRequired).length

    return (
      <button
        type="button"
        onClick={onClick}
        className="flex items-center text-xs px-3 py-1.5 w-full hover:bg-surface-raised/50 transition-colors text-left"
      >
        <span className="text-emerald-400">✓</span>
        <span className="text-on-surface-muted truncate ml-2">{label}</span>
        {found.length > 1 && (
          <span className="text-on-surface-dim shrink-0 ml-2">+{found.length - 1}</span>
        )}
        {optionalCount > 0 && (
          <span className="text-amber-400 shrink-0 ml-2">({optionalCount} optional)</span>
        )}
        <ProvisionActionInline
          provision={firstFound}
          disabled={disabled}
          onOpenFolder={onOpenFolder}
        />
      </button>
    )
  }

  return null
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
  const [provisionsOpen, setProvisionsOpen] = useState(false)
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
      return `ring-1 ${config.ringColor} bg-surface-alt`
    }
    return enabled ? 'bg-surface-alt' : 'bg-surface-alt/50'
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

  const logo = getEmulatorLogo(emulator.id)

  return (
    <div className={`rounded-element overflow-hidden relative ${cardClasses}`}>
      {changeType && <ChangeNotch type={changeType} />}

      <div className={`flex items-center gap-4 p-3 ${!enabled ? 'opacity-60' : ''}`}>
        {logo && (
          <div className="hidden min-[720px]:flex items-center justify-center w-10 h-10 shrink-0">
            <img src={logo} alt="" className="w-full h-full object-contain" />
          </div>
        )}
        <div className="flex-1 space-y-1">
          <div className="flex items-center gap-2">
            <span className="text-on-surface font-medium text-sm">{emulator.name}</span>
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
          <div className="flex items-center gap-2 text-xs text-on-surface-muted">
            {installedVersion ? (
              <>
                {execLine && (
                  <>
                    <button
                      type="button"
                      onClick={handleLaunch}
                      disabled={!enabled}
                      className={enabled ? 'hover:text-accent' : 'cursor-not-allowed'}
                    >
                      Launch
                    </button>
                    <span className="text-on-surface-faint">·</span>
                  </>
                )}
                {paths && (
                  <>
                    <button
                      type="button"
                      onClick={() => setPathsOpen(true)}
                      disabled={!enabled}
                      className={enabled ? 'hover:text-accent' : 'cursor-not-allowed'}
                    >
                      Paths
                    </button>
                    {provisions.length > 0 && <span className="text-on-surface-faint">·</span>}
                  </>
                )}
                {provisions.length > 0 && (
                  <button
                    type="button"
                    onClick={() => setProvisionsOpen(true)}
                    disabled={!enabled}
                    className={enabled ? 'hover:text-accent' : 'cursor-not-allowed'}
                  >
                    Provisions
                  </button>
                )}
              </>
            ) : (
              (emulator.downloadBytes || emulator.coreBytes) && (
                <span className="text-on-surface-dim">
                  {emulator.downloadBytes ? formatBytes(emulator.downloadBytes) : ''}
                  {sharedPackage && emulator.downloadBytes && (
                    <span className="text-accent ml-1">(shared)</span>
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
        <div className={`border-t border-outline/50 ${!enabled ? 'opacity-60' : ''}`}>
          <ProvisionsSummary
            provisions={provisions}
            disabled={!enabled}
            onOpenFolder={handleOpenFolder}
            onClick={() => setProvisionsOpen(true)}
            {...(execLine && onLaunch && { onLaunch: handleLaunch })}
          />
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

      {provisions.length > 0 && (
        <ProvisionsModal
          open={provisionsOpen}
          onClose={() => setProvisionsOpen(false)}
          emulatorName={emulator.name}
          provisions={provisions}
          disabled={!enabled}
          {...(execLine && onLaunch && { onLaunch: handleLaunch })}
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
    return (
      <span className="text-xs text-on-surface-muted tabular-nums font-mono">{defaultVersion}</span>
    )
  }

  const isPinned = pinnedVersion !== null

  return (
    <select
      value={pinnedVersion ?? ''}
      onChange={(e) => onChange(e.target.value === '' ? null : e.target.value)}
      disabled={disabled}
      className={`
        bg-surface-raised rounded-control px-2 py-1 text-xs text-on-surface-secondary
        outline-2 outline-offset-1 focus:outline-solid focus:outline-accent
        ${isPinned ? 'ring-2 ring-amber-500' : 'border border-outline-strong'}
        ${disabled ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'}
        tabular-nums font-mono
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
