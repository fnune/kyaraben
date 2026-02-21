import { useCallback, useState } from 'react'
import { Button } from '@/lib/Button'
import { formatBytes } from '@/lib/changeUtils'
import { Modal } from '@/lib/Modal'
import type { SyncLocalChange } from '@/types/daemon'

function formatAction(action: string): { text: string; className: string } | null {
  switch (action) {
    case 'changed':
      return { text: 'Added or changed locally', className: 'text-status-ok' }
    case 'deleted':
      return { text: 'Deleted locally', className: 'text-on-surface-muted' }
    case '':
      return null
    default:
      return {
        text: action.charAt(0).toUpperCase() + action.slice(1),
        className: 'text-on-surface-muted',
      }
  }
}

export interface LocalFilesActionsProps {
  readonly folderLabel: string
  readonly count: number
  readonly changes: SyncLocalChange[] | null
  readonly changesLoading: boolean
  readonly changesError: string | null
  readonly onRevert: () => Promise<void>
}

export function LocalFilesActions({
  folderLabel,
  count,
  changes,
  changesLoading,
  changesError,
  onRevert,
}: LocalFilesActionsProps) {
  const [showCopyModal, setShowCopyModal] = useState(false)
  const [showRevertModal, setShowRevertModal] = useState(false)
  const [isReverting, setIsReverting] = useState(false)

  const hasNonDeletedChanges = changes?.some((c) => c.action === 'changed') ?? true

  const handleConfirmRevert = useCallback(async () => {
    setIsReverting(true)
    try {
      await onRevert()
      setShowRevertModal(false)
    } finally {
      setIsReverting(false)
    }
  }, [onRevert])

  return (
    <>
      <div className="flex items-center justify-between text-xs">
        <span className="text-on-surface-muted">
          {count} local change{count === 1 ? '' : 's'}
        </span>
        <div className="flex gap-2">
          {hasNonDeletedChanges && (
            <button
              type="button"
              onClick={() => setShowCopyModal(true)}
              className="text-accent hover:underline"
            >
              Copy to primary...
            </button>
          )}
          <button
            type="button"
            onClick={() => setShowRevertModal(true)}
            className="text-accent hover:underline"
          >
            Revert...
          </button>
        </div>
      </div>

      <Modal
        open={showCopyModal}
        onClose={() => setShowCopyModal(false)}
        title="Copy files to primary"
      >
        <div className="space-y-4">
          <p className="text-sm text-on-surface-muted">
            To keep these files synced across devices, add them on your primary device:
          </p>
          <ol className="list-decimal list-inside text-sm text-on-surface-muted space-y-2">
            <li>Open the folder on your primary device</li>
            <li>Copy the files you want to keep into the same location</li>
            <li>The files will synchronize back to this device automatically</li>
          </ol>
          <p className="text-sm text-on-surface-muted">
            Files added on secondary devices are not synced to avoid accidental data loss.
          </p>
          <div className="flex justify-end">
            <Button variant="secondary" onClick={() => setShowCopyModal(false)}>
              Close
            </Button>
          </div>
        </div>
      </Modal>

      <Modal
        open={showRevertModal}
        onClose={() => setShowRevertModal(false)}
        title={`Revert local changes in ${folderLabel}`}
      >
        <div className="space-y-4">
          <p className="text-sm text-on-surface-muted">
            This will undo all local changes to match the primary device:
          </p>
          <ul className="list-disc list-inside text-sm text-on-surface-muted space-y-1">
            <li>Added or changed files will be restored to their original version</li>
            <li>Deleted files will be re-downloaded</li>
          </ul>

          {changesLoading && <p className="text-sm text-on-surface-muted">Loading changes...</p>}

          {!changesLoading && changesError && (
            <p className="text-sm text-status-error">{changesError}</p>
          )}

          {!changesLoading && changes && changes.length > 0 && (
            <div className="max-h-48 overflow-y-auto border border-outline rounded p-2 bg-surface">
              {changes.map((c) => {
                const action = formatAction(c.action)
                const isDirectory = c.type === 'directory'
                const showSize = c.action !== 'deleted' && !isDirectory && c.size > 0
                return (
                  <div key={c.path} className="py-0.5 text-xs truncate">
                    {action && <span className={action.className}>{action.text}: </span>}
                    <span className="text-on-surface-muted">{c.path}</span>
                    {showSize && (
                      <span className="text-on-surface-muted ml-1">({formatBytes(c.size)})</span>
                    )}
                  </div>
                )
              })}
            </div>
          )}

          {changes && changes.length === 0 && (
            <p className="text-sm text-on-surface-muted">No details available.</p>
          )}

          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setShowRevertModal(false)}>
              Cancel
            </Button>
            <Button variant="danger" onClick={handleConfirmRevert} disabled={isReverting}>
              {isReverting ? 'Reverting...' : 'Revert changes'}
            </Button>
          </div>
        </div>
      </Modal>
    </>
  )
}
