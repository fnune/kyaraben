import { useEffect, useState } from 'react'
import { PathText } from '@/lib/PathText'
import { useToast } from '@/lib/ToastContext'
import { StorageSelector } from './StorageSelector'

export interface SettingsProps {
  readonly collection: string
  readonly onCollectionChange: (value: string) => void
}

export function Settings({ collection, onCollectionChange }: SettingsProps) {
  const [opening, setOpening] = useState(false)
  const [folderExists, setFolderExists] = useState(false)
  const { showToast } = useToast()

  useEffect(() => {
    if (!collection) {
      setFolderExists(false)
      return
    }
    window.electron.invoke('path_exists', collection).then((exists) => {
      setFolderExists(Boolean(exists))
    })
  }, [collection])

  const handleOpenFolder = async () => {
    setOpening(true)
    try {
      const error = await window.electron.invoke('open_path', collection)
      if (error) {
        showToast(`Could not open folder: ${error}.`, 'error')
      } else {
        showToast(
          <span>
            Opening <PathText>{collection}</PathText>.
          </span>,
        )
      }
    } finally {
      setOpening(false)
    }
  }

  return (
    <StorageSelector
      collection={collection}
      onCollectionChange={onCollectionChange}
      onOpenFolder={handleOpenFolder}
      folderExists={folderExists}
      opening={opening}
    />
  )
}
