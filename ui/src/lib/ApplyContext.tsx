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
import type { PreflightResponse } from '@/types/daemon'
import type { ApplyStatus, ProgressStep } from '@/types/ui'

const PROGRESS_STEP_LABELS: Readonly<Record<string, string>> = {
  summary: 'Applying configuration',
  store: 'Setting up emulation folder',
  build: 'Installing emulators',
  cleanup: 'Cleaning up',
  finalize: 'Finalizing',
}

interface ApplyConfig {
  userStore: string
  systems: Record<string, string[]>
  emulators: Record<string, { version?: string }>
  frontends?: Record<string, { enabled: boolean; version?: string }>
  summaryMessage?: string
}

function hasConflicts(data: PreflightResponse): boolean {
  return (data.diffs ?? []).some(
    (d) => d.userModified && d.hasChanges && (d.userChanges?.length ?? 0) > 0,
  )
}

interface ApplyContextValue {
  status: ApplyStatus
  progressSteps: readonly ProgressStep[]
  error: string | null
  preflightData: PreflightResponse | null
  logPosition: number | null
  apply: (config: ApplyConfig) => Promise<boolean>
  confirmApply: () => Promise<boolean>
  cancel: () => Promise<void>
  reset: () => void
  onCompleteRef: MutableRefObject<(() => void) | null>
}

const ApplyContext = createContext<ApplyContextValue | null>(null)

export function ApplyProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<ApplyStatus>('idle')
  const [progressSteps, setProgressSteps] = useState<readonly ProgressStep[]>([])
  const [error, setError] = useState<string | null>(null)
  const [preflightData, setPreflightData] = useState<PreflightResponse | null>(null)
  const [logPosition, setLogPosition] = useState<number | null>(null)
  const onCompleteRef = useRef<(() => void) | null>(null)
  const summaryMessageRef = useRef<string | null>(null)
  const { showToast } = useToast()

  const runApply = useCallback(async (): Promise<boolean> => {
    setStatus('applying')
    setProgressSteps([])
    setLogPosition(null)

    const MAX_OUTPUT_LINES = 10000
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
      setStatus('success')
      onCompleteRef.current?.()

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
      showToast('Installation failed.', 'error')
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
        userStore: config.userStore,
        systems: config.systems,
        emulators: config.emulators,
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

      if (hasConflicts(preflightResult.data)) {
        setPreflightData(preflightResult.data)
        setStatus('reviewing')
        return false
      }

      return runApply()
    },
    [runApply],
  )

  const confirmApply = useCallback(async (): Promise<boolean> => {
    setPreflightData(null)
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
    setLogPosition(null)
  }, [])

  return (
    <ApplyContext.Provider
      value={{
        status,
        progressSteps,
        error,
        preflightData,
        logPosition,
        apply,
        confirmApply,
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
