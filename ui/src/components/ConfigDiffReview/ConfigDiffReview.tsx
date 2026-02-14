import { BottomBar } from '@/lib/BottomBar'
import { Button } from '@/lib/Button'
import { openPath } from '@/lib/daemon'
import { useScrollToTop } from '@/lib/useScrollToTop'
import type { ConfigChangeDetail, ConfigFileDiff, PreflightResponse } from '@/types/daemon'

export interface ConfigDiffReviewProps {
  readonly data: PreflightResponse
  readonly onConfirm: () => void
  readonly onCancel: () => void
}

function changeLabel(type: string): string {
  switch (type) {
    case 'add':
      return '+'
    case 'modify':
      return '~'
    case 'remove':
      return '-'
    default:
      return ' '
  }
}

function changeColor(type: string): string {
  switch (type) {
    case 'add':
      return 'text-status-ok'
    case 'modify':
      return 'text-status-warning'
    case 'remove':
      return 'text-status-error'
    default:
      return 'text-on-surface-muted'
  }
}

function ChangeRow({ change }: { readonly change: ConfigChangeDetail }) {
  const prefix = change.section ? `${change.section}.` : ''
  const key = `${prefix}${change.key}`

  return (
    <div className={`font-mono text-xs ${changeColor(change.type)} pl-4`}>
      <span>{changeLabel(change.type)} </span>
      <span>{key}</span>
      {change.type === 'modify' && (
        <span className="text-on-surface-dim">
          {' '}
          {change.oldValue} &rarr; {change.newValue}
        </span>
      )}
      {change.type === 'add' && <span className="text-on-surface-dim"> = {change.newValue}</span>}
    </div>
  )
}

function FileDiff({ diff }: { readonly diff: ConfigFileDiff }) {
  const hasConflict = diff.userModified && diff.hasChanges && (diff.userChanges?.length ?? 0) > 0

  return (
    <div className="border border-outline rounded-control p-3 space-y-2">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          {diff.isNewFile && (
            <span className="text-xs font-medium px-1.5 py-0.5 rounded bg-status-ok/10 text-status-ok">
              CREATE
            </span>
          )}
          {!diff.isNewFile && diff.hasChanges && (
            <span className="text-xs font-medium px-1.5 py-0.5 rounded bg-status-warning/10 text-status-warning">
              MODIFY
            </span>
          )}
          {!diff.isNewFile && !diff.hasChanges && (
            <span className="text-xs font-medium px-1.5 py-0.5 rounded bg-surface-raised/50 text-on-surface-muted">
              UNCHANGED
            </span>
          )}
          <span className="text-sm text-on-surface-secondary truncate font-mono">{diff.path}</span>
        </div>
        <button
          type="button"
          onClick={() => openPath(diff.path)}
          className="text-xs text-accent hover:text-accent-hover hover:underline whitespace-nowrap"
        >
          Open file
        </button>
      </div>

      {hasConflict && diff.userChanges && (
        <div className="bg-status-warning/10 border border-status-warning/30 rounded px-3 py-2">
          <p className="text-xs text-status-warning font-medium mb-1">
            You modified keys managed by kyaraben (will be overwritten):
          </p>
          {diff.userChanges.map((uc) => (
            <div key={uc.key} className="font-mono text-xs text-status-warning/80 pl-2">
              {uc.key}: {uc.currentValue} &rarr; {uc.baselineValue}
            </div>
          ))}
        </div>
      )}

      {diff.changes && diff.changes.length > 0 && (
        <div className="space-y-0.5">
          {diff.changes.map((change, i) => (
            <ChangeRow key={`${change.key}-${i}`} change={change} />
          ))}
        </div>
      )}
    </div>
  )
}

export function ConfigDiffReview({ data, onConfirm, onCancel }: ConfigDiffReviewProps) {
  useScrollToTop()

  const conflictDiffs = data.diffs.filter(
    (d) => d.userModified && d.hasChanges && (d.userChanges?.length ?? 0) > 0,
  )
  const otherDiffs = data.diffs.filter(
    (d) => !(d.userModified && d.hasChanges && (d.userChanges?.length ?? 0) > 0),
  )

  return (
    <div className="p-6 pb-24">
      <h2 className="text-lg font-semibold text-on-surface mb-1">Config conflicts detected</h2>
      <p className="text-sm text-on-surface-muted mb-4">
        You manually changed config files managed by kyaraben. Applying will overwrite those
        changes.
      </p>

      {conflictDiffs.length > 0 && (
        <div className="space-y-3 mb-6">
          <h3 className="text-sm font-medium text-status-warning">
            Files with conflicts ({conflictDiffs.length})
          </h3>
          {conflictDiffs.map((diff) => (
            <FileDiff key={diff.path} diff={diff} />
          ))}
        </div>
      )}

      {otherDiffs.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-sm font-medium text-on-surface-muted">
            Other changes ({otherDiffs.length})
          </h3>
          {otherDiffs.map((diff) => (
            <FileDiff key={diff.path} diff={diff} />
          ))}
        </div>
      )}

      <BottomBar>
        <button
          type="button"
          onClick={onCancel}
          className="text-accent hover:text-accent-hover hover:underline"
        >
          Cancel
        </button>
        <Button onClick={onConfirm}>Continue and override</Button>
      </BottomBar>
    </div>
  )
}
