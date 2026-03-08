import type { SyncDevice, SyncStatusResponse } from '@shared/daemon'
import { useCallback, useState } from 'react'
import syncthingLogo from '@/assets/syncthing.svg'
import { Button } from '@/lib/Button'
import { useOpenUrl } from '@/lib/hooks/useOpenUrl'
import { Input } from '@/lib/Input'
import { CopyIcon, TrashIcon } from '@/lib/icons'
import { Modal } from '@/lib/Modal'
import { Spinner } from '@/lib/Spinner'

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
  const openUrl = useOpenUrl()

  const getStatusDisplay = () => {
    if (device.paused) {
      return { dotClass: 'bg-status-warning', label: 'paused' }
    }
    if (device.connected) {
      if (device.completion !== undefined && device.completion < 100) {
        return { dotClass: 'bg-accent', label: `${device.completion}% synced` }
      }
      const connectionLabel = device.connectionType ? ` (${device.connectionType})` : ''
      return { dotClass: 'bg-status-ok', label: `synced${connectionLabel}` }
    }
    return { dotClass: 'bg-outline', label: 'offline' }
  }

  const { dotClass, label } = getStatusDisplay()
  const hasConnectivityIssue = device.connectivityIssue === 'port_unreachable'

  return (
    <div className="py-2 border-b border-outline last:border-0">
      <div className="flex items-center justify-between">
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
      {hasConnectivityIssue && (
        <div className="flex items-start gap-2 mt-1 p-2 bg-status-warning/10 border border-status-warning/30 rounded text-xs">
          <span className="text-status-warning w-3 text-center shrink-0">!</span>
          <span className="text-on-surface-muted">
            Port unreachable on peer device.{' '}
            <a
              href="https://docs.syncthing.net/users/firewall.html"
              onClick={(e) => {
                e.preventDefault()
                openUrl('https://docs.syncthing.net/users/firewall.html')
              }}
              className="text-accent underline"
            >
              Learn more about firewall configuration
            </a>
          </span>
        </div>
      )}
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

function PairingCodeDisplay({ code }: { readonly code: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }, [code])

  return (
    <div className="flex flex-col items-center justify-center py-4">
      <code className="bg-surface-raised px-4 py-3 rounded-card font-mono text-3xl sm:text-5xl lg:text-7xl tracking-[0.2em] sm:tracking-[0.4em] text-accent select-all slashed-zero max-w-full overflow-x-auto">
        {code}
      </code>
      <button
        type="button"
        onClick={handleCopy}
        className="mt-2 flex items-center gap-2 px-3 py-2 text-sm text-on-surface-muted hover:text-on-surface rounded"
        title={copied ? 'Copied!' : 'Copy code'}
      >
        <CopyIcon className="w-4 h-4" />
        {copied ? 'Copied!' : 'Copy code'}
      </button>
    </div>
  )
}

function PairingSection({
  title,
  children,
}: {
  readonly title: string
  readonly children: React.ReactNode
}) {
  return (
    <div className="mt-4 pt-4 border-t border-outline">
      <h4 className="text-sm font-medium text-on-surface mb-2">{title}</h4>
      {children}
    </div>
  )
}

function ConnectionError({ message }: { readonly message: string | null }) {
  if (!message) return null
  return (
    <div className="p-3 bg-status-error/10 border border-status-error/30 rounded text-sm text-status-error">
      {message}
    </div>
  )
}

function WaitingForPeer({ className }: { readonly className?: string }) {
  return (
    <div className={`flex items-center gap-3 ${className ?? ''}`}>
      <Spinner />
      <span className="text-sm text-on-surface-muted">Waiting for device to connect...</span>
    </div>
  )
}

interface PairingContentProps {
  readonly status: SyncStatusResponse
  readonly connectionProgress: string | null
  readonly connectionError: string | null
  readonly isConnecting: boolean
  readonly isPairing: boolean
  readonly pairingDeviceId: string | null
  readonly pairingCode: string | null
  readonly onConnectToDevice: (deviceId: string) => Promise<{ ok: boolean; error?: string }>
  readonly onStartPairing: () => Promise<void>
  readonly onStopPairing: () => Promise<void>
  readonly onClearConnectionError: () => void
}

