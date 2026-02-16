import type { SyncProgress, SyncState } from '@/types/daemon'
import {
  SyncStateConflict,
  SyncStateDisconnected,
  SyncStateError,
  SyncStateSynced,
  SyncStateSyncing,
} from '@/types/daemon'

interface SyncStatusBannerProps {
  readonly state: SyncState
  readonly progress: SyncProgress | null
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`
}

function getStateConfig(state: SyncState) {
  switch (state) {
    case SyncStateSynced:
      return {
        icon: '✓',
        label: 'All files synced',
        bgClass: 'bg-status-ok/10',
        textClass: 'text-status-ok',
        dotClass: 'bg-status-ok',
      }
    case SyncStateSyncing:
      return {
        icon: '↻',
        label: 'Syncing',
        bgClass: 'bg-accent/10',
        textClass: 'text-accent',
        dotClass: 'bg-accent',
        animate: true,
      }
    case SyncStateDisconnected:
      return {
        icon: '○',
        label: 'No devices connected',
        bgClass: 'bg-outline/10',
        textClass: 'text-on-surface-muted',
        dotClass: 'bg-outline',
      }
    case SyncStateConflict:
      return {
        icon: '⚠',
        label: 'Sync conflicts detected',
        bgClass: 'bg-status-warn/10',
        textClass: 'text-status-warn',
        dotClass: 'bg-status-warn',
      }
    case SyncStateError:
      return {
        icon: '✕',
        label: 'Sync error',
        bgClass: 'bg-status-error/10',
        textClass: 'text-status-error',
        dotClass: 'bg-status-error',
      }
    default:
      return {
        icon: '●',
        label: 'Unknown state',
        bgClass: 'bg-outline/10',
        textClass: 'text-on-surface-muted',
        dotClass: 'bg-outline',
      }
  }
}

export function SyncStatusBanner({ state, progress }: SyncStatusBannerProps) {
  const config = getStateConfig(state)
  const showProgress = state === SyncStateSyncing && progress && progress.needFiles > 0

  let label = config.label
  if (showProgress && progress) {
    label = `Syncing ${progress.needFiles} file${progress.needFiles === 1 ? '' : 's'} (${formatBytes(progress.needBytes)})`
  }

  return (
    <div className={`p-4 rounded-card ${config.bgClass}`}>
      <div className="flex items-center gap-3">
        <span
          className={`text-lg ${config.textClass} ${config.animate ? 'animate-spin' : ''}`}
          style={config.animate ? { animationDuration: '2s' } : undefined}
        >
          {config.icon}
        </span>
        <span className={`font-medium ${config.textClass}`}>{label}</span>
      </div>
      {showProgress && progress && (
        <div className="mt-3">
          <div className="h-2 bg-outline/30 rounded-full overflow-hidden">
            <div
              className="h-full bg-accent transition-all duration-300"
              style={{ width: `${progress.percent}%` }}
            />
          </div>
          <div className="mt-1 text-xs text-on-surface-muted text-right">{progress.percent}%</div>
        </div>
      )}
    </div>
  )
}
