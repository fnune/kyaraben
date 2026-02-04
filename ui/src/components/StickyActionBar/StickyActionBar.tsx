import { Button } from '@/lib/Button'
import type { ChangeSummary } from '@/lib/changeUtils'

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
  if (changes.total === 0) return null

  return (
    <div className="fixed bottom-0 left-0 right-0 bg-white/95 backdrop-blur border-t border-gray-200 px-6 py-4 z-40">
      <div className="max-w-3xl mx-auto flex items-center justify-between">
        <div className="flex items-center gap-3 text-sm">
          {changes.installs > 0 && (
            <span className="text-emerald-600 font-medium">+{changes.installs}</span>
          )}
          {changes.removes > 0 && (
            <span className="text-red-500 font-medium">−{changes.removes}</span>
          )}
          {changes.upgrades > 0 && (
            <span className="text-sky-500 font-medium">↑{changes.upgrades}</span>
          )}
          {changes.downgrades > 0 && (
            <span className="text-amber-500 font-medium">↓{changes.downgrades}</span>
          )}
          <span className="text-gray-400">
            {changes.total} change{changes.total !== 1 ? 's' : ''}
          </span>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="secondary" onClick={onDiscard} disabled={applying}>
            Discard
          </Button>
          <Button onClick={onApply} disabled={applying}>
            {applying ? 'Applying...' : 'Apply'}
          </Button>
        </div>
      </div>
    </div>
  )
}
