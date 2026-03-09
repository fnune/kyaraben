import { useState } from 'react'
import { ActionBanner } from '@/lib/ActionBanner'
import { Button } from '@/lib/Button'
import { useOpenUrl } from '@/lib/hooks/useOpenUrl'

const STORAGE_KEY = 'kyaraben-alpha-warning-dismissed'

export function AlphaWarningBanner() {
  const openUrl = useOpenUrl()
  const [dismissed, setDismissed] = useState(() => {
    return localStorage.getItem(STORAGE_KEY) === 'true'
  })

  if (dismissed) return null

  const handleDismiss = () => {
    localStorage.setItem(STORAGE_KEY, 'true')
    setDismissed(true)
  }

  const handleLearnMore = () => {
    openUrl('https://kyaraben.org/using-the-app/')
  }

  return (
    <ActionBanner
      variant="warning"
      title={
        <>
          Kyaraben is pre-alpha software. On first run, it may overwrite existing emulator
          configurations and saves stored in emulator directories. If you have an existing setup,
          back everything up first. For the smoothest experience, start fresh without existing
          emulator installations.
        </>
      }
      actions={
        <>
          <Button variant="secondary" onClick={handleLearnMore}>
            Learn more
          </Button>
          <Button onClick={handleDismiss}>Dismiss</Button>
        </>
      }
    />
  )
}
