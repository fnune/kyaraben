import {
  createContext,
  type MutableRefObject,
  type ReactNode,
  useCallback,
  useContext,
  useRef,
  useState,
} from 'react'
import * as daemon from '@/lib/daemon'
import { installApp } from '@/lib/daemon'
import { useToast } from '@/lib/ToastContext'
import type { PreflightResponse, SyncPendingResponse } from '@/types/daemon'
import type { LogEntry } from '@/types/logging.gen'
import type { ApplyStatus, ProgressStep } from '@/types/ui'

const PROGRESS_STEP_LABELS: Readonly<Record<string, string>> = {
  'sync-pause': 'Pausing synchronization',
  summary: 'Applying configuration',
  store: 'Setting up collection',
  build: 'Installing emulators',
  cleanup: 'Cleaning up',
  finalize: 'Finalizing',
  'sync-resume': 'Resuming synchronization',
}

interface ApplyConfig {
  collection: string
  graphics?: { preset: string; bezels: boolean; target?: string }
  savestate?: { resume: string }
  controller?: { nintendoConfirm: string }
  systems: Record<string, string[]>
  emulators: Record<string, { version?: string; preset?: string | null; resume?: string | null }>
  frontends?: Record<string, { enabled: boolean; version?: string }>
  summaryMessage?: string
}

function hasUserConflicts(data: PreflightResponse): boolean {
  return (data.diffs ?? []).some(
    (d) => d.userModified && d.hasChanges && (d.userChanges?.length ?? 0) > 0,
  )
}

function toEmulatorConfRequest(
  emulators: Record<string, { version?: string; preset?: string | null; resume?: string | null }>,
): Record<string, { version?: string; preset?: string; resume?: string }> {
  const result: Record<string, { version?: string; preset?: string; resume?: string }> = {}
  for (const [id, conf] of Object.entries(emulators)) {
    result[id] = {
      ...(conf.version !== undefined && { version: conf.version }),
      ...(conf.preset !== null && conf.preset !== undefined && { preset: conf.preset }),
      ...(conf.resume !== null && conf.resume !== undefined && { resume: conf.resume }),
    }
  }
  return result
}

function hasKyarabenUpdates(data: PreflightResponse): boolean {
  return (data.diffs ?? []).some((d) => d.kyarabenChanged && (d.kyarabenUpdates?.length ?? 0) > 0)
}

function needsReview(data: PreflightResponse): boolean {
  return hasUserConflicts(data) || hasKyarabenUpdates(data)
}

interface ApplyContextValue {
  status: ApplyStatus
  progressSteps: readonly ProgressStep[]
  error: string | null
  preflightData: PreflightResponse | null
  syncPendingData: SyncPendingResponse | null
  logPosition: number | null
  apply: (config: ApplyConfig) => Promise<boolean>
  confirmApply: () => Promise<boolean>
  confirmSyncPending: () => Promise<boolean>
  cancel: () => Promise<void>
  reset: () => void
  onCompleteRef: MutableRefObject<(() => void | Promise<void>) | null>
}

const ApplyContext = createContext<ApplyContextValue | null>(null)