function PairingContent({
  status,
  connectionProgress,
  connectionError,
  isConnecting,
  isPairing,
  pairingDeviceId,
  pairingCode,
  onConnectToDevice,
  onStartPairing,
  onStopPairing,
  onClearConnectionError,
}: PairingContentProps) {
  const [deviceIdMode, setDeviceIdMode] = useState(false)
  const [isStartingPairing, setIsStartingPairing] = useState(false)
  const [joinCode, setJoinCode] = useState('')
  const [otherDeviceId, setOtherDeviceId] = useState('')

  const handleStartPairing = useCallback(async () => {
    setIsStartingPairing(true)
    try {
      await onStartPairing()
    } finally {
      setIsStartingPairing(false)
    }
  }, [onStartPairing])

  const handleCancel = useCallback(async () => {
    setIsStartingPairing(false)
    setDeviceIdMode(false)
    await onStopPairing()
  }, [onStopPairing])

  const handleSwitchToDeviceIdMode = useCallback(() => {
    setDeviceIdMode(true)
  }, [])

  const handleSwitchToPairingCodeMode = useCallback(() => {
    setDeviceIdMode(false)
  }, [])

  const handleCodeChange = useCallback(
    (value: string) => {
      setJoinCode(value)
      if (connectionError) {
        onClearConnectionError()
      }
    },
    [connectionError, onClearConnectionError],
  )

  const handleOtherDeviceIdChange = useCallback(
    (value: string) => {
      setOtherDeviceId(value)
      if (connectionError) {
        onClearConnectionError()
      }
    },
    [connectionError, onClearConnectionError],
  )

  const handleConnectWithCode = useCallback(async () => {
    const trimmed = joinCode.trim().toUpperCase()
    if (trimmed) {
      const result = await onConnectToDevice(trimmed)
      if (result.ok) {
        setJoinCode('')
      }
    }
  }, [joinCode, onConnectToDevice])

  const handleConnectWithDeviceId = useCallback(async () => {
    const trimmed = otherDeviceId.trim().toUpperCase()
    if (trimmed) {
      const result = await onConnectToDevice(trimmed)
      if (result.ok) {
        setOtherDeviceId('')
      }
    }
  }, [otherDeviceId, onConnectToDevice])

  const isGeneratingCode = isStartingPairing || (isPairing && !pairingCode)
  const hasCode = isPairing && pairingCode
  const deviceId = pairingDeviceId || status.deviceId
  const isWaitingForPeer = isPairing || isStartingPairing

  if (isConnecting) {
    return (
      <PairingSection title="Connecting">
        <div className="flex items-center gap-3">
          <Spinner />
          <span className="text-sm text-on-surface-muted">
            {connectionProgress || 'Connecting...'}
          </span>
        </div>
      </PairingSection>
    )
  }

  if (deviceIdMode) {
    return (
      <PairingSection title="Pair with device ID">
        <div className="space-y-4">
          {deviceId && (
            <div>
              <p className="text-sm text-on-surface-muted mb-2">Your device ID:</p>
              <DeviceIdDisplay deviceId={deviceId} />
            </div>
          )}

          <div>
            <p className="text-sm text-on-surface-muted mb-2">Other device's ID:</p>
            <div className="flex gap-2">
              <Input
                value={otherDeviceId}
                onChange={handleOtherDeviceIdChange}
                placeholder="ABC1234-DEF5678-GHI9012-JKL3456-MNO7890-PQR1234-STU5678-VWX9012"
                className="flex-1 font-mono"
              />
              <Button
                onClick={handleConnectWithDeviceId}
                disabled={!otherDeviceId.trim() || isConnecting}
              >
                Connect
              </Button>
            </div>
          </div>

          <ConnectionError message={connectionError} />

          {isWaitingForPeer && <WaitingForPeer />}

          <div className="flex items-center gap-3">
            {isWaitingForPeer && (
              <Button variant="secondary" onClick={handleCancel}>
                Cancel
              </Button>
            )}
            <button
              type="button"
              onClick={handleSwitchToPairingCodeMode}
              className="text-sm text-accent hover:underline"
            >
              Use pairing code instead
            </button>
          </div>
        </div>
      </PairingSection>
    )
  }

  if (isGeneratingCode) {
    return (
      <PairingSection title="Generating pairing code">
        <div className="flex items-center gap-3">
          <Spinner />
          <span className="text-sm text-on-surface-muted">Please wait...</span>
        </div>
        <div className="mt-4 flex items-center gap-3">
          <Button variant="secondary" onClick={handleCancel}>
            Cancel
          </Button>
          <button
            type="button"
            onClick={handleSwitchToDeviceIdMode}
            className="text-sm text-accent hover:underline"
          >
            Use device ID instead
          </button>
        </div>
      </PairingSection>
    )
  }

  if (hasCode) {
    return (
      <PairingSection title="Pairing code">
        <p className="text-sm text-on-surface-muted mb-3">
          Enter this code on the other device to pair.
        </p>
        <PairingCodeDisplay code={pairingCode} />
        <WaitingForPeer className="mt-4" />
        <div className="mt-4 flex items-center gap-3">
          <Button variant="secondary" onClick={handleCancel}>
            Cancel
          </Button>
          <button
            type="button"
            onClick={handleSwitchToDeviceIdMode}
            className="text-sm text-accent hover:underline"
          >
            Use device ID instead
          </button>
        </div>
      </PairingSection>
    )
  }

  const canSubmit = joinCode.trim().length > 0 && !isConnecting

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (canSubmit) {
      handleConnectWithCode()
    }
  }

  return (
    <PairingSection title="Pair a device">
      <div className="space-y-4">
        <div>
          <p className="text-sm text-on-surface-muted mb-2">
            Generate a code for the other device:
          </p>
          <Button onClick={handleStartPairing} disabled={isStartingPairing}>
            Generate pairing code
          </Button>
        </div>

        <div>
          <p className="text-sm text-on-surface-muted mb-2">Or enter a code from another device:</p>
          <form onSubmit={handleSubmit} className="flex gap-2">
            <Input
              value={joinCode}
              onChange={handleCodeChange}
              placeholder="ABC123"
              className="flex-1 font-mono uppercase slashed-zero"
              enterKeyHint="go"
            />
            <Button type="submit" disabled={!canSubmit}>
              Connect
            </Button>
          </form>
        </div>

        <ConnectionError message={connectionError} />

        <button
          type="button"
          onClick={handleSwitchToDeviceIdMode}
          className="text-sm text-accent hover:underline"
        >
          Use device ID instead
        </button>
      </div>
    </PairingSection>
  )
}

