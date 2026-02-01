import { useEffect } from 'react'

export interface ToastProps {
  message: string
  type?: 'error' | 'success' | 'info'
  onDismiss: () => void
  duration?: number
}

export function Toast({ message, type = 'info', onDismiss, duration = 4000 }: ToastProps) {
  useEffect(() => {
    const timer = setTimeout(onDismiss, duration)
    return () => clearTimeout(timer)
  }, [onDismiss, duration])

  const bgColor = {
    error: 'bg-red-600',
    success: 'bg-green-600',
    info: 'bg-gray-800',
  }[type]

  return (
    <div
      className={`fixed bottom-4 right-4 ${bgColor} text-white px-4 py-3 rounded-lg shadow-lg max-w-sm`}
    >
      <div className="flex items-center gap-3">
        <span className="text-sm">{message}</span>
        <button type="button" onClick={onDismiss} className="text-white/70 hover:text-white">
          ✕
        </button>
      </div>
    </div>
  )
}
