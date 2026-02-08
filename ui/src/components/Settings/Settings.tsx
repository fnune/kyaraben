import { useEffect, useState } from 'react'
import { IconButton } from '@/lib/IconButton'
import { Input } from '@/lib/Input'
import { FolderIcon } from '@/lib/icons'
import { useToast } from '@/lib/ToastContext'

export interface SettingsProps {
  readonly userStore: string
  readonly onUserStoreChange: (value: string) => void
}

export function Settings({ userStore, onUserStoreChange }: SettingsProps) {
  const [opening, setOpening] = useState(false)
  const [folderExists, setFolderExists] = useState(false)
  const { showToast } = useToast()

  useEffect(() => {
    if (!userStore) {
      setFolderExists(false)
      return
    }
    window.electron.invoke('path_exists', userStore).then((exists) => {
      setFolderExists(Boolean(exists))
    })
  }, [userStore])

  const handleOpenFolder = async () => {
    setOpening(true)
    try {
      const error = await window.electron.invoke('open_path', userStore)
      if (error) {
        showToast(`Could not open folder: ${error}`, 'error')
      } else {
        showToast(`Opening ${userStore}`)
      }
    } finally {
      setOpening(false)
    }
  }

  return (
    <div>
      <span
        id="user-store-label"
        className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block"
      >
        Emulation folder
      </span>
      <div className="mt-1 flex gap-2">
        <div className="flex-1">
          <Input value={userStore} onChange={onUserStoreChange} placeholder="~/Emulation" />
        </div>
        <IconButton
          icon={<FolderIcon className="w-5 h-5 text-on-surface-muted" />}
          label={folderExists ? 'Open folder' : 'Folder does not exist'}
          loading={opening}
          disabled={!folderExists}
          onClick={handleOpenFolder}
        />
      </div>
    </div>
  )
}
