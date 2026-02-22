import { useCallback, useState } from 'react'
import syncthingLogo from '@/assets/syncthing.svg'
import { Button } from '@/lib/Button'
import { useOpenUrl } from '@/lib/hooks/useOpenUrl'
import { Input } from '@/lib/Input'
import { CopyIcon, TrashIcon } from '@/lib/icons'
import { Modal } from '@/lib/Modal'
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
      if (device.completion !== undefined && device.completion < 100) {
        return { dotClass: 'bg-accent', label: `${device.completion}% synced` }
      }
      return { dotClass: 'bg-status-ok', label: 'synced' }
    }
    return { dotClass: 'bg-outline', label: 'offline' }
  }

  const { dotClass, label } = getStatusDisplay()

  return (
    <div className="flex items-center justify-between py-2 border-b border-outline last:border-0">
      <div className="flex items-center gap-2">
        <span className={`w-2 h-2 rounded-full ${dotClass}`} />
        <span className="font-medium text-on-surface">{device.name || 'kyaraben device'}</span>
        <span className="text-on-surface-muted">{label}</span>
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

function PairingCodeDisplay({ code }: { readonly code: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }, [code])

  return (
    <div className="flex flex-col items-center justify-center py-8">
      <code className="bg-surface-raised px-8 py-6 rounded-card font-mono text-5xl tracking-[0.4em] text-accent select-all slashed-zero">
        {code}
      </code>
      <button
        type="button"
        onClick={handleCopy}
        className="mt-4 flex items-center gap-2 px-3 py-2 text-sm text-on-surface-muted hover:text-on-surface rounded"
        title={copied ? 'Copied!' : 'Copy code'}
      >
        <CopyIcon className="w-4 h-4" />
        {copied ? 'Copied!' : 'Copy code'}
      </button>
    </div>
  )
}

interface PrimaryPairingContentProps {
  readonly status: SyncStatusResponse
  readonly isPairing: boolean
  readonly pairingDeviceId: string | null
  readonly pairingCode: string | null
  readonly onStartPairing: () => Promise<void>
  readonly onStopPairing: () => Promise<void>
}

