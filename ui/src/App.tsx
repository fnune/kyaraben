import { useCallback, useEffect, useState } from 'react'
import { ApplyProgressBar } from '@/components/ApplyProgressBar/ApplyProgressBar'
import { InstallationView } from '@/components/InstallationView/InstallationView'
import { Sidebar } from '@/components/Sidebar/Sidebar'
import { SyncView } from '@/components/SyncView/SyncView'
import { SystemsView } from '@/components/SystemsView/SystemsView'
import { ApplyProvider, useApply } from '@/lib/ApplyContext'
import { BottomBarSlot, BottomBarSlotProvider } from '@/lib/BottomBarSlot'
import * as daemon from '@/lib/daemon'
import { ToastProvider } from '@/lib/ToastContext'
import type {
  DoctorResponse,
  EmulatorID,
  SyncStatusResponse,
  System,
  SystemID,
} from '@/types/daemon'
import type { View } from '@/types/ui'

function AppContent() {
  const [currentView, setCurrentView] = useState<View>('systems')
  const [systems, setSystems] = useState<readonly System[]>([])
  const [systemEmulators, setSystemEmulators] = useState<Map<SystemID, EmulatorID[]>>(new Map())
  const [emulatorVersions, setEmulatorVersions] = useState<Map<EmulatorID, string | null>>(
    new Map(),
  )
  const [installedVersions, setInstalledVersions] = useState<Map<EmulatorID, string>>(new Map())
  const [installedExecLines, setInstalledExecLines] = useState<Map<EmulatorID, string>>(new Map())
  const [managedConfigs, setManagedConfigs] = useState<Map<EmulatorID, string[]>>(new Map())
  const [provisions, setProvisions] = useState<DoctorResponse>({})
  const [userStore, setUserStore] = useState('~/Emulation')
  const [syncStatus, setSyncStatus] = useState<SyncStatusResponse | null>(null)

  const { status: applyStatus, onCompleteRef } = useApply()

  const refreshAfterApply = useCallback(async () => {
    const [doctorResult, statusResult] = await Promise.all([daemon.runDoctor(), daemon.getStatus()])

    if (doctorResult.ok) {
      setProvisions(doctorResult.data)
    }

    if (statusResult.ok) {
      const versions = new Map<EmulatorID, string>()
      const execLines = new Map<EmulatorID, string>()
      const configs = new Map<EmulatorID, string[]>()
      for (const emu of statusResult.data.installedEmulators ?? []) {
        versions.set(emu.id, emu.version)
        if (emu.execLine) {
          execLines.set(emu.id, emu.execLine)
        }
        if (emu.managedConfigs) {
          configs.set(emu.id, emu.managedConfigs)
        }
      }
      setInstalledVersions(versions)
      setInstalledExecLines(execLines)
      setManagedConfigs(configs)
    }
  }, [])

  useEffect(() => {
    onCompleteRef.current = refreshAfterApply
  }, [onCompleteRef, refreshAfterApply])

  useEffect(() => {
    async function init() {
      const [systemsResult, configResult, statusResult] = await Promise.all([
        daemon.getSystems(),
        daemon.getConfig(),
        daemon.getStatus(),
      ])

      if (systemsResult.ok) {
        setSystems(systemsResult.data)
      }

      if (configResult.ok) {
        setUserStore(configResult.data.userStore)
        const newSystemEmulators = new Map<SystemID, EmulatorID[]>()
        const newEmulatorVersions = new Map<EmulatorID, string | null>()

        for (const [sysId, emulatorIds] of Object.entries(configResult.data.systems)) {
          if (emulatorIds && emulatorIds.length > 0) {
            newSystemEmulators.set(sysId as SystemID, emulatorIds as EmulatorID[])
          }
        }

        if (configResult.data.emulators) {
          for (const [emuId, conf] of Object.entries(configResult.data.emulators)) {
            if (conf.version) {
              newEmulatorVersions.set(emuId as EmulatorID, conf.version)
            }
          }
        }

        setSystemEmulators(newSystemEmulators)
        setEmulatorVersions(newEmulatorVersions)
      }

      if (statusResult.ok) {
        const versions = new Map<EmulatorID, string>()
        const execLines = new Map<EmulatorID, string>()
        const configs = new Map<EmulatorID, string[]>()
        for (const emu of statusResult.data.installedEmulators ?? []) {
          versions.set(emu.id, emu.version)
          if (emu.execLine) {
            execLines.set(emu.id, emu.execLine)
          }
          if (emu.managedConfigs) {
            configs.set(emu.id, emu.managedConfigs)
          }
        }
        setInstalledVersions(versions)
        setInstalledExecLines(execLines)
        setManagedConfigs(configs)
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
  }, [])

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

  const handleDiscard = useCallback(async () => {
    const [configResult, statusResult] = await Promise.all([daemon.getConfig(), daemon.getStatus()])

    if (configResult.ok) {
      const newSystemEmulators = new Map<SystemID, EmulatorID[]>()
      const newEmulatorVersions = new Map<EmulatorID, string | null>()

      for (const [sysId, emulatorIds] of Object.entries(configResult.data.systems)) {
        if (emulatorIds && emulatorIds.length > 0) {
          newSystemEmulators.set(sysId as SystemID, emulatorIds as EmulatorID[])
        }
      }

      if (configResult.data.emulators) {
        for (const [emuId, conf] of Object.entries(configResult.data.emulators)) {
          if (conf.version) {
            newEmulatorVersions.set(emuId as EmulatorID, conf.version)
          }
        }
      }

      setSystemEmulators(newSystemEmulators)
      setEmulatorVersions(newEmulatorVersions)
    }

    if (statusResult.ok) {
      const versions = new Map<EmulatorID, string>()
      const execLines = new Map<EmulatorID, string>()
      const configs = new Map<EmulatorID, string[]>()
      for (const emu of statusResult.data.installedEmulators ?? []) {
        versions.set(emu.id, emu.version)
        if (emu.execLine) {
          execLines.set(emu.id, emu.execLine)
        }
        if (emu.managedConfigs) {
          configs.set(emu.id, emu.managedConfigs)
        }
      }
      setInstalledVersions(versions)
      setInstalledExecLines(execLines)
      setManagedConfigs(configs)
    }
  }, [])

  const renderView = () => {
    switch (currentView) {
      case 'systems':
        return (
          <SystemsView
            systems={systems}
            systemEmulators={systemEmulators}
            enabledEmulators={enabledEmulators}
            emulatorVersions={emulatorVersions}
            installedVersions={installedVersions}
            installedExecLines={installedExecLines}
            managedConfigs={managedConfigs}
            provisions={provisions}
            userStore={userStore}
            onUserStoreChange={setUserStore}
            onEmulatorToggle={handleEmulatorToggle}
            onVersionChange={handleVersionChange}
            onDiscard={handleDiscard}
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
    }
  }

  const showApplyProgressBar = applyStatus === 'applying' && currentView !== 'systems'

  return (
    <div className="h-dvh bg-gray-900 flex flex-col overflow-hidden">
      <div className="flex-1 flex flex-col min-[720px]:flex-row min-h-0">
        <Sidebar currentView={currentView} onNavigate={setCurrentView} syncStatus={syncStatus} />
        <main className="flex-1 overflow-y-auto">{renderView()}</main>
      </div>

      <BottomBarSlot />

      {showApplyProgressBar && (
        <ApplyProgressBar onNavigateToSystems={() => setCurrentView('systems')} />
      )}
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
