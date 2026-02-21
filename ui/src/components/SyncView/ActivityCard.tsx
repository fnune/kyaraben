import { formatBytes } from '@/lib/changeUtils'
import { ProgressBar, ProgressRail } from '@/lib/progressWidgets'
import type { SyncFolder, SyncState } from '@/types/daemon'

function formatTimeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

export interface ActivityCardProps {
  readonly state: SyncState | undefined
  readonly folders: SyncFolder[] | undefined
  readonly lastSyncedAt: Date | null
  readonly hasPairedDevices: boolean
}

function ScanningProgress({ folders }: { readonly folders: SyncFolder[] }) {
  const currentFolder = folders[0]
  if (!currentFolder) return null

  const totalNeedBytes = folders.reduce((sum, f) => sum + f.needSize, 0)
  const totalGlobalBytes = folders.reduce((sum, f) => sum + f.globalSize, 0)
  const hasProgressData = totalGlobalBytes > 0
  const percent = hasProgressData
    ? Math.round(((totalGlobalBytes - totalNeedBytes) / totalGlobalBytes) * 100)
    : null
  const queueCount = folders.length > 1 ? folders.length - 1 : 0

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="text-on-surface-muted animate-pulse text-sm leading-none">◉</span>
          <span className="text-sm font-medium text-on-surface">{currentFolder.label}</span>
        </div>
        <span className="text-xs text-on-surface-muted">
          {hasProgressData ? `Scanning ${formatBytes(totalNeedBytes)} remaining` : 'Scanning...'}
        </span>
      </div>
      <ProgressRail className="h-2 bg-outline/30 rounded-full overflow-hidden">
        {percent !== null ? (
          <ProgressBar percent={percent} />
        ) : (
          <div className="h-full w-1/4 bg-on-surface-muted/50 rounded-full animate-shimmer" />
        )}
      </ProgressRail>
      {queueCount > 0 && (
        <div className="mt-2 text-xs text-on-surface-muted">
          +{queueCount} folder{queueCount === 1 ? '' : 's'} in queue
        </div>
      )}
    </div>
  )
}

function SyncingProgress({ folders }: { readonly folders: SyncFolder[] }) {
  const currentFolder = folders[0]
  if (!currentFolder) return null

  const totalNeedBytes = folders.reduce((sum, f) => sum + f.needSize, 0)
  const totalGlobalBytes = folders.reduce((sum, f) => sum + f.globalSize, 0)
  const percent =
    totalGlobalBytes > 0
      ? Math.round(((totalGlobalBytes - totalNeedBytes) / totalGlobalBytes) * 100)
      : 100

  const queueCount = folders.length > 1 ? folders.length - 1 : 0

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span
            className="text-accent text-sm leading-none animate-spin"
            style={{ animationDuration: '2s' }}
          >
            ↻
          </span>
          <span className="text-sm font-medium text-on-surface">{currentFolder.label}</span>
        </div>
        <span className="text-xs text-on-surface-muted">
          {formatBytes(totalNeedBytes)} remaining
        </span>
      </div>
      <ProgressRail className="h-2 bg-outline/30 rounded-full">
        <ProgressBar percent={percent} />
      </ProgressRail>
      {queueCount > 0 && (
        <div className="mt-2 text-xs text-on-surface-muted">
          +{queueCount} folder{queueCount === 1 ? '' : 's'} in queue
        </div>
      )}
    </div>
  )
}

export function ActivityCard({ folders, lastSyncedAt, hasPairedDevices }: ActivityCardProps) {
  const scanningFolders = folders?.filter((f) => f.state === 'scanning') ?? []
  const syncingFolders =
    folders?.filter((f) => f.state === 'syncing' || (f.state === 'idle' && f.needSize > 0)) ?? []
  const errorFolders = folders?.filter((f) => f.state === 'error') ?? []

  const isScanning = scanningFolders.length > 0
  const isSyncing = syncingFolders.length > 0
  const isError = errorFolders.length > 0

  if (!hasPairedDevices) {
    return (
      <div className="p-4 bg-surface-alt rounded-card">
        <div className="flex items-center gap-2">
          <span className="text-on-surface-muted">○</span>
          <span className="text-sm text-on-surface-muted">Waiting for device connection</span>
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="p-4 bg-status-error/10 rounded-card">
        <div className="flex items-center gap-2 mb-2">
          <span className="text-status-error">✕</span>
          <span className="text-sm font-medium text-status-error">Synchronization error</span>
        </div>
        <div className="ml-5 space-y-1">
          {errorFolders.map((folder) => (
            <div key={folder.id} className="text-sm">
              <span className="text-status-error">{folder.label}</span>
              {folder.error && <span className="text-on-surface-muted">: {folder.error}</span>}
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (!isScanning && !isSyncing) {
    return (
      <div className="p-4 bg-surface-alt rounded-card">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-status-ok">✓</span>
            <span className="text-sm font-medium text-on-surface">All synchronized</span>
          </div>
          {lastSyncedAt && (
            <span className="text-xs text-on-surface-muted">
              Last synchronized: {formatTimeAgo(lastSyncedAt)}
            </span>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className="p-4 bg-surface-alt rounded-card space-y-3">
      {isSyncing && <SyncingProgress folders={syncingFolders} />}
      {isScanning && <ScanningProgress folders={scanningFolders} />}
    </div>
  )
}
