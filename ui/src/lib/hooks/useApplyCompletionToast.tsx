import { useEffect, useRef } from 'react'
import { useToast } from '@/lib/ToastContext'
import type { ApplyStatus, View } from '@/types/ui'
import { VIEW_CATALOG, VIEW_LABELS } from '@/types/ui'

export function useApplyCompletionToast(
  applyStatus: ApplyStatus,
  currentView: View,
  onNavigateToCatalog: () => void,
) {
  const { showToast } = useToast()
  const lastApplyStatus = useRef<ApplyStatus>(applyStatus)

  useEffect(() => {
    if (applyStatus === lastApplyStatus.current) return
    if (applyStatus === 'success') {
      if (currentView !== VIEW_CATALOG) {
        showToast(
          <span>
            Installation complete.{' '}
            <button
              type="button"
              className="underline hover:no-underline"
              onClick={onNavigateToCatalog}
            >
              Go to {VIEW_LABELS[VIEW_CATALOG].toLowerCase()}
            </button>
          </span>,
          'success',
          Infinity,
        )
      } else {
        showToast('Installation complete.', 'success')
      }
    }
    lastApplyStatus.current = applyStatus
  }, [applyStatus, currentView, showToast, onNavigateToCatalog])
}
