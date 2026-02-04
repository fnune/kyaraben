import { useMemo } from 'react'
import { Settings } from '@/components/Settings/Settings'
import { StickyActionBar } from '@/components/StickyActionBar/StickyActionBar'
import { SystemCard } from '@/components/SystemCard/SystemCard'
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
  readonly provisions: DoctorResponse
  readonly userStore: string
  readonly onUserStoreChange: (value: string) => void
  readonly onEmulatorToggle: (emulatorId: EmulatorID, enabled: boolean) => void
  readonly onVersionChange: (emulatorId: EmulatorID, version: string | null) => void
  readonly onApply: () => void
  readonly onCancel: () => void
  readonly onError: (message: string) => void
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

  return Array.from(groups.entries()).filter(([, systems]) => systems.length > 0)
}

export function SystemsView({
  systems,
  enabledEmulators,
  emulatorVersions,
  installedVersions,
  provisions,
  userStore,
  onUserStoreChange,
  onEmulatorToggle,
  onVersionChange,
  onApply,
  onCancel,
  onError,
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

        summary = addChange(summary, changeType)
      }
    }

    return summary
  }, [systems, enabledEmulators, emulatorVersions, installedVersions])

  const groupedSystems = useMemo(() => groupSystemsByManufacturer(systems), [systems])

  if (showProgress) {
    const errorMessage = applyStatus === 'error' && error ? error : undefined

    return (
      <div className="p-6">
        <ProgressSteps
          steps={progressSteps}
          {...(errorMessage && { error: errorMessage })}
          {...(applyStatus === 'cancelled' && { cancelled: true })}
        />
        <div className="flex gap-2">
          {isApplying && (
            <Button onClick={onCancel} variant="secondary">
              Cancel
            </Button>
          )}
          {!isApplying && <Button onClick={onReset}>Done</Button>}
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 pb-24">
      <Settings userStore={userStore} onUserStoreChange={onUserStoreChange} onError={onError} />

      <div className="space-y-8 mt-6">
        {groupedSystems.map(([manufacturer, manufacturerSystems]) => (
          <section key={manufacturer}>
            <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wide mb-3">
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
