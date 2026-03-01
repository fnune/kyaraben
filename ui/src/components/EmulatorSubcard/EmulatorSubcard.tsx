import { useState } from 'react'
import { ChangeNotch } from '@/components/ChangeNotch/ChangeNotch'
import { EmulatorLogo, getEmulatorLogo } from '@/components/EmulatorLogo/EmulatorLogo'
import { EmulatorSettingsModal } from '@/components/EmulatorSettingsModal/EmulatorSettingsModal'
import { PathsModal } from '@/components/PathsModal/PathsModal'
import { ProvisionSummary } from '@/components/ProvisionSummary/ProvisionSummary'
import {
  getKindLabel,
  ProvisionActionInline,
  ProvisionsModal,
} from '@/components/ProvisionsModal/ProvisionsModal'
import { CHANGE_CONFIG, formatBytes, getChangeType } from '@/lib/changeUtils'
import { PathText } from '@/lib/PathText'
import { Select } from '@/lib/Select'
import { useToast } from '@/lib/ToastContext'
import { ToggleSwitch } from '@/lib/ToggleSwitch'
import type {
  EmulatorPaths,
  EmulatorRef,
  ManagedConfigInfo,
  ProvisionResult,
  SystemID,
} from '@/types/daemon'

function isNonEmpty<T>(arr: readonly T[]): arr is readonly [T, ...T[]] {
  return arr.length > 0
}

export interface EmulatorSubcardProps {
  readonly emulator: EmulatorRef
  readonly systemId: SystemID
  readonly enabled: boolean
  readonly enabledElsewhere?: boolean
  readonly pinnedVersion: string | null
  readonly installedVersion: string | null
  readonly provisions: readonly ProvisionResult[]
  readonly managedConfigs?: readonly ManagedConfigInfo[]
  readonly paths?: EmulatorPaths
  readonly execLine?: string
  readonly sharedPackage?: boolean
  readonly shaders?: string | null
  readonly graphics: { shaders: string }
  readonly onToggle: (enabled: boolean) => void
  readonly onVersionChange: (version: string | null) => void
  readonly onShaderChange?: (value: string | null) => void
  readonly onLaunch?: () => void
}

function ProvisionsSummary({
  provisions,
  disabled,
  onOpenFolder,
  onClick,
  onLaunch,
}: {
  readonly provisions: readonly [ProvisionResult, ...ProvisionResult[]]
  readonly disabled: boolean
  readonly onOpenFolder: (path: string) => void
  readonly onClick: () => void
  readonly onLaunch?: () => void
}) {
  const unsatisfiedRequired = provisions.filter(
    (p) => p.status !== 'found' && p.groupRequired && !p.groupSatisfied,
  )

  const firstUnsatisfied = unsatisfiedRequired[0]
  const hasError = firstUnsatisfied !== undefined

  const getText = () => {
    if (hasError) {
      const kindLabel = getKindLabel(firstUnsatisfied.kind)
      const label = firstUnsatisfied.description
        ? `${kindLabel} (${firstUnsatisfied.description})`
        : kindLabel
      const statusLabel = firstUnsatisfied.groupSize > 1 ? 'at least one required' : 'required'
      return (
        <>
          <span className="text-on-surface-muted truncate">{label}</span>
          <span className="hidden md:inline text-on-surface-dim">, {statusLabel}</span>
          {unsatisfiedRequired.length > 1 && (
            <span className="text-on-surface-dim shrink-0 ml-2">
              +{unsatisfiedRequired.length - 1}
            </span>
          )}
        </>
      )
    }
    return null
  }

  const actionProvision = firstUnsatisfied ?? provisions[0]

  return (
    <button
      type="button"
      onClick={onClick}
      className="flex items-center text-xs px-3 py-1.5 w-full hover:bg-surface-raised/50 transition-colors text-left"
    >
      {hasError ? (
        <ProvisionSummary provisions={provisions} overrideLabel={getText()} />
      ) : (
        <ProvisionSummary provisions={provisions} />
      )}
      <ProvisionActionInline
        provision={actionProvision}
        disabled={disabled}
        onOpenFolder={onOpenFolder}
        {...(onLaunch && hasError && { onLaunch })}
      />
    </button>
  )
}