function PrimaryPairingContent({
  status,
  isPairing,
  pairingDeviceId,
  pairingCode,
  onStartPairing,
  onStopPairing,
}: PrimaryPairingContentProps) {
  const [showDeviceId, setShowDeviceId] = useState(false)
  const [isStartingPairing, setIsStartingPairing] = useState(false)
  const hasDevices = (status.devices?.length ?? 0) > 0

  const handleStartPairing = useCallback(async () => {
    setIsStartingPairing(true)
    try {
      await onStartPairing()
    } finally {
      setIsStartingPairing(false)
    }
  }, [onStartPairing])

  if (isStartingPairing || (isPairing && !(pairingCode || pairingDeviceId))) {
    return (
      <div className="mt-4 pt-4 border-t border-outline">
        <h4 className="text-sm font-medium text-on-surface mb-2">Starting pairing</h4>
        <div className="flex items-center gap-3">
          <Spinner />
          <span className="text-sm text-on-surface-muted">Connecting to relay server...</span>
        </div>
      </div>
    )
  }

  if (isPairing && (pairingCode || pairingDeviceId)) {
    return (
      <div className="mt-4 pt-4 border-t border-outline">
        <h4 className="text-sm font-medium text-on-surface mb-2">Pairing mode</h4>
        {pairingCode ? (
          <>
            <p className="text-sm text-on-surface-muted mb-3">
              Enter this code on your secondary device to pair.
            </p>
            <PairingCodeDisplay code={pairingCode} />
          </>
        ) : (
          <>
            <p className="text-sm text-on-surface-muted mb-3">
              Share this device ID with your secondary device.
            </p>
            {pairingDeviceId && <DeviceIdDisplay deviceId={pairingDeviceId} />}
          </>
        )}
        <div className="flex items-center gap-3 mt-4">
          <Spinner />
          <span className="text-sm text-on-surface-muted">Waiting for devices to connect...</span>
        </div>
        <div className="mt-4 flex items-center gap-3">
          <Button variant="secondary" onClick={onStopPairing}>
            Stop pairing
          </Button>
          {pairingCode && pairingDeviceId && (
            <button
              type="button"
              onClick={() => setShowDeviceId(!showDeviceId)}
              className="text-sm text-accent hover:underline"
            >
              {showDeviceId ? 'Hide device ID' : 'Show device ID'}
            </button>
          )}
        </div>
        {showDeviceId && pairingDeviceId && (
          <div className="mt-3">
            <DeviceIdDisplay deviceId={pairingDeviceId} />
          </div>
        )}
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
      <Button onClick={handleStartPairing} disabled={isStartingPairing}>
        Start pairing
      </Button>
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
  readonly onClearConnectionError: () => void
}

function SecondaryDiscoveryContent({
  status,
  discoveredDevices,
  connectionProgress,
  connectionError,
  isDiscovering,
  isConnecting,
  onConnectToDevice,
  onClearConnectionError,
}: SecondaryDiscoveryContentProps) {
  const [pairingCode, setPairingCode] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [manualDeviceId, setManualDeviceId] = useState('')
  const hasDevices = (status.devices?.length ?? 0) > 0

  const handleCodeChange = useCallback(
    (value: string) => {
      setPairingCode(value)
      if (connectionError) {
        onClearConnectionError()
      }
    },
    [connectionError, onClearConnectionError],
  )

  const handleConnectWithCode = useCallback(async () => {
    const trimmed = pairingCode.trim().toUpperCase()
    if (trimmed) {
      const result = await onConnectToDevice(trimmed)
      if (result.ok) {
        setPairingCode('')
      }
    }
  }, [pairingCode, onConnectToDevice])

  const handleConnectManual = useCallback(async () => {
    const trimmed = manualDeviceId.trim().toUpperCase()
    if (trimmed) {
      const result = await onConnectToDevice(trimmed)
      if (result.ok) {
        setManualDeviceId('')
        setShowAdvanced(false)
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

  const canSubmit = pairingCode.trim().length > 0 && !isConnecting

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (canSubmit) {
      handleConnectWithCode()
    }
  }

  return (
    <div className="mt-4 pt-4 border-t border-outline">
      <h4 className="text-sm font-medium text-on-surface mb-2">Connect to primary</h4>
      <p className="text-sm text-on-surface-muted mb-3">
        Enter the pairing code from your primary device.
      </p>

      <form onSubmit={handleSubmit} className="flex gap-3 mb-4">
        <Input
          value={pairingCode}
          onChange={handleCodeChange}
          placeholder="ABC123"
          className="flex-1 font-mono text-xl tracking-[0.2em] uppercase text-center slashed-zero py-3"
          enterKeyHint="go"
        />
        <Button type="submit" disabled={!canSubmit} className="px-6">
          Connect
        </Button>
      </form>

      {connectionError && (
        <div className="p-3 bg-status-error/10 border border-status-error/30 rounded text-sm text-status-error mb-4">
          {connectionError}
        </div>
      )}

      <button
        type="button"
        onClick={() => setShowAdvanced(!showAdvanced)}
        className="text-sm text-accent hover:underline"
      >
        {showAdvanced ? 'Hide advanced options' : 'Advanced options'}
      </button>

      {showAdvanced && (
        <div className="mt-4 space-y-4">
          {isDiscovering && discoveredDevices.length === 0 && (
            <div className="flex items-center gap-3">
              <Spinner />
              <span className="text-sm text-on-surface-muted">
                Searching for devices on local network...
              </span>
            </div>
          )}

          {discoveredDevices.length > 0 && (
            <div>
              <p className="text-sm text-on-surface-muted mb-2">Devices found on local network:</p>
              <div className="border border-outline rounded-card px-3 bg-surface">
                {discoveredDevices.map((device) => (
                  <DiscoveredDeviceRow
                    key={device.deviceId}
                    device={device}
                    onConnect={() => onConnectToDevice(device.deviceId)}
                    isConnecting={isConnecting}
                  />
                ))}
              </div>
            </div>
          )}

          <div>
            <p className="text-sm text-on-surface-muted mb-2">Or enter a device ID manually:</p>
            <div className="space-y-3">
              <Input
                value={manualDeviceId}
                onChange={setManualDeviceId}
                placeholder="XXXXXXX-XXXXXXX-XXXXXXX-..."
              />
              <Button
                onClick={handleConnectManual}
                disabled={!manualDeviceId.trim() || isConnecting}
              >
                Connect with device ID
              </Button>
            </div>
          </div>
        </div>
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
  readonly pairingCode: string | null
  readonly onRemoveDevice: (deviceId: string) => Promise<void>
  readonly onConnectToDevice: (deviceId: string) => Promise<{ ok: boolean; error?: string }>
  readonly onStartPairing: () => Promise<void>
  readonly onStopPairing: () => Promise<void>
  readonly onClearConnectionError: () => void
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
  pairingCode,
  onRemoveDevice,
  onConnectToDevice,
  onStartPairing,
  onStopPairing,
  onClearConnectionError,
}: StatusCardProps) {
  const [deviceToRemove, setDeviceToRemove] = useState<SyncDevice | null>(null)
  const connectedDevice = status.devices?.find((d) => d.connected)
  const hasDevices = (status.devices?.length ?? 0) > 0
  const openUrl = useOpenUrl()

  const handleConfirmRemove = useCallback(async () => {
    if (deviceToRemove) {
      await onRemoveDevice(deviceToRemove.id)
      setDeviceToRemove(null)
    }
  }, [deviceToRemove, onRemoveDevice])

  return (
    <div className="p-4 bg-surface-alt rounded-card">
      <Modal
        open={deviceToRemove !== null}
        onClose={() => setDeviceToRemove(null)}
        title="Remove device"
      >
        <p className="text-on-surface-secondary mb-4">
          Are you sure you want to remove{' '}
          <span className="font-medium text-on-surface">
            {deviceToRemove?.name || 'this device'}
          </span>
          ?
        </p>
        <p className="text-sm text-on-surface-muted mb-6">
          This device will no longer synchronize with you. You can pair again later if needed.
        </p>
        <div className="flex gap-3 justify-end">
          <Button variant="secondary" onClick={() => setDeviceToRemove(null)}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleConfirmRemove}>
            Remove device
          </Button>
        </div>
      </Modal>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <StatusBadge label={status.mode ?? 'unknown'} ok={true} />
          <StatusBadge label="running" ok={true} />
          {connectedDevice && <StatusBadge label="connected" ok={true} />}
        </div>
        <button
          type="button"
          onClick={() => openUrl('https://syncthing.net')}
          className="flex items-center gap-1 text-xs text-on-surface-muted hover:text-on-surface transition-colors"
        >
          <span>powered by</span>
          <img src={syncthingLogo} alt="Syncthing" className="h-5" />
        </button>
      </div>

      {hasDevices && (
        <div className="mt-3 border border-outline rounded-card px-3 bg-surface">
          {status.devices?.map((device) => (
            <DeviceRow key={device.id} device={device} onRemove={() => setDeviceToRemove(device)} />
          ))}
        </div>
      )}

      {status.mode === 'primary' && (
        <PrimaryPairingContent
          status={status}
          isPairing={isPairing}
          pairingDeviceId={pairingDeviceId}
          pairingCode={pairingCode}
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
          onClearConnectionError={onClearConnectionError}
        />
      )}
    </div>
  )
}
