import { useCallback, useEffect, useState } from 'react'
import { InstallationView } from '@/components/InstallationView/InstallationView'
import { Sidebar } from '@/components/Sidebar/Sidebar'
import { SyncView } from '@/components/SyncView/SyncView'
import { SystemsView } from '@/components/SystemsView/SystemsView'
import * as daemon from '@/lib/daemon'
import { BottomBarProvider } from '@/lib/BottomBarContext'
import { ToastProvider } from '@/lib/ToastContext'
import type {
  DoctorResponse,
  EmulatorID,
  SyncStatusResponse,
  System,
  SystemID,
} from '@/types/daemon'
import type { ApplyStatus, ProgressStep, View } from '@/types/ui'

const PROGRESS_STEP_LABELS: Readonly<Record<string, string>> = {
  store: 'Setting up emulation folder',
  build: 'Installing emulators',
  desktop: 'Adding to application menu',
  config: 'Configuring emulators',
}

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
  const [applyStatus, setApplyStatus] = useState<ApplyStatus>('idle')
  const [progressSteps, setProgressSteps] = useState<readonly ProgressStep[]>([])
  const [error, setError] = useState<string | null>(null)
  const [syncStatus, setSyncStatus] = useState<SyncStatusResponse | null>(null)


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
        for (const emu of statusResult.data.installedEmulators) {
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

  const handleApply = useCallback(async () => {
    setApplyStatus('applying')
    setProgressSteps([])
    setError(null)

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

    const configResult = await daemon.setConfig({
      userStore,
      systems: systemsConfig,
      emulators: emulatorsConfig,
    })

    if (!configResult.ok) {
      setError(configResult.error.message)
      setApplyStatus('error')
      return
    }

    const MAX_OUTPUT_LINES = 5

    const progressHandler = (...args: unknown[]) => {
      const data = args[0] as { step: string; message: string; output?: string }

      setProgressSteps((prev) => {
        const existing = prev.find((s) => s.id === data.step)
        const isNewStep = !existing

        return (
          isNewStep
            ? [
                ...prev,
                {
                  id: data.step,
                  label: PROGRESS_STEP_LABELS[data.step] ?? data.step,
                  status: 'in_progress' as const,
                },
              ]
            : prev
        ).map((s) => {
          if (s.id === data.step) {
            return {
              ...s,
              status: 'in_progress' as const,
              ...(data.message && { message: data.message }),
              ...(data.output && {
                output: [...(s.output ?? []), data.output].slice(-MAX_OUTPUT_LINES),
              }),
            }
          }
          if (isNewStep && s.status === 'in_progress') {
            return { ...s, status: 'completed' as const }
          }
          return s
        })
      })
    }

    window.electron.on('apply:progress', progressHandler)

    try {
      const applyResult = await daemon.apply()

      if (!applyResult.ok) {
        setError(applyResult.error.message)
        setApplyStatus('error')
        setProgressSteps((prev) =>
          prev.map((s) => ({ ...s, status: s.status === 'in_progress' ? 'error' : s.status })),
        )
        return
      }

      if (applyResult.data.cancelled) {
        setApplyStatus('cancelled')
        setProgressSteps((prev) =>
          prev.map((s) => ({ ...s, status: s.status === 'in_progress' ? 'cancelled' : s.status })),
        )
        return
      }

      setProgressSteps((prev) => prev.map((s) => ({ ...s, status: 'completed' as const })))
      setApplyStatus('success')

      // Refresh doctor and status after successful apply
      const [doctorResult, statusResult] = await Promise.all([
        daemon.runDoctor(),
        daemon.getStatus(),
      ])

      if (doctorResult.ok) {
        setProvisions(doctorResult.data)
      }

      if (statusResult.ok) {
        const versions = new Map<EmulatorID, string>()
        const execLines = new Map<EmulatorID, string>()
        const configs = new Map<EmulatorID, string[]>()
        for (const emu of statusResult.data.installedEmulators) {
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
    } catch (err) {
      console.error('Apply failed:', err)
      const message = err instanceof Error ? err.message : String(err)
      setError(message)
      setApplyStatus('error')
      setProgressSteps((prev) =>
        prev.map((s) => ({ ...s, status: s.status === 'in_progress' ? 'error' : s.status })),
      )
    } finally {
      window.electron.off('apply:progress', progressHandler)
    }
  }, [systemEmulators, emulatorVersions, userStore])

  const handleCancel = useCallback(async () => {
    await daemon.cancelApply()
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

  const handleReset = useCallback(() => {
    setApplyStatus('idle')
    setProgressSteps([])
    setError(null)
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
      for (const emu of statusResult.data.installedEmulators) {
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
            onApply={handleApply}
            onCancel={handleCancel}
            applyStatus={applyStatus}
            progressSteps={progressSteps}
            error={error}
            onReset={handleReset}
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

  return (
    <div className="h-dvh bg-gray-900 flex flex-col min-[720px]:flex-row overflow-hidden">
      <Sidebar currentView={currentView} onNavigate={setCurrentView} syncStatus={syncStatus} />

      <main className="flex-1 overflow-y-auto">{renderView()}</main>
    </div>
  )
}

export function App() {
  return (
    <BottomBarProvider>
      <ToastProvider>
        <AppContent />
      </ToastProvider>
    </BottomBarProvider>
  )
}
