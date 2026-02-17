import { useCallback, useEffect, useState } from 'react'
import { formatBytes } from '@/lib/changeUtils'
import { getStorageDevices, selectDirectory } from '@/lib/daemon'
import { IconButton } from '@/lib/IconButton'
import { Input } from '@/lib/Input'
import { FolderIcon } from '@/lib/icons'
import { collapseTilde, expandTilde } from '@/lib/paths'
import { RadioCard } from '@/lib/RadioCard'
import type { StorageDevice } from '@/types/daemon'

export interface StorageSelectorProps {
  readonly userStore: string
  readonly onUserStoreChange: (value: string) => void
  readonly onOpenFolder: () => void
  readonly folderExists: boolean
  readonly opening: boolean
}

function formatStorageDescription(device: StorageDevice): string {
  const free = formatBytes(device.freeBytes)
  const total = formatBytes(device.totalBytes)
  return `${free} free of ${total}`
}

export function StorageSelector({
  userStore,
  onUserStoreChange,
  onOpenFolder,
  folderExists,
  opening,
}: StorageSelectorProps) {
  const [devices, setDevices] = useState<StorageDevice[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [homeDir, setHomeDir] = useState<string>('')

  useEffect(() => {
    setLoading(true)
    getStorageDevices().then((result) => {
      if (result.ok) {
        setDevices(result.data.devices)

        const internal = result.data.devices.find((d) => d.id === 'internal')
        const home = internal ? internal.path.replace(/\/Emulation$/, '') : ''
        setHomeDir(home)

        const expandedUserStore = expandTilde(userStore, home)
        const matching = result.data.devices.find((d) => d.path === expandedUserStore)
        if (matching) {
          setSelectedId(matching.id)
        } else if (userStore) {
          setSelectedId('custom')
        }
      }
      setLoading(false)
    })
  }, [userStore])

  const handleSelect = useCallback(
    (device: StorageDevice) => {
      setSelectedId(device.id)
      onUserStoreChange(collapseTilde(device.path, homeDir))
    },
    [onUserStoreChange, homeDir],
  )

  const handleCustomSelect = useCallback(async () => {
    const result = await selectDirectory()
    if (result.ok && result.data) {
      setSelectedId('custom')
      onUserStoreChange(collapseTilde(result.data, homeDir))
    }
  }, [onUserStoreChange, homeDir])

  if (loading) {
    return (
      <div>
        <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block">
          Emulation folder
        </span>
        <div className="mt-2 text-on-surface-muted">Detecting storage...</div>
      </div>
    )
  }

  return (
    <div>
      <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block">
        Emulation folder
      </span>

      <div className="mt-2 grid grid-cols-3 gap-2">
        {devices.map((device) => (
          <RadioCard
            key={device.id}
            title={device.label}
            description={formatStorageDescription(device)}
            selected={selectedId === device.id}
            onSelect={() => handleSelect(device)}
            className="flex-1 min-w-0 p-3"
          />
        ))}
        <RadioCard
          title="Custom folder"
          description="Choose location"
          selected={selectedId === 'custom'}
          onSelect={handleCustomSelect}
          className="flex-1 min-w-0 p-3"
        />
        {devices.length === 1 && <div className="flex-1" />}
      </div>

      <p className="mt-2 text-xs text-on-surface-muted">
        Changing storage will not move existing files. Edit the path below to customize the folder
        name.
      </p>

      <div className="mt-2 flex gap-2">
        <div className="flex-1">
          <Input value={userStore} onChange={onUserStoreChange} placeholder="~/Emulation" />
        </div>
        <IconButton
          icon={<FolderIcon className="w-5 h-5 text-on-surface-muted" />}
          label={folderExists ? 'Open folder' : 'Folder does not exist'}
          loading={opening}
          disabled={!folderExists}
          onClick={onOpenFolder}
        />
      </div>
    </div>
  )
}
