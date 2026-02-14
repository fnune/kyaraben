import { useEffect, useState } from 'react'
import { SpeedBadge } from '@/components/SpeedBadge/SpeedBadge'
import { useApply } from '@/lib/ApplyContext'
import { BOTTOM_BAR_HEIGHT } from '@/lib/BottomBar'
import { BottomBarPortal } from '@/lib/BottomBarSlot'
import { getDownloadSpeedBytes, getStepSubtitle } from '@/lib/progressUtils'
import { ProgressBar, ProgressRail, Shimmer } from '@/lib/progressWidgets'
import { useOpenLog } from '@/lib/useOpenLog'
import { VIEW_CATALOG } from '@/types/ui'

export interface ApplyProgressBarProps {
  readonly currentView: string
  readonly onNavigateToCatalog: () => void
}

export function ApplyProgressBar({ currentView, onNavigateToCatalog }: ApplyProgressBarProps) {
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
  const subtitle = currentStep ? getStepSubtitle(currentStep) : null
  const showSpeed = currentStep?.id === 'build' && currentStep.status === 'in_progress'
  const downloadSpeedBytes = currentStep ? getDownloadSpeedBytes(currentStep) : 0
  const computedPercent =
    currentStep?.bytesTotal && currentStep.bytesTotal > 0
      ? Math.min(
          100,
          Math.floor(((currentStep.bytesDownloaded ?? 0) * 100) / currentStep.bytesTotal),
        )
      : undefined
  const progressPercent = currentStep?.progressPercent ?? computedPercent
  const showViewProgress = currentView !== VIEW_CATALOG
  const showInlineProgress = currentView !== VIEW_CATALOG

  return (
    <BottomBarPortal>
      <div
        className={`relative bg-surface-alt/95 backdrop-blur-sm border-t border-outline px-6 ${BOTTOM_BAR_HEIGHT} flex items-center`}
      >
        {showInlineProgress && (
          <ProgressRail className="absolute left-0 right-0 bottom-0 h-1 pointer-events-none">
            {progressPercent !== undefined ? (
              <ProgressBar percent={progressPercent} />
            ) : (
              <Shimmer />
            )}
          </ProgressRail>
        )}
        <div className="flex-1 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-4 h-4 border-2 border-accent border-t-transparent rounded-full animate-spin" />
            <span className="text-sm text-on-surface-secondary truncate max-w-md">
              <span className="font-medium">{label}</span>
              {subtitle && <span className="text-on-surface-dim ml-1">{subtitle}</span>}
            </span>
          </div>
          <div className="flex items-center gap-4">
            <SpeedBadge speedBytes={downloadSpeedBytes} show={showSpeed && showInlineProgress} />
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
                onClick={onNavigateToCatalog}
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
