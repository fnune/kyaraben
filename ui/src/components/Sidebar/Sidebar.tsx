import { SettingsIcon } from '@/lib/icons'
import type { SyncState, SyncStatusResponse } from '@/types/daemon'
import {
  SyncStateConflict,
  SyncStateDisabled,
  SyncStateDisconnected,
  SyncStateError,
  SyncStateSynced,
  SyncStateSyncing,
} from '@/types/daemon'
import { VIEW_CATALOG, VIEW_INSTALLATION, VIEW_LABELS, VIEW_SYNC, type View } from '@/types/ui'

export interface SidebarProps {
  readonly currentView: View
  readonly onNavigate: (view: View) => void
  readonly syncStatus: SyncStatusResponse | null
  readonly version: string | null
}

interface NavItemProps {
  readonly label: string
  readonly active: boolean
  readonly onClick: () => void
  readonly indicator?: React.ReactNode
}

const syncDotColors: Record<SyncState, string> = {
  [SyncStateDisabled]: 'bg-on-surface-faint',
  [SyncStateSynced]: 'bg-status-ok',
  [SyncStateSyncing]: 'bg-accent animate-pulse',
  [SyncStateDisconnected]: 'bg-outline',
  [SyncStateConflict]: 'bg-status-warning',
  [SyncStateError]: 'bg-status-error',
}

function getSyncState(status: SyncStatusResponse | null): SyncState {
  if (!status?.enabled) return SyncStateDisabled
  if (!status.running) return SyncStateDisconnected
  return status.state ?? SyncStateSynced
}

function NavItem({ label, active, onClick, indicator }: NavItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`px-4 py-2 text-sm tracking-wide flex items-center gap-2 ${
        active
          ? 'bg-accent-muted text-accent min-[720px]:border-l-2 min-[720px]:border-accent max-[719px]:border-b-2 max-[719px]:border-accent'
          : 'text-on-surface-secondary hover:bg-surface-raised min-[720px]:border-l-2 max-[719px]:border-b-2 border-transparent'
      }`}
    >
      <span>{label}</span>
      {indicator}
    </button>
  )
}

export function Sidebar({ currentView, onNavigate, syncStatus, version }: SidebarProps) {
  const syncState = getSyncState(syncStatus)
  const syncDotColor = syncDotColors[syncState]

  return (
    <aside className="bg-surface-alt border-b min-[720px]:border-b-0 min-[720px]:border-r border-outline flex flex-row min-[720px]:flex-col min-[720px]:w-56">
      <div className="p-4 border-r min-[720px]:border-r-0 min-[720px]:border-b border-outline">
        <h1 className="font-heading text-lg font-semibold text-accent tracking-wide">Kyaraben</h1>
      </div>

      <nav className="flex-1 flex flex-row min-[720px]:flex-col min-[720px]:py-2">
        <NavItem
          label={VIEW_LABELS[VIEW_CATALOG]}
          active={currentView === VIEW_CATALOG}
          onClick={() => onNavigate(VIEW_CATALOG)}
        />
        <NavItem
          label={VIEW_LABELS[VIEW_SYNC]}
          active={currentView === VIEW_SYNC}
          onClick={() => onNavigate(VIEW_SYNC)}
          indicator={<span className={`w-2 h-2 rounded-full ${syncDotColor}`} />}
        />
      </nav>

      <div className="flex items-center justify-between p-4 min-[720px]:border-t border-l min-[720px]:border-l-0 border-outline">
        <span className="text-xs text-on-surface-dim font-mono truncate">{version ?? ''}</span>
        <button
          type="button"
          onClick={() => onNavigate(VIEW_INSTALLATION)}
          className={`p-1 rounded-sm transition-colors ${
            currentView === VIEW_INSTALLATION
              ? 'text-accent bg-accent-muted'
              : 'text-on-surface-dim hover:text-on-surface-secondary hover:bg-surface-raised'
          }`}
          title={VIEW_LABELS[VIEW_INSTALLATION]}
          aria-label={VIEW_LABELS[VIEW_INSTALLATION]}
        >
          <SettingsIcon className="w-4 h-4" />
        </button>
      </div>
    </aside>
  )
}
