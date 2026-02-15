import { useCallback, useState } from 'react'
import { Button } from '@/lib/Button'
import { formatBytes } from '@/lib/changeUtils'
import { getSyncLocalChanges, openPath, revertSyncFolder } from '@/lib/daemon'
import { Input } from '@/lib/Input'
import { FolderIcon, TrashIcon } from '@/lib/icons'
import { Spinner } from '@/lib/Spinner'
import type {
  SyncDevice,
  SyncFolder,
  SyncLocalChange,
  SyncMode,
  SyncStatusResponse,
} from '@/types/daemon'

export interface SyncViewProps {
  readonly status: SyncStatusResponse | null
  readonly onRemoveDevice: (deviceId: string) => Promise<void>
  readonly onStartPairing: () => Promise<void>
  readonly onCancelPairing: () => Promise<void>
  readonly onJoinPrimary: (code: string) => Promise<{ ok: boolean; error?: string }>
  readonly onEnableSync: (mode: SyncMode) => Promise<void>
  readonly onRefresh: () => void
  readonly pairingCode: string | null
  readonly pairingProgress: string | null
  readonly pairingError: string | null
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
}: {
  readonly folder: SyncFolder
  readonly onRefresh: () => void
}) {
  const [showChanges, setShowChanges] = useState(false)
  const [changes, setChanges] = useState<SyncLocalChange[] | null>(null)
  const [loadingChanges, setLoadingChanges] = useState(false)
  const [reverting, setReverting] = useState(false)
  const [changesError, setChangesError] = useState<string | null>(null)

  const isSyncing = folder.state === 'syncing' || folder.needSize > 0
  const hasLocalChanges = folder.receiveOnlyChanges > 0
  const isReceiveOnly = folder.type === 'receiveonly'
  const sizeDiffers = isReceiveOnly && folder.localSize !== folder.globalSize
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
    if (hasLocalChanges) return 'bg-status-warn'
    if (isSyncing) return 'bg-status-warn'
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
        <div className="mt-2 ml-4 p-2 bg-status-warn/10 border border-status-warn/30 rounded text-xs">
          <div className="flex items-center justify-between">
            <span className="text-status-warn">
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
            <div className="mt-2 max-h-32 overflow-y-auto">
              {changes.map((c, i) => (
                <div key={i} className="py-0.5 text-on-surface-muted truncate">
                  <span className="text-on-surface-secondary">{c.action}</span> {c.path}
                  {c.size > 0 && <span className="text-on-surface-muted"> ({formatBytes(c.size)})</span>}
                </div>
              ))}
            </div>
          )}
          {showChanges && changes && changes.length === 0 && (
            <div className="mt-2 text-on-surface-muted">No details available</div>
          )}
          {showChanges && changesError && (
            <div className="mt-2 text-status-error">{changesError}</div>
          )}
        </div>
      )}
    </div>
  )
}

function ModeCard({
  title,
  description,
  selected,
  onSelect,
}: {
  readonly title: string
  readonly description: string
  readonly selected: boolean
  readonly onSelect: () => void
}) {
  return (
    <button
      type="button"
      onClick={onSelect}
      className={`w-full text-left p-4 rounded-card border-2 transition-colors ${
        selected
          ? 'border-accent bg-accent/5'
          : 'border-outline bg-surface hover:border-outline-hover'
      }`}
    >
      <div className="grid grid-cols-[1rem_1fr] gap-x-3 gap-y-1 items-center">
        <div
          className={`w-4 h-4 rounded-full border-2 ${
            selected ? 'border-accent bg-accent' : 'border-outline'
          }`}
          style={selected ? { boxShadow: 'inset 0 0 0 3px var(--color-surface)' } : undefined}
        />
        <span className={`font-medium ${selected ? 'text-accent' : 'text-on-surface'}`}>
          {title}
        </span>
        <div />
        <p className="text-sm text-on-surface-muted">{description}</p>
      </div>
    </button>
  )
}

