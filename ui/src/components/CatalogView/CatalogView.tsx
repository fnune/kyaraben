import { useCallback, useMemo, useRef, useState } from 'react'
import { ConfigDiffReview } from '@/components/ConfigDiffReview/ConfigDiffReview'
import { FrontendCard } from '@/components/FrontendCard/FrontendCard'
import { ManufacturerNav } from '@/components/ManufacturerNav/ManufacturerNav'
import { SearchInput } from '@/components/SearchInput/SearchInput'
import { ControllerSettings } from '@/components/Settings/ControllerSettings'
import { GraphicsSettings } from '@/components/Settings/GraphicsSettings'
import { Settings } from '@/components/Settings/Settings'
import { StickyActionBar } from '@/components/StickyActionBar/StickyActionBar'
import { SYSTEM_YEARS, SystemCard } from '@/components/SystemCard/SystemCard'
import { useApply } from '@/lib/ApplyContext'
import { BottomBar } from '@/lib/BottomBar'
import { Button } from '@/lib/Button'
import {
  addChange,
  calculateEmulatorSizes,
  type EmulatorChangeInput,
  emptyChangeSummary,
  formatChangeSummary,
  getChangeType,
  withConfigChanges,
} from '@/lib/changeUtils'
import { ProgressSteps } from '@/lib/ProgressSteps'
import { useOpenLog } from '@/lib/useOpenLog'
import type {
  DoctorResponse,
  EmulatorID,
  EmulatorPaths,
  FrontendID,
  FrontendRef,
  ManagedConfigInfo,
  System,
  SystemID,
} from '@/types/daemon'
import { MANUFACTURER_ORDER, type Manufacturer } from '@/types/ui'

export interface CatalogViewProps {
  readonly systems: readonly System[]
  readonly frontends: readonly FrontendRef[]
  readonly systemEmulators: Map<SystemID, EmulatorID[]>
  readonly enabledEmulators: ReadonlySet<EmulatorID>
  readonly enabledFrontends: Map<FrontendID, boolean>
  readonly emulatorVersions: Map<EmulatorID, string | null>
  readonly emulatorShaders: Map<EmulatorID, string | null>
  readonly graphics: { shaders: string }
  readonly controller: { nintendoConfirm: string }
  readonly frontendVersions: Map<FrontendID, string | null>
  readonly installedVersions: Map<EmulatorID, string>
  readonly installedFrontendVersions: Map<FrontendID, string>
  readonly installedFrontendExecLines: Map<FrontendID, string>
  readonly installedExecLines: Map<EmulatorID, string>
  readonly managedConfigs: Map<EmulatorID, ManagedConfigInfo[]>
  readonly installedPaths: Map<EmulatorID, Record<string, EmulatorPaths>>
  readonly provisions: DoctorResponse
  readonly collection: string
  readonly configChanges: readonly string[]
  readonly onCollectionChange: (value: string) => void
  readonly onEmulatorToggle: (systemId: SystemID, emulatorId: EmulatorID, enabled: boolean) => void
  readonly onVersionChange: (emulatorId: EmulatorID, version: string | null) => void
  readonly onShaderChange: (emulatorId: EmulatorID, shaders: string | null) => void
  readonly onFrontendToggle: (frontendId: FrontendID, enabled: boolean) => void
  readonly onFrontendVersionChange: (frontendId: FrontendID, version: string | null) => void
  readonly onGraphicsShadersChange: (value: string) => void
  readonly onControllerNintendoConfirmChange: (value: string) => void
  readonly onDiscard: () => void
  readonly onEnableAll: () => void
  readonly upgradeAvailable?: boolean
  readonly onReapply?: () => void
}

