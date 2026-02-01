import { useMemo } from 'react'
import { Settings } from '@/components/Settings/Settings'
import { StickyActionBar } from '@/components/StickyActionBar/StickyActionBar'
import { SYSTEM_YEARS, SystemCard } from '@/components/SystemCard/SystemCard'
import { BottomBar } from '@/lib/BottomBar'
import { Button } from '@/lib/Button'
import { addChange, emptyChangeSummary, getChangeType } from '@/lib/changeUtils'
import { ProgressSteps } from '@/lib/ProgressSteps'
import type { DoctorResponse, EmulatorID, System } from '@/types/daemon'
import type { ApplyStatus, ProgressStep } from '@/types/ui'
import { MANUFACTURER_ORDER } from '@/types/ui'

export interface SystemsViewProps {
  readonly systems: readonly System[]
  readonly enabledEmulators: ReadonlySet<EmulatorID>
  readonly emulatorVersions: Map<EmulatorID, string | null>
  readonly installedVersions: Map<EmulatorID, string>
  readonly installedExecLines: Map<EmulatorID, string>
  readonly provisions: DoctorResponse
  readonly userStore: string
  readonly onUserStoreChange: (value: string) => void
  readonly onEmulatorToggle: (emulatorId: EmulatorID, enabled: boolean) => void
  readonly onVersionChange: (emulatorId: EmulatorID, version: string | null) => void
  readonly onApply: () => void
  readonly onCancel: () => void
  readonly applyStatus: ApplyStatus
  readonly progressSteps: readonly ProgressStep[]
  readonly error: string | null
  readonly onReset: () => void
  readonly onDiscard: () => void
}

function groupSystemsByManufacturer(systems: readonly System[]) {
  const groups = new Map<string, System[]>()

  for (const manufacturer of MANUFACTURER_ORDER) {
    groups.set(manufacturer, [])
  }

  for (const system of systems) {
    const group = groups.get(system.manufacturer) ?? groups.get('Other')
    if (group) {
      group.push(system)
    }
  }

  for (const group of groups.values()) {
    group.sort((a, b) => (SYSTEM_YEARS[a.id] ?? 0) - (SYSTEM_YEARS[b.id] ?? 0))
  }

  return Array.from(groups.entries()).filter(([, systems]) => systems.length > 0)
}

export function SystemsView({
  systems,
  enabledEmulators,
  emulatorVersions,
  installedVersions,
  installedExecLines,
  provisions,
  userStore,
  onUserStoreChange,
  onEmulatorToggle,
  onVersionChange,
  onApply,
  onCancel,
  applyStatus,
  progressSteps,
  error,
  onReset,
  onDiscard,
}: SystemsViewProps) {
  const isApplying = applyStatus === 'applying'
  const showProgress = applyStatus !== 'idle'

  const changes = useMemo(() => {
    let summary = emptyChangeSummary()
    const seenEmulators = new Set<EmulatorID>()

    for (const system of systems) {
      for (const emulator of system.emulators) {
        if (seenEmulators.has(emulator.id)) continue
        seenEmulators.add(emulator.id)

        const enabled = enabledEmulators.has(emulator.id)
        const installedVersion = installedVersions.get(emulator.id) ?? null
        const pinnedVersion = emulatorVersions.get(emulator.id) ?? null
        const effectiveVersion = pinnedVersion ?? emulator.defaultVersion ?? null

        const changeType = getChangeType(
          enabled,
          installedVersion,
          effectiveVersion,
          emulator.availableVersions,
        )

        summary = addChange(summary, changeType, emulator.downloadBytes)
      }
    }

    return summary
  }, [systems, enabledEmulators, emulatorVersions, installedVersions])

  const groupedSystems = useMemo(() => groupSystemsByManufacturer(systems), [systems])

  if (showProgress) {
    const errorMessage = applyStatus === 'error' && error ? error : undefined

    return (
      <div className="p-6 pb-24">
        <ProgressSteps
          steps={progressSteps}
          {...(errorMessage && { error: errorMessage })}
          {...(applyStatus === 'cancelled' && { cancelled: true })}
        />
        <BottomBar>
          <button
            type="button"
            onClick={onCancel}
            disabled={!isApplying}
            className="text-blue-400 hover:text-blue-300 hover:underline text-sm disabled:text-gray-600 disabled:no-underline disabled:cursor-not-allowed"
          >
            Cancel
          </button>
          <Button onClick={onReset} disabled={isApplying}>
            Done
          </Button>
        </BottomBar>
      </div>
    )
  }

  return (
    <div className="p-6 pb-24">
      <Settings userStore={userStore} onUserStoreChange={onUserStoreChange} />

      <div className="space-y-8 mt-6">
        {groupedSystems.map(([manufacturer, manufacturerSystems]) => (
          <section key={manufacturer}>
            <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">
              {manufacturer}
            </h2>
            <div className="space-y-4">
              {manufacturerSystems.map((system) => (
                <SystemCard
                  key={system.id}
                  system={system}
                  enabledEmulators={enabledEmulators}
                  emulatorVersions={emulatorVersions}
                  installedVersions={installedVersions}
                  installedExecLines={installedExecLines}
                  provisions={provisions}
                  userStore={userStore}
                  onEmulatorToggle={onEmulatorToggle}
                  onVersionChange={onVersionChange}
                />
              ))}
            </div>
          </section>
        ))}
      </div>

      <StickyActionBar
        changes={changes}
        onApply={onApply}
        onDiscard={onDiscard}
        applying={isApplying}
      />
    </div>
  )
}
