import { createContext, type ReactNode, useCallback, useContext, useState } from 'react'

interface Toast {
  id: number
  content: ReactNode
  type: 'error' | 'success' | 'info'
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
      setToasts((prev) => [...prev, { id, content, type }])
      setTimeout(() => {
        setToasts((prev) => prev.filter((t) => t.id !== id))
      }, duration)
    },
    [],
  )

  const dismiss = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const getStyles = (type: Toast['type']) => {
    switch (type) {
      case 'error':
        return 'bg-red-900/90 text-red-100 border-red-700/50'
      case 'success':
        return 'bg-emerald-900/90 text-emerald-100 border-emerald-700/50'
      default:
        return 'bg-gray-800/90 text-gray-200 border-gray-700/50'
    }
  }

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      <div className="fixed top-4 left-1/2 -translate-x-1/2 z-50 flex flex-col gap-2">
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className={`
              ${getStyles(toast.type)}
              px-3 py-1.5 rounded border text-xs backdrop-blur-xs
              shadow-xs flex items-center gap-2
            `}
          >
            {toast.content}
            <button
              type="button"
              onClick={() => dismiss(toast.id)}
              className="opacity-50 hover:opacity-100 transition-opacity"
            >
              ✕
            </button>
          </div>
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
