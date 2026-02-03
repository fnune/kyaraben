import { useEffect, useState } from 'react'
import { Button } from '@/lib/Button'
import { getInstallStatus, getStatus, getUninstallPreview, installApp } from '@/lib/daemon'
import type { InstallStatus, UninstallPreviewResponse } from '@/types/daemon'

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="p-4 bg-gray-800 rounded-lg">
      <h3 className="text-sm font-medium text-gray-100 mb-3">{title}</h3>
      {children}
    </div>
  )
}

function PathItem({
  path,
  variant = 'default',
}: {
  path: string
  variant?: 'default' | 'preserved'
}) {
  const bgColor = variant === 'preserved' ? 'bg-green-500/10' : 'bg-gray-700'
  const textColor = variant === 'preserved' ? 'text-green-300' : 'text-gray-300'
  return <li className={`font-mono text-xs ${bgColor} ${textColor} px-2 py-1 rounded`}>{path}</li>
}

function EmptyState({ message }: { message: string }) {
  return <p className="text-sm text-gray-500 italic">{message}</p>
}

export function InstallationView() {
  const [preview, setPreview] = useState<UninstallPreviewResponse | null>(null)
  const [installStatus, setInstallStatus] = useState<InstallStatus | null>(null)
  const [healthWarning, setHealthWarning] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [installing, setInstalling] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    setError(null)
    Promise.all([getUninstallPreview(), getInstallStatus(), getStatus()]).then(
      ([previewResult, installResult, statusResult]) => {
        if (previewResult.ok) {
          setPreview(previewResult.data)
        } else {
          setError(previewResult.error.message)
        }
        if (installResult.ok) {
          setInstallStatus(installResult.data)
        }
        if (statusResult.ok && statusResult.data.healthWarning) {
          setHealthWarning(statusResult.data.healthWarning)
        }
        setLoading(false)
      },
    )
  }, [])

  const handleInstall = async () => {
    setInstalling(true)
    const result = await installApp()
    if (result.ok) {
      const statusResult = await getInstallStatus()
      if (statusResult.ok) {
        setInstallStatus(statusResult.data)
      }
    }
    setInstalling(false)
  }

  if (loading) {
    return (
      <div className="p-6">
        <p className="text-gray-400">Loading installation info...</p>
      </div>
    )
  }

  if (error || !preview) {
    return (
      <div className="p-6">
        <p className="text-red-400">Failed to load installation info: {error ?? 'Unknown error'}</p>
      </div>
    )
  }

  const desktopFiles = preview.desktopFiles ?? []
  const iconFiles = preview.iconFiles ?? []
  const configFiles = preview.configFiles ?? []

  return (
    <div className="p-6 space-y-6">
      {healthWarning === 'orphaned_artifacts' && (
        <div className="p-4 bg-red-900/30 border border-red-700/50 rounded-lg">
          <h3 className="text-sm font-medium text-red-300 mb-2">Installation state corrupted</h3>
          <p className="text-sm text-red-200/80 mb-3">
            Kyaraben found installation artifacts but the manifest tracking them is missing or
            empty. This can happen if files were manually deleted or corrupted.
          </p>
          <p className="text-sm text-red-200/80">
            To fix this, click Apply in the Systems tab to restore the installation state. Please
            also consider{' '}
            <a
              href="https://github.com/fnune/kyaraben/issues"
              target="_blank"
              rel="noopener noreferrer"
              className="underline hover:no-underline"
            >
              reporting this issue
            </a>
            .
          </p>
        </div>
      )}

      <Section title="Kyaraben">
        {installStatus?.installed ? (
          <div className="space-y-2">
            <p className="text-sm text-green-400">Installed</p>
            <ul className="space-y-1">
              {installStatus.appPath && <PathItem path={installStatus.appPath} />}
              {installStatus.cliPath && <PathItem path={installStatus.cliPath} />}
              {installStatus.desktopPath && <PathItem path={installStatus.desktopPath} />}
            </ul>
          </div>
        ) : (
          <div className="space-y-3">
            <p className="text-sm text-gray-400">
              Install Kyaraben to your applications menu and add the CLI to your{' '}
              <code className="text-gray-300">$PATH</code>.
            </p>
            <Button onClick={handleInstall} disabled={installing}>
              {installing ? 'Installing...' : 'Install'}
            </Button>
          </div>
        )}
      </Section>

      <Section title="State directory">
        {preview.stateDirExists ? (
          <ul className="space-y-1">
            <PathItem path={preview.stateDir} />
          </ul>
        ) : (
          <EmptyState message="Not created yet" />
        )}
      </Section>

      <Section title="Desktop files">
        {desktopFiles.length > 0 ? (
          <ul className="space-y-1">
            {desktopFiles.map((f) => (
              <PathItem key={f} path={f} />
            ))}
          </ul>
        ) : (
          <EmptyState message="No desktop files installed" />
        )}
      </Section>

      <Section title="Icons">
        {iconFiles.length > 0 ? (
          <ul className="space-y-1">
            {iconFiles.map((f) => (
              <PathItem key={f} path={f} />
            ))}
          </ul>
        ) : (
          <EmptyState message="No icons installed" />
        )}
      </Section>

      <Section title="Managed config files">
        {configFiles.length > 0 ? (
          <ul className="space-y-1">
            {configFiles.map((f) => (
              <PathItem key={f} path={f} />
            ))}
          </ul>
        ) : (
          <EmptyState message="No config files managed" />
        )}
      </Section>

      <Section title="Preserved on uninstall">
        <p className="text-sm text-gray-400 mb-2">
          These directories will not be removed when uninstalling:
        </p>
        <ul className="space-y-1">
          <PathItem
            path={`${preview.preserved.userStore} (ROMs, saves, BIOS)`}
            variant="preserved"
          />
          <PathItem path={`${preview.preserved.configDir} (config)`} variant="preserved" />
        </ul>
      </Section>

      <div className="border-t border-gray-700 pt-6">
        <h3 className="text-sm font-medium text-gray-100 mb-2">Uninstall</h3>
        <p className="text-sm text-gray-400 mb-3">
          To remove Kyaraben and all managed files (except preserved data), run:
        </p>
        <code className="block bg-gray-700 text-gray-300 px-3 py-2 rounded text-sm font-mono">
          {preview.stateDir}/bin/kyaraben uninstall
        </code>
      </div>
    </div>
  )
}
