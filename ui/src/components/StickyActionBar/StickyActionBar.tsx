import { useEffect, useState } from 'react'
import { BottomBar } from '@/lib/BottomBar'
import { Button } from '@/lib/Button'
import { type ChangeSummary, formatBytes } from '@/lib/changeUtils'

export interface StickyActionBarProps {
  readonly changes: ChangeSummary
  readonly onApply: () => void
  readonly onDiscard: () => void
  readonly applying?: boolean
}

export function StickyActionBar({
  changes,
  onApply,
  onDiscard,
  applying = false,
}: StickyActionBarProps) {
  const [confirmingDiscard, setConfirmingDiscard] = useState(false)

  useEffect(() => {
    if (!confirmingDiscard) return
    const timer = setTimeout(() => setConfirmingDiscard(false), 3000)
    return () => clearTimeout(timer)
  }, [confirmingDiscard])

  if (changes.total === 0 && !changes.hasConfigChanges) return null

  const handleDiscard = () => {
    if (confirmingDiscard) {
      onDiscard()
      setConfirmingDiscard(false)
    } else {
      setConfirmingDiscard(true)
    }
  }

  const netBytes = changes.downloadBytes - changes.freeBytes
  const hasSize = changes.downloadBytes > 0 || changes.freeBytes > 0

  return (
    <BottomBar>
      <div className="flex items-center gap-4 text-sm">
        {hasSize && (
          <div className="flex items-center gap-2 font-mono">
            {changes.downloadBytes > 0 && (
              <span className="text-emerald-400">+{formatBytes(changes.downloadBytes)}</span>
            )}
            {changes.freeBytes > 0 && (
              <span className="text-red-400">-{formatBytes(changes.freeBytes)}</span>
            )}
            {changes.downloadBytes > 0 && changes.freeBytes > 0 && (
              <span className="text-gray-500">
                ({netBytes >= 0 ? '+' : '-'}
                {formatBytes(netBytes)})
              </span>
            )}
          </div>
        )}
        {changes.hasConfigChanges && (
          <button
            type="button"
            onClick={handleDiscard}
            disabled={applying}
            className="text-blue-400 hover:text-blue-300 hover:underline disabled:opacity-50"
          >
            {confirmingDiscard ? 'Click again to confirm' : 'Discard changes'}
          </button>
        )}
      </div>

      <Button onClick={onApply} disabled={applying}>
        {applying ? 'Applying...' : 'Apply'}
      </Button>
    </BottomBar>
  )
}
