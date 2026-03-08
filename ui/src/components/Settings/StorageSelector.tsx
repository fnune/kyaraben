import type { StorageDevice } from '@shared/daemon'
import { useCallback, useEffect, useState } from 'react'
import { formatBytes } from '@/lib/changeUtils'
import { getStorageDevices, selectDirectory } from '@/lib/daemon'
import { IconButton } from '@/lib/IconButton'
import { Input } from '@/lib/Input'
import { FolderIcon } from '@/lib/icons'
import { collapseTilde, expandTilde } from '@/lib/paths'
import { RadioCard } from '@/lib/RadioCard'

export interface StorageSelectorProps {
  readonly collection: string
  readonly onCollectionChange: (value: string) => void
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
  collection,
  onCollectionChange,
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
        setHomeDir(internal ? internal.path.replace(/\/Emulation$/, '') : '')
      }
      setLoading(false)
    })
  }, [])

  useEffect(() => {
    if (!homeDir || devices.length === 0) return
    const expandedCollection = expandTilde(collection, homeDir)
    const matching = devices.find((d) => d.path === expandedCollection)
    if (matching) {
      setSelectedId(matching.id)
    } else if (collection) {
      setSelectedId('custom')
    }
  }, [collection, homeDir, devices])

  const handleSelect = useCallback(
    (device: StorageDevice) => {
      setSelectedId(device.id)
      onCollectionChange(collapseTilde(device.path, homeDir))
    },
    [onCollectionChange, homeDir],
  )

  const handleCustomSelect = useCallback(async () => {
    const result = await selectDirectory()
    if (result.ok && result.data) {
      setSelectedId('custom')
      onCollectionChange(collapseTilde(result.data, homeDir))
    }
  }, [onCollectionChange, homeDir])

  if (loading) {
    return (
      <div>
        <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block">
          Collection
        </span>
        <div className="mt-2 text-on-surface-muted">Detecting storage...</div>
      </div>
    )
  }

  return (
    <div>
      <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block">
        Collection
      </span>

      <div className="mt-2 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2">
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
          <Input value={collection} onChange={onCollectionChange} placeholder="~/Emulation" />
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
