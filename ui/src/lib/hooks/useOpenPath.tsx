import { useCallback } from 'react'
import { openPath } from '@/lib/daemon'
import { PathText } from '@/lib/PathText'
import { useToast } from '@/lib/ToastContext'

export function useOpenPath() {
  const { showToast } = useToast()

  return useCallback(
    async (path: string) => {
      const result = await openPath(path)
      if (result.ok) {
        showToast(
          <span>
            Opening <PathText>{path}</PathText>.
          </span>,
          'info',
          2000,
        )
      } else {
        showToast(`Could not open: ${result.error.message}.`, 'error')
      }
    },
    [showToast],
  )
}
