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
import { useUpdateChecker } from '@/lib/hooks/useUpdateChecker'
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
import type { View } from '@/types/ui'

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

function parseConfigResponse(data: ConfigResponse) {
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

  return { systemEmulators, emulatorVersions, enabledFrontends, frontendVersions }
}

function AppContent() {
  const [currentView, setCurrentView] = useState<View>('systems')
  const [systems, setSystems] = useState<readonly System[]>([])
  const [frontends, setFrontends] = useState<readonly FrontendRef[]>([])
  const [systemEmulators, setSystemEmulators] = useState<Map<SystemID, EmulatorID[]>>(new Map())
  const [enabledFrontends, setEnabledFrontends] = useState<Map<FrontendID, boolean>>(new Map())
  const [emulatorVersions, setEmulatorVersions] = useState<Map<EmulatorID, string | null>>(
    new Map(),
  )
  const [frontendVersions, setFrontendVersions] = useState<Map<FrontendID, string | null>>(
    new Map(),
  )
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
  const [userStore, setUserStore] = useState('~/Emulation')
  const [syncStatus, setSyncStatus] = useState<SyncStatusResponse | null>(null)

  const savedSystemEmulators = useRef<Map<SystemID, EmulatorID[]>>(new Map())
  const savedEmulatorVersions = useRef<Map<EmulatorID, string | null>>(new Map())
  const savedEnabledFrontends = useRef<Map<FrontendID, boolean>>(new Map())
  const savedFrontendVersions = useRef<Map<FrontendID, string | null>>(new Map())
  const savedUserStore = useRef('~/Emulation')

  const { onCompleteRef } = useApply()
  const { showToast } = useToast()

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
      savedSystemEmulators.current = new Map(parsed.systemEmulators)
      savedEmulatorVersions.current = new Map(parsed.emulatorVersions)
      savedEnabledFrontends.current = new Map(parsed.enabledFrontends)
      savedFrontendVersions.current = new Map(parsed.frontendVersions)
      savedUserStore.current = configResult.data.userStore
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
        setUserStore(configResult.data.userStore)
        savedUserStore.current = configResult.data.userStore
        const parsed = parseConfigResponse(configResult.data)
        setSystemEmulators(parsed.systemEmulators)
        setEmulatorVersions(parsed.emulatorVersions)
        setEnabledFrontends(parsed.enabledFrontends)
        setFrontendVersions(parsed.frontendVersions)
        savedSystemEmulators.current = new Map(parsed.systemEmulators)
        savedEmulatorVersions.current = new Map(parsed.emulatorVersions)
        savedEnabledFrontends.current = new Map(parsed.enabledFrontends)
        savedFrontendVersions.current = new Map(parsed.frontendVersions)
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

  const handleEmulatorToggle = useCallback(
    (emulatorId: EmulatorID, enabled: boolean) => {
      setSystemEmulators((prev) => {
        const next = new Map(prev)

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
        return next
      })
    },
    [systems],
  )

  const enabledEmulators = new Set(Array.from(systemEmulators.values()).flat())

  const handleVersionChange = useCallback((emulatorId: EmulatorID, version: string | null) => {
    setEmulatorVersions((prev) => {
      const next = new Map(prev)
      if (version === null) {
        next.delete(emulatorId)
      } else {
        next.set(emulatorId, version)
      }
      return next
    })
  }, [])

  const handleFrontendToggle = useCallback((frontendId: FrontendID, enabled: boolean) => {
    setEnabledFrontends((prev) => {
      const next = new Map(prev)
      next.set(frontendId, enabled)
      return next
    })
  }, [])

  const handleFrontendVersionChange = useCallback(
    (frontendId: FrontendID, version: string | null) => {
      setFrontendVersions((prev) => {
        const next = new Map(prev)
        if (version === null) {
          next.delete(frontendId)
        } else {
          next.set(frontendId, version)
        }
        return next
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
    setSystemEmulators(newSystemEmulators)

    showToast('All systems enabled. Use "Discard changes" to undo.', 'success')
  }, [systems, showToast])

  const handleDiscard = useCallback(async () => {
    const [configResult, statusResult] = await Promise.all([daemon.getConfig(), daemon.getStatus()])

    if (configResult.ok) {
      const parsed = parseConfigResponse(configResult.data)
      setSystemEmulators(parsed.systemEmulators)
      setEmulatorVersions(parsed.emulatorVersions)
      setEnabledFrontends(parsed.enabledFrontends)
      setFrontendVersions(parsed.frontendVersions)
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
    if (userStore !== savedUserStore.current) return true

    if (systemEmulators.size !== savedSystemEmulators.current.size) return true
    for (const [sysId, emuIds] of systemEmulators) {
      const savedIds = savedSystemEmulators.current.get(sysId)
      if (!savedIds || emuIds.length !== savedIds.length) return true
      if (!emuIds.every((id, i) => savedIds[i] === id)) return true
    }
    for (const sysId of savedSystemEmulators.current.keys()) {
      if (!systemEmulators.has(sysId)) return true
    }

    if (emulatorVersions.size !== savedEmulatorVersions.current.size) return true
    for (const [emuId, version] of emulatorVersions) {
      if (savedEmulatorVersions.current.get(emuId) !== version) return true
    }
    for (const emuId of savedEmulatorVersions.current.keys()) {
      if (!emulatorVersions.has(emuId)) return true
    }

    if (enabledFrontends.size !== savedEnabledFrontends.current.size) return true
    for (const [feId, enabled] of enabledFrontends) {
      if (savedEnabledFrontends.current.get(feId) !== enabled) return true
    }
    for (const feId of savedEnabledFrontends.current.keys()) {
      if (!enabledFrontends.has(feId)) return true
    }

    if (frontendVersions.size !== savedFrontendVersions.current.size) return true
    for (const [feId, version] of frontendVersions) {
      if (savedFrontendVersions.current.get(feId) !== version) return true
    }
    for (const feId of savedFrontendVersions.current.keys()) {
      if (!frontendVersions.has(feId)) return true
    }

    return false
  })()

  const renderView = () => {
    switch (currentView) {
      case 'systems':
        return (
          <SystemsView
            systems={systems}
            frontends={frontends}
            systemEmulators={systemEmulators}
            enabledEmulators={enabledEmulators}
            enabledFrontends={enabledFrontends}
            emulatorVersions={emulatorVersions}
            frontendVersions={frontendVersions}
            installedVersions={installedVersions}
            installedFrontendVersions={installedFrontendVersions}
            installedExecLines={installedExecLines}
            managedConfigs={managedConfigs}
            installedPaths={installedPaths}
            provisions={provisions}
            userStore={userStore}
            hasConfigChanges={hasConfigChanges}
            onUserStoreChange={setUserStore}
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
        <main className="flex-1 overflow-y-auto">{renderView()}</main>
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
