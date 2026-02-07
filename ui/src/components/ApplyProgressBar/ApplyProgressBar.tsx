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
  const { status, progressSteps, cancel } = useApply()
  const handleOpenLog = useOpenLog()
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
        className={`bg-gray-800/95 backdrop-blur-sm border-t border-gray-700 px-6 ${BOTTOM_BAR_HEIGHT} flex items-center`}
      >
        <div className="flex-1 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-4 h-4 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
            <span className="text-sm text-gray-300 truncate max-w-md">
              {label}
              {detail && <span className="text-gray-500 ml-2">— {detail}</span>}
            </span>
          </div>
          <div className="flex items-center gap-4">
            <button
              type="button"
              onClick={handleOpenLog}
              className="text-gray-400 hover:text-gray-300 hover:underline text-sm"
            >
              Open log in terminal
            </button>
            <button
              type="button"
              onClick={handleCancel}
              disabled={cancelling}
              className={`text-sm ${cancelling ? 'text-gray-500 cursor-not-allowed' : confirmingCancel ? 'text-red-400 hover:text-red-300 hover:underline' : 'text-gray-400 hover:text-gray-300 hover:underline'}`}
            >
              {cancelling ? 'Canceling...' : confirmingCancel ? 'Confirm cancel' : 'Cancel'}
            </button>
            {showViewProgress && (
              <button
                type="button"
                onClick={onNavigateToSystems}
                className="text-blue-400 hover:text-blue-300 hover:underline text-sm"
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
