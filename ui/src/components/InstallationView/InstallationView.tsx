import { useEffect, useState } from 'react'
import { getUninstallPreview } from '@/lib/daemon'
import type { UninstallPreviewResponse } from '@/types/daemon'

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="p-4 bg-gray-50 rounded-lg">
      <h3 className="text-sm font-medium text-gray-900 mb-3">{title}</h3>
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
  const bgColor = variant === 'preserved' ? 'bg-green-50' : 'bg-gray-100'
  return <li className={`font-mono text-xs ${bgColor} px-2 py-1 rounded`}>{path}</li>
}

function EmptyState({ message }: { message: string }) {
  return <p className="text-sm text-gray-500 italic">{message}</p>
}

export function InstallationView() {
  const [preview, setPreview] = useState<UninstallPreviewResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    setError(null)
    getUninstallPreview().then((result) => {
      if (result.ok) {
        setPreview(result.data)
      } else {
        setError(result.error.message)
      }
      setLoading(false)
    })
  }, [])

  if (loading) {
    return (
      <div className="p-6">
        <p className="text-gray-600">Loading installation info...</p>
      </div>
    )
  }

  if (error || !preview) {
    return (
      <div className="p-6">
        <p className="text-red-600">Failed to load installation info: {error ?? 'Unknown error'}</p>
      </div>
    )
  }

  const desktopFiles = preview.desktopFiles ?? []
  const iconFiles = preview.iconFiles ?? []
  const configFiles = preview.configFiles ?? []

  return (
    <div className="p-6 space-y-6">
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
        <p className="text-sm text-gray-600 mb-2">
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

      <div className="border-t border-gray-200 pt-6">
        <h3 className="text-sm font-medium text-gray-900 mb-2">Uninstall</h3>
        <p className="text-sm text-gray-600 mb-3">
          To remove Kyaraben and all managed files (except preserved data), run:
        </p>
        <code className="block bg-gray-100 px-3 py-2 rounded text-sm font-mono">
          kyaraben uninstall
        </code>
      </div>
    </div>
  )
}