export function ApplyProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<ApplyStatus>('idle')
  const [progressSteps, setProgressSteps] = useState<readonly ProgressStep[]>([])
  const [error, setError] = useState<string | null>(null)
  const [preflightData, setPreflightData] = useState<PreflightResponse | null>(null)
  const [syncPendingData, setSyncPendingData] = useState<SyncPendingResponse | null>(null)
  const [logPosition, setLogPosition] = useState<number | null>(null)
  const onCompleteRef = useRef<(() => void | Promise<void>) | null>(null)
  const summaryMessageRef = useRef<string | null>(null)
  const { showToast } = useToast()

  const runApply = useCallback(async (): Promise<boolean> => {
    setStatus('applying')
    setProgressSteps([])
    setLogPosition(null)

    const MAX_OUTPUT_LINES = 10000
    const MAX_LOG_ENTRIES = 10000
    let logPositionCaptured = false

    const progressHandler = (data: {
      step: string
      message?: string
      output?: string
      buildPhase?: string
      packageName?: string
      progressPercent?: number
      bytesDownloaded?: number
      bytesTotal?: number
      bytesPerSecond?: number
      logPosition?: number
      logEntry?: LogEntry
    }) => {
      if (!logPositionCaptured && data.logPosition !== undefined) {
        setLogPosition(data.logPosition)
        logPositionCaptured = true
      }
      setProgressSteps((prev) => {
        const existing = prev.find((s) => s.id === data.step)
        const isNewStep = !existing

        const effectiveMessage =
          data.step === 'summary' && summaryMessageRef.current
            ? summaryMessageRef.current
            : data.message

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
            const nextProgressPercent =
              data.progressPercent !== undefined
                ? Math.max(s.progressPercent ?? 0, data.progressPercent)
                : s.progressPercent
            return {
              ...s,
              status: 'in_progress' as const,
              ...(effectiveMessage && { message: effectiveMessage }),
              ...(data.output && {
                output: [...(s.output ?? []), data.output].slice(-MAX_OUTPUT_LINES),
              }),
              ...(data.logEntry && {
                logEntries: [...(s.logEntries ?? []), data.logEntry].slice(-MAX_LOG_ENTRIES),
              }),
              ...(data.buildPhase && { buildPhase: data.buildPhase }),
              ...(data.packageName && { packageName: data.packageName }),
              ...(nextProgressPercent !== undefined && { progressPercent: nextProgressPercent }),
              ...(data.bytesDownloaded !== undefined && { bytesDownloaded: data.bytesDownloaded }),
              ...(data.bytesTotal !== undefined && { bytesTotal: data.bytesTotal }),
              ...(data.bytesPerSecond !== undefined && { bytesPerSecond: data.bytesPerSecond }),
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
        setStatus('error')
        setProgressSteps((prev) =>
          prev.map((s) => ({ ...s, status: s.status === 'in_progress' ? 'error' : s.status })),
        )
        return false
      }

      if (applyResult.data.cancelled) {
        setStatus('cancelled')
        setProgressSteps((prev) =>
          prev.map((s) => ({
            ...s,
            status: s.status === 'in_progress' ? 'cancelled' : s.status,
          })),
        )
        showToast('Installation cancelled.', 'info')
        return false
      }

      setProgressSteps((prev) => prev.map((s) => ({ ...s, status: 'completed' as const })))
      await onCompleteRef.current?.()
      setStatus('success')

      installApp().catch((err) => {
        console.error('Failed to install Kyaraben:', err)
      })

      return true
    } catch (err) {
      console.error('Apply failed:', err)
      const message = err instanceof Error ? err.message : String(err)
      setError(message)
      setStatus('error')
      setProgressSteps((prev) =>
        prev.map((s) => ({ ...s, status: s.status === 'in_progress' ? 'error' : s.status })),
      )
      showToast(`Installation failed: ${message}`, 'error')
      return false
    } finally {
      window.electron.off('apply:progress')
    }
  }, [showToast])

  const apply = useCallback(
    async (config: ApplyConfig): Promise<boolean> => {
      setError(null)
      setPreflightData(null)
      summaryMessageRef.current = config.summaryMessage ?? null

      const configResult = await daemon.setConfig({
        collection: config.collection,
        systems: config.systems,
        emulators: toEmulatorConfRequest(config.emulators),
        ...(config.graphics && { graphics: config.graphics }),
        ...(config.savestate && { savestate: config.savestate }),
        ...(config.controller && { controller: config.controller }),
        ...(config.frontends && { frontends: config.frontends }),
      })

      if (!configResult.ok) {
        setError(configResult.error.message)
        setStatus('error')
        return false
      }

      const preflightResult = await daemon.preflight()

      if (!preflightResult.ok) {
        setError(preflightResult.error.message)
        setStatus('error')
        return false
      }

      if (needsReview(preflightResult.data)) {
        setPreflightData(preflightResult.data)
        setStatus('reviewing')
        return false
      }

      const syncPendingResult = await daemon.getSyncPending()
      if (syncPendingResult.ok && syncPendingResult.data.pending) {
        setSyncPendingData(syncPendingResult.data)
        setStatus('confirming_sync')
        return false
      }

      return runApply()
    },
    [runApply],
  )

  const confirmApply = useCallback(async (): Promise<boolean> => {
    setPreflightData(null)

    const syncPendingResult = await daemon.getSyncPending()
    if (syncPendingResult.ok && syncPendingResult.data.pending) {
      setSyncPendingData(syncPendingResult.data)
      setStatus('confirming_sync')
      return false
    }

    return runApply()
  }, [runApply])

  const confirmSyncPending = useCallback(async (): Promise<boolean> => {
    setSyncPendingData(null)
    return runApply()
  }, [runApply])

  const cancel = useCallback(async () => {
    await daemon.cancelApply()
  }, [])

  const reset = useCallback(() => {
    setStatus('idle')
    setProgressSteps([])
    setError(null)
    setPreflightData(null)
    setSyncPendingData(null)
    setLogPosition(null)
  }, [])

  return (
    <ApplyContext.Provider
      value={{
        status,
        progressSteps,
        error,
        preflightData,
        syncPendingData,
        logPosition,
        apply,
        confirmApply,
        confirmSyncPending,
        cancel,
        reset,
        onCompleteRef,
      }}
    >
      {children}
    </ApplyContext.Provider>
  )
}

export function useApply() {
  const context = useContext(ApplyContext)
  if (!context) {
    throw new Error('useApply must be used within an ApplyProvider')
  }
  return context
}
