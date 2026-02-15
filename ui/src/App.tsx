import { useCallback, useEffect, useRef, useState } from 'react'
import { ApplyAfterUpdateBanner } from '@/components/ApplyAfterUpdateBanner/ApplyAfterUpdateBanner'
import { ApplyProgressBar } from '@/components/ApplyProgressBar/ApplyProgressBar'
import { InstallationView } from '@/components/InstallationView/InstallationView'
import { Sidebar } from '@/components/Sidebar/Sidebar'
import { SyncView } from '@/components/SyncView/SyncView'
import { SystemsView } from '@/components/SystemsView/SystemsView'
import { UpdateBanner } from '@/components/UpdateBanner/UpdateBanner'
import { ApplyProvider, useApply } from '@/lib/ApplyContext'
import { BottomBarSlot, BottomBarSlotProvider } from '@/lib/BottomBarSlot'
import * as daemon from '@/lib/daemon'
import { useOnWindowFocus } from '@/lib/hooks/useOnWindowFocus'
import { useUpdateChecker } from '@/lib/hooks/useUpdateChecker'
import { type FoundProvision, getNewlyFoundProvisions } from '@/lib/provisions'
import { Spinner } from '@/lib/Spinner'
import { ToastProvider, useToast } from '@/lib/ToastContext'
import type {
  ConfigResponse,
  DoctorResponse,
  EmulatorID,
  EmulatorPaths,
  FrontendID,
  FrontendRef,
  ManagedConfigInfo,
  StatusResponse,
  SyncStatusResponse,
  System,
  SystemID,
} from '@/types/daemon'
import type { ApplyStatus, View } from '@/types/ui'

function parseStatusResponse(data: StatusResponse) {
  const versions = new Map<EmulatorID, string>()
  const execLines = new Map<EmulatorID, string>()
  const configs = new Map<EmulatorID, ManagedConfigInfo[]>()
  const paths = new Map<EmulatorID, Record<string, EmulatorPaths>>()
  for (const emu of data.installedEmulators ?? []) {
    versions.set(emu.id, emu.version)
    if (emu.execLine) {
      execLines.set(emu.id, emu.execLine)
    }
    if (emu.managedConfigs) {
      configs.set(emu.id, emu.managedConfigs)
    }
    if (emu.paths) {
      paths.set(emu.id, emu.paths)
    }
  }

  const feVersions = new Map<FrontendID, string>()
  for (const fe of data.installedFrontends ?? []) {
    feVersions.set(fe.id, fe.version)
  }

  return { versions, execLines, configs, feVersions, paths }
}

interface ConfigState {
  userStore: string
  systemEmulators: Map<SystemID, EmulatorID[]>
  emulatorVersions: Map<EmulatorID, string | null>
  enabledFrontends: Map<FrontendID, boolean>
  frontendVersions: Map<FrontendID, string | null>
}

function emptyConfigState(): ConfigState {
  return {
    userStore: '',
    systemEmulators: new Map(),
    emulatorVersions: new Map(),
    enabledFrontends: new Map(),
    frontendVersions: new Map(),
  }
}

function keyForProvision(provision: FoundProvision) {
  return provision.id
}

function cloneConfigState(state: ConfigState): ConfigState {
  return {
    userStore: state.userStore,
    systemEmulators: new Map(state.systemEmulators),
    emulatorVersions: new Map(state.emulatorVersions),
    enabledFrontends: new Map(state.enabledFrontends),
    frontendVersions: new Map(state.frontendVersions),
  }
}

