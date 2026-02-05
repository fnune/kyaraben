import { useEffect, useState } from 'react'
import { Button } from '@/lib/Button'
import {
  getInstallStatus,
  getStatus,
  getUninstallPreview,
  installApp,
  launchCliUninstall,
  openPath,
  readFile,
} from '@/lib/daemon'
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
  return (
    <li className={`font-mono text-xs ${bgColor} ${textColor} px-2 py-1 rounded-sm`}>{path}</li>
  )
}

function EmptyState({ message }: { message: string }) {
  return <p className="text-sm text-gray-500 italic">{message}</p>
}

function UninstallPendingOverlay() {
  return (
    <div className="fixed inset-0 bg-gray-900 flex items-center justify-center z-50">
      <div className="text-center max-w-md px-6">
        <div className="text-6xl mb-6">👋</div>
        <h1 className="text-2xl font-medium text-gray-100 mb-4">Ready to uninstall</h1>
        <p className="text-gray-400 mb-2">Kyaraben will uninstall when you close this window.</p>
        <p className="text-gray-500 text-sm">You'll receive a notification when it's done.</p>
      </div>
    </div>
  )
}

export function InstallationView() {
  const [preview, setPreview] = useState<UninstallPreviewResponse | null>(null)
  const [installStatus, setInstallStatus] = useState<InstallStatus | null>(null)
  const [healthWarning, setHealthWarning] = useState<string | null>(null)
  const [configContent, setConfigContent] = useState<string | null>(null)
  const [configPath, setConfigPath] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [installing, setInstalling] = useState(false)
  const [uninstallPending, setUninstallPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    setError(null)
    Promise.all([getUninstallPreview(), getInstallStatus(), getStatus()]).then(
      async ([previewResult, installResult, statusResult]) => {
        if (previewResult.ok) {
          setPreview(previewResult.data)
          const path = `${previewResult.data.preserved.configDir}/config.toml`
          setConfigPath(path)
          const contentResult = await readFile(path)
          if (contentResult.ok) {
            setConfigContent(contentResult.data)
          }
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

  const handleUninstall = async () => {
    if (
      !window.confirm(
        'Are you sure you want to uninstall Kyaraben? Your ROMs, saves, and configuration will be preserved.',
      )
    ) {
      return
    }
    const result = await launchCliUninstall()
    if (!result.ok) {
      setError(result.error.message)
    } else if (!result.data.success) {
      setError(result.data.error ?? 'Failed to launch uninstaller')
    } else {
      setUninstallPending(true)
    }
  }

  if (uninstallPending) {
    return <UninstallPendingOverlay />
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

      <Section title="Actions">
        <div className="space-y-4">
          {!installStatus?.installed && (
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-300">Install to PATH</p>
                <p className="text-xs text-gray-500">
                  Add Kyaraben to your applications menu and <code>$PATH</code>
                </p>
              </div>
              <Button onClick={handleInstall} disabled={installing}>
                {installing ? 'Installing...' : 'Install'}
              </Button>
            </div>
          )}
          {configPath && (
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-300">Edit configuration</p>
                <p className="text-xs text-gray-500">{configPath}</p>
              </div>
              <Button variant="secondary" onClick={() => openPath(configPath)}>
                Open
              </Button>
            </div>
          )}
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-300">Uninstall Kyaraben</p>
              <p className="text-xs text-gray-500">
                Remove all managed files (preserves ROMs and saves)
              </p>
            </div>
            <Button variant="secondary" onClick={handleUninstall}>
              Uninstall
            </Button>
          </div>
        </div>
      </Section>

      <Section title="Configuration">
        {configContent ? (
          <pre className="bg-gray-900 text-gray-300 text-xs font-mono p-3 rounded-sm overflow-x-auto max-h-64 overflow-y-auto">
            {configContent}
          </pre>
        ) : (
          <EmptyState message="Config file not found" />
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

      {installStatus?.installed && (
        <Section title="Kyaraben installation">
          <p className="text-sm text-green-400 mb-2">Installed</p>
          <ul className="space-y-1">
            {installStatus.appPath && <PathItem path={installStatus.appPath} />}
            {installStatus.cliPath && <PathItem path={installStatus.cliPath} />}
            {installStatus.desktopPath && <PathItem path={installStatus.desktopPath} />}
          </ul>
        </Section>
      )}

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

      {preview.retroArchCoresDir && (
        <Section title="RetroArch cores">
          <ul className="space-y-1">
            <PathItem path={preview.retroArchCoresDir} />
          </ul>
          {preview.retroArchCoreFiles && preview.retroArchCoreFiles.length > 0 && (
            <div className="mt-2 text-xs text-gray-400">
              {preview.retroArchCoreFiles.length} core
              {preview.retroArchCoreFiles.length !== 1 ? 's' : ''} installed:{' '}
              {preview.retroArchCoreFiles.map((f) => f.replace('_libretro.so', '')).join(', ')}
            </div>
          )}
        </Section>
      )}
    </div>
  )
}
