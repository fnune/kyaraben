import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react'

interface Toast {
  id: number
  content: ReactNode
  type: 'error' | 'success' | 'info' | 'warning'
  expiresAt: number
}

interface ToastContextValue {
  showToast: (content: ReactNode, type?: Toast['type'], duration?: number) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

let nextId = 0

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const showToast = useCallback(
    (content: ReactNode, type: Toast['type'] = 'info', duration = 5000) => {
      const id = nextId++
      const expiresAt = Date.now() + duration
      setToasts((prev) => [...prev, { id, content, type, expiresAt }])
    },
    [],
  )

  const dismiss = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const getStyles = (type: Toast['type']) => {
    switch (type) {
      case 'error':
        return 'bg-status-error/90 text-white border-status-error/50'
      case 'success':
        return 'bg-status-ok/90 text-white border-status-ok/50'
      case 'warning':
        return 'bg-status-warning/90 text-white border-status-warning/50'
      default:
        return 'bg-on-surface/90 text-surface border-on-surface-muted/50'
    }
  }

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      <div className="fixed top-4 left-1/2 -translate-x-1/2 z-50 flex flex-col gap-2">
        {toasts.map((toast) => (
          <ToastItem key={toast.id} toast={toast} onDismiss={dismiss} getStyles={getStyles} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  const context = useContext(ToastContext)
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider')
  }
  return context
}

function ToastItem({
  toast,
  onDismiss,
  getStyles,
}: {
  readonly toast: Toast
  readonly onDismiss: (id: number) => void
  readonly getStyles: (type: Toast['type']) => string
}) {
  const timeoutRef = useRef<number | null>(null)
  const remainingRef = useRef<number>(toast.expiresAt - Date.now())

  const clearTimer = useCallback(() => {
    if (timeoutRef.current !== null) {
      window.clearTimeout(timeoutRef.current)
      timeoutRef.current = null
    }
  }, [])

  const startTimer = useCallback(() => {
    clearTimer()
    const remaining = remainingRef.current
    if (remaining <= 0) {
      onDismiss(toast.id)
      return
    }
    if (!Number.isFinite(remaining)) {
      return
    }
    timeoutRef.current = window.setTimeout(() => {
      onDismiss(toast.id)
    }, remaining)
  }, [onDismiss, toast.id, clearTimer])

  useEffect(() => {
    remainingRef.current = toast.expiresAt - Date.now()
    startTimer()
    return () => clearTimer()
  }, [startTimer, toast.expiresAt, clearTimer])

  return (
    <output
      className={`
        ${getStyles(toast.type)}
        px-3 py-1.5 rounded border text-xs backdrop-blur-xs
        shadow-xs flex items-center gap-2
      `}
      onMouseEnter={() => {
        remainingRef.current = toast.expiresAt - Date.now()
        clearTimer()
      }}
      onMouseLeave={() => {
        startTimer()
      }}
    >
      {toast.content}
      <button
        type="button"
        onClick={() => onDismiss(toast.id)}
        className="opacity-50 hover:opacity-100 transition-opacity"
      >
        ✕
      </button>
    </output>
  )
}
