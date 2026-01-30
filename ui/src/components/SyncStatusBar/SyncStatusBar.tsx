import type { SyncState, SyncStatusResponse } from '@/types/daemon'

export interface SyncStatusBarProps {
  readonly status: SyncStatusResponse | null
  readonly onOpenSettings: () => void
}

const stateConfig: Record<SyncState, { label: string; color: string; bgColor: string }> = {
  disabled: { label: 'Sync disabled', color: 'text-gray-500', bgColor: 'bg-gray-100' },
  synced: { label: 'Synced', color: 'text-green-700', bgColor: 'bg-green-50' },
  syncing: { label: 'Syncing...', color: 'text-blue-700', bgColor: 'bg-blue-50' },
  disconnected: { label: 'Disconnected', color: 'text-red-700', bgColor: 'bg-red-50' },
  conflict: { label: 'Conflicts', color: 'text-yellow-700', bgColor: 'bg-yellow-50' },
  error: { label: 'Error', color: 'text-red-700', bgColor: 'bg-red-50' },
}

function getDisplayState(status: SyncStatusResponse | null): SyncState {
  if (!status?.enabled) return 'disabled'
  if (!status.running) return 'disconnected'
  return status.state ?? 'synced'
}

export function SyncStatusBar({ status, onOpenSettings }: SyncStatusBarProps) {
  const state = getDisplayState(status)
  const config = stateConfig[state]

  const connectedCount = status?.devices?.filter((d) => d.connected).length ?? 0
  const totalDevices = status?.devices?.length ?? 0

  return (
    <button
      type="button"
      onClick={onOpenSettings}
      className={`w-full px-4 py-2 flex items-center justify-between ${config.bgColor} hover:opacity-90 transition-opacity`}
    >
      <div className="flex items-center gap-2">
        <span className={`text-sm font-medium ${config.color}`}>{config.label}</span>
        {status?.enabled && status.running && totalDevices > 0 && (
          <span className="text-xs text-gray-500">
            ({connectedCount}/{totalDevices} devices)
          </span>
        )}
      </div>
      <div className="flex items-center gap-2">
        {status?.enabled && status.mode && (
          <span className="text-xs text-gray-500 capitalize">{status.mode}</span>
        )}
        <span className="text-xs text-gray-400">Click to configure</span>
      </div>
    </button>
  )
}
