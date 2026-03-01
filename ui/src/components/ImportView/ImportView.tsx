import { useCallback, useEffect, useState } from 'react'
import { Button } from '@/lib/Button'
import * as daemon from '@/lib/daemon'
import { useHomeDir } from '@/lib/HomeDirContext'
import { useOnWindowFocus } from '@/lib/hooks/useOnWindowFocus'
import { useOpenPath } from '@/lib/hooks/useOpenPath'
import { IconButton } from '@/lib/IconButton'
import { Input } from '@/lib/Input'
import { FolderIcon } from '@/lib/icons'
import { PathText } from '@/lib/PathText'
import { collapseTilde } from '@/lib/paths'
import { Select } from '@/lib/Select'
import { Spinner } from '@/lib/Spinner'
import { useToast } from '@/lib/ToastContext'
import type {
  ImportDataComparison,
  ImportEmulatorReport,
  ImportFrontendReport,
  ImportScanResponse,
  ImportSystemReport,
} from '@/types/daemon'

const STORAGE_KEY_SOURCE = 'kyaraben-import-source-path'
const STORAGE_KEY_ESDE = 'kyaraben-import-esde-path'
const STORAGE_KEY_LAYOUT = 'kyaraben-import-layout'

const LAYOUTS = [{ value: 'emudeck', label: 'EmuDeck' }]

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / 1024 ** i).toFixed(1)} ${units[i]}`
}

function formatDelta(bytes: number, positive: boolean): string {
  const prefix = positive ? '+' : '-'
  if (bytes === 0) return `${prefix}0`
  return `${prefix}${formatBytes(bytes)}`
}

function PathRow({
  path,
  exists,
  onOpen,
}: {
  path: string
  exists: boolean
  onOpen: (path: string) => void
}) {
  return (
    <div className="flex items-center gap-1 min-w-0">
      <span className="text-sm text-on-surface-secondary truncate">
        <PathText>{path}</PathText>
      </span>
      {exists && (
        <button
          type="button"
          onClick={() => onOpen(path)}
          title="Open folder"
          className="p-0.5 text-on-surface-muted hover:text-on-surface-secondary rounded shrink-0"
        >
          <FolderIcon className="w-3.5 h-3.5" />
        </button>
      )}
    </div>
  )
}

function DataComparisonCard({
  comparison,
  mode,
  enabled,
  onOpenPath,
  onNavigateToCatalog,
}: {
  comparison: ImportDataComparison
  mode: string
  enabled: boolean
  onOpenPath: (path: string) => void
  onNavigateToCatalog: () => void
}) {
  const foundLabel = mode === 'reorganize' ? 'Current:' : 'Found at:'
  const expectLabel = mode === 'reorganize' ? 'Expected:' : 'Kyaraben expects:'
  const missingLabel = mode === 'reorganize' ? 'Needs to move' : 'Missing from Kyaraben'
  const extraLabel = mode === 'reorganize' ? 'Already in place' : 'Only in Kyaraben'

  const dataTypeLabels: Record<string, string> = {
    roms: 'ROMs',
    bios: 'BIOS',
    saves: 'Saves',
    states: 'States',
    screenshots: 'Screenshots',
    gamelists: 'Gamelists',
    media: 'Scraped media',
  }

  return (
    <div className="py-3 border-b border-outline last:border-b-0">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium text-on-surface">
          {dataTypeLabels[comparison.dataType] ?? comparison.dataType}
        </span>
        <span className="text-xs font-mono text-on-surface-muted">
          {formatDelta(comparison.diff.kyarabenDelta, true)}{' '}
          {formatDelta(comparison.diff.sourceDelta, false)}
        </span>
      </div>

      <div className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-xs">
        <span className="text-on-surface-dim">{foundLabel}</span>
        <div>
          {comparison.source.path ? (
            <>
              <PathRow
                path={comparison.source.path}
                exists={comparison.source.exists}
                onOpen={onOpenPath}
              />
              {comparison.source.exists ? (
                <span className="text-on-surface-muted ml-2">
                  {comparison.source.fileCount} files, {formatBytes(comparison.source.totalSize)}
                </span>
              ) : (
                <span className="text-on-surface-dim ml-2">(not found)</span>
              )}
            </>
          ) : (
            <span className="text-sm text-on-surface-dim">
              Not found in your existing collection
            </span>
          )}
        </div>

        <span className="text-on-surface-dim">{expectLabel}</span>
        <div>
          <PathRow
            path={comparison.kyaraben.path}
            exists={comparison.kyaraben.exists}
            onOpen={onOpenPath}
          />
          {comparison.kyaraben.exists ? (
            <span className="text-on-surface-muted ml-2">
              {comparison.kyaraben.fileCount} files, {formatBytes(comparison.kyaraben.totalSize)}
            </span>
          ) : !enabled ? (
            <button
              type="button"
              onClick={onNavigateToCatalog}
              className="text-accent hover:underline ml-2"
            >
              Enable in Catalog and apply
            </button>
          ) : (
            <span className="text-on-surface-dim ml-2">(not found)</span>
          )}
        </div>
      </div>

      {comparison.notes && comparison.notes.length > 0 && (
        <div className="mt-2 p-2 bg-status-warning/10 border border-status-warning/30 rounded text-xs text-on-surface">
          {comparison.notes.map((note) => (
            <p key={note}>{note}</p>
          ))}
        </div>
      )}

      {comparison.diff.onlyInSource && comparison.diff.onlyInSource.length > 0 && (
        <div className="mt-2">
          <span className="text-xs text-on-surface-dim">
            {missingLabel} ({comparison.diff.onlyInSource.length} files,{' '}
            {formatBytes(comparison.diff.sourceDelta)}):
          </span>
          <ul className="mt-1 text-xs text-on-surface-muted font-mono">
            {comparison.diff.onlyInSource.slice(0, 3).map((f) => (
              <li key={f.relPath} className="truncate">
                {f.relPath}
              </li>
            ))}
            {comparison.diff.onlyInSource.length > 3 && (
              <li className="text-on-surface-dim">
                ... and {comparison.diff.onlyInSource.length - 3} more
              </li>
            )}
          </ul>
        </div>
      )}

      {comparison.diff.onlyInKyaraben && comparison.diff.onlyInKyaraben.length > 0 && (
        <div className="mt-2">
          <span className="text-xs text-on-surface-dim">
            {extraLabel} ({comparison.diff.onlyInKyaraben.length} files,{' '}
            {formatBytes(comparison.diff.kyarabenDelta)}):
          </span>
          <ul className="mt-1 text-xs text-on-surface-muted font-mono">
            {comparison.diff.onlyInKyaraben.slice(0, 3).map((f) => (
              <li key={f.relPath} className="truncate">
                {f.relPath}
              </li>
            ))}
            {comparison.diff.onlyInKyaraben.length > 3 && (
              <li className="text-on-surface-dim">
                ... and {comparison.diff.onlyInKyaraben.length - 3} more
              </li>
            )}
          </ul>
        </div>
      )}
    </div>
  )
}

function EmulatorSection({
  emulator,
  mode,
  onOpenPath,
  onNavigateToCatalog,
}: {
  emulator: ImportEmulatorReport
  mode: string
  onOpenPath: (path: string) => void
  onNavigateToCatalog: () => void
}) {
  const [expanded, setExpanded] = useState(true)
  const hasData = emulator.emulatorData.some((d) => d.source.exists || d.kyaraben.exists)

  if (!hasData) return null

  return (
    <div className="ml-4 mt-2 border-l border-outline pl-4">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="text-sm text-on-surface-secondary hover:text-on-surface flex items-center gap-2"
      >
        <span>{expanded ? '▼' : '▶'}</span>
        <span>{emulator.emulatorName}</span>
        {!emulator.enabled && (
          <span className="text-xs text-on-surface-dim bg-surface-alt px-1.5 py-0.5 rounded">
            not enabled
          </span>
        )}
      </button>
      {expanded && (
        <div className="mt-2">
          {emulator.emulatorData.map((data) => (
            <DataComparisonCard
              key={data.dataType}
              comparison={data}
              mode={mode}
              enabled={emulator.enabled}
              onOpenPath={onOpenPath}
              onNavigateToCatalog={onNavigateToCatalog}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function SystemSection({
  system,
  mode,
  onOpenPath,
  onNavigateToCatalog,
}: {
  system: ImportSystemReport
  mode: string
  onOpenPath: (path: string) => void
  onNavigateToCatalog: () => void
}) {
  const [expanded, setExpanded] = useState(true)
  const hasSystemData = system.systemData.some((d) => d.source.exists || d.kyaraben.exists)
  const hasEmulatorData = system.emulators.some((e) =>
    e.emulatorData.some((d) => d.source.exists || d.kyaraben.exists),
  )

  if (!hasSystemData && !hasEmulatorData) return null

  return (
    <div className="bg-surface-alt rounded-card overflow-hidden">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="w-full px-4 py-3 flex items-center justify-between bg-surface-raised hover:bg-surface-alt transition-colors"
      >
        <div className="flex items-center gap-2">
          <span className="font-medium text-on-surface">{system.systemName}</span>
          {!system.enabled && (
            <span className="text-xs text-on-surface-dim bg-surface-alt px-1.5 py-0.5 rounded">
              not enabled
            </span>
          )}
        </div>
        <span className="text-on-surface-dim">{expanded ? '▼' : '▶'}</span>
      </button>
      {expanded && (
        <div className="px-4 pb-4">
          {system.systemData.map((data) => (
            <DataComparisonCard
              key={data.dataType}
              comparison={data}
              mode={mode}
              enabled={system.enabled}
              onOpenPath={onOpenPath}
              onNavigateToCatalog={onNavigateToCatalog}
            />
          ))}
          {system.emulators.map((emu) => (
            <EmulatorSection
              key={emu.emulator}
              emulator={emu}
              mode={mode}
              onOpenPath={onOpenPath}
              onNavigateToCatalog={onNavigateToCatalog}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function FrontendSection({
  frontend,
  mode,
  onOpenPath,
  onNavigateToCatalog,
}: {
  frontend: ImportFrontendReport
  mode: string
  onOpenPath: (path: string) => void
  onNavigateToCatalog: () => void
}) {
  const [expanded, setExpanded] = useState(true)
  const hasData = frontend.frontendData.some((d) => d.source.exists || d.kyaraben.exists)

  if (!hasData) return null

  return (
    <div className="bg-surface-alt rounded-card overflow-hidden">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="w-full px-4 py-3 flex items-center justify-between bg-surface-raised hover:bg-surface-alt transition-colors"
      >
        <span className="font-medium text-on-surface">{frontend.frontendName} (frontend)</span>
        <span className="text-on-surface-dim">{expanded ? '▼' : '▶'}</span>
      </button>
      {expanded && (
        <div className="px-4 pb-4">
          {frontend.frontendData.map((data) => (
            <DataComparisonCard
              key={data.dataType}
              comparison={data}
              mode={mode}
              enabled={true}
              onOpenPath={onOpenPath}
              onNavigateToCatalog={onNavigateToCatalog}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export interface ImportViewProps {
  readonly onNavigateToCatalog: () => void
}

export function ImportView({ onNavigateToCatalog }: ImportViewProps) {
  const [sourcePath, setSourcePath] = useState(() => localStorage.getItem(STORAGE_KEY_SOURCE) ?? '')
  const [esdePath, setEsdePath] = useState(() => localStorage.getItem(STORAGE_KEY_ESDE) ?? '')
  const [layout, setLayout] = useState(() => localStorage.getItem(STORAGE_KEY_LAYOUT) ?? 'emudeck')
  const [report, setReport] = useState<ImportScanResponse | null>(null)
  const [scanning, setScanning] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { showToast } = useToast()
  const openPath = useOpenPath()
  const homeDir = useHomeDir()

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY_SOURCE, sourcePath)
  }, [sourcePath])

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY_ESDE, esdePath)
  }, [esdePath])

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY_LAYOUT, layout)
  }, [layout])

  const handleScan = useCallback(async () => {
    if (!sourcePath) {
      showToast('Please enter a source path', 'error')
      return
    }

    setScanning(true)
    setError(null)

    const result = await daemon.importScan({
      sourcePath,
      layout,
      ...(esdePath && { esdePath }),
    })

    setScanning(false)

    if (result.ok) {
      setReport(result.data)
    } else {
      const msg = result.error.message
      if (msg.includes('no such file') || msg.includes('not exist')) {
        setError('That folder does not exist. Check the path and try again.')
      } else {
        setError(msg)
      }
    }
  }, [sourcePath, esdePath, layout, showToast])

  const handleSelectSource = useCallback(async () => {
    const result = await daemon.selectDirectory()
    if (result.ok && result.data) {
      setSourcePath(collapseTilde(result.data, homeDir))
    }
  }, [homeDir])

  const handleSelectEsde = useCallback(async () => {
    const result = await daemon.selectDirectory()
    if (result.ok && result.data) {
      setEsdePath(collapseTilde(result.data, homeDir))
    }
  }, [homeDir])

  useOnWindowFocus(() => {
    if (sourcePath && report) {
      handleScan()
    }
  })

  if (!report) {
    return (
      <div className="p-6">
        <h1 className="text-xl font-semibold text-on-surface mb-2">Import</h1>
        <p className="text-sm text-on-surface-muted mb-6">
          Bring your existing collection into Kyaraben. This tool analyzes your files and shows what
          can be imported.
        </p>

        <div className="space-y-4">
          <div>
            <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block mb-2">
              Existing collection
            </span>
            <div className="flex gap-2">
              <div className="flex-1">
                <Input value={sourcePath} onChange={setSourcePath} placeholder="~/Emulation" />
              </div>
              <IconButton
                icon={<FolderIcon className="w-5 h-5 text-on-surface-muted" />}
                label="Browse"
                onClick={handleSelectSource}
              />
            </div>
          </div>

          <div>
            <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block mb-2">
              ES-DE data (optional)
            </span>
            <div className="flex gap-2">
              <div className="flex-1">
                <Input value={esdePath} onChange={setEsdePath} placeholder="~/ES-DE" />
              </div>
              <IconButton
                icon={<FolderIcon className="w-5 h-5 text-on-surface-muted" />}
                label="Browse"
                onClick={handleSelectEsde}
              />
            </div>
            <p className="mt-1 text-xs text-on-surface-muted">
              Leave empty to auto-detect ~/ES-DE if it exists.
            </p>
          </div>

          <div>
            <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block mb-2">
              Layout hint
            </span>
            <Select value={layout} options={LAYOUTS} onChange={setLayout} />
            <p className="mt-1 text-xs text-on-surface-muted">
              How your existing collection is organized.
            </p>
          </div>

          {error && (
            <div className="p-4 bg-status-warning/10 border border-status-warning/30 rounded-card">
              <p className="text-sm text-on-surface">{error}</p>
            </div>
          )}

          <Button onClick={handleScan} disabled={scanning || !sourcePath}>
            {scanning ? (
              <>
                <Spinner />
                <span className="ml-2">Scanning...</span>
              </>
            ) : (
              'Scan'
            )}
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6">
      <div className="mb-6 p-4 bg-status-warning/10 border border-status-warning/30 rounded-card">
        <p className="text-sm text-on-surface">Back up your collection before copying anything.</p>
      </div>

      <div className="mb-6 flex flex-wrap items-center gap-4">
        <div className="flex-1 min-w-0">
          <div className="text-sm text-on-surface-dim">
            {report.mode === 'reorganize' ? 'Reorganizing:' : 'Existing collection:'}
          </div>
          <PathRow path={report.sourcePath} exists={true} onOpen={openPath} />
        </div>
        {report.mode !== 'reorganize' && (
          <div className="flex-1 min-w-0">
            <div className="text-sm text-on-surface-dim">Kyaraben collection:</div>
            <PathRow path={report.kyarabenPath} exists={true} onOpen={openPath} />
          </div>
        )}
        <div className="flex gap-2">
          <Button variant="secondary" onClick={() => setReport(null)}>
            Change
          </Button>
          <Button onClick={handleScan} disabled={scanning}>
            {scanning ? <Spinner /> : 'Rescan'}
          </Button>
        </div>
      </div>

      <div className="mb-6 text-sm text-on-surface-muted">
        Overall: {formatDelta(report.summary.totalOnlyInKyaraben, true)} in Kyaraben,{' '}
        {formatDelta(report.summary.totalOnlyInSource, false)} to import
      </div>

      <div className="space-y-4">
        {[...report.systems]
          .sort((a, b) => (a.enabled === b.enabled ? 0 : a.enabled ? -1 : 1))
          .map((sys) => (
            <SystemSection
              key={sys.system}
              system={sys}
              mode={report.mode}
              onOpenPath={openPath}
              onNavigateToCatalog={onNavigateToCatalog}
            />
          ))}
        {report.frontends?.map((fe) => (
          <FrontendSection
            key={fe.frontend}
            frontend={fe}
            mode={report.mode}
            onOpenPath={openPath}
            onNavigateToCatalog={onNavigateToCatalog}
          />
        ))}
      </div>
    </div>
  )
}