function groupSystemsByManufacturer(systems: readonly System[]): [Manufacturer, System[]][] {
  const groups = new Map<Manufacturer, System[]>()

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

const SEARCH_ALIASES: Record<string, readonly string[]> = {
  nes: ['nes', 'famicom'],
  snes: ['snes', 'super famicom', 'sfc'],
  n64: ['n64'],
  gb: ['gameboy', 'game boy', 'dmg'],
  gbc: ['gameboy color', 'game boy color', 'gbc'],
  gba: ['gameboy advance', 'game boy advance', 'gba'],
  nds: ['nds', 'ds', 'nintendo ds'],
  n3ds: ['3ds', 'nintendo 3ds'],
  psx: ['psx', 'ps1', 'playstation 1', 'playstation1'],
  ps2: ['ps2', 'playstation2'],
  ps3: ['ps3', 'playstation3'],
  psp: ['psp'],
  psvita: ['vita', 'ps vita', 'psvita'],
  genesis: ['genesis', 'mega drive', 'megadrive', 'md'],
  mastersystem: ['sms', 'master system'],
  gamegear: ['gg', 'game gear'],
  saturn: ['saturn'],
  dreamcast: ['dc'],
  pcengine: ['pce', 'pc engine', 'turbografx', 'tg16'],
  ngp: ['ngp', 'neo geo pocket', 'neogeo pocket'],
  neogeo: ['neo geo', 'neogeo', 'aes', 'mvs'],
  atari2600: ['2600', 'vcs'],
  c64: ['c64', 'commodore'],
  gamecube: ['gc', 'gcn', 'gamecube'],
  wii: ['wii'],
  wiiu: ['wii u', 'wiiu'],
  switch: ['switch', 'nx'],
  xbox: ['xbox', 'og xbox'],
  xbox360: ['360', 'xbox 360'],
  arcade: ['arcade', 'mame'],
}

function matchesSearch(system: System, query: string): boolean {
  const lowerQuery = query.toLowerCase()
  if (system.name.toLowerCase().includes(lowerQuery)) return true
  if (system.manufacturer.toLowerCase().includes(lowerQuery)) return true
  for (const emulator of system.emulators) {
    if (emulator.name.toLowerCase().includes(lowerQuery)) return true
  }
  const aliases = SEARCH_ALIASES[system.id]
  if (aliases) {
    for (const alias of aliases) {
      if (alias.includes(lowerQuery) || lowerQuery.includes(alias)) return true
    }
  }
  return false
}

export function CatalogView({
  systems,
  frontends,
  systemEmulators,
  enabledEmulators,
  enabledFrontends,
  emulatorVersions,
  emulatorShaders,
  graphics,
  controller,
  frontendVersions,
  installedVersions,
  installedFrontendVersions,
  installedFrontendExecLines,
  installedExecLines,
  managedConfigs,
  installedPaths,
  provisions,
  collection,
  configChanges,
  onCollectionChange,
  onEmulatorToggle,
  onVersionChange,
  onShaderChange,
  onFrontendToggle,
  onFrontendVersionChange,
  onGraphicsShadersChange,
  onControllerNintendoConfirmChange,
  onDiscard,
  onEnableAll,
  upgradeAvailable,
  onReapply,
}: CatalogViewProps) {
  const {
    status: applyStatus,
    progressSteps,
    error,
    preflightData,
    syncPendingData,
    apply,
    confirmApply,
    confirmSyncPending,
    reset,
    logPosition,
  } = useApply()
  const openLog = useOpenLog()
  const isApplying = applyStatus === 'applying'
  const showProgress =
    applyStatus !== 'idle' && applyStatus !== 'reviewing' && applyStatus !== 'confirming_sync'

  const [searchQuery, setSearchQuery] = useState('')
  const stickyHeaderRef = useRef<HTMLDivElement>(null)
  const manufacturerRefs = useRef<Map<Manufacturer, HTMLElement>>(new Map())

  const handleApply = useCallback(
    async (changeSummary: ReturnType<typeof emptyChangeSummary>) => {
      const systemsConfig: Record<string, string[]> = {}
      for (const [sysId, emuIds] of systemEmulators) {
        systemsConfig[sysId] = emuIds
      }

      const emulatorsConfig: Record<string, { version?: string; shaders?: string | null }> = {}
      for (const [emuId, version] of emulatorVersions) {
        if (version) {
          emulatorsConfig[emuId] = { version }
        }
      }
      for (const [emuId, shaders] of emulatorShaders) {
        emulatorsConfig[emuId] = { ...emulatorsConfig[emuId], shaders }
      }

      const frontendsConfig: Record<string, { enabled: boolean; version?: string }> = {}
      for (const [feId, enabled] of enabledFrontends) {
        const version = frontendVersions.get(feId)
        frontendsConfig[feId] = { enabled, ...(version && { version }) }
      }

      const summaryMessage = formatChangeSummary(changeSummary)
      await apply({
        collection,
        systems: systemsConfig,
        emulators: emulatorsConfig,
        frontends: frontendsConfig,
        ...(graphics.shaders && { graphics }),
        ...(controller.nintendoConfirm && { controller }),
        ...(summaryMessage && { summaryMessage }),
      })
    },
    [
      apply,
      systemEmulators,
      emulatorVersions,
      emulatorShaders,
      enabledFrontends,
      frontendVersions,
      collection,
      graphics,
      controller,
    ],
  )

  const changes = useMemo(() => {
    let summary = emptyChangeSummary()
    const seenEmulators = new Set<EmulatorID>()

    interface EmulatorInputWithName extends EmulatorChangeInput {
      name: string
    }

    const emulatorInputs: EmulatorInputWithName[] = []

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

        emulatorInputs.push({
          id: emulator.id,
          name: emulator.name,
          packageName: emulator.packageName ?? emulator.id,
          downloadBytes: emulator.downloadBytes ?? 0,
          coreBytes: emulator.coreBytes ?? 0,
          changeType,
          isInstalled: installedVersion !== null,
        })
      }
    }

    const emulatorSizes = calculateEmulatorSizes(emulatorInputs)
    for (const input of emulatorInputs) {
      summary = addChange(summary, input.changeType, emulatorSizes.get(input.id) ?? 0, {
        id: input.id,
        name: input.name,
      })
    }

    for (const frontend of frontends) {
      const enabled = enabledFrontends.get(frontend.id) ?? false
      const installedVersion = installedFrontendVersions.get(frontend.id) ?? null
      const pinnedVersion = frontendVersions.get(frontend.id) ?? null
      const effectiveVersion = pinnedVersion ?? frontend.defaultVersion ?? null

      const changeType = getChangeType(
        enabled,
        installedVersion,
        effectiveVersion,
        frontend.availableVersions,
      )

      const downloadBytes = frontend.downloadBytes ?? 0
      summary = addChange(summary, changeType, downloadBytes, {
        id: frontend.id,
        name: frontend.name,
      })
    }

    return withConfigChanges(summary, configChanges)
  }, [
    systems,
    frontends,
    enabledEmulators,
    enabledFrontends,
    emulatorVersions,
    frontendVersions,
    installedVersions,
    installedFrontendVersions,
    configChanges,
  ])

  const filteredSystems = useMemo(() => {
    if (searchQuery) {
      return systems.filter((system) => matchesSearch(system, searchQuery))
    }
    return systems
  }, [systems, searchQuery])

  const groupedSystems = useMemo(
    () => groupSystemsByManufacturer(filteredSystems),
    [filteredSystems],
  )

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

  const allManufacturers = useMemo(
    () => groupSystemsByManufacturer(systems).map(([manufacturer]) => manufacturer),
    [systems],
  )

  const enabledManufacturers = useMemo(
    () => new Set(groupedSystems.map(([manufacturer]) => manufacturer)),
    [groupedSystems],
  )

  const handleManufacturerClick = useCallback((manufacturer: Manufacturer) => {
    const element = manufacturerRefs.current.get(manufacturer)
    const header = stickyHeaderRef.current
    const scrollContainer = document.getElementById('main-content')
    if (element && scrollContainer) {
      const headerHeight = header?.offsetHeight ?? 0
      const containerRect = scrollContainer.getBoundingClientRect()
      const elementRect = element.getBoundingClientRect()
      const scrollTop =
        scrollContainer.scrollTop + elementRect.top - containerRect.top - headerHeight - 16
      scrollContainer.scrollTo({ top: scrollTop, behavior: 'smooth' })
    }
  }, [])

  if (applyStatus === 'reviewing' && preflightData) {
    return <ConfigDiffReview data={preflightData} onConfirm={confirmApply} onCancel={reset} />
  }

  if (applyStatus === 'confirming_sync' && syncPendingData) {
    const totalMB = (syncPendingData.totalBytes / (1024 * 1024)).toFixed(1)
    return (
      <div className="p-6">
        <div className="max-w-lg mx-auto bg-surface-alt rounded-card p-6">
          <h2 className="text-lg font-medium text-on-surface mb-4">Synchronization in progress</h2>
          <p className="text-sm text-on-surface-muted mb-4">
            There are {syncPendingData.totalFiles} files ({totalMB} MB) still synchronizing.
            Applying now will temporarily pause synchronization while emulators are configured.
          </p>
          <p className="text-sm text-on-surface-muted mb-6">
            Synchronization will resume automatically after apply completes.
          </p>
          <div className="flex gap-3">
            <button
              type="button"
              onClick={reset}
              className="flex-1 px-4 py-2 text-sm font-medium text-on-surface bg-surface rounded-card hover:bg-outline"
            >
              Wait for synchronization
            </button>
            <button
              type="button"
              onClick={confirmSyncPending}
              className="flex-1 px-4 py-2 text-sm font-medium text-white bg-accent rounded-card hover:bg-accent-hover"
            >
              Apply anyway
            </button>
          </div>
        </div>
      </div>
    )
  }

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
                onClick={() => openLog(logPosition ?? undefined)}
                className="text-on-surface-muted hover:text-on-surface-secondary hover:underline text-sm"
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

  const hasNoResults = searchQuery && filteredSystems.length === 0

  return (
    <div className="pb-24">
      <div className="p-6 pb-0">
        <Settings collection={collection} onCollectionChange={onCollectionChange} />

        <div className="mt-6">
          <GraphicsSettings shaders={graphics.shaders} onShadersChange={onGraphicsShadersChange} />
        </div>

        <div className="mt-6">
          <ControllerSettings
            nintendoConfirm={controller.nintendoConfirm}
            onNintendoConfirmChange={onControllerNintendoConfirmChange}
          />
        </div>

        {frontends.length > 0 && (
          <div className="isolate">
            <div className="mt-6" data-section="frontends">
              <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest">
                Frontends
              </span>
            </div>
            <div className="space-y-3 mt-3">
              {frontends.map((frontend) => (
                <FrontendCard
                  key={frontend.id}
                  frontend={frontend}
                  enabled={enabledFrontends.get(frontend.id) ?? false}
                  pinnedVersion={frontendVersions.get(frontend.id) ?? null}
                  installedVersion={installedFrontendVersions.get(frontend.id) ?? null}
                  execLine={installedFrontendExecLines.get(frontend.id)}
                  onToggle={(enabled) => onFrontendToggle(frontend.id, enabled)}
                  onVersionChange={(version) => onFrontendVersionChange(frontend.id, version)}
                />
              ))}
            </div>
          </div>
        )}

        <div className="mt-6 flex items-center justify-between">
          <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest">
            Systems
          </span>
          <button
            type="button"
            onClick={onEnableAll}
            className="text-sm text-accent hover:text-accent-hover"
            title="Enable all systems with their default emulators"
          >
            Enable all systems
          </button>
        </div>
      </div>

      <div
        ref={stickyHeaderRef}
        className="sticky top-0 z-10 bg-surface border-b border-outline px-6 py-2 space-y-2"
      >
        <ManufacturerNav
          manufacturers={allManufacturers}
          enabledManufacturers={enabledManufacturers}
          onManufacturerClick={handleManufacturerClick}
        />

        <SearchInput
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="Search systems, manufacturers, or emulators..."
        />
      </div>

      <div className="px-6 pt-6 isolate min-h-[calc(100vh-10rem)]">
        {hasNoResults ? (
          <div className="mt-4 text-center text-on-surface-muted">
            <p>No systems match your search.</p>
          </div>
        ) : (
          <div className="space-y-8">
            {groupedSystems.map(([manufacturer, manufacturerSystems]) => (
              <section
                key={manufacturer}
                ref={(el) => {
                  if (el) {
                    manufacturerRefs.current.set(manufacturer, el)
                  } else {
                    manufacturerRefs.current.delete(manufacturer)
                  }
                }}
              >
                <h2 className="font-heading text-sm font-semibold text-on-surface-dim uppercase tracking-widest mb-3 border-l-2 border-accent pl-2">
                  {manufacturer}
                </h2>
                <div className="space-y-4">
                  {manufacturerSystems.map((system) => {
                    const systemEmuIds = systemEmulators.get(system.id) ?? []
                    return (
                      <SystemCard
                        key={system.id}
                        system={system}
                        systemEnabledEmulators={new Set(systemEmuIds)}
                        globalEnabledEmulators={enabledEmulators}
                        emulatorVersions={emulatorVersions}
                        emulatorShaders={emulatorShaders}
                        graphics={graphics}
                        installedVersions={installedVersions}
                        installedExecLines={installedExecLines}
                        managedConfigs={managedConfigs}
                        installedPaths={installedPaths}
                        provisions={provisions}
                        sharedPackages={sharedPackages}
                        onEmulatorToggle={onEmulatorToggle}
                        onVersionChange={onVersionChange}
                        onShaderChange={onShaderChange}
                      />
                    )
                  })}
                </div>
              </section>
            ))}
          </div>
        )}
      </div>

      <StickyActionBar
        changes={changes}
        onApply={handleApply}
        onDiscard={onDiscard}
        applying={isApplying}
        {...(upgradeAvailable && { upgradeAvailable })}
        {...(onReapply && { onReapply })}
      />
    </div>
  )
}
