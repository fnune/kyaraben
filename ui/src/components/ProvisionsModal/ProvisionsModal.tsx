import { CopyIcon, FolderIcon, PlayIcon } from '@/lib/icons'
import { Modal } from '@/lib/Modal'
import { useToast } from '@/lib/ToastContext'
import type { ProvisionResult } from '@/types/daemon'

export interface ProvisionsModalProps {
  readonly open: boolean
  readonly onClose: () => void
  readonly emulatorName: string
  readonly provisions: readonly ProvisionResult[]
  readonly onLaunch?: () => void
  readonly disabled: boolean
}

const KIND_LABELS: Record<string, string> = {
  bios: 'BIOS',
  keys: 'Keys',
  firmware: 'Firmware',
}

export function getKindLabel(kind: string): string {
  return KIND_LABELS[kind] ?? kind
}

interface ProvisionActionProps {
  readonly provision: ProvisionResult
  readonly disabled: boolean
  readonly onOpenFolder: (path: string) => void
  readonly onLaunch?: () => void
  readonly variant: 'inline' | 'row'
}

function ProvisionAction({
  provision,
  disabled,
  onOpenFolder,
  onLaunch,
  variant,
}: ProvisionActionProps) {
  const expectedPath = provision.expectedPath ?? ''
  const isFound = provision.status === 'found'

  const isInline = variant === 'inline'
  const iconClass = isInline ? 'w-4 h-4 shrink-0' : 'w-3.5 h-3.5 shrink-0'
  const textClass = isInline
    ? 'flex items-center gap-1 text-accent hover:text-accent-hover transition-colors shrink-0 ml-auto'
    : 'flex items-center gap-1 text-xs text-accent hover:text-accent-hover transition-colors truncate'
  const disabledClass = disabled ? 'opacity-50 cursor-not-allowed' : ''

  if (provision.importViaUI) {
    if (onLaunch && !disabled) {
      return (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            onLaunch()
          }}
          className={textClass}
        >
          <PlayIcon className={iconClass} />
          <span className={isInline ? 'hidden sm:inline' : 'truncate'}>Import in emulator</span>
        </button>
      )
    }
    if (!isFound) {
      return isInline ? (
        <span className="text-on-surface-dim shrink-0 ml-auto hidden sm:inline">
          Import after install
        </span>
      ) : (
        <span className="text-xs text-on-surface-dim truncate">Import after install</span>
      )
    }
    return null
  }

  if (!expectedPath) return null

  return (
    <button
      type="button"
      onClick={(e) => {
        e.stopPropagation()
        onOpenFolder(expectedPath)
      }}
      disabled={disabled}
      className={`${textClass} ${disabledClass}`}
    >
      <FolderIcon className={iconClass} />
      {isInline ? (
        <>
          <span className="hidden sm:inline">Open</span>
          <span className="hidden md:inline truncate max-w-32">{expectedPath}</span>
        </>
      ) : (
        <span className="truncate">Open {expectedPath}</span>
      )}
    </button>
  )
}

export interface ProvisionActionInlineProps {
  readonly provision: ProvisionResult
  readonly disabled: boolean
  readonly onOpenFolder: (path: string) => void
  readonly onLaunch?: () => void
}

export function ProvisionActionInline(props: ProvisionActionInlineProps) {
  return <ProvisionAction {...props} variant="inline" />
}

interface ProvisionGroup {
  message: string | null
  provisions: ProvisionResult[]
}

function groupProvisions(provisions: readonly ProvisionResult[]): ProvisionGroup[] {
  const groups: Map<string, ProvisionGroup> = new Map()

  for (const p of provisions) {
    const key = p.groupMessage ?? ''
    if (!groups.has(key)) {
      groups.set(key, { message: p.groupMessage ?? null, provisions: [] })
    }
    groups.get(key)?.provisions.push(p)
  }

  return Array.from(groups.values())
}

