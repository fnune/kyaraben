import { FolderIcon } from '@/lib/icons'
import { Modal } from '@/lib/Modal'
import type { EmulatorPaths, ManagedConfigInfo, ManagedRegionInfo } from '@/types/daemon'

export interface PathsModalProps {
  readonly open: boolean
  readonly onClose: () => void
  readonly emulatorName: string
  readonly paths: EmulatorPaths
  readonly managedConfigs?: readonly ManagedConfigInfo[]
}

function formatRegion(region: ManagedRegionInfo): string {
  if (region.type === 'file') {
    return 'entire file'
  }
  if (region.keyPrefix) {
    return `[${region.section}] ${region.keyPrefix}*`
  }
  return `[${region.section}]`
}

export function PathsModal({
  open,
  onClose,
  emulatorName,
  paths: emulatorPaths,
  managedConfigs,
}: PathsModalProps) {
  const paths: Array<{ label: string; path: string }> = [
    { label: 'ROMs', path: emulatorPaths.roms },
    ...(emulatorPaths.bios ? [{ label: 'BIOS', path: emulatorPaths.bios }] : []),
    ...(emulatorPaths.saves ? [{ label: 'Saves', path: emulatorPaths.saves }] : []),
    ...(emulatorPaths.states ? [{ label: 'Savestates', path: emulatorPaths.states }] : []),
    ...(emulatorPaths.screenshots
      ? [{ label: 'Screenshots', path: emulatorPaths.screenshots }]
      : []),
  ]

  const handleOpenFolder = (path: string) => {
    window.electron.invoke('open_path', path)
  }

  return (
    <Modal open={open} onClose={onClose} title={`${emulatorName} Paths`}>
      <div className="space-y-4">
        {paths.map(({ label, path }) => (
          <div key={label}>
            <p className="text-sm text-on-surface-muted mb-1">{label}</p>
            <div className="flex items-center gap-2">
              <code className="flex-1 text-sm bg-surface-raised px-2 py-1.5 rounded-sm text-on-surface-secondary select-all truncate">
                {path}
              </code>
              <button
                type="button"
                onClick={() => handleOpenFolder(path)}
                className="p-1.5 bg-outline hover:bg-outline-strong text-on-surface-secondary rounded-sm transition-colors shrink-0"
                aria-label={`Open ${label} folder`}
              >
                <FolderIcon />
              </button>
            </div>
          </div>
        ))}

        {managedConfigs && managedConfigs.length > 0 && (
          <div>
            <p className="text-sm text-on-surface-muted mb-1">Managed settings</p>
            <p className="text-xs text-on-surface-dim mb-2">
              These settings are controlled by kyaraben. Changing them may cause issues.
            </p>
            <div className="space-y-2">
              {managedConfigs.map((config) => (
                <div key={config.path}>
                  <code className="block text-xs text-on-surface-dim truncate">{config.path}</code>
                  {config.managedRegions && config.managedRegions.length > 0 && (
                    <div className="mt-0.5 ml-2 space-y-0.5">
                      {config.managedRegions.map((region, i) => (
                        <code
                          key={`${config.path}-${region.type}-${region.section ?? ''}-${i}`}
                          className="block text-xs text-on-surface-dim/70"
                        >
                          {formatRegion(region)}
                        </code>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}
