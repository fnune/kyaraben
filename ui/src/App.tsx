import { useCallback, useEffect, useState } from 'react'
import { Settings } from '@/components/Settings/Settings'
import { SyncSettings } from '@/components/SyncSettings/SyncSettings'
import { SyncStatusBar } from '@/components/SyncStatusBar/SyncStatusBar'
import { SystemGrid } from '@/components/SystemGrid/SystemGrid'
import { useDaemon } from '@/hooks/useDaemon'
import { Button } from '@/lib/Button'
import { ProgressSteps } from '@/lib/ProgressSteps'
import { Toast } from '@/lib/Toast'
import type {
  DoctorResponse,
  EmulatorID,
  SyncStatusResponse,
  System,
  SystemID,
} from '@/types/daemon'
import type { ApplyStatus, ProgressStep } from '@/types/ui'

interface ToastState {
  message: string
  type: 'error' | 'success' | 'info'
}

const PROGRESS_STEP_LABELS: Readonly<Record<string, string>> = {
  store: 'Setting up emulation folder',
  build: 'Installing emulators',
  desktop: 'Adding to application menu',
  config: 'Configuring emulators',
}

export function App() {
  const daemon = useDaemon()

  const [systems, setSystems] = useState<readonly System[]>([])
  const [selections, setSelections] = useState<Map<SystemID, EmulatorID>>(new Map())
  const [provisions, setProvisions] = useState<DoctorResponse>({})
  const [userStore, setUserStore] = useState('~/Emulation')
  const [applyStatus, setApplyStatus] = useState<ApplyStatus>('idle')
  const [progressSteps, setProgressSteps] = useState<readonly ProgressStep[]>([])
  const [error, setError] = useState<string | null>(null)
  const [syncStatus, setSyncStatus] = useState<SyncStatusResponse | null>(null)
  const [showSyncSettings, setShowSyncSettings] = useState(false)
  const [toast, setToast] = useState<ToastState | null>(null)

  const showToast = useCallback((message: string, type: ToastState['type'] = 'info') => {
    setToast({ message, type })
  }, [])

  // biome-ignore lint/correctness/useExhaustiveDependencies: init runs once on mount, daemon methods are stable
  useEffect(() => {
    async function init() {
      const [systemsResult, configResult] = await Promise.all([
        daemon.getSystems(),
        daemon.getConfig(),
      ])

      if (systemsResult.ok) {
        setSystems(systemsResult.data)
      }

      if (configResult.ok) {
        setUserStore(configResult.data.userStore)
        const newSelections = new Map<SystemID, EmulatorID>()
        for (const [sysId, emuId] of Object.entries(configResult.data.systems)) {
          newSelections.set(sysId as SystemID, emuId as EmulatorID)
        }
        setSelections(newSelections)
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

  const handleToggle = useCallback(
    (systemId: SystemID, enabled: boolean) => {
      setSelections((prev) => {
        const next = new Map(prev)
        if (enabled) {
          const system = systems.find((s) => s.id === systemId)
          const defaultEmulator = system?.emulators[0]
          if (defaultEmulator) {
            next.set(systemId, defaultEmulator.id)
          }
        } else {
          next.delete(systemId)
        }
        return next
      })
    },
    [systems],
  )

  const handleApply = useCallback(async () => {
    setApplyStatus('applying')
    setProgressSteps([])
    setError(null)

    const systemsConfig: Partial<Record<SystemID, EmulatorID>> = {}
    for (const [sysId, emuId] of selections) {
      systemsConfig[sysId] = emuId
    }

    const configResult = await daemon.setConfig({
      userStore,
      systems: systemsConfig,
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

      setProgressSteps((prev) => prev.map((s) => ({ ...s, status: 'completed' as const })))
      setApplyStatus('success')

      const doctorResult = await daemon.runDoctor()
      if (doctorResult.ok) {
        setProvisions(doctorResult.data)
      }
    } catch (err) {
      console.error('Apply failed:', err)
      setError(err instanceof Error ? err.message : String(err))
      setApplyStatus('error')
      setProgressSteps((prev) =>
        prev.map((s) => ({ ...s, status: s.status === 'in_progress' ? 'error' : s.status })),
      )
    } finally {
      window.electron.off('apply:progress', progressHandler)
    }
  }, [daemon, selections, userStore])

  const handleAddDevice = useCallback(
    async (deviceId: string, name: string) => {
      const result = await daemon.addSyncDevice({ deviceId, name })
      if (result.ok) {
        const syncResult = await daemon.getSyncStatus()
        if (syncResult.ok) {
          setSyncStatus(syncResult.data)
        }
      }
    },
    [daemon],
  )

  const handleRemoveDevice = useCallback(
    async (deviceId: string) => {
      const result = await daemon.removeSyncDevice({ deviceId })
      if (result.ok) {
        const syncResult = await daemon.getSyncStatus()
        if (syncResult.ok) {
          setSyncStatus(syncResult.data)
        }
      }
    },
    [daemon],
  )

  const isApplying = applyStatus === 'applying'
  const showProgress = applyStatus !== 'idle'

  const handleReset = useCallback(() => {
    setApplyStatus('idle')
    setProgressSteps([])
    setError(null)
  }, [])

  return (
    <div className="min-h-screen bg-white">
      <SyncStatusBar status={syncStatus} onOpenSettings={() => setShowSyncSettings(true)} />

      <header className="border-b border-gray-200 py-6 px-8">
        <h1 className="text-2xl font-bold text-gray-900">Kyaraben</h1>
        <p className="text-gray-500">Declarative emulation manager</p>
      </header>

      <main className="max-w-5xl mx-auto px-8 py-6">
        {showProgress ? (
          <>
            <ProgressSteps steps={progressSteps} error={error ?? undefined} />
            {!isApplying && <Button onClick={handleReset}>Done</Button>}
          </>
        ) : (
          <>
            <Settings
              userStore={userStore}
              onUserStoreChange={setUserStore}
              onError={(msg) => showToast(msg, 'error')}
            />

            <SystemGrid
              systems={systems}
              selections={selections}
              provisions={provisions}
              onToggle={handleToggle}
            />

            <div className="mt-6">
              <Button onClick={handleApply} disabled={selections.size === 0}>
                Apply
              </Button>
            </div>
          </>
        )}
      </main>

      {showSyncSettings && (
        <SyncSettings
          status={syncStatus}
          onAddDevice={handleAddDevice}
          onRemoveDevice={handleRemoveDevice}
          onClose={() => setShowSyncSettings(false)}
        />
      )}

      {toast && (
        <Toast message={toast.message} type={toast.type} onDismiss={() => setToast(null)} />
      )}
    </div>
  )
}
