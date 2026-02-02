import { useEffect, useState } from 'react'
import { IconButton } from '@/lib/IconButton'
import { Input } from '@/lib/Input'

export interface SettingsProps {
  readonly userStore: string
  readonly onUserStoreChange: (value: string) => void
  readonly onError?: (message: string) => void
}

const FolderIcon = (
  <svg
    className="w-5 h-5 text-gray-600"
    fill="none"
    stroke="currentColor"
    viewBox="0 0 24 24"
    aria-hidden="true"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z"
    />
  </svg>
)

export function Settings({ userStore, onUserStoreChange, onError }: SettingsProps) {
  const [opening, setOpening] = useState(false)
  const [folderExists, setFolderExists] = useState(false)

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
        onError?.(`Could not open folder: ${error}`)
      }
    } finally {
      setOpening(false)
    }
  }

  return (
    <div className="p-4 bg-gray-50 rounded-lg mb-6">
      <span id="user-store-label" className="text-sm font-medium text-gray-700 block">
        Emulation folder
      </span>
      <div className="mt-1 flex gap-2">
        <div className="flex-1">
          <Input value={userStore} onChange={onUserStoreChange} placeholder="~/Emulation" />
        </div>
        <IconButton
          icon={FolderIcon}
          label={folderExists ? 'Open folder' : 'Folder does not exist'}
          loading={opening}
          disabled={!folderExists}
          onClick={handleOpenFolder}
        />
      </div>
    </div>
  )
}
