import { useCallback, useEffect, useRef, useState } from 'react'
import { ApplyProgressBar } from '@/components/ApplyProgressBar/ApplyProgressBar'
import { CatalogView } from '@/components/CatalogView/CatalogView'
import { InstallationView } from '@/components/InstallationView/InstallationView'
import { Sidebar } from '@/components/Sidebar/Sidebar'
import { SyncView } from '@/components/SyncView/SyncView'
import { UpdateBanner } from '@/components/UpdateBanner/UpdateBanner'
import { ApplyProvider, useApply } from '@/lib/ApplyContext'
import { BottomBarSlot, BottomBarSlotProvider } from '@/lib/BottomBarSlot'
import * as daemon from '@/lib/daemon'
import { useOnWindowFocus } from '@/lib/hooks/useOnWindowFocus'
import { useSyncPairing } from '@/lib/hooks/useSyncPairing'
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
  System,
  SystemID,
} from '@/types/daemon'
import {
  type ApplyStatus,
  VIEW_CATALOG,
  VIEW_INSTALLATION,
  VIEW_LABELS,
  VIEW_SYNC,
  type View,
} from '@/types/ui'

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
  const feExecLines = new Map<FrontendID, string>()
  for (const fe of data.installedFrontends ?? []) {
    feVersions.set(fe.id, fe.version)
    if (fe.execLine) {
      feExecLines.set(fe.id, fe.execLine)
    }
  }

  return { versions, execLines, configs, feVersions, feExecLines, paths }
}

interface ConfigState {
  userStore: string
  graphicsShaders: string
  systemEmulators: Map<SystemID, EmulatorID[]>
  emulatorVersions: Map<EmulatorID, string | null>
  emulatorShaders: Map<EmulatorID, string | null>
  enabledFrontends: Map<FrontendID, boolean>
  frontendVersions: Map<FrontendID, string | null>
}

function emptyConfigState(): ConfigState {
  return {
    userStore: '',
    graphicsShaders: '',
    systemEmulators: new Map(),
    emulatorVersions: new Map(),
    emulatorShaders: new Map(),
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
    graphicsShaders: state.graphicsShaders,
    systemEmulators: new Map(state.systemEmulators),
    emulatorVersions: new Map(state.emulatorVersions),
    emulatorShaders: new Map(state.emulatorShaders),
    enabledFrontends: new Map(state.enabledFrontends),
    frontendVersions: new Map(state.frontendVersions),
  }
}

function parseConfigResponse(data: ConfigResponse): ConfigState {
  const systemEmulators = new Map<SystemID, EmulatorID[]>()
  const emulatorVersions = new Map<EmulatorID, string | null>()
  const emulatorShaders = new Map<EmulatorID, string | null>()
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
      if (conf.shaders !== undefined) {
        emulatorShaders.set(emuId as EmulatorID, conf.shaders ?? null)
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
    graphicsShaders: data.graphics?.shaders ?? '',
    systemEmulators,
    emulatorVersions,
    emulatorShaders,
    enabledFrontends,
    frontendVersions,
  }
}

