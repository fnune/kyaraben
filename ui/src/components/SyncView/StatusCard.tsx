import { useCallback, useState } from 'react'
import { Button } from '@/lib/Button'
import { Input } from '@/lib/Input'
import { CopyIcon, TrashIcon } from '@/lib/icons'
import { Spinner } from '@/lib/Spinner'
import type { SyncDevice, SyncDiscoveredDevice, SyncStatusResponse } from '@/types/daemon'

function StatusBadge({ label, ok }: { label: string; ok: boolean }) {
  const capitalizedLabel = label.charAt(0).toUpperCase() + label.slice(1)
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-sm text-xs ${
        ok ? 'bg-status-ok/10 text-status-ok' : 'bg-outline/50 text-on-surface-muted'
      }`}
    >
      <span className={`w-1.5 h-1.5 rounded-full ${ok ? 'bg-status-ok' : 'bg-outline'}`} />
      {capitalizedLabel}
    </span>
  )
}

function DeviceRow({
  device,
  onRemove,
}: {
  readonly device: SyncDevice
  readonly onRemove: () => void
}) {
  const getStatusDisplay = () => {
    if (device.paused) {
      return { dotClass: 'bg-status-warning', label: 'paused' }
    }
    if (device.connected) {
      return { dotClass: 'bg-status-ok', label: 'connected' }
    }
    return { dotClass: 'bg-outline', label: 'offline' }
  }

  const { dotClass, label } = getStatusDisplay()

  return (
    <div className="flex items-center justify-between py-2 border-b border-outline last:border-0">
      <div className="flex items-center gap-2">
        <span className={`w-2 h-2 rounded-full ${dotClass}`} />
        <span className="font-medium text-on-surface">{device.name || 'Unknown device'}</span>
        <span className="text-xs text-on-surface-muted">{label}</span>
      </div>
      <button
        type="button"
        onClick={onRemove}
        className="p-1.5 text-on-surface-muted hover:text-on-surface-secondary rounded"
        title="Remove device"
      >
        <TrashIcon className="w-4 h-4" />
      </button>
    </div>
  )
}

function formatDeviceIdGroup(group: string, index: number) {
  if (index === 0) {
    return (
      <span key={index} className="text-accent font-semibold">
        {group}
      </span>
    )
  }
  return (
    <span key={index} className="text-on-surface-muted">
      {group}
    </span>
  )
}

function DeviceIdDisplay({
  deviceId,
  showCopy = true,
}: {
  readonly deviceId: string
  readonly showCopy?: boolean
}) {
  const [copied, setCopied] = useState(false)
  const groups = deviceId.split('-')

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(deviceId)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }, [deviceId])

  return (
    <div className="flex items-center gap-2">
      <code className="bg-surface-raised px-3 py-2 rounded-sm font-mono text-sm tracking-wide">
        {groups.map((group, i) => (
          <span key={`${i}-${group}`}>
            {formatDeviceIdGroup(group, i)}
            {i < groups.length - 1 && <span className="text-outline mx-0.5">-</span>}
          </span>
        ))}
      </code>
      {showCopy && (
        <button
          type="button"
          onClick={handleCopy}
          className="p-2 text-on-surface-muted hover:text-on-surface rounded"
          title={copied ? 'Copied!' : 'Copy device ID'}
        >
          <CopyIcon className="w-4 h-4" />
        </button>
      )}
    </div>
  )
}

function DiscoveredDeviceRow({
  device,
  onConnect,
  isConnecting,
}: {
  readonly device: SyncDiscoveredDevice
  readonly onConnect: () => void
  readonly isConnecting: boolean
}) {
  const groups = device.deviceId.split('-')

  return (
    <div className="flex items-center justify-between py-3 border-b border-outline last:border-0">
      <code className="font-mono text-sm">
        <span className="text-accent font-semibold">{groups[0]}</span>
        <span className="text-outline mx-0.5">-</span>
        <span className="text-on-surface-muted">{groups[1]}</span>
        <span className="text-on-surface-muted">...</span>
      </code>
      <Button size="sm" onClick={onConnect} disabled={isConnecting}>
        {isConnecting ? 'Connecting...' : 'Connect'}
      </Button>
    </div>
  )
}

interface PrimaryPairingContentProps {
  readonly status: SyncStatusResponse
  readonly isPairing: boolean
  readonly pairingDeviceId: string | null
  readonly onStartPairing: () => Promise<void>
  readonly onStopPairing: () => Promise<void>
}

function PrimaryPairingContent({
  status,
  isPairing,
  pairingDeviceId,
  onStartPairing,
  onStopPairing,
}: PrimaryPairingContentProps) {
  const hasDevices = (status.devices?.length ?? 0) > 0

  if (isPairing && pairingDeviceId) {
    return (
      <div className="mt-4 pt-4 border-t border-outline">
        <h4 className="text-sm font-medium text-on-surface mb-2">Pairing mode</h4>
        <p className="text-sm text-on-surface-muted mb-3">
          Share this device ID with your secondary device. It will appear in the Sync tab.
        </p>
        <DeviceIdDisplay deviceId={pairingDeviceId} />
        <div className="flex items-center gap-3 mt-4">
          <Spinner />
          <span className="text-sm text-on-surface-muted">Waiting for devices to connect...</span>
        </div>
        <div className="mt-4">
          <Button variant="secondary" onClick={onStopPairing}>
            Stop pairing
          </Button>
        </div>
      </div>
    )
  }

  if (hasDevices) {
    return null
  }

  return (
    <div className="mt-4 pt-4 border-t border-outline">
      <h4 className="text-sm font-medium text-on-surface mb-2">Pair a device</h4>
      <p className="text-sm text-on-surface-muted mb-3">
        Start pairing to allow secondary devices to connect.
      </p>
      <Button onClick={onStartPairing}>Start pairing</Button>
    </div>
  )
}

interface SecondaryDiscoveryContentProps {
  readonly status: SyncStatusResponse
  readonly discoveredDevices: SyncDiscoveredDevice[]
  readonly connectionProgress: string | null
  readonly connectionError: string | null
  readonly isDiscovering: boolean
  readonly isConnecting: boolean
  readonly onConnectToDevice: (deviceId: string) => Promise<{ ok: boolean; error?: string }>
}

function SecondaryDiscoveryContent({
  status,
  discoveredDevices,
  connectionProgress,
  connectionError,
  isDiscovering,
  isConnecting,
  onConnectToDevice,
}: SecondaryDiscoveryContentProps) {
  const [manualDeviceId, setManualDeviceId] = useState('')
  const [showManualInput, setShowManualInput] = useState(false)
  const hasDevices = (status.devices?.length ?? 0) > 0

  const handleConnectManual = useCallback(async () => {
    const trimmed = manualDeviceId.trim().toUpperCase()
    if (trimmed) {
      const result = await onConnectToDevice(trimmed)
      if (result.ok) {
        setManualDeviceId('')
        setShowManualInput(false)
      }
    }
  }, [manualDeviceId, onConnectToDevice])

  if (hasDevices) {
    return null
  }

  if (isConnecting) {
    return (
      <div className="mt-4 pt-4 border-t border-outline">
        <h4 className="text-sm font-medium text-on-surface mb-2">Connecting to primary</h4>
        <div className="flex items-center gap-3">
          <Spinner />
          <span className="text-sm text-on-surface-muted">
            {connectionProgress || 'Connecting...'}
          </span>
        </div>
      </div>
    )
  }

  return (
    <div className="mt-4 pt-4 border-t border-outline">
      <h4 className="text-sm font-medium text-on-surface mb-2">Connect to primary</h4>
      <p className="text-sm text-on-surface-muted mb-3">
        Select a kyaraben primary device from your network to start syncing.
      </p>

      {isDiscovering && discoveredDevices.length === 0 && (
        <div className="flex items-center gap-3 mb-4">
          <Spinner />
          <span className="text-sm text-on-surface-muted">Searching for devices...</span>
        </div>
      )}

      {discoveredDevices.length > 0 && (
        <div className="border border-outline rounded-card px-3 bg-surface mb-4">
          {discoveredDevices.map((device) => (
            <DiscoveredDeviceRow
              key={device.deviceId}
              device={device}
              onConnect={() => onConnectToDevice(device.deviceId)}
              isConnecting={isConnecting}
            />
          ))}
        </div>
      )}

      {connectionError && (
        <div className="p-3 bg-status-error/10 border border-status-error/30 rounded text-sm text-status-error mb-4">
          {connectionError}
        </div>
      )}

      {showManualInput ? (
        <div className="space-y-3">
          <p className="text-sm text-on-surface-muted">
            Enter the device ID from the primary device:
          </p>
          <Input
            value={manualDeviceId}
            onChange={setManualDeviceId}
            placeholder="XXXXXXX-XXXXXXX-XXXXXXX-..."
          />
          <div className="flex gap-2">
            <Button onClick={handleConnectManual} disabled={!manualDeviceId.trim()}>
              Connect
            </Button>
            <Button variant="secondary" onClick={() => setShowManualInput(false)}>
              Cancel
            </Button>
          </div>
        </div>
      ) : (
        <button
          type="button"
          onClick={() => setShowManualInput(true)}
          className="text-sm text-accent hover:underline"
        >
          Enter device ID manually
        </button>
      )}
    </div>
  )
}

export interface StatusCardProps {
  readonly status: SyncStatusResponse
  readonly discoveredDevices: SyncDiscoveredDevice[]
  readonly connectionProgress: string | null
  readonly connectionError: string | null
  readonly isDiscovering: boolean
  readonly isConnecting: boolean
  readonly isPairing: boolean
  readonly pairingDeviceId: string | null
  readonly onRemoveDevice: (deviceId: string) => Promise<void>
  readonly onConnectToDevice: (deviceId: string) => Promise<{ ok: boolean; error?: string }>
  readonly onStartPairing: () => Promise<void>
  readonly onStopPairing: () => Promise<void>
}

export function StatusCard({
  status,
  discoveredDevices,
  connectionProgress,
  connectionError,
  isDiscovering,
  isConnecting,
  isPairing,
  pairingDeviceId,
  onRemoveDevice,
  onConnectToDevice,
  onStartPairing,
  onStopPairing,
}: StatusCardProps) {
  const connectedDevice = status.devices?.find((d) => d.connected)
  const hasDevices = (status.devices?.length ?? 0) > 0

  return (
    <div className="p-4 bg-surface-alt rounded-card">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <StatusBadge label={status.mode ?? 'unknown'} ok={true} />
          <StatusBadge label="running" ok={true} />
        </div>
      </div>

      <div className="mt-3">
        {connectedDevice ? (
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-status-ok" />
            <span className="text-sm text-on-surface">
              Connected to {connectedDevice.name || 'Unknown device'}
            </span>
          </div>
        ) : hasDevices ? (
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-outline" />
            <span className="text-sm text-on-surface-muted">Device offline</span>
          </div>
        ) : (
          <span className="text-sm text-on-surface-muted">No device connected</span>
        )}
      </div>

      {hasDevices && (
        <div className="mt-3 border border-outline rounded-card px-3 bg-surface">
          {status.devices?.map((device) => (
            <DeviceRow key={device.id} device={device} onRemove={() => onRemoveDevice(device.id)} />
          ))}
        </div>
      )}

      {status.mode === 'primary' && (
        <PrimaryPairingContent
          status={status}
          isPairing={isPairing}
          pairingDeviceId={pairingDeviceId}
          onStartPairing={onStartPairing}
          onStopPairing={onStopPairing}
        />
      )}

      {status.mode === 'secondary' && (
        <SecondaryDiscoveryContent
          status={status}
          discoveredDevices={discoveredDevices}
          connectionProgress={connectionProgress}
          connectionError={connectionError}
          isDiscovering={isDiscovering}
          isConnecting={isConnecting}
          onConnectToDevice={onConnectToDevice}
        />
      )}
    </div>
  )
}
