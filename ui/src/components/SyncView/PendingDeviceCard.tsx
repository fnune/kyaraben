import { useState } from 'react'
import { Button } from '@/lib/Button'
import type { PendingDevice } from '@/lib/hooks/useSyncPairing'

export interface PendingDeviceCardProps {
  readonly pendingDevice: PendingDevice
  readonly onAccept: (accept: boolean) => Promise<void>
}

export function PendingDeviceCard({ pendingDevice, onAccept }: PendingDeviceCardProps) {
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
    <div className="p-4 bg-status-success/10 border border-status-success/30 rounded-card">
      <h3 className="text-sm font-medium text-on-surface mb-2">Device wants to connect</h3>
      <p className="text-sm text-on-surface-muted mb-4">
        <span className="font-medium text-on-surface">{displayName}</span> is requesting to sync
        with this device.
      </p>
      <div className="flex gap-2">
        <Button onClick={handleAccept} disabled={isProcessing}>
          {isProcessing ? 'Processing...' : 'Accept'}
        </Button>
        <Button variant="secondary" onClick={handleReject} disabled={isProcessing}>
          Reject
        </Button>
      </div>
    </div>
  )
}
