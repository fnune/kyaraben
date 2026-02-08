import { useState } from 'react'
import { BugReport } from '@/components/BugReport/BugReport'
import type { SyncState, SyncStatusResponse } from '@/types/daemon'
import {
  SyncStateConflict,
  SyncStateDisabled,
  SyncStateDisconnected,
  SyncStateError,
  SyncStateSynced,
  SyncStateSyncing,
} from '@/types/daemon'
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
  [SyncStateDisabled]: 'bg-gray-400',
  [SyncStateSynced]: 'bg-green-500',
  [SyncStateSyncing]: 'bg-blue-500 animate-pulse',
  [SyncStateDisconnected]: 'bg-red-500',
  [SyncStateConflict]: 'bg-yellow-500',
  [SyncStateError]: 'bg-red-500',
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

export function Sidebar({ currentView, onNavigate, syncStatus }: SidebarProps) {
  const [bugReportOpen, setBugReportOpen] = useState(false)
  const syncState = getSyncState(syncStatus)
  const syncDotColor = syncDotColors[syncState]

  return (
    <aside className="bg-surface-alt border-b min-[720px]:border-b-0 min-[720px]:border-r border-outline flex flex-row min-[720px]:flex-col min-[720px]:w-56">
      <div className="p-4 border-r min-[720px]:border-r-0 min-[720px]:border-b border-outline">
        <h1 className="font-heading text-lg font-semibold text-on-surface tracking-wider uppercase">
          Kyaraben
        </h1>
      </div>

      <nav className="flex-1 flex flex-row min-[720px]:flex-col min-[720px]:py-2">
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
        <NavItem
          label="Debug"
          active={currentView === 'debug'}
          onClick={() => onNavigate('debug')}
        />
      </nav>

      <div className="hidden min-[720px]:block p-4 border-t border-outline">
        <div className="flex items-center justify-between">
          <span className="text-xs text-on-surface-dim font-mono">v0.1.0</span>
          <button
            type="button"
            onClick={() => setBugReportOpen(true)}
            className="text-xs text-on-surface-dim hover:text-on-surface-secondary"
          >
            Report a problem
          </button>
        </div>
      </div>

      <BugReport open={bugReportOpen} onClose={() => setBugReportOpen(false)} />
    </aside>
  )
}
