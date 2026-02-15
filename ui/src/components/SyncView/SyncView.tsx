import { useCallback, useState } from 'react'
import { Button } from '@/lib/Button'
import { formatBytes } from '@/lib/changeUtils'
import { getSyncLocalChanges, openPath, revertSyncFolder } from '@/lib/daemon'
import { Input } from '@/lib/Input'
import { CopyIcon, FolderIcon, TrashIcon } from '@/lib/icons'
import { RadioCard } from '@/lib/RadioCard'
import { Spinner } from '@/lib/Spinner'
import type {
  SyncDevice,
  SyncDiscoveredDevice,
  SyncFolder,
  SyncLocalChange,
  SyncMode,
  SyncStatusResponse,
} from '@/types/daemon'

export interface SyncViewProps {
  readonly status: SyncStatusResponse | null
  readonly discoveredDevices: SyncDiscoveredDevice[]
  readonly connectionProgress: string | null
  readonly connectionError: string | null
  readonly isDiscovering: boolean
  readonly isConnecting: boolean
  readonly isPairing: boolean
  readonly pairingDeviceId: string | null
  readonly onRemoveDevice: (deviceId: string) => Promise<void>
  readonly onConnectToDevice: (deviceId: string) => Promise<{ ok: boolean; error?: string }>
  readonly onEnableSync: (mode: SyncMode) => Promise<void>
  readonly onResetSync: () => Promise<void>
  readonly onStartPairing: () => Promise<void>
  readonly onStopPairing: () => Promise<void>
  readonly onRefresh: () => void
  readonly isEnabling: boolean
  readonly enableError: string | null
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

function FolderRow({
  folder,
  onRefresh,
  hasPairedDevices,
}: {
  readonly folder: SyncFolder
  readonly onRefresh: () => void
  readonly hasPairedDevices: boolean
}) {
  const [showChanges, setShowChanges] = useState(false)
  const [changes, setChanges] = useState<SyncLocalChange[] | null>(null)
  const [loadingChanges, setLoadingChanges] = useState(false)
  const [reverting, setReverting] = useState(false)
  const [changesError, setChangesError] = useState<string | null>(null)

  const isSyncing = hasPairedDevices && (folder.state === 'syncing' || folder.needSize > 0)
  const hasLocalChanges = hasPairedDevices && folder.receiveOnlyChanges > 0
  const isReceiveOnly = folder.type === 'receiveonly'
  const sizeDiffers = hasPairedDevices && isReceiveOnly && folder.localSize !== folder.globalSize
  const percent =
    folder.globalSize > 0 ? Math.round((folder.localSize / folder.globalSize) * 100) : 100

  const handleShowChanges = useCallback(async () => {
    if (showChanges) {
      setShowChanges(false)
      return
    }
    setLoadingChanges(true)
    setChangesError(null)
    const result = await getSyncLocalChanges({ folderId: folder.id })
    if (result.ok) {
      setChanges(result.data.changes)
    } else {
      setChangesError(result.error?.message ?? 'Failed to load changes')
    }
    setLoadingChanges(false)
    setShowChanges(true)
  }, [folder.id, showChanges])

  const handleRevert = useCallback(async () => {
    setReverting(true)
    await revertSyncFolder({ folderId: folder.id })
    setReverting(false)
    setShowChanges(false)
    setChanges(null)
    onRefresh()
  }, [folder.id, onRefresh])

  const getStatusIndicator = () => {
    if (hasLocalChanges) return 'bg-status-warning'
    if (isSyncing) return 'bg-accent animate-pulse'
    return 'bg-status-ok'
  }

  return (
    <div className="py-2 border-b border-outline last:border-0">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <span className={`w-2 h-2 rounded-full flex-shrink-0 ${getStatusIndicator()}`} />
          <span className="font-medium text-on-surface truncate">{folder.label}</span>
        </div>
        <div className="flex items-center gap-2 text-xs text-on-surface-muted flex-shrink-0">
          {isSyncing ? (
            <span>
              {percent}% ({formatBytes(folder.needSize)} left)
            </span>
          ) : sizeDiffers ? (
            <span>
              {formatBytes(folder.localSize)} / {formatBytes(folder.globalSize)}
            </span>
          ) : (
            <span>{formatBytes(folder.globalSize)}</span>
          )}
          <button
            type="button"
            onClick={() => openPath(folder.path)}
            className="p-1 text-on-surface-muted hover:text-on-surface-secondary rounded"
            title="Open folder"
          >
            <FolderIcon className="w-4 h-4" />
          </button>
        </div>
      </div>
      {hasLocalChanges && (
        <div className="mt-2 ml-4 p-2 bg-status-warning/10 border border-status-warning/30 rounded text-xs">
          <div className="flex items-center justify-between">
            <span className="text-status-warning">
              {folder.receiveOnlyChanges} local change{folder.receiveOnlyChanges === 1 ? '' : 's'}{' '}
              differ from remote
            </span>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={handleShowChanges}
                className="text-accent hover:underline"
                disabled={loadingChanges}
              >
                {loadingChanges ? 'Loading...' : showChanges ? 'Hide' : 'Show'}
              </button>
              <button
                type="button"
                onClick={handleRevert}
                className="text-accent hover:underline"
                disabled={reverting}
              >
                {reverting ? 'Reverting...' : 'Revert'}
              </button>
            </div>
          </div>
          {showChanges && changes && changes.length > 0 && (
            <div className="mt-1 max-h-32 overflow-y-auto">
              {changes.map((c) => (
                <div key={c.path} className="py-0.5 text-on-surface-muted truncate">
                  <span className="text-status-error">missing</span> {c.path}
                  {c.size > 0 && <span> ({formatBytes(c.size)})</span>}
                </div>
              ))}
            </div>
          )}
          {showChanges && changes && changes.length === 0 && (
            <div className="mt-1 text-on-surface-muted">No details available</div>
          )}
          {showChanges && changesError && (
            <div className="mt-1 text-status-error">{changesError}</div>
          )}
        </div>
      )}
    </div>
  )
}

function DisabledState({
  status,
  onEnable,
  onReset,
  isEnabling,
  enableError,
}: {
  readonly status: SyncStatusResponse | null
  readonly onEnable: (mode: SyncMode) => Promise<void>
  readonly onReset: () => Promise<void>
  readonly isEnabling: boolean
  readonly enableError: string | null
}) {
  const [selectedMode, setSelectedMode] = useState<SyncMode>('primary')
  const [isResetting, setIsResetting] = useState(false)
  const [showResetConfirm, setShowResetConfirm] = useState(false)

  const handleEnable = useCallback(async () => {
    await onEnable(selectedMode)
  }, [onEnable, selectedMode])

  const handleReset = useCallback(async () => {
    setIsResetting(true)
    try {
      await onReset()
    } finally {
      setIsResetting(false)
      setShowResetConfirm(false)
    }
  }, [onReset])

  const hasOrphanedState =
    status && !status.enabled && (status.installed || status.serviceInstalled)

  return (
    <div className="p-6 space-y-6">
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
            <div className="space-y-3">
              <RadioCard
                title="Primary"
                description="Your main device with the ROM collection. Sends ROMs and BIOS to secondaries, syncs saves both ways."
                selected={selectedMode === 'primary'}
                onSelect={() => setSelectedMode('primary')}
                className="w-full p-4"
              />
              <RadioCard
                title="Secondary"
                description="Receives ROMs from primary (read-only). Play anywhere and saves sync back automatically."
                selected={selectedMode === 'secondary'}
                onSelect={() => setSelectedMode('secondary')}
                className="w-full p-4"
              />
            </div>
            <Button onClick={handleEnable}>Enable sync</Button>
            {enableError && (
              <div className="p-4 bg-status-error/10 border border-status-error/30 rounded-card">
                <p className="text-sm text-status-error">{enableError}</p>
              </div>
            )}
          </div>
        )}
      </Section>

