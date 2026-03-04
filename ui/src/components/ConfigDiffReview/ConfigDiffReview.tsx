import { useOpenPath } from '@/lib/hooks/useOpenPath'
import { PathText } from '@/lib/PathText'
import { useScrollToTop } from '@/lib/useScrollToTop'
import type {
  ConfigChangeDetail,
  ConfigFileDiff,
  KyarabenUpdateDetail,
  ManagedRegionInfo,
  PreflightResponse,
} from '@/types/daemon'

export interface ConfigDiffReviewProps {
  readonly data: PreflightResponse
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

function isWholeFile(regions?: readonly ManagedRegionInfo[]): boolean {
  return regions?.some((r) => r.type === 'file') ?? false
}

function hasOnlyEmptyKeys(changes?: readonly { key: string }[]): boolean {
  return changes?.every((c) => c.key === '') ?? false
}

function conflictMessage(regions?: readonly ManagedRegionInfo[]): string {
  if (isWholeFile(regions)) {
    return 'This entire file is managed by Kyaraben and will be rewritten:'
  }
  return 'You modified settings managed by Kyaraben (will be overwritten):'
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

function KyarabenUpdateRow({ update }: { readonly update: KyarabenUpdateDetail }) {
  return (
    <div className="font-mono text-xs text-status-ok/80 pl-2">
      {update.key}: {update.oldValue} &rarr; {update.newValue}
    </div>
  )
}

function FileDiff({ diff }: { readonly diff: ConfigFileDiff }) {
  const openPath = useOpenPath()
  const hasUserConflict =
    diff.userModified && diff.hasChanges && (diff.userChanges?.length ?? 0) > 0
  const hasKyarabenUpdates = diff.kyarabenChanged && (diff.kyarabenUpdates?.length ?? 0) > 0

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
          <span className="text-sm text-on-surface-secondary truncate">
            <PathText>{diff.path}</PathText>
          </span>
        </div>
        <button
          type="button"
          onClick={() => openPath(diff.path)}
          className="text-xs text-accent hover:text-accent-hover hover:underline whitespace-nowrap"
        >
          Open file
        </button>
      </div>

      {hasKyarabenUpdates && diff.kyarabenUpdates && (
        <div className="bg-status-ok/10 border border-status-ok/30 rounded px-3 py-2">
          <p className="text-xs text-status-ok font-medium mb-1">
            Kyaraben has new defaults for these settings:
          </p>
          {diff.kyarabenUpdates.map((ku) => (
            <KyarabenUpdateRow key={ku.key} update={ku} />
          ))}
        </div>
      )}

      {hasUserConflict && diff.userChanges && (
        <div className="bg-status-warning/10 border border-status-warning/30 rounded px-3 py-2">
          <p className="text-xs text-status-warning font-medium mb-1">
            {conflictMessage(diff.managedRegions)}
          </p>
          {isWholeFile(diff.managedRegions) || hasOnlyEmptyKeys(diff.userChanges) ? (
            <p className="text-xs text-status-warning/80 pl-2 italic">File content was modified</p>
          ) : (
            diff.userChanges.map((uc) => (
              <div key={uc.key} className="font-mono text-xs text-status-warning/80 pl-2">
                {uc.key}: {uc.writtenValue} &rarr; {uc.currentValue}
              </div>
            ))
          )}
        </div>
      )}

      {!hasUserConflict && !hasKyarabenUpdates && diff.changes && diff.changes.length > 0 && (
        <div className="space-y-0.5">
          {diff.changes.map((change, i) => (
            <ChangeRow key={`${change.key}-${i}`} change={change} />
          ))}
        </div>
      )}
    </div>
  )
}

export function ConfigDiffReview({ data }: ConfigDiffReviewProps) {
  useScrollToTop()

  const userConflictDiffs = data.diffs.filter(
    (d) => d.userModified && d.hasChanges && (d.userChanges?.length ?? 0) > 0,
  )
  const kyarabenUpdateDiffs = data.diffs.filter(
    (d) =>
      d.kyarabenChanged &&
      (d.kyarabenUpdates?.length ?? 0) > 0 &&
      !(d.userModified && d.hasChanges && (d.userChanges?.length ?? 0) > 0),
  )
  const otherDiffs = data.diffs.filter(
    (d) =>
      !(d.userModified && d.hasChanges && (d.userChanges?.length ?? 0) > 0) &&
      !(d.kyarabenChanged && (d.kyarabenUpdates?.length ?? 0) > 0),
  )

  const hasUserConflicts = userConflictDiffs.length > 0

  const title = hasUserConflicts ? 'Config conflicts detected' : 'Kyaraben has updated its defaults'

  const description = hasUserConflicts
    ? 'You manually changed config files managed by Kyaraben. Applying will overwrite those changes.'
    : 'Kyaraben has new default values for some settings. Review the changes below.'

  return (
    <div className="p-6 pb-24">
      <h2 className="text-lg font-semibold text-on-surface mb-1">{title}</h2>
      <p className="text-sm text-on-surface-muted mb-4">{description}</p>

      {userConflictDiffs.length > 0 && (
        <div className="space-y-3 mb-6">
          <h3 className="text-sm font-medium text-status-warning">
            Your changes (will be overwritten) ({userConflictDiffs.length})
          </h3>
          {userConflictDiffs.map((diff) => (
            <FileDiff key={diff.path} diff={diff} />
          ))}
        </div>
      )}

      {kyarabenUpdateDiffs.length > 0 && (
        <div className="space-y-3 mb-6">
          <h3 className="text-sm font-medium text-status-ok">
            Kyaraben updates ({kyarabenUpdateDiffs.length})
          </h3>
          {kyarabenUpdateDiffs.map((diff) => (
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
    </div>
  )
}