function DisabledState({
  onEnable,
  isEnabling,
  enableError,
}: {
  readonly onEnable: (mode: SyncMode) => Promise<void>
  readonly isEnabling: boolean
  readonly enableError: string | null
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
            <div className="space-y-3">
              <ModeCard
                title="Primary"
                description="Your main device with the ROM collection. Sends ROMs and BIOS to secondaries, syncs saves both ways."
                selected={selectedMode === 'primary'}
                onSelect={() => setSelectedMode('primary')}
              />
              <ModeCard
                title="Secondary"
                description="Receives ROMs from primary (read-only). Play anywhere and saves sync back automatically."
                selected={selectedMode === 'secondary'}
                onSelect={() => setSelectedMode('secondary')}
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
    </div>
  )
}

function PairingSection({
  status,
  pairingCode,
  pairingProgress,
  pairingError,
  onStartPairing,
  onCancelPairing,
  onJoinPrimary,
}: {
  readonly status: SyncStatusResponse
  readonly pairingCode: string | null
  readonly pairingProgress: string | null
  readonly pairingError: string | null
  readonly onStartPairing: () => Promise<void>
  readonly onCancelPairing: () => Promise<void>
  readonly onJoinPrimary: (code: string) => Promise<{ ok: boolean; error?: string }>
}) {
  const [joinCode, setJoinCode] = useState('')
  const [isJoining, setIsJoining] = useState(false)
  const [isStartingPairing, setIsStartingPairing] = useState(false)
  const [localError, setLocalError] = useState<string | null>(null)
  const isPairing = status.pairing || pairingCode !== null
  const isRunning = status.running ?? false
  const hasDevices = (status.devices?.length ?? 0) > 0

  const displayError = localError || pairingError

  const handleJoin = useCallback(
    async (code: string) => {
      setIsJoining(true)
      setLocalError(null)
      try {
        const result = await onJoinPrimary(code)
        if (result.ok) {
          setJoinCode('')
        } else {
          setLocalError(result.error ?? 'Failed to join primary')
        }
      } finally {
        setIsJoining(false)
      }
    },
    [onJoinPrimary],
  )

  const handleStartPairing = useCallback(async () => {
    setIsStartingPairing(true)
    setLocalError(null)
    try {
      await onStartPairing()
    } finally {
      setIsStartingPairing(false)
    }
  }, [onStartPairing])

  if (!isRunning) {
    return (
      <Section title="Syncthing not running">
        <div className="space-y-3">
          <div className="flex items-center gap-3">
            <Spinner />
            <span className="text-sm text-on-surface-muted">Waiting for syncthing to start...</span>
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
    if (hasDevices) {
      return null
    }

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
          {displayError && <p className="text-sm text-status-error">{displayError}</p>}
          <Button onClick={() => handleJoin(joinCode.trim())} disabled={!joinCode.trim()}>
            Join primary
          </Button>
        </div>
      </Section>
    )
  }

  if (isStartingPairing) {
    return (
      <Section title="Pair a device">
        <div className="flex items-center gap-3">
          <Spinner />
          <span className="text-sm text-on-surface-muted">Starting pairing...</span>
        </div>
      </Section>
    )
  }

  return (
    <Section title="Pair a device">
      <p className="text-sm text-on-surface-muted mb-3">
        Start pairing to connect another device on your local network.
      </p>
      <Button onClick={handleStartPairing}>Start pairing</Button>
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
  onRemoveDevice,
  onStartPairing,
  onCancelPairing,
  onJoinPrimary,
  onEnableSync,
  onRefresh,
  pairingCode,
  pairingProgress,
  pairingError,
  isEnabling,
  enableError,
}: SyncViewProps) {
  if (!status?.enabled) {
    return (
      <DisabledState onEnable={onEnableSync} isEnabling={isEnabling} enableError={enableError} />
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
        <StatusBadge label={status.running ? 'running' : 'stopped'} ok={status.running ?? false} />
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
        pairingError={pairingError}
        onStartPairing={onStartPairing}
        onCancelPairing={onCancelPairing}
        onJoinPrimary={onJoinPrimary}
      />

      {sortedFolders.length > 0 && (
        <Section title="Synced folders">
          <div className="border border-outline rounded-card px-3 bg-surface">
            {sortedFolders.map((folder) => (
              <FolderRow key={folder.id} folder={folder} onRefresh={onRefresh} />
            ))}
          </div>
        </Section>
      )}

      {status.guiURL && (
        <Section title="Advanced" collapsible defaultCollapsed>
          <a
            href={status.guiURL}
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-accent hover:underline"
          >
            Open Syncthing web interface
          </a>
        </Section>
      )}
    </div>
  )
}
