import { FolderIcon } from '@/lib/icons'
import { Modal } from '@/lib/Modal'
import type { EmulatorPaths, ManagedConfigInfo } from '@/types/daemon'

export interface PathsModalProps {
  readonly open: boolean
  readonly onClose: () => void
  readonly emulatorName: string
  readonly paths: EmulatorPaths
  readonly managedConfigs?: readonly ManagedConfigInfo[]
}

export function PathsModal({
  open,
  onClose,
  emulatorName,
  paths: emulatorPaths,
  managedConfigs,
}: PathsModalProps) {
  const paths = [
    { label: 'ROMs', path: emulatorPaths.roms },
    { label: 'BIOS', path: emulatorPaths.bios },
    { label: 'Saves', path: emulatorPaths.saves },
    { label: 'Savestates', path: emulatorPaths.states },
    { label: 'Screenshots', path: emulatorPaths.screenshots },
    ...(emulatorPaths.opaque ? [{ label: 'Emulator data', path: emulatorPaths.opaque }] : []),
  ]

  const handleOpenFolder = (path: string) => {
    window.electron.invoke('open_path', path)
  }

  return (
    <Modal open={open} onClose={onClose} title={`${emulatorName} Paths`}>
      <div className="space-y-4">
        {paths.map(({ label, path }) => (
          <div key={label}>
            <p className="text-sm text-gray-400 mb-1">{label}</p>
            <div className="flex items-center gap-2">
              <code className="flex-1 text-sm bg-gray-700 px-2 py-1.5 rounded-sm text-gray-300 select-all truncate">
                {path}
              </code>
              <button
                type="button"
                onClick={() => handleOpenFolder(path)}
                className="p-1.5 bg-gray-600 hover:bg-gray-500 text-gray-300 rounded-sm transition-colors shrink-0"
                aria-label={`Open ${label} folder`}
              >
                <FolderIcon />
              </button>
            </div>
          </div>
        ))}

        {managedConfigs && managedConfigs.length > 0 && (
          <div>
            <p className="text-sm text-gray-400 mb-1">Managed settings</p>
            <p className="text-xs text-gray-500 mb-2">
              These settings are controlled by kyaraben. Changing them may cause issues.
            </p>
            <div className="space-y-3">
              {managedConfigs.map((config) => (
                <div key={config.path}>
                  <code className="block text-xs text-gray-500 mb-1 truncate">{config.path}</code>
                  <div className="bg-gray-700 rounded-sm px-2 py-1.5 space-y-0.5">
                    {config.keys.map((key) => (
                      <div key={key.key} className="flex text-xs gap-2">
                        <span className="text-gray-400 shrink-0">{key.key}</span>
                        <span className="text-gray-500">=</span>
                        <span className="text-gray-300 truncate">{key.value}</span>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}
