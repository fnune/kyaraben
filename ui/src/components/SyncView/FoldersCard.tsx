import type { SyncFolder, SyncLocalChange } from '@shared/daemon'
import { useCallback, useEffect, useState } from 'react'
import { formatBytes } from '@/lib/changeUtils'
import { getSyncLocalChanges, revertSyncFolder } from '@/lib/daemon'
import { useOnWindowFocus } from '@/lib/hooks/useOnWindowFocus'
import { useOpenPath } from '@/lib/hooks/useOpenPath'
import { FolderIcon } from '@/lib/icons'
import { LocalFilesActions } from './LocalFilesActions'

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

interface FolderRowProps {
  readonly folder: SyncFolder
  readonly onRefresh: () => void
  readonly hasPairedDevices: boolean
}

function FolderRow({ folder, onRefresh, hasPairedDevices }: FolderRowProps) {
  const [showChanges, setShowChanges] = useState(false)
  const [changes, setChanges] = useState<SyncLocalChange[] | null>(null)
  const [loadingChanges, setLoadingChanges] = useState(false)
  const [changesError, setChangesError] = useState<string | null>(null)
  const openPath = useOpenPath()

  const isSyncing = hasPairedDevices && (folder.state === 'syncing' || folder.needSize > 0)
  const hasLocalChanges = hasPairedDevices && folder.receiveOnlyChanges > 0
  const hasConflicts = (folder.conflictCount ?? 0) > 0
  const isReceiveOnly = folder.type === 'receiveonly'
  const sizeDiffers = hasPairedDevices && isReceiveOnly && folder.localSize !== folder.globalSize
  const percent =
    folder.globalSize > 0
      ? Math.round(((folder.globalSize - folder.needSize) / folder.globalSize) * 100)
      : 100

  const fetchChanges = useCallback(async () => {
    setLoadingChanges(true)
    setChangesError(null)
    const result = await getSyncLocalChanges({ folderId: folder.id })
    if (result.ok) {
      setChanges(result.data.changes)
    } else {
      setChangesError(result.error?.message ?? 'Failed to load changes')
    }
    setLoadingChanges(false)
  }, [folder.id])

  useEffect(() => {
    if (hasLocalChanges && changes === null) {
      fetchChanges()
    }
  }, [hasLocalChanges, changes, fetchChanges])

  useOnWindowFocus(
    useCallback(() => {
      if (showChanges) {
        fetchChanges()
      }
    }, [showChanges, fetchChanges]),
  )

  const handleShowChanges = useCallback(() => {
    setShowChanges(!showChanges)
  }, [showChanges])

  const handleRevert = useCallback(async () => {
    await revertSyncFolder({ folderId: folder.id })
    setShowChanges(false)
    setChanges(null)
    onRefresh()
  }, [folder.id, onRefresh])

  const isError = folder.state === 'error'

  const getStatusIndicator = () => {
    if (isError) return 'bg-status-error'
    if (hasConflicts) return 'bg-status-warning'
    if (hasLocalChanges) return 'bg-on-surface-muted'
    if (isSyncing) return 'bg-accent animate-pulse'
    return 'bg-status-ok'
  }

  return (
    <div className="py-2 border-b border-outline last:border-0">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <span className={`w-2 h-2 rounded-full flex-shrink-0 ${getStatusIndicator()}`} />
          <span
            className={`font-medium truncate ${isError ? 'text-status-error' : hasLocalChanges ? 'text-on-surface' : 'text-on-surface-muted'}`}
          >
            {folder.label}
          </span>
        </div>
        <div className="flex items-center gap-2 text-xs text-on-surface-muted flex-shrink-0">
          {isError ? (
            <span className="text-status-error">Error</span>
          ) : isSyncing ? (
            <span>
              {percent}% ({formatBytes(folder.needSize)} left)
            </span>
          ) : sizeDiffers ? (
            <span>
              {formatBytes(folder.localSize)} local / {formatBytes(folder.globalSize)} remote
            </span>
          ) : (
            <span>{formatBytes(folder.globalSize)}</span>
          )}
          <button
            type="button"
            onClick={() => openPath(folder.path)}
            className="p-1 text-on-surface-muted hover:text-on-surface-secondary rounded"
            title="Open folder"
          >
            <FolderIcon className="w-4 h-4" />
          </button>
        </div>
      </div>
      {isError && folder.error && (
        <div className="mt-2 ml-4 text-xs text-status-error">{folder.error}</div>
      )}
      {hasConflicts && (
        <div className="mt-2 ml-4 text-xs text-status-warning">
          {folder.conflictCount} conflict {folder.conflictCount === 1 ? 'file' : 'files'}
        </div>
      )}
      {hasLocalChanges && (
        <div className="mt-2 ml-4">
          <LocalFilesActions
            folderLabel={folder.label}
            count={folder.receiveOnlyChanges}
            changes={changes}
            changesLoading={loadingChanges}
            changesError={changesError}
            onRevert={handleRevert}
          />
          <button
            type="button"
            onClick={handleShowChanges}
            className="mt-1 text-xs text-accent hover:underline"
            disabled={loadingChanges}
          >
            {loadingChanges ? 'Loading...' : showChanges ? 'Hide details' : 'Show details'}
          </button>
          {showChanges && changes && changes.length > 0 && (
            <div className="mt-1 max-h-32 overflow-y-auto text-xs">
              {changes.map((c) => {
                const action = formatAction(c.action)
                const isDirectory = c.type === 'directory'
                const showSize = c.action !== 'deleted' && !isDirectory && c.size > 0
                return (
                  <div key={c.path} className="py-0.5 truncate">
                    {action && <span className={action.className}>{action.text}: </span>}
                    <span className="text-on-surface-muted">{c.path}</span>
                    {showSize && (
                      <span className="text-on-surface-muted"> ({formatBytes(c.size)})</span>
                    )}
                  </div>
                )
              })}
            </div>
          )}
          {showChanges && changes && changes.length === 0 && (
            <div className="mt-1 text-xs text-on-surface-muted">No details available</div>
          )}
          {showChanges && changesError && (
            <div className="mt-1 text-xs text-status-error">{changesError}</div>
          )}
        </div>
      )}
    </div>
  )
}

