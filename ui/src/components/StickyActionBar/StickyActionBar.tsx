import { useEffect, useState } from 'react'
import { BottomBar } from '@/lib/BottomBar'
import { Button } from '@/lib/Button'
import { CHANGE_CONFIG, type ChangeSummary, formatBytes, getChangeGroups } from '@/lib/changeUtils'

export interface StickyActionBarProps {
  readonly changes: ChangeSummary
  readonly onApply: (changes: ChangeSummary) => void
  readonly onDiscard: () => void
  readonly applying?: boolean
  readonly upgradeAvailable?: boolean
  readonly onReapply?: () => void
}

export function StickyActionBar({
  changes,
  onApply,
  onDiscard,
  applying = false,
  upgradeAvailable = false,
  onReapply,
}: StickyActionBarProps) {
  const [confirmingDiscard, setConfirmingDiscard] = useState(false)

  useEffect(() => {
    if (!confirmingDiscard) return
    const timer = setTimeout(() => setConfirmingDiscard(false), 3000)
    return () => clearTimeout(timer)
  }, [confirmingDiscard])

  const hasChanges = changes.total > 0 || changes.hasConfigChanges
  if (!hasChanges && !upgradeAvailable) return null

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
  const changeGroups = getChangeGroups(changes)
  const upgradeOnly = upgradeAvailable && !hasChanges

  return (
    <BottomBar>
      <div className="flex items-center gap-4 text-sm min-w-0 overflow-hidden">
        {upgradeOnly ? (
          <span className="text-on-surface-secondary">
            Kyaraben was updated. Apply to get the latest emulator configs.
          </span>
        ) : (
          <>
            {changeGroups.length > 0 && (
              <div className="flex items-center gap-3 min-w-0">
                {changeGroups.map(({ type, items }) => {
                  const config = CHANGE_CONFIG[type]
                  const names = items.map((i) => i.name).join(', ')
                  return (
                    <span
                      key={type}
                      className={`flex items-center gap-1.5 min-w-0 ${config.textColor}`}
                    >
                      <span className="shrink-0">{config.icon}</span>
                      <span className="truncate">{names}</span>
                    </span>
                  )
                })}
              </div>
            )}
            {hasSize && (
              <div className="flex items-center gap-2 font-mono shrink-0">
                {changes.downloadBytes > 0 && (
                  <span className="text-status-ok">+{formatBytes(changes.downloadBytes)}</span>
                )}
                {changes.freeBytes > 0 && (
                  <span className="text-status-error">-{formatBytes(changes.freeBytes)}</span>
                )}
                {changes.downloadBytes > 0 && changes.freeBytes > 0 && (
                  <span className="text-on-surface-dim">
                    ({netBytes >= 0 ? '+' : '-'}
                    {formatBytes(netBytes)})
                  </span>
                )}
              </div>
            )}
          </>
        )}
      </div>

      <div className="flex items-center gap-4 shrink-0 ml-4">
        {changes.hasConfigChanges && (
          <button
            type="button"
            onClick={handleDiscard}
            disabled={applying}
            className="text-sm text-accent hover:text-accent-hover hover:underline disabled:opacity-50"
          >
            {confirmingDiscard ? 'Click again to confirm' : 'Discard changes'}
          </button>
        )}
        <Button
          onClick={upgradeOnly && onReapply ? onReapply : () => onApply(changes)}
          disabled={applying}
        >
          {applying ? 'Applying...' : 'Apply'}
        </Button>
      </div>
    </BottomBar>
  )
}