export function EmulatorSubcard({
  emulator,
  systemId,
  enabled,
  enabledElsewhere,
  pinnedVersion,
  installedVersion,
  provisions,
  managedConfigs,
  paths,
  execLine,
  sharedPackage,
  shaders,
  graphics,
  onToggle,
  onVersionChange,
  onShaderChange,
  onLaunch,
}: EmulatorSubcardProps) {
  const [pathsOpen, setPathsOpen] = useState(false)
  const [provisionsOpen, setProvisionsOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const { showToast } = useToast()

  const hasSettings = (emulator.supportedSettings?.length ?? 0) > 0
  const supportsShaders = emulator.supportedSettings?.includes('shaders') ?? false

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
      showToast(`Launching ${emulator.name}.`)
    }
  }

  const handleOpenFolder = (path: string) => {
    window.electron.invoke('open_path', path)
    showToast(
      <span>
        Opening <PathText>{path}</PathText>.
      </span>,
    )
  }

  const logo = getEmulatorLogo(emulator.id)

  return (
    <div className={`rounded-element overflow-hidden relative ${cardClasses}`}>
      {changeType && <ChangeNotch type={changeType} />}

      <div className="flex items-center gap-4 p-3">
        {logo && (
          <EmulatorLogo
            emulatorId={emulator.id}
            emulatorName={emulator.name}
            className="hidden min-[720px]:flex items-center justify-center w-10 h-10 shrink-0"
          />
        )}
        <div className="flex-1 space-y-0.5">
          <div className="flex flex-col gap-2 min-[400px]:flex-row min-[400px]:items-center">
            <span className="text-on-surface font-medium text-sm">{emulator.name}</span>
            <div className="flex items-center gap-3 min-[400px]:ml-auto">
              <VersionSelector
                defaultVersion={emulator.defaultVersion}
                availableVersions={emulator.availableVersions}
                pinnedVersion={pinnedVersion}
                onChange={onVersionChange}
                disabled={!enabled}
              />
              <ToggleSwitch enabled={enabled} onChange={onToggle} />
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
                {hasSettings && (
                  <>
                    {(paths || provisions.length > 0) && (
                      <span className="text-on-surface-faint">·</span>
                    )}
                    <button
                      type="button"
                      onClick={() => setSettingsOpen(true)}
                      disabled={!enabled}
                      className={enabled ? 'hover:text-accent' : 'cursor-not-allowed'}
                    >
                      Settings
                    </button>
                  </>
                )}
                {!enabled && enabledElsewhere && (
                  <>
                    <span className="text-on-surface-faint">·</span>
                    <span className="text-on-surface-dim">used by other systems</span>
                  </>
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
      </div>

      {isNonEmpty(provisions) && (
        <div className="border-t border-outline/50">
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

      {hasSettings && (
        <EmulatorSettingsModal
          open={settingsOpen}
          onClose={() => setSettingsOpen(false)}
          emulatorId={emulator.id}
          emulatorName={emulator.name}
          systemId={systemId}
          supportsShaders={supportsShaders}
          shaders={shaders ?? null}
          graphics={graphics}
          onShaderChange={onShaderChange ?? ((_: string | null) => undefined)}
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
  const options = [
    { value: '', label: `${defaultVersion} (auto)` },
    ...availableVersions.map((v) => ({ value: v, label: `${v} (pin)` })),
  ]

  return (
    <Select
      value={pinnedVersion ?? ''}
      options={options}
      onChange={(v) => onChange(v === '' ? null : v)}
      disabled={disabled}
      className={isPinned ? '[&>button]:ring-2 [&>button]:ring-status-warning' : ''}
    />
  )
}
