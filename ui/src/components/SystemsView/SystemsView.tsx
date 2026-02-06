import { useCallback, useMemo } from 'react'
import { Settings } from '@/components/Settings/Settings'
import { StickyActionBar } from '@/components/StickyActionBar/StickyActionBar'
import { SYSTEM_YEARS, SystemCard } from '@/components/SystemCard/SystemCard'
import { useApply } from '@/lib/ApplyContext'
import { BottomBar } from '@/lib/BottomBar'
import { Button } from '@/lib/Button'
import { addChange, emptyChangeSummary, getChangeType } from '@/lib/changeUtils'
import { ProgressSteps } from '@/lib/ProgressSteps'
import { useOpenLog } from '@/lib/useOpenLog'
import type { DoctorResponse, EmulatorID, System, SystemID } from '@/types/daemon'
import { MANUFACTURER_ORDER } from '@/types/ui'

export interface SystemsViewProps {
  readonly systems: readonly System[]
  readonly systemEmulators: Map<SystemID, EmulatorID[]>
  readonly enabledEmulators: ReadonlySet<EmulatorID>
  readonly emulatorVersions: Map<EmulatorID, string | null>
  readonly installedVersions: Map<EmulatorID, string>
  readonly installedExecLines: Map<EmulatorID, string>
  readonly managedConfigs: Map<EmulatorID, string[]>
  readonly provisions: DoctorResponse
  readonly userStore: string
  readonly onUserStoreChange: (value: string) => void
  readonly onEmulatorToggle: (emulatorId: EmulatorID, enabled: boolean) => void
  readonly onVersionChange: (emulatorId: EmulatorID, version: string | null) => void
  readonly onDiscard: () => void
  readonly onEnableAll: () => void
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
  systemEmulators,
  enabledEmulators,
  emulatorVersions,
  installedVersions,
  installedExecLines,
  managedConfigs,
  provisions,
  userStore,
  onUserStoreChange,
  onEmulatorToggle,
  onVersionChange,
  onDiscard,
  onEnableAll,
}: SystemsViewProps) {
  const { status: applyStatus, progressSteps, error, apply, reset } = useApply()
  const handleOpenLog = useOpenLog()
  const isApplying = applyStatus === 'applying'
  const showProgress = applyStatus !== 'idle'

  const handleApply = useCallback(async () => {
    const systemsConfig: Record<string, string[]> = {}
    for (const [sysId, emuIds] of systemEmulators) {
      systemsConfig[sysId] = emuIds
    }

    const emulatorsConfig: Record<string, { version?: string }> = {}
    for (const [emuId, version] of emulatorVersions) {
      if (version) {
        emulatorsConfig[emuId] = { version }
      }
    }

    await apply({
      userStore,
      systems: systemsConfig,
      emulators: emulatorsConfig,
    })
  }, [apply, systemEmulators, emulatorVersions, userStore])

  const changes = useMemo(() => {
    let summary = emptyChangeSummary()
    const seenEmulators = new Set<EmulatorID>()
    const seenPackages = new Set<string>()

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

        const packageName = emulator.packageName ?? emulator.id
        const isNewPackage = !seenPackages.has(packageName)
        if (isNewPackage && (changeType === 'install' || changeType === 'upgrade')) {
          seenPackages.add(packageName)
        }

        const packageBytes = isNewPackage ? (emulator.downloadBytes ?? 0) : 0
        const coreBytes = emulator.coreBytes ?? 0
        summary = addChange(summary, changeType, packageBytes + coreBytes)
      }
    }

    return summary
  }, [systems, enabledEmulators, emulatorVersions, installedVersions])

  const groupedSystems = useMemo(() => groupSystemsByManufacturer(systems), [systems])

  const sharedPackages = useMemo(() => {
    const packageSystems = new Map<string, Set<string>>()

    for (const system of systems) {
      for (const emulator of system.emulators) {
        if (!enabledEmulators.has(emulator.id)) continue

        const packageName = emulator.packageName ?? emulator.id
        const systemIds = packageSystems.get(packageName) ?? new Set<string>()
        systemIds.add(system.id)
        packageSystems.set(packageName, systemIds)
      }
    }

    const shared = new Set<string>()
    for (const [pkg, systemIds] of packageSystems) {
      if (systemIds.size > 1) shared.add(pkg)
    }
    return shared
  }, [systems, enabledEmulators])

  if (showProgress) {
    const errorMessage = applyStatus === 'error' && error ? error : undefined
    const isDone =
      applyStatus === 'success' || applyStatus === 'error' || applyStatus === 'cancelled'

    return (
      <div className="p-6 pb-24">
        <ProgressSteps
          steps={progressSteps}
          {...(errorMessage && { error: errorMessage })}
          {...(applyStatus === 'cancelled' && { cancelled: true })}
        />
        {isDone && (
          <BottomBar>
            <span />
            <div className="flex items-center gap-4">
              <button
                type="button"
                onClick={handleOpenLog}
                className="text-gray-400 hover:text-gray-300 hover:underline text-sm"
              >
                Open log in terminal
              </button>
              <Button onClick={reset}>Done</Button>
            </div>
          </BottomBar>
        )}
      </div>
    )
  }

  return (
    <div className="p-6 pb-24">
      <Settings userStore={userStore} onUserStoreChange={onUserStoreChange} />

      <div className="mt-6 flex items-center justify-between">
        <span className="text-sm font-medium text-gray-300">Emulators</span>
        <button
          type="button"
          onClick={onEnableAll}
          className="text-sm text-blue-400 hover:text-blue-300"
          title="Enable all systems with their default emulators"
        >
          Enable all
        </button>
      </div>

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
                  managedConfigs={managedConfigs}
                  provisions={provisions}
                  userStore={userStore}
                  sharedPackages={sharedPackages}
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
        onApply={handleApply}
        onDiscard={onDiscard}
        applying={isApplying}
      />
    </div>
  )
}
