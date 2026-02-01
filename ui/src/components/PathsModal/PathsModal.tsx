import { Modal } from '@/lib/Modal'

export interface PathsModalProps {
  readonly open: boolean
  readonly onClose: () => void
  readonly emulatorName: string
  readonly emulatorId: string
  readonly userStore: string
  readonly managedConfigs?: readonly string[]
}

export function PathsModal({
  open,
  onClose,
  emulatorName,
  emulatorId,
  userStore,
  managedConfigs,
}: PathsModalProps) {
  const paths = [
    { label: 'Config', path: `~/.config/${emulatorId}/` },
    { label: 'Data', path: `${userStore}/data/${emulatorId}/` },
    { label: 'Saves', path: `${userStore}/saves/${emulatorId}/` },
  ]

  const handleOpenFolder = (path: string) => {
    const expandedPath = path.replace(/^~/, userStore.replace(/^~/, ''))
    window.electron.invoke('open_path', expandedPath)
  }

  return (
    <Modal open={open} onClose={onClose} title={`${emulatorName} Paths`}>
      <div className="space-y-4">
        {paths.map(({ label, path }) => (
          <div key={label}>
            <p className="text-sm text-gray-400 mb-1">{label}</p>
            <div className="flex items-center gap-2">
              <code className="flex-1 text-sm bg-gray-700 px-2 py-1.5 rounded text-gray-300 select-all truncate">
                {path}
              </code>
              <button
                type="button"
                onClick={() => handleOpenFolder(path)}
                className="px-3 py-1.5 text-sm bg-gray-600 hover:bg-gray-500 text-gray-300 rounded transition-colors flex-shrink-0"
              >
                Open
              </button>
            </div>
          </div>
        ))}

        {managedConfigs && managedConfigs.length > 0 && (
          <div>
            <p className="text-sm text-gray-400 mb-1">Managed configs</p>
            <div className="space-y-1.5">
              {managedConfigs.map((configPath) => (
                <code
                  key={configPath}
                  className="block text-xs bg-gray-700 px-2 py-1.5 rounded text-gray-400 select-all truncate"
                >
                  {configPath}
                </code>
              ))}
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}
