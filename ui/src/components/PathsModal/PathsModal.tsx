import { Modal } from '@/lib/Modal'

export interface PathsModalProps {
  readonly open: boolean
  readonly onClose: () => void
  readonly emulatorName: string
  readonly emulatorId: string
  readonly userStore: string
}

export function PathsModal({
  open,
  onClose,
  emulatorName,
  emulatorId,
  userStore,
}: PathsModalProps) {
  const paths = [
    { label: 'Config', path: `~/.config/${emulatorId}/` },
    { label: 'Data', path: `${userStore}/data/${emulatorId}/` },
    { label: 'Saves', path: `${userStore}/saves/${emulatorId}/` },
  ]

  const handleOpenFolder = (path: string) => {
    // Expand ~ to home directory for the API
    const expandedPath = path.replace(/^~/, userStore.replace(/^~/, ''))
    window.electron.invoke('open_path', expandedPath)
  }

  return (
    <Modal open={open} onClose={onClose} title={`${emulatorName} Paths`}>
      <div className="space-y-4">
        {paths.map(({ label, path }) => (
          <div key={label}>
            <p className="text-sm text-gray-500 mb-1">{label}</p>
            <div className="flex items-center gap-2">
              <code className="flex-1 text-sm bg-gray-100 px-2 py-1.5 rounded text-gray-600 select-all truncate">
                {path}
              </code>
              <button
                type="button"
                onClick={() => handleOpenFolder(path)}
                className="px-3 py-1.5 text-sm bg-gray-200 hover:bg-gray-300 text-gray-700 rounded transition-colors flex-shrink-0"
              >
                Open
              </button>
            </div>
          </div>
        ))}
      </div>
    </Modal>
  )
}
