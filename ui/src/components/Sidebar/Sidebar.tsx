import type { SyncState, SyncStatusResponse } from '@shared/daemon'
import {
  SyncStateConflict,
  SyncStateDisabled,
  SyncStateDisconnected,
  SyncStateError,
  SyncStateSynced,
  SyncStateSyncing,
} from '@shared/daemon'
import {
  VIEW_CATALOG,
  VIEW_IMPORT,
  VIEW_INSTALLATION,
  VIEW_LABELS,
  VIEW_PREFERENCES,
  VIEW_SYNC,
  type View,
} from '@shared/ui'
import { AppIcon } from '@/components/Logo/AppIcon'
import { Logo } from '@/components/Logo/Logo'
import { useOpenUrl } from '@/lib/hooks/useOpenUrl'
import { ExternalLinkIcon, SettingsIcon } from '@/lib/icons'

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
      className={`px-4 py-2 text-sm tracking-wide flex items-center gap-2 min-w-0 ${
        active
          ? 'bg-accent-muted text-accent min-[720px]:border-l-2 min-[720px]:border-accent max-[719px]:border-b-2 max-[719px]:border-accent'
          : 'text-on-surface-secondary hover:bg-surface-raised min-[720px]:border-l-2 max-[719px]:border-b-2 border-transparent'
      }`}
    >
      <span className="truncate">{label}</span>
      {indicator && <span className="shrink-0">{indicator}</span>}
    </button>
  )
}

export function Sidebar({ currentView, onNavigate, syncStatus, version }: SidebarProps) {
  const syncState = getSyncState(syncStatus)
  const syncDotColor = syncDotColors[syncState]
  const openUrl = useOpenUrl()

  return (
    <aside className="bg-surface-alt border-b min-[720px]:border-b-0 min-[720px]:border-r border-outline flex flex-row min-[720px]:flex-col min-[720px]:w-56">
      <div className="p-4 border-r min-[720px]:border-r-0 min-[720px]:border-b border-outline flex items-center">
        <AppIcon className="w-8 h-8 min-[720px]:hidden" />
        <Logo className="hidden min-[720px]:block w-full" />
      </div>

      <nav className="flex-1 min-w-0 flex flex-row min-[720px]:flex-col min-[720px]:py-2 overflow-hidden">
        <NavItem
          label={VIEW_LABELS[VIEW_CATALOG]}
          active={currentView === VIEW_CATALOG}
          onClick={() => onNavigate(VIEW_CATALOG)}
        />
        <NavItem
          label={VIEW_LABELS[VIEW_PREFERENCES]}
          active={currentView === VIEW_PREFERENCES}
          onClick={() => onNavigate(VIEW_PREFERENCES)}
        />
        <NavItem
          label={VIEW_LABELS[VIEW_SYNC]}
          active={currentView === VIEW_SYNC}
          onClick={() => onNavigate(VIEW_SYNC)}
          indicator={<span className={`w-2 h-2 rounded-full ${syncDotColor}`} />}
        />
        <NavItem
          label={VIEW_LABELS[VIEW_IMPORT]}
          active={currentView === VIEW_IMPORT}
          onClick={() => onNavigate(VIEW_IMPORT)}
        />
      </nav>

      <div className="hidden min-[720px]:block min-[720px]:py-2 min-[720px]:mt-auto">
        <button
          type="button"
          onClick={() => openUrl('https://kyaraben.dev')}
          className="w-full px-4 py-2 text-sm tracking-wide flex items-center gap-2 text-on-surface-secondary hover:bg-surface-raised min-[720px]:border-l-2 border-transparent"
        >
          <span>Documentation</span>
          <ExternalLinkIcon className="w-3 h-3" />
        </button>
        <button
          type="button"
          onClick={() => openUrl('https://kyaraben.dev/support/')}
          className="w-full px-4 py-2 text-sm tracking-wide flex items-center gap-2 text-on-surface-secondary hover:bg-surface-raised min-[720px]:border-l-2 border-transparent"
        >
          <span>Support the project</span>
          <ExternalLinkIcon className="w-3 h-3" />
        </button>
      </div>

      <div className="flex items-center justify-end min-[720px]:justify-between p-4 min-[720px]:border-t border-l min-[720px]:border-l-0 border-outline">
        <span className="hidden min-[720px]:block text-xs text-on-surface-dim font-mono truncate">
          {version ?? ''}
        </span>
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