      {hasOrphanedState && (
        <Section title="Orphaned syncthing state">
          <p className="text-sm text-on-surface-muted mb-3">
            Syncthing files from a previous installation were detected. This can happen after an
            incomplete uninstall or if sync was disabled manually.
          </p>
          {showResetConfirm ? (
            <div className="space-y-3">
              <div className="p-3 bg-status-warning/10 border border-status-warning/30 rounded text-sm">
                <p className="text-on-surface mb-2">This will remove:</p>
                <ul className="list-disc list-inside text-on-surface-muted space-y-1">
                  <li>Syncthing service (if running)</li>
                  <li>Syncthing configuration and database</li>
                  <li>Device pairing information</li>
                </ul>
                <p className="mt-2 text-on-surface-muted">
                  Your ROMs, saves, and other emulation data will not be affected.
                </p>
              </div>
              <div className="flex gap-2">
                <Button variant="danger" onClick={handleReset} disabled={isResetting}>
                  {isResetting ? 'Resetting...' : 'Confirm reset'}
                </Button>
                <Button variant="secondary" onClick={() => setShowResetConfirm(false)}>
                  Cancel
                </Button>
              </div>
            </div>
          ) : (
            <Button variant="secondary" onClick={() => setShowResetConfirm(true)}>
              Clean up syncthing state
            </Button>
          )}
        </Section>
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

function PrimaryPairingSection({
  status,
  isPairing,
  pairingDeviceId,
  onStartPairing,
  onStopPairing,
}: {
  readonly status: SyncStatusResponse
  readonly isPairing: boolean
  readonly pairingDeviceId: string | null
  readonly onStartPairing: () => Promise<void>
  readonly onStopPairing: () => Promise<void>
}) {
  const hasDevices = (status.devices?.length ?? 0) > 0

  if (isPairing && pairingDeviceId) {
    return (
      <Section title="Pairing mode">
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
      </Section>
    )
  }

  if (hasDevices) {
    return null
  }

  return (
    <Section title="Pair a device">
      <p className="text-sm text-on-surface-muted mb-3">
        Start pairing to allow secondary devices to connect. Your device ID will be shown and
        secondary devices on the network will be able to discover and connect to this device.
      </p>
      <Button onClick={onStartPairing}>Start pairing</Button>
    </Section>
  )
}

function SecondaryDiscoverySection({
  status,
  discoveredDevices,
  connectionProgress,
  connectionError,
  isDiscovering,
  isConnecting,
  onConnectToDevice,
}: {
  readonly status: SyncStatusResponse
  readonly discoveredDevices: SyncDiscoveredDevice[]
  readonly connectionProgress: string | null
  readonly connectionError: string | null
  readonly isDiscovering: boolean
  readonly isConnecting: boolean
  readonly onConnectToDevice: (deviceId: string) => Promise<{ ok: boolean; error?: string }>
}) {
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
      <Section title="Connecting to primary">
        <div className="flex items-center gap-3">
          <Spinner />
          <span className="text-sm text-on-surface-muted">
            {connectionProgress || 'Connecting...'}
          </span>
        </div>
      </Section>
    )
  }

  return (
    <Section title="Connect to primary">
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
    </Section>
  )
}

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

export function SyncView({
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
  onEnableSync,
  onResetSync,
  onStartPairing,
  onStopPairing,
  onRefresh,
  isEnabling,
  enableError,
}: SyncViewProps) {
  const [isResetting, setIsResetting] = useState(false)
  const [showResetConfirm, setShowResetConfirm] = useState(false)

  const handleReset = useCallback(async () => {
    setIsResetting(true)
    try {
      await onResetSync()
    } finally {
      setIsResetting(false)
      setShowResetConfirm(false)
    }
  }, [onResetSync])

  if (!status?.enabled) {
    return (
      <DisabledState
        status={status}
        onEnable={onEnableSync}
        onReset={onResetSync}
        isEnabling={isEnabling}
        enableError={enableError}
      />
    )
  }

  if (!status.running) {
    return (
      <div className="p-6">
        <Section title="Syncthing not running">
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <Spinner />
              <span className="text-sm text-on-surface-muted">
                Waiting for syncthing to start...
              </span>
            </div>
            {status.serviceError && (
              <details className="text-sm">
                <summary className="text-on-surface-muted cursor-pointer hover:text-on-surface">
                  Service logs
                </summary>
                <pre className="mt-2 text-xs text-on-surface-muted bg-surface-raised p-3 rounded-sm overflow-x-auto whitespace-pre-wrap">
                  {status.serviceError}
                </pre>
              </details>
            )}
          </div>
        </Section>
      </div>
    )
  }

  const connectedCount = status.devices?.filter((d) => d.connected).length ?? 0
  const totalDevices = status.devices?.length ?? 0

  const sortedFolders = status.folders
    ? [...status.folders].sort((a, b) => a.label.localeCompare(b.label))
    : []

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center gap-2">
        <StatusBadge label={status.mode ?? 'unknown'} ok={true} />
        <StatusBadge label="running" ok={true} />
      </div>

      {status.mode === 'primary' && (
        <PrimaryPairingSection
          status={status}
          isPairing={isPairing}
          pairingDeviceId={pairingDeviceId}
          onStartPairing={onStartPairing}
          onStopPairing={onStopPairing}
        />
      )}

      {status.mode === 'secondary' && (
        <SecondaryDiscoverySection
          status={status}
          discoveredDevices={discoveredDevices}
          connectionProgress={connectionProgress}
          connectionError={connectionError}
          isDiscovering={isDiscovering}
          isConnecting={isConnecting}
          onConnectToDevice={onConnectToDevice}
        />
      )}

      <Section title="Paired devices">
        {totalDevices === 0 ? (
          <p className="text-sm text-on-surface-muted">
            No devices paired yet.{' '}
            {status.mode === 'primary'
              ? 'Secondary devices will appear here when they connect.'
              : 'Connect to a primary device above to start syncing.'}
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

      {sortedFolders.length > 0 && (
        <Section title="Synced folders">
          <div className="border border-outline rounded-card px-3 bg-surface">
            {sortedFolders.map((folder) => (
              <FolderRow
                key={folder.id}
                folder={folder}
                onRefresh={onRefresh}
                hasPairedDevices={totalDevices > 0}
              />
            ))}
          </div>
        </Section>
      )}

      <Section title="Advanced" collapsible defaultCollapsed>
        <div className="space-y-4">
          {status.guiURL && (
            <a
              href={status.guiURL}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-accent hover:underline block"
            >
              Open Syncthing web interface
            </a>
          )}
          <div>
            <h4 className="text-sm font-medium text-on-surface mb-2">Reset sync</h4>
            {showResetConfirm ? (
              <div className="space-y-3">
                <div className="p-3 bg-status-warning/10 border border-status-warning/30 rounded text-sm">
                  <p className="text-on-surface mb-2">This will:</p>
                  <ul className="list-disc list-inside text-on-surface-muted space-y-1">
                    <li>Stop and remove the syncthing service</li>
                    <li>Delete syncthing configuration and database</li>
                    <li>Remove all device pairings</li>
                    <li>Disable sync in your kyaraben config</li>
                  </ul>
                  <p className="mt-2 text-on-surface-muted">
                    Your ROMs, saves, and other emulation data will not be affected. You can
                    re-enable sync afterwards.
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button variant="danger" onClick={handleReset} disabled={isResetting}>
                    {isResetting ? 'Resetting...' : 'Confirm reset'}
                  </Button>
                  <Button variant="secondary" onClick={() => setShowResetConfirm(false)}>
                    Cancel
                  </Button>
                </div>
              </div>
            ) : (
              <div>
                <p className="text-sm text-on-surface-muted mb-2">
                  Remove all syncthing state and start fresh.
                </p>
                <Button variant="secondary" onClick={() => setShowResetConfirm(true)}>
                  Reset sync
                </Button>
              </div>
            )}
          </div>
        </div>
      </Section>
    </div>
  )
}
