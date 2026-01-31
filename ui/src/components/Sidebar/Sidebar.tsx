import type { SyncState, SyncStatusResponse } from '@/types/daemon'
import type { View } from '@/types/ui'

export interface SidebarProps {
  readonly currentView: View
  readonly onNavigate: (view: View) => void
  readonly syncStatus: SyncStatusResponse | null
}

interface NavItemProps {
  readonly label: string
  readonly active: boolean
  readonly onClick: () => void
  readonly indicator?: React.ReactNode
}

const syncDotColors: Record<SyncState, string> = {
  disabled: 'bg-gray-400',
  synced: 'bg-green-500',
  syncing: 'bg-blue-500 animate-pulse',
  disconnected: 'bg-red-500',
  conflict: 'bg-yellow-500',
  error: 'bg-red-500',
}

function getSyncState(status: SyncStatusResponse | null): SyncState {
  if (!status?.enabled) return 'disabled'
  if (!status.running) return 'disconnected'
  return status.state ?? 'synced'
}

function NavItem({ label, active, onClick, indicator }: NavItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`w-full text-left px-4 py-2 text-sm flex items-center justify-between ${
        active
          ? 'bg-blue-50 text-blue-700 border-l-2 border-blue-600'
          : 'text-gray-700 hover:bg-gray-100 border-l-2 border-transparent'
      }`}
    >
      <span>{label}</span>
      {indicator}
    </button>
  )
}

export function Sidebar({ currentView, onNavigate, syncStatus }: SidebarProps) {
  const syncState = getSyncState(syncStatus)
  const syncDotColor = syncDotColors[syncState]

  return (
    <aside className="w-56 bg-gray-50 border-r border-gray-200 flex flex-col">
      <div className="p-4 border-b border-gray-200">
        <h1 className="text-lg font-semibold text-gray-900">Kyaraben</h1>
      </div>

      <nav className="flex-1 py-2">
        <NavItem
          label="Systems"
          active={currentView === 'systems'}
          onClick={() => onNavigate('systems')}
        />
        <NavItem
          label="Installation"
          active={currentView === 'installation'}
          onClick={() => onNavigate('installation')}
        />
        <NavItem
          label="Sync"
          active={currentView === 'sync'}
          onClick={() => onNavigate('sync')}
          indicator={<span className={`w-2 h-2 rounded-full ${syncDotColor}`} />}
        />
      </nav>

      <div className="p-4 border-t border-gray-200">
        <span className="text-xs text-gray-400">v0.1.0</span>
      </div>
    </aside>
  )
}
