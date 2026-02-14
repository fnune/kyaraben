import { useState, useCallback } from 'react'
import { Button } from '@/lib/Button'
import { Input } from '@/lib/Input'
import type {
  SyncDevice,
  SyncDiscoveredPrimary,
  SyncStatusResponse,
} from '@/types/daemon'

export interface SyncViewProps {
  readonly status: SyncStatusResponse | null
  readonly onAddDevice: (deviceId: string, name: string) => Promise<void>
  readonly onRemoveDevice: (deviceId: string) => Promise<void>
  readonly onStartPairing: () => Promise<void>
  readonly onCancelPairing: () => Promise<void>
  readonly onJoinPrimary: (code: string, pairingAddr: string) => Promise<void>
  readonly onPause: () => Promise<void>
  readonly onResume: () => Promise<void>
  readonly pairingCode: string | null
  readonly pairingProgress: string | null
  readonly discoveredPrimaries: readonly SyncDiscoveredPrimary[]
}

function truncateDeviceId(id: string): string {
  if (id.length > 20) {
    return `${id.slice(0, 7)}...${id.slice(-7)}`
  }
  return id
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="p-4 bg-surface-alt rounded-card">
      <h3 className="text-sm font-medium text-on-surface mb-3">{title}</h3>
      {children}
    </div>
  )
}

function DeviceRow({
  device,
  onRemove,
}: {
  readonly device: SyncDevice
  readonly onRemove: () => void
}) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-outline last:border-0">
      <div className="flex items-center gap-2">
        <span
          className={`w-2 h-2 rounded-full ${device.connected ? 'bg-status-ok' : 'bg-outline'}`}
        />
        <span className="font-medium text-on-surface">{device.name}</span>
        <span className="text-xs text-on-surface-dim">{truncateDeviceId(device.id)}</span>
      </div>
      <button
        type="button"
        onClick={onRemove}
        className="text-xs text-status-error hover:text-status-error"
      >
        Remove
      </button>
    </div>
  )
}

function DisabledState() {
  return (
    <div className="p-6">
      <Section title="Sync is not enabled">
        <p className="text-sm text-on-surface-muted mb-4">Enable sync in your config.toml:</p>
        <pre className="bg-surface-raised text-on-surface-secondary p-3 rounded-sm text-sm font-mono">
          {`[sync]
enabled = true
mode = "primary"  # or "secondary"`}
        </pre>
      </Section>
    </div>
  )
}

function PairingSection({
  status,
  pairingCode,
  pairingProgress,
  discoveredPrimaries,
  onStartPairing,
  onCancelPairing,
  onJoinPrimary,
}: {
  readonly status: SyncStatusResponse
  readonly pairingCode: string | null
  readonly pairingProgress: string | null
  readonly discoveredPrimaries: readonly SyncDiscoveredPrimary[]
  readonly onStartPairing: () => Promise<void>
  readonly onCancelPairing: () => Promise<void>
  readonly onJoinPrimary: (code: string, pairingAddr: string) => Promise<void>
}) {
  const [joinCode, setJoinCode] = useState('')
  const [joinAddr, setJoinAddr] = useState('')
  const [isJoining, setIsJoining] = useState(false)
  const isPairing = status.pairing || pairingCode !== null

  const handleJoin = useCallback(
    async (code: string, addr: string) => {
      setIsJoining(true)
      try {
        await onJoinPrimary(code, addr)
        setJoinCode('')
        setJoinAddr('')
      } finally {
        setIsJoining(false)
      }
    },
    [onJoinPrimary],
  )

  if (isPairing) {
    return (
      <Section title="Pairing in progress">
        {pairingCode && (
          <div className="mb-4">
            <p className="text-sm text-on-surface-muted mb-2">
              Enter this code on the other device:
            </p>
            <code className="block bg-surface-raised text-on-surface px-4 py-3 rounded-sm text-2xl font-mono text-center tracking-widest">
              {pairingCode}
            </code>
          </div>
        )}
        {pairingProgress && (
          <p className="text-sm text-on-surface-muted mb-3">{pairingProgress}</p>
        )}
        <Button variant="secondary" onClick={onCancelPairing}>
          Cancel pairing
        </Button>
      </Section>
    )
  }

  if (status.mode === 'secondary') {
    return (
      <Section title="Join a primary device">
        <div className="space-y-3">
          <p className="text-sm text-on-surface-muted">
            Enter the pairing code shown on the primary device.
          </p>
          {discoveredPrimaries.length > 0 && (
            <div className="space-y-2">
              <p className="text-xs text-on-surface-dim">Found on network:</p>
              {discoveredPrimaries.map((primary) => (
                <button
                  key={primary.pairingAddr}
                  type="button"
                  onClick={() => setJoinAddr(primary.pairingAddr)}
                  className={`w-full text-left px-3 py-2 rounded-sm text-sm border ${
                    joinAddr === primary.pairingAddr
                      ? 'border-accent bg-surface-raised'
                      : 'border-outline bg-surface'
                  }`}
                >
                  {primary.hostname}
                  <span className="text-xs text-on-surface-dim ml-2">
                    ({primary.pairingAddr})
                  </span>
                </button>
              ))}
            </div>
          )}
          <Input
            value={joinCode}
            onChange={setJoinCode}
            placeholder="Pairing code"
          />
          {discoveredPrimaries.length === 0 && (
            <Input
              value={joinAddr}
              onChange={setJoinAddr}
              placeholder="Primary address (e.g. 192.168.1.100:43210)"
            />
          )}
          <Button
            onClick={() => handleJoin(joinCode.trim(), joinAddr.trim())}
            disabled={!joinCode.trim() || !joinAddr.trim() || isJoining}
          >
            {isJoining ? 'Joining...' : 'Join primary'}
          </Button>
        </div>
      </Section>
    )
  }

  return (
    <Section title="Pair a device">
      <p className="text-sm text-on-surface-muted mb-3">
        Start pairing to connect another device on your local network.
      </p>
      <Button onClick={onStartPairing}>Start pairing</Button>
    </Section>
  )
}

