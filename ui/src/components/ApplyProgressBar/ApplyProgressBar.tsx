import { useEffect, useState } from 'react'
import { useApply } from '@/lib/ApplyContext'
import { BOTTOM_BAR_HEIGHT } from '@/lib/BottomBar'
import { BottomBarPortal } from '@/lib/BottomBarSlot'
import { useOpenLog } from '@/lib/useOpenLog'

export interface ApplyProgressBarProps {
  readonly currentView: string
  readonly onNavigateToSystems: () => void
}

export function ApplyProgressBar({ currentView, onNavigateToSystems }: ApplyProgressBarProps) {
  const { status, progressSteps, cancel, logPosition } = useApply()
  const openLog = useOpenLog()
  const [confirmingCancel, setConfirmingCancel] = useState(false)
  const [cancelling, setCancelling] = useState(false)

  useEffect(() => {
    if (status !== 'applying') {
      setCancelling(false)
      setConfirmingCancel(false)
    }
  }, [status])

  useEffect(() => {
    if (!confirmingCancel) return
    const timer = setTimeout(() => setConfirmingCancel(false), 3000)
    return () => clearTimeout(timer)
  }, [confirmingCancel])

  const handleCancel = () => {
    if (cancelling) return
    if (confirmingCancel) {
      setCancelling(true)
      setConfirmingCancel(false)
      cancel()
    } else {
      setConfirmingCancel(true)
    }
  }

  if (status !== 'applying') return null

  const currentStep = [...progressSteps].reverse().find((s) => s.status === 'in_progress')
  const label = currentStep?.label ?? 'Installing...'
  const detail = currentStep?.message
  const showViewProgress = currentView !== 'systems'

  return (
    <BottomBarPortal>
      <div
        className={`bg-surface-alt/95 backdrop-blur-sm border-t border-outline px-6 ${BOTTOM_BAR_HEIGHT} flex items-center`}
      >
        <div className="flex-1 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-4 h-4 border-2 border-accent border-t-transparent rounded-full animate-spin" />
            <span className="text-sm text-on-surface-secondary truncate max-w-md">
              {label}
              {detail && <span className="text-on-surface-dim ml-2">— {detail}</span>}
            </span>
          </div>
          <div className="flex items-center gap-4">
            <button
              type="button"
              onClick={() => openLog(logPosition ?? undefined)}
              className="text-on-surface-muted hover:text-on-surface-secondary hover:underline text-sm"
            >
              Open log in terminal
            </button>
            <button
              type="button"
              onClick={handleCancel}
              disabled={cancelling}
              className={`text-sm ${cancelling ? 'text-on-surface-dim cursor-not-allowed' : confirmingCancel ? 'text-status-error hover:text-status-error hover:underline' : 'text-on-surface-muted hover:text-on-surface-secondary hover:underline'}`}
            >
              {cancelling ? 'Canceling...' : confirmingCancel ? 'Confirm cancel' : 'Cancel'}
            </button>
            {showViewProgress && (
              <button
                type="button"
                onClick={onNavigateToSystems}
                className="text-accent hover:text-accent-hover hover:underline text-sm"
              >
                View progress
              </button>
            )}
          </div>
        </div>
      </div>
    </BottomBarPortal>
  )
}
