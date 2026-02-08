import { useEffect } from 'react'
import { BottomBar } from '@/lib/BottomBar'
import { Button } from '@/lib/Button'
import { openPath } from '@/lib/daemon'
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
      return 'text-emerald-400'
    case 'modify':
      return 'text-amber-400'
    case 'remove':
      return 'text-red-400'
    default:
      return 'text-gray-400'
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
        <span className="text-gray-500">
          {' '}
          {change.oldValue} &rarr; {change.newValue}
        </span>
      )}
      {change.type === 'add' && <span className="text-gray-500"> = {change.newValue}</span>}
    </div>
  )
}

function FileDiff({ diff }: { readonly diff: ConfigFileDiff }) {
  const hasConflict = diff.userModified && diff.hasChanges && (diff.userChanges?.length ?? 0) > 0

  return (
    <div className="border border-gray-700 rounded-md p-3 space-y-2">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          {diff.isNewFile && (
            <span className="text-xs font-medium px-1.5 py-0.5 rounded bg-emerald-900/50 text-emerald-400">
              CREATE
            </span>
          )}
          {!diff.isNewFile && diff.hasChanges && (
            <span className="text-xs font-medium px-1.5 py-0.5 rounded bg-amber-900/50 text-amber-400">
              MODIFY
            </span>
          )}
          {!diff.isNewFile && !diff.hasChanges && (
            <span className="text-xs font-medium px-1.5 py-0.5 rounded bg-gray-700/50 text-gray-400">
              UNCHANGED
            </span>
          )}
          <span className="text-sm text-gray-300 truncate font-mono">{diff.path}</span>
        </div>
        <button
          type="button"
          onClick={() => openPath(diff.path)}
          className="text-xs text-blue-400 hover:text-blue-300 hover:underline whitespace-nowrap"
        >
          Open file
        </button>
      </div>

      {hasConflict && diff.userChanges && (
        <div className="bg-amber-950/30 border border-amber-800/50 rounded px-3 py-2">
          <p className="text-xs text-amber-400 font-medium mb-1">
            You modified keys managed by kyaraben (will be overwritten):
          </p>
          {diff.userChanges.map((uc) => (
            <div key={uc.key} className="font-mono text-xs text-amber-300/80 pl-2">
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
  useEffect(() => {
    window.scrollTo(0, 0)
  }, [])

  const conflictDiffs = data.diffs.filter(
    (d) => d.userModified && d.hasChanges && (d.userChanges?.length ?? 0) > 0,
  )
  const otherDiffs = data.diffs.filter(
    (d) => !(d.userModified && d.hasChanges && (d.userChanges?.length ?? 0) > 0),
  )

  return (
    <div className="p-6 pb-24">
      <h2 className="text-lg font-semibold text-gray-200 mb-1">Config conflicts detected</h2>
      <p className="text-sm text-gray-400 mb-4">
        You manually changed config files managed by kyaraben. Applying will overwrite those
        changes.
      </p>

      {conflictDiffs.length > 0 && (
        <div className="space-y-3 mb-6">
          <h3 className="text-sm font-medium text-amber-400">
            Files with conflicts ({conflictDiffs.length})
          </h3>
          {conflictDiffs.map((diff) => (
            <FileDiff key={diff.path} diff={diff} />
          ))}
        </div>
      )}

      {otherDiffs.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-sm font-medium text-gray-400">Other changes ({otherDiffs.length})</h3>
          {otherDiffs.map((diff) => (
            <FileDiff key={diff.path} diff={diff} />
          ))}
        </div>
      )}

      <BottomBar>
        <button
          type="button"
          onClick={onCancel}
          className="text-blue-400 hover:text-blue-300 hover:underline"
        >
          Cancel
        </button>
        <Button onClick={onConfirm}>Continue and override</Button>
      </BottomBar>
    </div>
  )
}
