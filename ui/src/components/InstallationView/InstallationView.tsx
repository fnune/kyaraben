import { useEffect, useState } from 'react'
import { BugReport } from '@/components/BugReport/BugReport'
import { Button } from '@/lib/Button'
import type { UpdateInfo } from '@/lib/daemon'
import {
  applyUpdate,
  checkForUpdates,
  downloadUpdate,
  getInstallStatus,
  getStatus,
  getUninstallPreview,
  installApp,
  launchCliUninstall,
  openPath,
  readFile,
} from '@/lib/daemon'
import { PathText } from '@/lib/PathText'
import type { InstallStatus, UninstallPreviewResponse } from '@/types/daemon'
import { VIEW_CATALOG, VIEW_LABELS } from '@/types/ui'

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="p-4 bg-surface-alt rounded-card">
      <h3 className="text-sm font-medium text-on-surface mb-1">{title}</h3>
      {children}
    </div>
  )
}

function SectionIntro({ children }: { children: React.ReactNode }) {
  return <p className="text-sm text-on-surface-muted mb-1">{children}</p>
}

function PathItem({
  path,
  variant = 'default',
}: {
  path: string
  variant?: 'default' | 'preserved'
}) {
  const bgColor = variant === 'preserved' ? 'bg-status-ok/10' : 'bg-surface-raised'
  const textColor = variant === 'preserved' ? 'text-status-ok' : 'text-on-surface-secondary'
  return (
    <li className={`font-mono text-xs ${bgColor} ${textColor} px-2 py-1 rounded-sm`}>{path}</li>
  )
}

function EmptyState({ message }: { message: string }) {
  return <p className="text-sm text-on-surface-dim italic">{message}</p>
}

