import { formatBytes } from '@/lib/changeUtils'
import { ProgressBar, ProgressRail } from '@/lib/progressWidgets'
import type { SyncFolder, SyncState } from '@/types/daemon'
import { SyncStateSyncing } from '@/types/daemon'

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

export function ActivityCard({
  state,
  folders,
  lastSyncedAt,
  hasPairedDevices,
}: ActivityCardProps) {
  const syncingFolders = folders?.filter((f) => f.needSize > 0) ?? []
  const isSyncing = state === SyncStateSyncing || syncingFolders.length > 0
  const currentFolder = syncingFolders[0]
  const queueCount = syncingFolders.length > 1 ? syncingFolders.length - 1 : 0

  const totalNeedBytes = syncingFolders.reduce((sum, f) => sum + f.needSize, 0)
  const totalGlobalBytes = syncingFolders.reduce((sum, f) => sum + f.globalSize, 0)
  const percent =
    totalGlobalBytes > 0
      ? Math.round(((totalGlobalBytes - totalNeedBytes) / totalGlobalBytes) * 100)
      : 100

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

  if (isSyncing && currentFolder) {
    return (
      <div className="p-4 bg-surface-alt rounded-card">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <span className="text-accent animate-spin" style={{ animationDuration: '2s' }}>
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

  return (
    <div className="p-4 bg-surface-alt rounded-card">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-status-ok">✓</span>
          <span className="text-sm font-medium text-on-surface">All synced</span>
        </div>
        {lastSyncedAt && (
          <span className="text-xs text-on-surface-muted">
            Last sync: {formatTimeAgo(lastSyncedAt)}
          </span>
        )}
      </div>
    </div>
  )
}
