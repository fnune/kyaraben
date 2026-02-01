import { useApply } from '@/lib/ApplyContext'
import { BOTTOM_BAR_HEIGHT } from '@/lib/BottomBar'
import { BottomBarPortal } from '@/lib/BottomBarSlot'

export interface ApplyProgressBarProps {
  readonly onNavigateToSystems: () => void
}

export function ApplyProgressBar({ onNavigateToSystems }: ApplyProgressBarProps) {
  const { progressSteps } = useApply()

  const currentStep = [...progressSteps].reverse().find((s) => s.status === 'in_progress')
  const label = currentStep?.label ?? 'Installing...'
  const detail = currentStep?.message

  return (
    <BottomBarPortal>
      <div
        className={`bg-gray-800/95 backdrop-blur border-t border-gray-700 px-6 ${BOTTOM_BAR_HEIGHT} flex items-center`}
      >
        <div className="flex-1 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-4 h-4 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
            <span className="text-sm text-gray-300 truncate max-w-md">
              {label}
              {detail && <span className="text-gray-500 ml-2">— {detail}</span>}
            </span>
          </div>
          <button
            type="button"
            onClick={onNavigateToSystems}
            className="text-blue-400 hover:text-blue-300 hover:underline text-sm"
          >
            View progress
          </button>
        </div>
      </div>
    </BottomBarPortal>
  )
}
