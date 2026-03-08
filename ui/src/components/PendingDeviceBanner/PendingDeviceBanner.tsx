import { useState } from 'react'
import { ActionBanner } from '@/lib/ActionBanner'
import { Button } from '@/lib/Button'
import type { PendingDevice } from '@/lib/hooks/useSyncPairing'

export interface PendingDeviceBannerProps {
  readonly pendingDevice: PendingDevice
  readonly onAccept: (accept: boolean) => Promise<void>
}

export function PendingDeviceBanner({ pendingDevice, onAccept }: PendingDeviceBannerProps) {
  const [isProcessing, setIsProcessing] = useState(false)

  const handleAccept = async () => {
    setIsProcessing(true)
    try {
      await onAccept(true)
    } finally {
      setIsProcessing(false)
    }
  }

  const handleReject = async () => {
    setIsProcessing(true)
    try {
      await onAccept(false)
    } finally {
      setIsProcessing(false)
    }
  }

  const displayName = pendingDevice.name || 'Unknown device'

  return (
    <ActionBanner
      variant="success"
      title="Device wants to connect"
      description={
        <p className="text-sm text-on-surface-muted">
          <span className="font-medium text-on-surface">{displayName}</span> is requesting to sync
          with this device.
        </p>
      }
      actions={
        <>
          <Button variant="secondary" onClick={handleReject} disabled={isProcessing}>
            Reject
          </Button>
          <Button onClick={handleAccept} disabled={isProcessing}>
            {isProcessing ? 'Processing...' : 'Accept'}
          </Button>
        </>
      }
    />
  )
}