export interface StatusCardProps {
  readonly status: SyncStatusResponse
  readonly connectionProgress: string | null
  readonly connectionError: string | null
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
  connectionProgress,
  connectionError,
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

      <div className="flex flex-col items-center gap-2 sm:flex-row sm:justify-between">
        <div className="flex items-center gap-2">
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

      {status.localConnectivityIssue && (
        <div className="flex items-start gap-2 mt-3 p-2 bg-status-warning/10 border border-status-warning/30 rounded text-xs">
          <span className="text-status-warning w-3 text-center shrink-0">!</span>
          <span className="text-on-surface-muted">
            {status.localConnectivityIssue === 'listen_error'
              ? 'Failed to listen on sync port.'
              : 'Other devices may not be able to connect.'}{' '}
            <a
              href="https://docs.syncthing.net/users/firewall.html"
              onClick={(e) => {
                e.preventDefault()
                openUrl('https://docs.syncthing.net/users/firewall.html')
              }}
              className="text-accent underline"
            >
              Learn more about firewall configuration
            </a>
          </span>
        </div>
      )}

      <PairingContent
        status={status}
        connectionProgress={connectionProgress}
        connectionError={connectionError}
        isConnecting={isConnecting}
        isPairing={isPairing}
        pairingDeviceId={pairingDeviceId}
        pairingCode={pairingCode}
        onConnectToDevice={onConnectToDevice}
        onStartPairing={onStartPairing}
        onStopPairing={onStopPairing}
        onClearConnectionError={onClearConnectionError}
      />
    </div>
  )
}
