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
import type { ApplyStatus, ProgressStep } from '@/types/ui'

const PROGRESS_STEP_LABELS: Readonly<Record<string, string>> = {
  store: 'Setting up emulation folder',
  build: 'Installing emulators',
  desktop: 'Adding to application menu',
  config: 'Configuring emulators',
}

interface ApplyConfig {
  userStore: string
  systems: Record<string, string[]>
  emulators: Record<string, { version?: string }>
}

interface ApplyContextValue {
  status: ApplyStatus
  progressSteps: readonly ProgressStep[]
  error: string | null
  apply: (config: ApplyConfig) => Promise<boolean>
  cancel: () => Promise<void>
  reset: () => void
  onCompleteRef: MutableRefObject<(() => void) | null>
}

const ApplyContext = createContext<ApplyContextValue | null>(null)

export function ApplyProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<ApplyStatus>('idle')
  const [progressSteps, setProgressSteps] = useState<readonly ProgressStep[]>([])
  const [error, setError] = useState<string | null>(null)
  const progressHandlerRef = useRef<((...args: unknown[]) => void) | null>(null)
  const onCompleteRef = useRef<(() => void) | null>(null)
  const { showToast } = useToast()

  const apply = useCallback(async (config: ApplyConfig): Promise<boolean> => {
    setStatus('applying')
    setProgressSteps([])
    setError(null)

    const configResult = await daemon.setConfig({
      userStore: config.userStore,
      systems: config.systems,
      emulators: config.emulators,
    })

    if (!configResult.ok) {
      setError(configResult.error.message)
      setStatus('error')
      return false
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

    progressHandlerRef.current = progressHandler
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
        showToast('Installation cancelled', 'info')
        return false
      }

      setProgressSteps((prev) => prev.map((s) => ({ ...s, status: 'completed' as const })))
      setStatus('success')
      onCompleteRef.current?.()

      installApp().catch((err) => {
        console.error('Failed to install Kyaraben:', err)
      })

      showToast('Installation complete', 'success')
      return true
    } catch (err) {
      console.error('Apply failed:', err)
      const message = err instanceof Error ? err.message : String(err)
      setError(message)
      setStatus('error')
      setProgressSteps((prev) =>
        prev.map((s) => ({ ...s, status: s.status === 'in_progress' ? 'error' : s.status })),
      )
      showToast('Installation failed', 'error')
      return false
    } finally {
      if (progressHandlerRef.current) {
        window.electron.off('apply:progress', progressHandlerRef.current)
        progressHandlerRef.current = null
      }
    }
  }, [showToast])

  const cancel = useCallback(async () => {
    await daemon.cancelApply()
  }, [])

  const reset = useCallback(() => {
    setStatus('idle')
    setProgressSteps([])
    setError(null)
  }, [])

  return (
    <ApplyContext.Provider
      value={{ status, progressSteps, error, apply, cancel, reset, onCompleteRef }}
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
