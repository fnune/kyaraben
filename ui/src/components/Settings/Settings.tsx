import { useEffect, useState } from 'react'
import { PathText } from '@/lib/PathText'
import { useToast } from '@/lib/ToastContext'
import { StorageSelector } from './StorageSelector'

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
        showToast(`Could not open folder: ${error}.`, 'error')
      } else {
        showToast(
          <span>
            Opening <PathText>{userStore}</PathText>.
          </span>,
        )
      }
    } finally {
      setOpening(false)
    }
  }

  return (
    <StorageSelector
      userStore={userStore}
      onUserStoreChange={onUserStoreChange}
      onOpenFolder={handleOpenFolder}
      folderExists={folderExists}
      opening={opening}
    />
  )
}
