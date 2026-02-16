import { useCallback, useState } from 'react'
import { Button } from '@/lib/Button'
import { Input } from '@/lib/Input'
import { Spinner } from '@/lib/Spinner'
import type { SyncDevice, SyncMode, SyncStatusResponse } from '@/types/daemon'
import { SyncStateSynced } from '@/types/daemon'
import { SyncStatusBanner } from './SyncStatusBanner'

export interface SyncViewProps {
  readonly status: SyncStatusResponse | null
  readonly onAddDevice: (deviceId: string, name: string) => Promise<void>
  readonly onRemoveDevice: (deviceId: string) => Promise<void>
  readonly onStartPairing: () => Promise<void>
  readonly onCancelPairing: () => Promise<void>
  readonly onJoinPrimary: (code: string) => Promise<void>
  readonly onPause: () => Promise<void>
  readonly onResume: () => Promise<void>
  readonly onEnableSync: (mode: SyncMode) => Promise<void>
  readonly pairingCode: string | null
  readonly pairingProgress: string | null
  readonly isEnabling: boolean
}

function Section({
  title,
  children,
  collapsible = false,
  defaultCollapsed = false,
}: {
  title: string
  children: React.ReactNode
  collapsible?: boolean
  defaultCollapsed?: boolean
}) {
  const [isCollapsed, setIsCollapsed] = useState(defaultCollapsed)

  return (
    <div className="p-4 bg-surface-alt rounded-card">
      {collapsible ? (
        <button
          type="button"
          onClick={() => setIsCollapsed(!isCollapsed)}
          className="flex items-center gap-2 w-full text-left"
        >
          <span
            className={`transition-transform text-xs text-on-surface-muted ${isCollapsed ? '' : 'rotate-90'}`}
          >
            ▶
          </span>
          <h3 className="text-sm font-medium text-on-surface">{title}</h3>
        </button>
      ) : (
        <h3 className="text-sm font-medium text-on-surface mb-3">{title}</h3>
      )}
      {(!collapsible || !isCollapsed) && (
        <div className={collapsible ? 'mt-3' : ''}>{children}</div>
      )}
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
  const getStatusDisplay = () => {
    if (device.paused) {
      return { dotClass: 'bg-status-warn', label: 'paused' }
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
        className="text-xs text-status-error hover:text-status-error"
      >
        Remove
      </button>
    </div>
  )
}

function DisabledState({
  onEnable,
  isEnabling,
}: {
  readonly onEnable: (mode: SyncMode) => Promise<void>
  readonly isEnabling: boolean
}) {
  const [selectedMode, setSelectedMode] = useState<SyncMode>('primary')

  const handleEnable = useCallback(async () => {
    await onEnable(selectedMode)
  }, [onEnable, selectedMode])

  return (
    <div className="p-6">
      <Section title="Enable sync">
        <p className="text-sm text-on-surface-muted mb-4">
          Sync your saves, states, and screenshots across devices using Syncthing.
        </p>

        {isEnabling ? (
          <div className="flex items-center gap-3">
            <Spinner />
            <span className="text-sm text-on-surface-muted">Installing syncthing...</span>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="sync-mode"
                  value="primary"
                  checked={selectedMode === 'primary'}
                  onChange={() => setSelectedMode('primary')}
                  className="accent-accent"
                />
                <span className="text-sm text-on-surface">Primary</span>
                <span className="text-xs text-on-surface-muted">
                  - this device hosts the main copy
                </span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="sync-mode"
                  value="secondary"
                  checked={selectedMode === 'secondary'}
                  onChange={() => setSelectedMode('secondary')}
                  className="accent-accent"
                />
                <span className="text-sm text-on-surface">Secondary</span>
                <span className="text-xs text-on-surface-muted">- syncs from a primary device</span>
              </label>
            </div>
            <Button onClick={handleEnable}>Enable sync</Button>
          </div>
        )}
      </Section>
    </div>
  )
}

function PairingSection({
  status,
  pairingCode,
  pairingProgress,
  onStartPairing,
  onCancelPairing,
  onJoinPrimary,
}: {
  readonly status: SyncStatusResponse
  readonly pairingCode: string | null
  readonly pairingProgress: string | null
  readonly onStartPairing: () => Promise<void>
  readonly onCancelPairing: () => Promise<void>
  readonly onJoinPrimary: (code: string) => Promise<void>
}) {
  const [joinCode, setJoinCode] = useState('')
  const [isJoining, setIsJoining] = useState(false)
  const isPairing = status.pairing || pairingCode !== null
  const isRunning = status.running ?? false

  const handleJoin = useCallback(
    async (code: string) => {
      setIsJoining(true)
      try {
        await onJoinPrimary(code)
        setJoinCode('')
      } finally {
        setIsJoining(false)
      }
    },
    [onJoinPrimary],
  )

  if (!isRunning) {
    return (
      <Section title="Pair a device">
        <div className="flex items-center gap-3">
          <Spinner />
          <span className="text-sm text-on-surface-muted">Waiting for syncthing to start...</span>
        </div>
      </Section>
    )
  }

  if (isPairing) {
    return (
      <Section title="Pairing in progress">
        {pairingCode && (
          <div className="mb-4">
            <p className="text-sm text-on-surface-muted mb-2">
              Enter this code on the other device:
            </p>
            <code className="block bg-surface-raised text-on-surface px-4 py-3 rounded-sm text-2xl font-mono text-center tracking-widest tabular-nums">
              {pairingCode}
            </code>
          </div>
        )}
        {pairingProgress && <p className="text-sm text-on-surface-muted mb-3">{pairingProgress}</p>}
        <Button variant="secondary" onClick={onCancelPairing}>
          Cancel pairing
        </Button>
      </Section>
    )
  }

  if (status.mode === 'secondary') {
    if (isJoining) {
      return (
        <Section title="Joining primary">
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <Spinner />
              <span className="text-sm text-on-surface-muted">
                {pairingProgress || 'Searching for primary on local network...'}
              </span>
            </div>
            <Button variant="secondary" onClick={onCancelPairing}>
              Cancel
            </Button>
          </div>
        </Section>
      )
    }

    return (
      <Section title="Join a primary device">
        <div className="space-y-3">
          <p className="text-sm text-on-surface-muted">
            Enter the pairing code shown on the primary device. The primary will be discovered
            automatically on your local network.
          </p>
          <Input value={joinCode} onChange={setJoinCode} placeholder="Pairing code (e.g. 6MDLRF)" />
          <Button onClick={() => handleJoin(joinCode.trim())} disabled={!joinCode.trim()}>
            Join primary
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

function StatusBadge({ label, ok }: { label: string; ok: boolean }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs ${
        ok ? 'bg-status-ok/10 text-status-ok' : 'bg-outline/50 text-on-surface-muted'
      }`}
    >
      <span className={`w-1.5 h-1.5 rounded-full ${ok ? 'bg-status-ok' : 'bg-outline'}`} />
      {label}
    </span>
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
  onEnableSync,
  pairingCode,
  pairingProgress,
  isEnabling,
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

  if (!status?.enabled) {
    return <DisabledState onEnable={onEnableSync} isEnabling={isEnabling} />
  }

  const connectedCount = status.devices?.filter((d) => d.connected).length ?? 0
  const totalDevices = status.devices?.length ?? 0
  const isPaused = status.paused ?? false
  const state = status.state ?? SyncStateSynced
  const progress = status.progress ?? null

  return (
    <div className="p-6 space-y-6">
      {status.running && (
        <SyncStatusBanner state={state} progress={progress} paused={isPaused} onResume={onResume} />
      )}

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <StatusBadge label={status.mode ?? 'unknown'} ok={true} />
          <StatusBadge
            label={status.running ? 'running' : 'stopped'}
            ok={status.running ?? false}
          />
        </div>
        {status.running && !isPaused && (
          <Button variant="secondary" onClick={onPause}>
            Pause sync
          </Button>
        )}
      </div>

      <Section title="Paired devices">
        {totalDevices === 0 ? (
          <p className="text-sm text-on-surface-muted">
            No devices paired yet. Pair a device to start syncing.
          </p>
        ) : (
          <>
            <p className="text-sm text-on-surface-muted mb-3">
              {connectedCount} of {totalDevices} device{totalDevices === 1 ? '' : 's'} connected
            </p>
            <div className="border border-outline rounded-card px-3 bg-surface">
              {status.devices?.map((device) => (
                <DeviceRow
                  key={device.id}
                  device={device}
                  onRemove={() => onRemoveDevice(device.id)}
                />
              ))}
            </div>
          </>
        )}
      </Section>

      <PairingSection
        status={status}
        pairingCode={pairingCode}
        pairingProgress={pairingProgress}
        onStartPairing={onStartPairing}
        onCancelPairing={onCancelPairing}
        onJoinPrimary={onJoinPrimary}
      />

      <Section title="Advanced" collapsible defaultCollapsed>
        <div className="space-y-4">
          {status.deviceId && (
            <div>
              <p className="text-sm text-on-surface-muted mb-2">
                This device ID (for manual pairing):
              </p>
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
            </div>
          )}

          <div>
            <p className="text-sm text-on-surface-muted mb-2">
              Add device by ID (for cross-network):
            </p>
            <div className="space-y-2">
              <Input value={newDeviceId} onChange={setNewDeviceId} placeholder="Device ID" />
              <Input
                value={newDeviceName}
                onChange={setNewDeviceName}
                placeholder="Friendly name (optional)"
              />
              <Button
                variant="secondary"
                onClick={handleAddDevice}
                disabled={!newDeviceId.trim() || isAdding}
              >
                {isAdding ? 'Adding...' : 'Add device'}
              </Button>
            </div>
          </div>
        </div>
      </Section>
    </div>
  )
}
