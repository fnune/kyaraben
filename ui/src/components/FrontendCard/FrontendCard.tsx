import type { FrontendRef } from '@shared/daemon.gen'
import { getFrontendLogo } from '@/components/FrontendLogo/FrontendLogo'
import {
  PackageCard,
  PackageCardHeader,
  useChangeType,
} from '@/components/PackageCardHeader/PackageCardHeader'
import { formatBytes } from '@/lib/changeUtils'
import { launchEmulator } from '@/lib/daemon'
import { useToast } from '@/lib/ToastContext'

export interface FrontendCardProps {
  readonly frontend: FrontendRef
  readonly enabled: boolean
  readonly selectedVersion: string
  readonly installedVersion: string | null
  readonly execLine?: string | undefined
  readonly onToggle: (enabled: boolean) => void
  readonly onVersionChange: (version: string) => void
}

export function FrontendCard({
  frontend,
  enabled,
  selectedVersion,
  installedVersion,
  execLine,
  onToggle,
  onVersionChange,
}: FrontendCardProps) {
  const { showToast } = useToast()
  const changeType = useChangeType(
    enabled,
    installedVersion,
    selectedVersion,
    frontend.defaultVersion,
    frontend.availableVersions,
  )

  const handleLaunch = () => {
    if (execLine) {
      launchEmulator(execLine)
      showToast(`Launching ${frontend.name}.`)
    }
  }

  const logo = getFrontendLogo(frontend.id)
  const logoElement = logo ? (
    <img src={logo} alt="" className="w-full h-full object-contain" />
  ) : undefined

  const secondaryContent = (
    <div className="flex items-center gap-2 text-xs text-on-surface-muted">
      {installedVersion
        ? execLine && (
            <button
              type="button"
              onClick={handleLaunch}
              disabled={!enabled}
              className={enabled ? 'hover:text-accent' : 'cursor-not-allowed'}
            >
              Launch
            </button>
          )
        : frontend.downloadBytes && (
            <span className="text-on-surface-dim">{formatBytes(frontend.downloadBytes)}</span>
          )}
    </div>
  )

  return (
    <PackageCard changeType={changeType} installed={!!installedVersion} enabled={enabled}>
      <PackageCardHeader
        name={frontend.name}
        logo={logoElement}
        defaultVersion={frontend.defaultVersion}
        availableVersions={frontend.availableVersions}
        selectedVersion={selectedVersion}
        enabled={enabled}
        onToggle={onToggle}
        onVersionChange={onVersionChange}
        secondaryContent={secondaryContent}
      />
    </PackageCard>
  )
}
