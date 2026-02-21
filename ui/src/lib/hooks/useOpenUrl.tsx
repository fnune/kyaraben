import { useCallback } from 'react'
import { openUrl } from '@/lib/daemon'
import { useToast } from '@/lib/ToastContext'

function truncateUrl(url: string, maxLength = 40): string {
  if (url.length <= maxLength) return url
  const start = url.slice(0, maxLength - 3)
  return `${start}...`
}

export function useOpenUrl() {
  const { showToast } = useToast()

  return useCallback(
    async (url: string) => {
      const result = await openUrl(url)
      if (result.ok) {
        showToast(
          <span>
            Opening <span className="font-mono">{truncateUrl(url)}</span>.
          </span>,
          'info',
          2000,
        )
      }
    },
    [showToast],
  )
}
