import { useCallback, useState } from 'react'
import { Button } from '@/lib/Button'
import { Spinner } from '@/lib/Spinner'
import type { SyncStatusResponse } from '@/types/daemon'
import { ActivityCard } from './ActivityCard'
import { FoldersCard } from './FoldersCard'
import { StatusCard } from './StatusCard'
import { SyncSettingsSection } from './SyncSettingsSection'

export interface SyncViewProps {
  readonly status: SyncStatusResponse | null
  readonly connectionProgress: string | null
  readonly connectionError: string | null
  readonly isConnecting: boolean
  readonly isPairing: boolean
  readonly pairingDeviceId: string | null
  readonly pairingCode: string | null
  readonly lastSyncedAt: Date | null
  readonly onRemoveDevice: (deviceId: string) => Promise<void>
  readonly onConnectToDevice: (deviceId: string) => Promise<{ ok: boolean; error?: string }>
  readonly onEnableSync: () => Promise<void>
  readonly onResetSync: () => Promise<void>
  readonly onStartPairing: () => Promise<void>
  readonly onStopPairing: () => Promise<void>
  readonly onClearConnectionError: () => void
  readonly onRefresh: () => void
  readonly isEnabling: boolean
  readonly enableError: string | null
}

function DisabledState({
  status,
  onEnable,
  onReset,
  isEnabling,
  enableError,
}: {
  readonly status: SyncStatusResponse | null
  readonly onEnable: () => Promise<void>
  readonly onReset: () => Promise<void>
  readonly isEnabling: boolean
  readonly enableError: string | null
}) {
  const [isResetting, setIsResetting] = useState(false)
  const [showResetConfirm, setShowResetConfirm] = useState(false)

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
      <div className="p-4 bg-surface-alt rounded-card">
        <h3 className="text-sm font-medium text-on-surface mb-3">Enable synchronization</h3>
        <p className="text-sm text-on-surface-muted mb-4">
          Synchronize your saves, states, and screenshots across devices using Syncthing.
        </p>

        {isEnabling ? (
          <div className="flex items-center gap-3">
            <Spinner />
            <span className="text-sm text-on-surface-muted">Installing syncthing...</span>
          </div>
        ) : (
          <div className="space-y-4">
            <Button onClick={onEnable}>Enable synchronization</Button>
            {enableError && (
              <div className="p-4 bg-status-error/10 border border-status-error/30 rounded-card">
                <p className="text-sm text-status-error">{enableError}</p>
              </div>
            )}
          </div>
        )}
      </div>

      {hasOrphanedState && (
        <div className="p-4 bg-surface-alt rounded-card">
          <h3 className="text-sm font-medium text-on-surface mb-3">Orphaned syncthing state</h3>
          <p className="text-sm text-on-surface-muted mb-3">
            Syncthing files from a previous installation were detected. This can happen after an
            incomplete uninstall or if synchronization was disabled manually.
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
        </div>
      )}
    </div>
  )
}

function NotRunningState({ serviceError }: { readonly serviceError: string | undefined }) {
  return (
    <div className="p-6">
      <div className="p-4 bg-surface-alt rounded-card">
        <h3 className="text-sm font-medium text-on-surface mb-3">Syncthing not running</h3>
        <div className="space-y-3">
          <div className="flex items-center gap-3">
            <Spinner />
            <span className="text-sm text-on-surface-muted">Waiting for syncthing to start...</span>
          </div>
          {serviceError && (
            <details className="text-sm">
              <summary className="text-on-surface-muted cursor-pointer hover:text-on-surface">
                Service logs
              </summary>
              <pre className="mt-2 text-xs text-on-surface-muted bg-surface-raised p-3 rounded-sm overflow-x-auto whitespace-pre-wrap">
                {serviceError}
              </pre>
            </details>
          )}
        </div>
      </div>
    </div>
  )
}

export function SyncView({
  status,
  connectionProgress,
  connectionError,
  isConnecting,
  isPairing,
  pairingDeviceId,
  pairingCode,
  lastSyncedAt,
  onRemoveDevice,
  onConnectToDevice,
  onEnableSync,
  onResetSync,
  onStartPairing,
  onStopPairing,
  onClearConnectionError,
  onRefresh,
  isEnabling,
  enableError,
}: SyncViewProps) {
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
    return <NotRunningState serviceError={status.serviceError} />
  }

  const hasPairedDevices = (status.devices?.length ?? 0) > 0

  return (
    <div className="p-6 space-y-4">
      <StatusCard
        status={status}
        connectionProgress={connectionProgress}
        connectionError={connectionError}
        isConnecting={isConnecting}
        isPairing={isPairing}
        pairingDeviceId={pairingDeviceId}
        pairingCode={pairingCode}
        onRemoveDevice={onRemoveDevice}
        onConnectToDevice={onConnectToDevice}
        onStartPairing={onStartPairing}
        onStopPairing={onStopPairing}
        onClearConnectionError={onClearConnectionError}
      />

      <ActivityCard
        state={status.state}
        folders={status.folders}
        lastSyncedAt={lastSyncedAt}
        hasPairedDevices={hasPairedDevices}
      />

      <FoldersCard
        folders={status.folders}
        onRefresh={onRefresh}
        hasPairedDevices={hasPairedDevices}
      />

      <SyncSettingsSection guiURL={status.guiURL} onReset={onResetSync} />
    </div>
  )
}