export function SyncView({
  status,
  onAddDevice,
  onRemoveDevice,
  onStartPairing,
  onCancelPairing,
  onJoinPrimary,
  onPause,
  onResume,
  pairingCode,
  pairingProgress,
  discoveredPrimaries,
}: SyncViewProps) {
  const [newDeviceId, setNewDeviceId] = useState('')
  const [newDeviceName, setNewDeviceName] = useState('')
  const [isAdding, setIsAdding] = useState(false)

  const handleAddDevice = async () => {
    if (!newDeviceId.trim()) return
    setIsAdding(true)
    try {
      await onAddDevice(newDeviceId.trim(), newDeviceName.trim())
      setNewDeviceId('')
      setNewDeviceName('')
    } finally {
      setIsAdding(false)
    }
  }

  const handleOpenGui = () => {
    if (status?.guiURL) {
      window.open(status.guiURL, '_blank')
    }
  }

  if (!status?.enabled) {
    return <DisabledState />
  }

  const connectedCount = status.devices?.filter((d) => d.connected).length ?? 0
  const totalDevices = status.devices?.length ?? 0
  const isSyncing = status.state === 'syncing'

  return (
    <div className="p-6 space-y-6">
      <Section title="Status">
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-on-surface-muted">Mode</span>
            <span className="font-medium text-on-surface capitalize">{status.mode}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-on-surface-muted">Running</span>
            <span
              className={`font-medium ${status.running ? 'text-status-ok' : 'text-status-error'}`}
            >
              {status.running ? 'Yes' : 'No'}
            </span>
          </div>
          {totalDevices > 0 && (
            <div className="flex justify-between">
              <span className="text-on-surface-muted">Connected devices</span>
              <span className="font-medium text-on-surface">
                {connectedCount}/{totalDevices}
              </span>
            </div>
          )}
          {status.running && (
            <div className="flex justify-end gap-2 pt-2">
              <Button variant="secondary" onClick={isSyncing ? onPause : onResume}>
                {isSyncing ? 'Pause sync' : 'Resume sync'}
              </Button>
            </div>
          )}
        </div>
      </Section>

      {status.deviceId && (
        <Section title="This device ID">
          <div className="flex items-center gap-2">
            <code className="flex-1 bg-surface-raised text-on-surface-secondary px-3 py-2 rounded-sm text-xs break-all font-mono">
              {status.deviceId}
            </code>
            <button
              type="button"
              onClick={() => {
                if (status.deviceId) {
                  navigator.clipboard.writeText(status.deviceId)
                }
              }}
              className="px-3 py-2 text-xs bg-surface-raised text-on-surface-secondary rounded-sm hover:bg-outline"
            >
              Copy
            </button>
          </div>
        </Section>
      )}

      {status.running && (
        <PairingSection
          status={status}
          pairingCode={pairingCode}
          pairingProgress={pairingProgress}
          discoveredPrimaries={discoveredPrimaries}
          onStartPairing={onStartPairing}
          onCancelPairing={onCancelPairing}
          onJoinPrimary={onJoinPrimary}
        />
      )}

      <Section title="Paired devices">
        <div className="flex items-center justify-between mb-3">
          <span className="text-sm text-on-surface-muted">
            {totalDevices === 0
              ? 'No devices paired'
              : `${totalDevices} device${totalDevices === 1 ? '' : 's'}`}
          </span>
          {status.guiURL && (
            <button
              type="button"
              onClick={handleOpenGui}
              className="text-xs text-accent hover:text-accent-hover"
            >
              Open Syncthing UI
            </button>
          )}
        </div>

        {status.devices && status.devices.length > 0 && (
          <div className="border border-outline rounded-card px-3 bg-surface">
            {status.devices.map((device) => (
              <DeviceRow
                key={device.id}
                device={device}
                onRemove={() => onRemoveDevice(device.id)}
              />
            ))}
          </div>
        )}
      </Section>

      <Section title="Add a device manually">
        <p className="text-sm text-on-surface-muted mb-3">
          For cross-network pairing, add a device by its Syncthing device ID.
        </p>
        <div className="space-y-3">
          <Input value={newDeviceId} onChange={setNewDeviceId} placeholder="Device ID" />
          <Input
            value={newDeviceName}
            onChange={setNewDeviceName}
            placeholder="Friendly name (optional)"
          />
          <Button onClick={handleAddDevice} disabled={!newDeviceId.trim() || isAdding}>
            {isAdding ? 'Adding...' : 'Add device'}
          </Button>
        </div>
      </Section>
    </div>
  )
}
