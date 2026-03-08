import type { DoctorResponse, EmulatorID, FrontendID } from '@shared/daemon'
import {
  VIEW_CATALOG,
  VIEW_IMPORT,
  VIEW_INSTALLATION,
  VIEW_PREFERENCES,
  VIEW_SYNC,
  type View,
} from '@shared/ui'
import { useEffect, useRef, useState } from 'react'
import { AlphaWarningBanner } from '@/components/AlphaWarningBanner/AlphaWarningBanner'
import { ApplyBar } from '@/components/ApplyBar/ApplyBar'
import { CatalogView } from '@/components/CatalogView/CatalogView'
import { ImportView } from '@/components/ImportView/ImportView'
import { InstallationView } from '@/components/InstallationView/InstallationView'
import { PendingDeviceBanner } from '@/components/PendingDeviceBanner/PendingDeviceBanner'
import { PreferencesView } from '@/components/PreferencesView/PreferencesView'
import { Sidebar } from '@/components/Sidebar/Sidebar'
import { SyncView } from '@/components/SyncView/SyncView'
import { UpdateBanner } from '@/components/UpdateBanner/UpdateBanner'
import { ApplyProvider, useApply } from '@/lib/ApplyContext'
import { BottomBarSlot, BottomBarSlotProvider } from '@/lib/BottomBarSlot'
import { ConfigProvider, useConfig } from '@/lib/ConfigContext'
import * as daemon from '@/lib/daemon'
import { useApplyStatusHandler } from '@/lib/hooks/useApplyStatusHandler'
import { useOnWindowFocus } from '@/lib/hooks/useOnWindowFocus'
import { useSyncPairing } from '@/lib/hooks/useSyncPairing'
import { useUpdateChecker } from '@/lib/hooks/useUpdateChecker'
import { type FoundProvision, getNewlyFoundProvisions } from '@/lib/provisions'
import { Spinner } from '@/lib/Spinner'
import { ToastProvider, useToast } from '@/lib/ToastContext'
import { useStatusData } from '@/lib/useStatusData'

function keyForProvision(provision: FoundProvision) {
  return provision.id
}

function AppContent() {
  const [currentView, setCurrentView] = useState<View>(VIEW_CATALOG)
  const [provisions, setProvisions] = useState<DoctorResponse>({})
  const [fontsReady, setFontsReady] = useState(false)

  const {
    setStatusResponse,
    installedExecLines,
    managedConfigs,
    installedPaths,
    installedFrontendExecLines,
    kyarabenVersion,
  } = useStatusData()

  const config = useConfig()
  const {
    refreshAfterApply,
    setSystems,
    setFrontends,
    initFromResponse,
    setInstalledVersions,
    setInstalledFrontendVersions,
    setUpgradeAvailable,
    enableAllSystems,
  } = config
  const { onCompleteRef } = useApply()
  const { showToast } = useToast()
  const { status: applyStatus } = useApply()
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
    pendingDevice,
    handleRemoveDevice,
    handleConnectToDevice,
    handleEnableSync,
    handleResetSync,
    handleStartPairing,
    handleStopPairing,
    handleToggleGlobalDiscovery,
    handleToggleRunning,
    handleToggleAutostart,
    handleAcceptDevice,
    clearConnectionError,
    refreshSyncStatus,
  } = useSyncPairing(showToast, currentView === VIEW_SYNC)

  useEffect(() => {
    document.fonts.ready.then(() => setFontsReady(true))
  }, [])

  useEffect(() => {
    onCompleteRef.current = async () => {
      const [doctorResult, statusResult] = await Promise.all([
        daemon.runDoctor(),
        daemon.getStatus(),
      ])

      if (doctorResult.ok) {
        setProvisions(doctorResult.data)
      }

      if (statusResult.ok) {
        setStatusResponse(statusResult.data)

        const emuVersions = new Map<EmulatorID, string>()
        for (const emu of statusResult.data.installedEmulators) {
          emuVersions.set(emu.id, emu.version)
        }
        setInstalledVersions(emuVersions)

        const feVersions = new Map<FrontendID, string>()
        for (const fe of statusResult.data.installedFrontends) {
          feVersions.set(fe.id, fe.version)
        }
        setInstalledFrontendVersions(feVersions)
      }

      await refreshAfterApply()
      clearApplyBannerDismissal()
    }
  }, [
    onCompleteRef,
    clearApplyBannerDismissal,
    setStatusResponse,
    refreshAfterApply,
    setInstalledVersions,
    setInstalledFrontendVersions,
  ])

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
        initFromResponse(configResult.data)
      } else {
        showToast(`Failed to load configuration: ${configResult.error.message}`, 'error')
      }

      if (statusResult.ok) {
        setStatusResponse(statusResult.data)

        const emuVersions = new Map<EmulatorID, string>()
        for (const emu of statusResult.data.installedEmulators) {
          emuVersions.set(emu.id, emu.version)
        }
        setInstalledVersions(emuVersions)

        const feVersions = new Map<FrontendID, string>()
        for (const fe of statusResult.data.installedFrontends) {
          feVersions.set(fe.id, fe.version)
        }
        setInstalledFrontendVersions(feVersions)

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
  }, [
    showToast,
    setShowApplyBanner,
    refreshSyncStatus,
    setStatusResponse,
    initFromResponse,
    setSystems,
    setFrontends,
    setInstalledVersions,
    setInstalledFrontendVersions,
  ])

  useEffect(() => {
    setUpgradeAvailable(showApplyBanner && !applyBannerDismissed)
  }, [showApplyBanner, applyBannerDismissed, setUpgradeAvailable])

  useApplyStatusHandler(applyStatus, currentView, () => setCurrentView(VIEW_CATALOG))

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

  const handleEnableAll = () => {
    enableAllSystems()
    showToast('All systems enabled. Use "Discard changes" to undo.', 'success')
  }

  const renderView = () => {
    if (!config.configReady) {
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
            provisions={provisions}
            installedExecLines={installedExecLines}
            installedFrontendExecLines={installedFrontendExecLines}
            managedConfigs={managedConfigs}
            installedPaths={installedPaths}
            onNavigateToPreferences={() => setCurrentView(VIEW_PREFERENCES)}
            onEnableAll={handleEnableAll}
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
            onToggleRunning={handleToggleRunning}
            onToggleAutostart={handleToggleAutostart}
            enableError={enableError}
            isEnabling={isEnabling}
          />
        )
      case VIEW_IMPORT:
        return <ImportView onNavigateToCatalog={() => setCurrentView(VIEW_CATALOG)} />
      case VIEW_PREFERENCES:
        return <PreferencesView />
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
      <AlphaWarningBanner />

      {updateInfo?.available && !updateDismissed && (
        <UpdateBanner
          updateInfo={updateInfo}
          onUpdate={handleUpdate}
          onDismiss={handleDismissUpdate}
          isDownloading={isDownloading}
          downloadProgress={downloadProgress}
        />
      )}

      {pendingDevice && (
        <PendingDeviceBanner pendingDevice={pendingDevice} onAccept={handleAcceptDevice} />
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
      <ApplyBar />
    </div>
  )
}

export function App() {
  return (
    <BottomBarSlotProvider>
      <ToastProvider>
        <ApplyProvider>
          <ConfigProvider>
            <AppContent />
          </ConfigProvider>
        </ApplyProvider>
      </ToastProvider>
    </BottomBarSlotProvider>
  )
}