function parseConfigResponse(data: ConfigResponse): ConfigState {
  const systemEmulators = new Map<SystemID, EmulatorID[]>()
  const emulatorVersions = new Map<EmulatorID, string | null>()
  const enabledFrontends = new Map<FrontendID, boolean>()
  const frontendVersions = new Map<FrontendID, string | null>()

  for (const [sysId, emulatorIds] of Object.entries(data.systems)) {
    if (emulatorIds && emulatorIds.length > 0) {
      systemEmulators.set(sysId as SystemID, emulatorIds as EmulatorID[])
    }
  }

  if (data.emulators) {
    for (const [emuId, conf] of Object.entries(data.emulators)) {
      if (conf.version) {
        emulatorVersions.set(emuId as EmulatorID, conf.version)
      }
    }
  }

  if (data.frontends) {
    for (const [feId, conf] of Object.entries(data.frontends)) {
      enabledFrontends.set(feId as FrontendID, conf.enabled)
      if (conf.version) {
        frontendVersions.set(feId as FrontendID, conf.version)
      }
    }
  }

  return {
    userStore: data.userStore,
    systemEmulators,
    emulatorVersions,
    enabledFrontends,
    frontendVersions,
  }
}

function AppContent() {
  const [currentView, setCurrentView] = useState<View>('systems')
  const [systems, setSystems] = useState<readonly System[]>([])
  const [frontends, setFrontends] = useState<readonly FrontendRef[]>([])
  const [installedVersions, setInstalledVersions] = useState<Map<EmulatorID, string>>(new Map())
  const [installedFrontendVersions, setInstalledFrontendVersions] = useState<
    Map<FrontendID, string>
  >(new Map())
  const [installedExecLines, setInstalledExecLines] = useState<Map<EmulatorID, string>>(new Map())
  const [managedConfigs, setManagedConfigs] = useState<Map<EmulatorID, ManagedConfigInfo[]>>(
    new Map(),
  )
  const [installedPaths, setInstalledPaths] = useState<
    Map<EmulatorID, Record<string, EmulatorPaths>>
  >(new Map())
  const [provisions, setProvisions] = useState<DoctorResponse>({})
  const [configState, setConfigState] = useState<ConfigState>(emptyConfigState)
  const [configReady, setConfigReady] = useState(false)
  const [syncStatus, setSyncStatus] = useState<SyncStatusResponse | null>(null)

  const savedConfigState = useRef<ConfigState>(emptyConfigState())

  const { onCompleteRef } = useApply()
  const { showToast } = useToast()
  const { status: applyStatus } = useApply()
  const lastApplyStatus = useRef<ApplyStatus | null>(null)
  const seenNotifications = useRef(new Set<string>())

  const {
    updateInfo,
    isDownloading,
    downloadProgress,
    updateDismissed,
    showApplyBanner,
    applyBannerDismissed,
    handleUpdate,
    handleDismissUpdate,
    handleDismissApplyBanner,
    handleApplyFromBanner,
    setShowApplyBanner,
    clearApplyBannerDismissal,
  } = useUpdateChecker(showToast, setCurrentView)

  const refreshAfterApply = useCallback(async () => {
    const [doctorResult, statusResult, configResult] = await Promise.all([
      daemon.runDoctor(),
      daemon.getStatus(),
      daemon.getConfig(),
    ])

    if (doctorResult.ok) {
      setProvisions(doctorResult.data)
    }

    if (statusResult.ok) {
      const { versions, execLines, configs, feVersions, paths } = parseStatusResponse(
        statusResult.data,
      )
      setInstalledVersions(versions)
      setInstalledExecLines(execLines)
      setManagedConfigs(configs)
      setInstalledFrontendVersions(feVersions)
      setInstalledPaths(paths)
    }

    if (configResult.ok) {
      const parsed = parseConfigResponse(configResult.data)
      savedConfigState.current = cloneConfigState(parsed)
    }
  }, [])

  useEffect(() => {
    onCompleteRef.current = () => {
      refreshAfterApply()
      clearApplyBannerDismissal()
    }
  }, [onCompleteRef, refreshAfterApply, clearApplyBannerDismissal])

  useEffect(() => {
    async function init() {
      const [systemsResult, frontendsResult, configResult, statusResult] = await Promise.all([
        daemon.getSystems(),
        daemon.getFrontends(),
        daemon.getConfig(),
        daemon.getStatus(),
      ])

      if (systemsResult.ok) {
        setSystems(systemsResult.data)
      }

      if (frontendsResult.ok) {
        setFrontends(frontendsResult.data)
      }

      if (configResult.ok) {
        const parsed = parseConfigResponse(configResult.data)
        setConfigState(parsed)
        savedConfigState.current = cloneConfigState(parsed)
        setConfigReady(true)
      } else {
        showToast('Failed to load configuration.', 'error')
      }

      if (statusResult.ok) {
        const { versions, execLines, configs, feVersions, paths } = parseStatusResponse(
          statusResult.data,
        )
        setInstalledVersions(versions)
        setInstalledExecLines(execLines)
        setManagedConfigs(configs)
        setInstalledFrontendVersions(feVersions)
        setInstalledPaths(paths)

        if (statusResult.data.healthWarning === 'orphaned_artifacts') {
          showToast(
            <span>
              Installation state may be corrupted.{' '}
              <button
                type="button"
                className="underline hover:no-underline"
                onClick={() => setCurrentView('installation')}
              >
                See details
              </button>
            </span>,
            'error',
          )
        }

        const running = statusResult.data.kyarabenVersion
        const manifest = statusResult.data.manifestKyarabenVersion
        if (running && manifest && running !== manifest) {
          setShowApplyBanner(true)
        }
      }

      const [doctorResult, syncResult] = await Promise.all([
        daemon.runDoctor(),
        daemon.getSyncStatus(),
      ])

      if (doctorResult.ok) {
        setProvisions(doctorResult.data)
      }

      if (syncResult.ok) {
        setSyncStatus(syncResult.data)
      }
    }

    init()
  }, [showToast, setShowApplyBanner])

  useEffect(() => {
    if (applyStatus === lastApplyStatus.current) return
    if (applyStatus === 'success') {
      if (currentView !== 'systems') {
        showToast(
          <span>
            Installation complete.{' '}
            <button
              type="button"
              className="underline hover:no-underline"
              onClick={() => setCurrentView('systems')}
            >
              Go to systems
            </button>
          </span>,
          'success',
          8000,
        )
      } else {
        showToast('Installation complete.', 'success')
      }
    }
    lastApplyStatus.current = applyStatus
  }, [applyStatus, currentView, showToast])

  useOnWindowFocus(async () => {
    const result = await daemon.runDoctor()
    if (result.ok) {
      setProvisions((prev) => {
        const newlyFound = getNewlyFoundProvisions(prev, result.data)
        const unseen = newlyFound.filter((prov) => {
          const key = keyForProvision(prov)
          if (seenNotifications.current.has(key)) {
            return false
          }
          seenNotifications.current.add(key)
          return true
        })
        if (unseen.length > 0) {
          const descriptions = unseen
            .map((prov) => {
              if (prov.description && prov.description !== prov.displayName) {
                return `${prov.displayName} (${prov.description})`
              }
              return prov.displayName
            })
            .join(', ')
          showToast(`Found ${descriptions}.`, 'success')
        }
        return result.data
      })
    }
  })

  const handleEmulatorToggle = useCallback(
    (emulatorId: EmulatorID, enabled: boolean) => {
      setConfigState((prev) => {
        const next = new Map(prev.systemEmulators)

        for (const system of systems) {
          const hasEmulator = system.emulators.some((e) => e.id === emulatorId)
          if (!hasEmulator) continue

          const current = next.get(system.id) ?? []

          if (enabled) {
            if (!current.includes(emulatorId)) {
              next.set(system.id, [...current, emulatorId])
            }
          } else {
            const filtered = current.filter((id) => id !== emulatorId)
            if (filtered.length === 0) {
              next.delete(system.id)
            } else {
              next.set(system.id, filtered)
            }
          }
        }
        return { ...prev, systemEmulators: next }
      })
    },
    [systems],
  )

  const enabledEmulators = new Set(Array.from(configState.systemEmulators.values()).flat())

  const handleVersionChange = useCallback((emulatorId: EmulatorID, version: string | null) => {
    setConfigState((prev) => {
      const next = new Map(prev.emulatorVersions)
      if (version === null) {
        next.delete(emulatorId)
      } else {
        next.set(emulatorId, version)
      }
      return { ...prev, emulatorVersions: next }
    })
  }, [])

  const handleFrontendToggle = useCallback((frontendId: FrontendID, enabled: boolean) => {
    setConfigState((prev) => {
      const next = new Map(prev.enabledFrontends)
      next.set(frontendId, enabled)
      return { ...prev, enabledFrontends: next }
    })
  }, [])

  const handleFrontendVersionChange = useCallback(
    (frontendId: FrontendID, version: string | null) => {
      setConfigState((prev) => {
        const next = new Map(prev.frontendVersions)
        if (version === null) {
          next.delete(frontendId)
        } else {
          next.set(frontendId, version)
        }
        return { ...prev, frontendVersions: next }
      })
    },
    [],
  )

  const handleAddDevice = useCallback(async (deviceId: string, name: string) => {
    const result = await daemon.addSyncDevice({ deviceId, name })
    if (result.ok) {
      const syncResult = await daemon.getSyncStatus()
      if (syncResult.ok) {
        setSyncStatus(syncResult.data)
      }
    }
  }, [])

  const handleRemoveDevice = useCallback(async (deviceId: string) => {
    const result = await daemon.removeSyncDevice({ deviceId })
    if (result.ok) {
      const syncResult = await daemon.getSyncStatus()
      if (syncResult.ok) {
        setSyncStatus(syncResult.data)
      }
    }
  }, [])

  const handleEnableAll = useCallback(() => {
    const newSystemEmulators = new Map<SystemID, EmulatorID[]>()
    for (const sys of systems) {
      newSystemEmulators.set(sys.id, [sys.defaultEmulatorId])
    }
    setConfigState((prev) => ({ ...prev, systemEmulators: newSystemEmulators }))

    showToast('All systems enabled. Use "Discard changes" to undo.', 'success')
  }, [systems, showToast])

  const handleDiscard = useCallback(async () => {
    const [configResult, statusResult] = await Promise.all([daemon.getConfig(), daemon.getStatus()])

    if (configResult.ok) {
      setConfigState(parseConfigResponse(configResult.data))
    }

    if (statusResult.ok) {
      const { versions, execLines, configs, feVersions, paths } = parseStatusResponse(
        statusResult.data,
      )
      setInstalledVersions(versions)
      setInstalledExecLines(execLines)
      setManagedConfigs(configs)
      setInstalledFrontendVersions(feVersions)
      setInstalledPaths(paths)
    }
  }, [])

  const hasConfigChanges = (() => {
    if (!configReady) return false
    if (configState.userStore !== savedConfigState.current.userStore) return true

    if (configState.systemEmulators.size !== savedConfigState.current.systemEmulators.size)
      return true
    for (const [sysId, emuIds] of configState.systemEmulators) {
      const savedIds = savedConfigState.current.systemEmulators.get(sysId)
      if (!savedIds || emuIds.length !== savedIds.length) return true
      if (!emuIds.every((id, i) => savedIds[i] === id)) return true
    }
    for (const sysId of savedConfigState.current.systemEmulators.keys()) {
      if (!configState.systemEmulators.has(sysId)) return true
    }

    if (configState.emulatorVersions.size !== savedConfigState.current.emulatorVersions.size)
      return true
    for (const [emuId, version] of configState.emulatorVersions) {
      if (savedConfigState.current.emulatorVersions.get(emuId) !== version) return true
    }
    for (const emuId of savedConfigState.current.emulatorVersions.keys()) {
      if (!configState.emulatorVersions.has(emuId)) return true
    }

    if (configState.enabledFrontends.size !== savedConfigState.current.enabledFrontends.size)
      return true
    for (const [feId, enabled] of configState.enabledFrontends) {
      if (savedConfigState.current.enabledFrontends.get(feId) !== enabled) return true
    }
    for (const feId of savedConfigState.current.enabledFrontends.keys()) {
      if (!configState.enabledFrontends.has(feId)) return true
    }

    if (configState.frontendVersions.size !== savedConfigState.current.frontendVersions.size)
      return true
    for (const [feId, version] of configState.frontendVersions) {
      if (savedConfigState.current.frontendVersions.get(feId) !== version) return true
    }
    for (const feId of savedConfigState.current.frontendVersions.keys()) {
      if (!configState.frontendVersions.has(feId)) return true
    }

    return false
  })()

  const renderView = () => {
    if (!configReady) {
      return (
        <div className="h-full flex items-center justify-center text-on-surface-muted">
          <div className="flex items-center gap-3">
            <Spinner />
            <span className="text-sm">Loading configuration...</span>
          </div>
        </div>
      )
    }
    switch (currentView) {
      case 'systems':
        return (
          <SystemsView
            systems={systems}
            frontends={frontends}
            systemEmulators={configState.systemEmulators}
            enabledEmulators={enabledEmulators}
            enabledFrontends={configState.enabledFrontends}
            emulatorVersions={configState.emulatorVersions}
            frontendVersions={configState.frontendVersions}
            installedVersions={installedVersions}
            installedFrontendVersions={installedFrontendVersions}
            installedExecLines={installedExecLines}
            managedConfigs={managedConfigs}
            installedPaths={installedPaths}
            provisions={provisions}
            userStore={configState.userStore}
            hasConfigChanges={hasConfigChanges}
            onUserStoreChange={(value) => setConfigState((prev) => ({ ...prev, userStore: value }))}
            onEmulatorToggle={handleEmulatorToggle}
            onVersionChange={handleVersionChange}
            onFrontendToggle={handleFrontendToggle}
            onFrontendVersionChange={handleFrontendVersionChange}
            onDiscard={handleDiscard}
            onEnableAll={handleEnableAll}
          />
        )
      case 'installation':
        return <InstallationView />
      case 'sync':
        return (
          <SyncView
            status={syncStatus}
            onAddDevice={handleAddDevice}
            onRemoveDevice={handleRemoveDevice}
          />
        )
      default:
        return null
    }
  }

  return (
    <div className="h-dvh bg-surface flex flex-col overflow-hidden font-body">
      {updateInfo?.available && !updateDismissed && (
        <UpdateBanner
          updateInfo={updateInfo}
          onUpdate={handleUpdate}
          onDismiss={handleDismissUpdate}
          isDownloading={isDownloading}
          downloadProgress={downloadProgress}
        />
      )}

      {showApplyBanner && !applyBannerDismissed && (
        <ApplyAfterUpdateBanner
          onApply={handleApplyFromBanner}
          onDismiss={handleDismissApplyBanner}
        />
      )}

      <div className="flex-1 flex flex-col min-[720px]:flex-row min-h-0">
        <Sidebar currentView={currentView} onNavigate={setCurrentView} syncStatus={syncStatus} />
        <main id="main-content" className="flex-1 overflow-y-auto">
          {renderView()}
        </main>
      </div>

      <BottomBarSlot />

      <ApplyProgressBar
        currentView={currentView}
        onNavigateToSystems={() => setCurrentView('systems')}
      />
    </div>
  )
}

export function App() {
  return (
    <BottomBarSlotProvider>
      <ToastProvider>
        <ApplyProvider>
          <AppContent />
        </ApplyProvider>
      </ToastProvider>
    </BottomBarSlotProvider>
  )
}
