import { useCallback } from 'react'
import { openLogTail } from '@/lib/daemon'
import { useToast } from '@/lib/ToastContext'

export function useOpenLog() {
  const { showToast } = useToast()

  return useCallback(
    async (position?: number) => {
      const result = await openLogTail(position)
      if (!result.ok) {
        showToast('Failed to open log.', 'error')
        return
      }
      if (!result.data.success && result.data.command) {
        showToast(`No terminal found. Run manually: ${result.data.command}.`, 'info', 10000)
      }
    },
    [showToast],
  )
}