function AppContent() {
  const [currentView, setCurrentView] = useState<View>(VIEW_CATALOG)
  const [systems, setSystems] = useState<readonly System[]>([])
  const [frontends, setFrontends] = useState<readonly FrontendRef[]>([])
  const [installedVersions, setInstalledVersions] = useState<Map<EmulatorID, string>>(new Map())
  const [installedFrontendVersions, setInstalledFrontendVersions] = useState<
    Map<FrontendID, string>
  >(new Map())
  const [installedFrontendExecLines, setInstalledFrontendExecLines] = useState<
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
  const [kyarabenVersion, setKyarabenVersion] = useState<string | null>(null)
  const [fontsReady, setFontsReady] = useState(false)

  const savedConfigState = useRef<ConfigState>(emptyConfigState())

  const { onCompleteRef, apply } = useApply()
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
    setShowApplyBanner,
    clearApplyBannerDismissal,
  } = useUpdateChecker(showToast)

  const handleReapply = useCallback(async () => {
    const saved = savedConfigState.current
    const systemsConfig: Record<string, string[]> = {}
    for (const [sysId, emuIds] of saved.systemEmulators) {
      systemsConfig[sysId] = emuIds
    }
    const emulatorsConfig: Record<string, { version?: string }> = {}
    for (const [emuId, version] of saved.emulatorVersions) {
      if (version) {
        emulatorsConfig[emuId] = { version }
      }
    }
    const frontendsConfig: Record<string, { enabled: boolean; version?: string }> = {}
    for (const [feId, enabled] of saved.enabledFrontends) {
      const version = saved.frontendVersions.get(feId)
      frontendsConfig[feId] = { enabled, ...(version && { version }) }
    }
    await apply({
      userStore: saved.userStore,
      systems: systemsConfig,
      emulators: emulatorsConfig,
      frontends: frontendsConfig,
    })
  }, [apply])

  const {
    syncStatus,
    connectionProgress,
    connectionError,
    enableError,
    isEnabling,
    isConnecting,
    isPairing,
    pairingDeviceId,
    pairingCode,
    lastSyncedAt,
    handleRemoveDevice,
    handleConnectToDevice,
    handleEnableSync,
    handleResetSync,
    handleStartPairing,
    handleStopPairing,
    handleToggleGlobalDiscovery,
    clearConnectionError,
    refreshSyncStatus,
  } = useSyncPairing(showToast, currentView === VIEW_SYNC)

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
      const { versions, execLines, configs, feVersions, feExecLines, paths } = parseStatusResponse(
        statusResult.data,
      )
      setInstalledVersions(versions)
      setInstalledExecLines(execLines)
      setManagedConfigs(configs)
      setInstalledFrontendVersions(feVersions)
      setInstalledFrontendExecLines(feExecLines)
      setInstalledPaths(paths)
      setKyarabenVersion(statusResult.data.kyarabenVersion)
    }

    if (configResult.ok) {
      const parsed = parseConfigResponse(configResult.data)
      savedConfigState.current = cloneConfigState(parsed)
    }
  }, [])

  useEffect(() => {
    document.fonts.ready.then(() => setFontsReady(true))
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
        showToast(`Failed to load configuration: ${configResult.error.message}`, 'error')
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
        setKyarabenVersion(statusResult.data.kyarabenVersion)

        if (statusResult.data.healthWarning === 'orphaned_artifacts') {
          showToast(
            <span>
              Installation state may be corrupted.{' '}
              <button
                type="button"
                className="underline hover:no-underline"
                onClick={() => setCurrentView(VIEW_INSTALLATION)}
              >
                See details
              </button>
            </span>,
            'error',
          )
        }

        if (statusResult.data.configWarnings && statusResult.data.configWarnings.length > 0) {
          const warnings = statusResult.data.configWarnings
          showToast(
            <div className="flex flex-col">
              <span>Config issues found (using defaults):</span>
              <ul className="list-disc ml-4 mt-1">
                {warnings.map((w) => (
                  <li key={w.field}>
                    <code>{w.field}</code>: {w.message}
                  </li>
                ))}
              </ul>
            </div>,
            'warning',
            Infinity,
          )
        }

        const running = statusResult.data.kyarabenVersion
        const manifest = statusResult.data.manifestKyarabenVersion
        if (running && manifest && running !== manifest) {
          setShowApplyBanner(true)
        }
      }

      const doctorResult = await daemon.runDoctor()
      if (doctorResult.ok) {
        setProvisions(doctorResult.data)
      }

      await refreshSyncStatus()
    }

    init()
  }, [showToast, setShowApplyBanner, refreshSyncStatus])

  useEffect(() => {
    if (applyStatus === lastApplyStatus.current) return
    if (applyStatus === 'success') {
      if (currentView !== VIEW_CATALOG) {
        showToast(
          <span>
            Installation complete.{' '}
            <button
              type="button"
              className="underline hover:no-underline"
              onClick={() => setCurrentView(VIEW_CATALOG)}
            >
              Go to {VIEW_LABELS[VIEW_CATALOG].toLowerCase()}
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
    const [doctorResult] = await Promise.all([daemon.runDoctor(), refreshSyncStatus()])
    if (doctorResult.ok) {
      setProvisions((prev) => {
        const newlyFound = getNewlyFoundProvisions(prev, doctorResult.data)
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
        return doctorResult.data
      })
    }
  })

  const handleEmulatorToggle = useCallback(
    (systemId: SystemID, emulatorId: EmulatorID, enabled: boolean) => {
      setConfigState((prev) => {
        const next = new Map(prev.systemEmulators)
        const current = next.get(systemId) ?? []

        if (enabled) {
          if (!current.includes(emulatorId)) {
            next.set(systemId, [...current, emulatorId])
          }
        } else {
          const filtered = current.filter((id) => id !== emulatorId)
          if (filtered.length === 0) {
            next.delete(systemId)
          } else {
            next.set(systemId, filtered)
          }
        }
        return { ...prev, systemEmulators: next }
      })
    },
    [],
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

  const handleShaderChange = useCallback((emulatorId: EmulatorID, shaders: string | null) => {
    setConfigState((prev) => {
      const next = new Map(prev.emulatorShaders)
      if (shaders === null) {
        next.delete(emulatorId)
      } else {
        next.set(emulatorId, shaders)
      }
      return { ...prev, emulatorShaders: next }
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

  const handleGraphicsShadersChange = useCallback((value: string) => {
    setConfigState((prev) => ({ ...prev, graphicsShaders: value }))
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
      const { versions, execLines, configs, feVersions, feExecLines, paths } = parseStatusResponse(
        statusResult.data,
      )
      setInstalledVersions(versions)
      setInstalledExecLines(execLines)
      setManagedConfigs(configs)
      setInstalledFrontendVersions(feVersions)
      setInstalledFrontendExecLines(feExecLines)
      setInstalledPaths(paths)
      setKyarabenVersion(statusResult.data.kyarabenVersion)
    }
  }, [])

  const configChanges = (() => {
    if (!configReady) return []
    const changes: string[] = []

    if (configState.userStore !== savedConfigState.current.userStore) {
      changes.push('Emulation folder')
    }

    if (configState.graphicsShaders !== savedConfigState.current.graphicsShaders) {
      changes.push('Shader settings')
    }

    const systemEmulatorsChanged = (() => {
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
      return false
    })()
    if (systemEmulatorsChanged) {
      changes.push('Enabled emulators')
    }

    const emulatorVersionsChanged = (() => {
      if (configState.emulatorVersions.size !== savedConfigState.current.emulatorVersions.size)
        return true
      for (const [emuId, version] of configState.emulatorVersions) {
        if (savedConfigState.current.emulatorVersions.get(emuId) !== version) return true
      }
      for (const emuId of savedConfigState.current.emulatorVersions.keys()) {
        if (!configState.emulatorVersions.has(emuId)) return true
      }
      return false
    })()
    if (emulatorVersionsChanged) {
      changes.push('Emulator versions')
    }

    const emulatorShadersChanged = (() => {
      if (configState.emulatorShaders.size !== savedConfigState.current.emulatorShaders.size)
        return true
      for (const [emuId, shaders] of configState.emulatorShaders) {
        if (savedConfigState.current.emulatorShaders.get(emuId) !== shaders) return true
      }
      for (const emuId of savedConfigState.current.emulatorShaders.keys()) {
        if (!configState.emulatorShaders.has(emuId)) return true
      }
      return false
    })()
    if (emulatorShadersChanged) {
      changes.push('Emulator shaders')
    }

    const enabledFrontendsChanged = (() => {
      if (configState.enabledFrontends.size !== savedConfigState.current.enabledFrontends.size)
        return true
      for (const [feId, enabled] of configState.enabledFrontends) {
        if (savedConfigState.current.enabledFrontends.get(feId) !== enabled) return true
      }
      for (const feId of savedConfigState.current.enabledFrontends.keys()) {
        if (!configState.enabledFrontends.has(feId)) return true
      }
      return false
    })()
    if (enabledFrontendsChanged) {
      changes.push('Enabled frontends')
    }

    const frontendVersionsChanged = (() => {
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
    if (frontendVersionsChanged) {
      changes.push('Frontend versions')
    }

    return changes
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
      case VIEW_CATALOG:
        return (
          <CatalogView
            systems={systems}
            frontends={frontends}
            systemEmulators={configState.systemEmulators}
            enabledEmulators={enabledEmulators}
            enabledFrontends={configState.enabledFrontends}
            emulatorVersions={configState.emulatorVersions}
            emulatorShaders={configState.emulatorShaders}
            graphics={{ shaders: configState.graphicsShaders }}
            frontendVersions={configState.frontendVersions}
            installedVersions={installedVersions}
            installedFrontendVersions={installedFrontendVersions}
            installedFrontendExecLines={installedFrontendExecLines}
            installedExecLines={installedExecLines}
            managedConfigs={managedConfigs}
            installedPaths={installedPaths}
            provisions={provisions}
            userStore={configState.userStore}
            configChanges={configChanges}
            onUserStoreChange={(value) => setConfigState((prev) => ({ ...prev, userStore: value }))}
            onEmulatorToggle={handleEmulatorToggle}
            onVersionChange={handleVersionChange}
            onShaderChange={handleShaderChange}
            onFrontendToggle={handleFrontendToggle}
            onFrontendVersionChange={handleFrontendVersionChange}
            onGraphicsShadersChange={handleGraphicsShadersChange}
            onDiscard={handleDiscard}
            onEnableAll={handleEnableAll}
            upgradeAvailable={showApplyBanner && !applyBannerDismissed}
            onReapply={handleReapply}
          />
        )
      case VIEW_INSTALLATION:
        return <InstallationView />
      case VIEW_SYNC:
        return (
          <SyncView
            status={syncStatus}
            connectionProgress={connectionProgress}
            connectionError={connectionError}
            isConnecting={isConnecting}
            isPairing={isPairing}
            pairingDeviceId={pairingDeviceId}
            pairingCode={pairingCode}
            lastSyncedAt={lastSyncedAt}
            onRemoveDevice={handleRemoveDevice}
            onConnectToDevice={handleConnectToDevice}
            onEnableSync={handleEnableSync}
            onResetSync={handleResetSync}
            onStartPairing={handleStartPairing}
            onStopPairing={handleStopPairing}
            onClearConnectionError={clearConnectionError}
            onRefresh={refreshSyncStatus}
            onToggleGlobalDiscovery={handleToggleGlobalDiscovery}
            enableError={enableError}
            isEnabling={isEnabling}
          />
        )
      default:
        return null
    }
  }

  if (!fontsReady) {
    return (
      <div className="h-dvh bg-surface flex items-center justify-center">
        <Spinner />
      </div>
    )
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

      <div className="flex-1 flex flex-col min-[720px]:flex-row min-h-0">
        <Sidebar
          currentView={currentView}
          onNavigate={setCurrentView}
          syncStatus={syncStatus}
          version={kyarabenVersion}
        />
        <main id="main-content" className="flex-1 overflow-y-scroll">
          {renderView()}
        </main>
      </div>

      <BottomBarSlot />

      <ApplyProgressBar
        currentView={currentView}
        onNavigateToCatalog={() => setCurrentView(VIEW_CATALOG)}
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