function UninstallPendingOverlay() {
  return (
    <div className="fixed inset-0 bg-surface flex items-center justify-center z-50">
      <div className="text-center max-w-md px-6">
        <div className="text-6xl mb-6">👋</div>
        <h1 className="text-2xl font-medium text-on-surface mb-4">Ready to uninstall</h1>
        <p className="text-on-surface-muted mb-2">
          Kyaraben will uninstall when you close this window.
        </p>
        <p className="text-on-surface-dim text-sm">You'll receive a notification when it's done.</p>
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
  const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null)
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [downloadingUpdate, setDownloadingUpdate] = useState(false)
  const [downloadProgress, setDownloadProgress] = useState(0)
  const [bugReportOpen, setBugReportOpen] = useState(false)

  useEffect(() => {
    return window.electron.on('update:progress', (data) => {
      setDownloadProgress(data.percent)
    })
  }, [])

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

  const handleCheckUpdate = async () => {
    setCheckingUpdate(true)
    setError(null)
    const result = await checkForUpdates()
    if (result.ok) {
      setUpdateInfo(result.data)
    } else {
      setError(result.error.message)
    }
    setCheckingUpdate(false)
  }

  const handleDownloadUpdate = async () => {
    if (!updateInfo?.downloadUrl) return

    setDownloadingUpdate(true)
    setDownloadProgress(0)
    setError(null)

    const downloadResult = await downloadUpdate(updateInfo.downloadUrl)
    if (!downloadResult.ok) {
      setError(downloadResult.error.message || 'Download failed')
      setDownloadingUpdate(false)
      return
    }
    if (!downloadResult.data.success || !downloadResult.data.path) {
      setError(downloadResult.data.error || 'Download failed')
      setDownloadingUpdate(false)
      return
    }

    const applyResult = await applyUpdate(downloadResult.data.path)
    if (!applyResult.ok) {
      setError(applyResult.error.message || 'Update failed')
      setDownloadingUpdate(false)
      return
    }
    if (!applyResult.data.success) {
      setError(applyResult.data.error || 'Update failed')
      setDownloadingUpdate(false)
    }
  }

  if (uninstallPending) {
    return <UninstallPendingOverlay />
  }

  if (loading) {
    return (
      <div className="p-6">
        <p className="text-on-surface-muted">Loading installation info...</p>
      </div>
    )
  }

  if (error || !preview) {
    return (
      <div className="p-6">
        <p className="text-status-error">
          Failed to load installation info: {error ?? 'Unknown error'}
        </p>
      </div>
    )
  }

  const desktopFiles = preview.desktopFiles ?? []
  const iconFiles = preview.iconFiles ?? []
  const configFiles = preview.configFiles ?? []

  return (
    <div className="p-6 space-y-6">
      {healthWarning === 'orphaned_artifacts' && (
        <div className="p-4 bg-status-error/10 border border-status-error/30 rounded-card">
          <h3 className="text-sm font-medium text-status-error mb-2">
            Installation state corrupted
          </h3>
          <p className="text-sm text-status-error/80 mb-3">
            Kyaraben found installation artifacts but the manifest tracking them is missing or
            empty. This can happen if files were manually deleted or corrupted.
          </p>
          <p className="text-sm text-status-error/80">
            To fix this, click Apply in the {VIEW_LABELS[VIEW_CATALOG]} view to restore the
            installation state. Please also consider{' '}
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

      <Section title="Updates">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-on-surface-secondary">Check for updates</p>
              <p className="text-xs text-on-surface-dim">
                {updateInfo
                  ? updateInfo.available
                    ? `New version available: ${updateInfo.latestVersion}`
                    : `You're on the latest version (${updateInfo.currentVersion})`
                  : 'Current version: 0.1.0'}
              </p>
            </div>
            <div className="flex gap-2">
              {updateInfo?.available && !downloadingUpdate && (
                <Button onClick={handleDownloadUpdate}>Update now</Button>
              )}
              <Button
                variant="secondary"
                onClick={handleCheckUpdate}
                disabled={checkingUpdate || downloadingUpdate}
              >
                {checkingUpdate ? 'Checking...' : 'Check'}
              </Button>
            </div>
          </div>
          {downloadingUpdate && (
            <div>
              <div className="h-1.5 bg-surface-raised rounded-full overflow-hidden">
                <div
                  className="h-full bg-accent transition-all duration-200"
                  style={{ width: `${downloadProgress}%` }}
                />
              </div>
              <p className="text-xs text-on-surface-muted mt-1">
                Downloading... {downloadProgress}%
              </p>
            </div>
          )}
        </div>
      </Section>

      <Section title="Actions">
        <div className="space-y-4">
          {!installStatus?.installed && (
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-on-surface-secondary">Add to applications menu</p>
                <p className="text-xs text-on-surface-dim">
                  Adds <pre className="inline">kyaraben</pre> and{' '}
                  <pre className="inline">kyaraben-ui</pre> to{' '}
                  <pre className="inline">~/.local/bin</pre>
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
                <p className="text-sm text-on-surface-secondary">Edit configuration</p>
                <p className="text-xs text-on-surface-dim">
                  <PathText>{configPath}</PathText>
                </p>
              </div>
              <Button variant="secondary" onClick={() => openPath(configPath)}>
                Open
              </Button>
            </div>
          )}
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-on-surface-secondary">Report a problem</p>
              <p className="text-xs text-on-surface-dim">
                Generate a bug report with system information
              </p>
            </div>
            <Button variant="secondary" onClick={() => setBugReportOpen(true)}>
              Report
            </Button>
          </div>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-on-surface-secondary">Uninstall Kyaraben</p>
              <p className="text-xs text-on-surface-dim">
                Remove all managed files (preserves your emulation folder)
              </p>
            </div>
            <Button variant="danger" onClick={handleUninstall}>
              Uninstall
            </Button>
          </div>
        </div>
      </Section>

      <BugReport open={bugReportOpen} onClose={() => setBugReportOpen(false)} />

      <Section title="Configuration">
        {configContent ? (
          <pre className="bg-surface text-on-surface-secondary text-xs font-mono p-3 rounded-sm overflow-x-auto max-h-64 overflow-y-auto">
            {configContent}
          </pre>
        ) : (
          <EmptyState message="Config file not found" />
        )}
      </Section>

      <Section title="Preserved on uninstall">
        <SectionIntro>These directories will not be removed when uninstalling:</SectionIntro>
        <ul className="space-y-1">
          <PathItem
            path={`${preview.preserved.userStore} (emulation folder)`}
            variant="preserved"
          />
          <PathItem path={`${preview.preserved.configDir} (config)`} variant="preserved" />
        </ul>
      </Section>

      {installStatus?.installed && (
        <Section title="Kyaraben installation">
          <p className="text-sm text-status-ok mb-2">Installed</p>
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
            <div className="mt-2 text-xs text-on-surface-muted">
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