export function ProvisionsModal({
  open,
  onClose,
  emulatorName,
  provisions,
  onLaunch,
  disabled,
}: ProvisionsModalProps) {
  const { showToast } = useToast()

  const handleOpenFolder = (path: string) => {
    window.electron.invoke('open_path', path)
    showToast(`Opening ${path}`)
  }

  const handleCopy = (filename: string) => {
    navigator.clipboard.writeText(filename)
    showToast(`Copied ${filename}`)
  }

  const found = provisions.filter((p) => p.status === 'found')
  const missing = provisions.filter((p) => p.status !== 'found')

  const foundGroups = groupProvisions(found)
  const missingGroups = groupProvisions(missing)

  return (
    <Modal open={open} onClose={onClose} title={`${emulatorName} provisions`}>
      <div className="space-y-4">
        {found.length > 0 && (
          <div>
            <p className="text-sm text-on-surface-muted mb-2">Found</p>
            <div className="space-y-3">
              {foundGroups.map((group) => (
                <div key={group.message ?? 'default'} className="space-y-1">
                  {group.message && <p className="text-xs text-on-surface-dim mb-1">{group.message}</p>}
                  {group.provisions.map((p) => (
                    <ProvisionRow
                      key={p.filename}
                      provision={p}
                      disabled={disabled}
                      onOpenFolder={handleOpenFolder}
                      onCopy={handleCopy}
                    />
                  ))}
                </div>
              ))}
            </div>
          </div>
        )}

        {missing.length > 0 && (
          <div>
            <p className="text-sm text-on-surface-muted mb-2">Missing</p>
            <div className="space-y-3">
              {missingGroups.map((group) => (
                <div key={group.message ?? 'default'} className="space-y-1">
                  {group.message && <p className="text-xs text-on-surface-dim mb-1">{group.message}</p>}
                  {group.provisions.map((p) => (
                    <ProvisionRow
                      key={p.filename}
                      provision={p}
                      disabled={disabled}
                      onOpenFolder={handleOpenFolder}
                      onCopy={handleCopy}
                      {...(onLaunch && { onLaunch })}
                    />
                  ))}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}

function ProvisionRow({
  provision,
  disabled,
  onOpenFolder,
  onCopy,
  onLaunch,
}: {
  readonly provision: ProvisionResult
  readonly disabled: boolean
  readonly onOpenFolder: (path: string) => void
  readonly onCopy: (filename: string) => void
  readonly onLaunch?: () => void
}) {
  const isFound = provision.status === 'found'
  const isOptional = !provision.groupRequired
  const isGroupSatisfied = provision.groupSatisfied
  const kindLabel = getKindLabel(provision.kind)

  const getStatusLabel = () => {
    if (isFound) return null
    if (isOptional) return 'optional'
    if (isGroupSatisfied) return 'not needed'
    if (provision.groupSize > 1) return 'at least one required'
    return 'required'
  }

  const getIcon = () => {
    if (isFound) return { icon: '✓', color: 'text-status-ok' }
    if (isGroupSatisfied && !isOptional) return { icon: '-', color: 'text-on-surface-dim' }
    if (isOptional) return { icon: '✗', color: 'text-status-warning' }
    return { icon: '✗', color: 'text-status-error' }
  }

  const { icon, color } = getIcon()
  const statusLabel = getStatusLabel()

  return (
    <div className="bg-surface-raised rounded-sm px-3 py-2 space-y-0.5">
      <div className="flex items-center gap-2">
        <span className={`${color} w-3 text-center shrink-0`}>{icon}</span>
        <span className="text-sm text-on-surface-secondary">
          {kindLabel}
          {provision.description && ` (${provision.description})`}
        </span>
        {statusLabel && <span className="text-xs text-on-surface-dim">{statusLabel}</span>}
      </div>
      <div className="flex items-center gap-1 ml-5">
        <code className="text-xs text-on-surface-dim truncate">{provision.filename}</code>
        <button
          type="button"
          onClick={() => onCopy(provision.filename)}
          disabled={disabled}
          className={`p-0.5 text-on-surface-muted rounded-sm transition-colors shrink-0 ${disabled ? 'opacity-50 cursor-not-allowed' : 'hover:text-on-surface'}`}
          aria-label={`Copy ${provision.filename}`}
        >
          <CopyIcon />
        </button>
        <div className="ml-auto">
          <ProvisionAction
            provision={provision}
            disabled={disabled}
            onOpenFolder={onOpenFolder}
            variant="row"
            {...(onLaunch && { onLaunch })}
          />
        </div>
      </div>
    </div>
  )
}