export interface FoldersCardProps {
  readonly folders: SyncFolder[] | undefined
  readonly onRefresh: () => void
  readonly hasPairedDevices: boolean
}

export function FoldersCard({ folders, onRefresh, hasPairedDevices }: FoldersCardProps) {
  const [isCollapsed, setIsCollapsed] = useState(true)

  const sortedFolders = folders ? [...folders].sort((a, b) => a.label.localeCompare(b.label)) : []
  const folderCount = sortedFolders.length
  const foldersWithLocalFiles = sortedFolders.filter(
    (f) => hasPairedDevices && f.receiveOnlyChanges > 0,
  )
  const localFilesCount = foldersWithLocalFiles.length

  if (folderCount === 0) {
    return null
  }

  return (
    <div className="p-4 bg-surface-alt rounded-card">
      <button
        type="button"
        onClick={() => setIsCollapsed(!isCollapsed)}
        className="flex items-center justify-between w-full text-left"
      >
        <div className="flex items-center gap-2">
          <span
            className={`transition-transform text-xs text-on-surface-muted ${isCollapsed ? '' : 'rotate-90'}`}
          >
            ▶
          </span>
          <h3 className="text-sm font-medium text-on-surface">Folders ({folderCount})</h3>
        </div>
        {isCollapsed && localFilesCount > 0 && (
          <span className="text-xs text-on-surface-muted">
            {localFilesCount} {localFilesCount === 1 ? 'has' : 'have'} local changes
          </span>
        )}
      </button>
      {!isCollapsed && (
        <div className="mt-3 border border-outline rounded-card px-3 bg-surface">
          {sortedFolders.map((folder) => (
            <FolderRow
              key={folder.id}
              folder={folder}
              onRefresh={onRefresh}
              hasPairedDevices={hasPairedDevices}
            />
          ))}
        </div>
      )}
    </div>
  )
}
