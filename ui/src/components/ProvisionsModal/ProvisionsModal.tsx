import { ProvisionSummary } from '@/components/ProvisionSummary/ProvisionSummary'
import { FolderIcon, PlayIcon } from '@/lib/icons'
import { Modal } from '@/lib/Modal'
import { PathText } from '@/lib/PathText'
import { collapseTilde } from '@/lib/paths'
import {
  OPTIONAL_PROVISION_COLOR,
  OPTIONAL_PROVISION_ICON,
  PROVISION_FOUND_COLOR,
  PROVISION_FOUND_ICON,
  PROVISION_MISSING_COLOR,
  PROVISION_MISSING_ICON,
  PROVISION_NOT_NEEDED_COLOR,
  PROVISION_NOT_NEEDED_ICON,
} from '@/lib/provisionStatus'
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
    ? 'flex items-center gap-1 text-accent hover:text-accent-hover transition-colors ml-auto min-w-0 overflow-hidden max-w-full'
    : 'flex items-center gap-1 text-xs text-accent hover:text-accent-hover transition-colors overflow-hidden min-w-0 max-w-full'
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
          <span className={isInline ? 'hidden sm:inline' : 'truncate min-w-0'}>
            Import in emulator
          </span>
        </button>
      )
    }
    if (!isFound) {
      return isInline ? (
        <span className="text-on-surface-dim shrink-0 ml-auto hidden sm:inline">
          Import after install
        </span>
      ) : (
        <span className="text-xs text-on-surface-dim truncate min-w-0">Import after install</span>
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
          <span className="hidden md:inline truncate max-w-32">
            <PathText>{expectedPath}</PathText>
          </span>
        </>
      ) : (
        <span className="truncate min-w-0">
          Open <PathText>{expectedPath}</PathText>
        </span>
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
    const displayPath = collapseTilde(path, window.electron.homeDir)
    showToast(`Opening ${displayPath}.`)
  }

  const found = provisions.filter((p) => p.status === 'found')
  const missing = provisions.filter((p) => p.status !== 'found')

  const foundGroups = groupProvisions(found)
  const missingGroups = groupProvisions(missing)

  return (
    <Modal open={open} onClose={onClose} title={`${emulatorName} provisions`}>
      <div className="space-y-4">
        <ProvisionSummary provisions={provisions} size="sm" />
        {found.length > 0 && (
          <div>
            <p className="text-sm text-on-surface-muted mb-2">Found</p>
            <div className="space-y-3">
              {foundGroups.map((group) => (
                <div key={group.message ?? 'default'} className="space-y-1">
                  {group.message && (
                    <p className="text-xs text-on-surface-dim mb-1">{group.message}</p>
                  )}
                  {group.provisions.map((p) => (
                    <ProvisionRow
                      key={p.filename}
                      provision={p}
                      disabled={disabled}
                      onOpenFolder={handleOpenFolder}
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
                  {group.message && (
                    <p className="text-xs text-on-surface-dim mb-1">{group.message}</p>
                  )}
                  {group.provisions.map((p) => (
                    <ProvisionRow
                      key={p.filename}
                      provision={p}
                      disabled={disabled}
                      onOpenFolder={handleOpenFolder}
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
  onLaunch,
}: {
  readonly provision: ProvisionResult
  readonly disabled: boolean
  readonly onOpenFolder: (path: string) => void
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
    if (isFound) return { icon: PROVISION_FOUND_ICON, color: PROVISION_FOUND_COLOR }
    if (isGroupSatisfied && !isOptional) {
      return { icon: PROVISION_NOT_NEEDED_ICON, color: PROVISION_NOT_NEEDED_COLOR }
    }
    if (isOptional) return { icon: OPTIONAL_PROVISION_ICON, color: OPTIONAL_PROVISION_COLOR }
    return { icon: PROVISION_MISSING_ICON, color: PROVISION_MISSING_COLOR }
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
        {statusLabel && (
          <span className="text-xs text-on-surface-dim hidden min-[450px]:inline">
            {statusLabel}
          </span>
        )}
      </div>
      <div className="flex flex-col min-[450px]:flex-row min-[450px]:items-center gap-1 ml-5 min-w-0">
        <span className="text-xs text-on-surface-dim truncate min-w-0 min-[450px]:flex-1">
          {isFound ? (
            <>
              Verified (
              <code className="text-xs">
                {provision.verifiedDisplayName || provision.displayName}
              </code>
              )
            </>
          ) : provision.importViaUI ? (
            <>
              Import <code className="text-xs">{provision.displayName}</code> via emulator
            </>
          ) : (
            provision.instructions
          )}
        </span>
        <div className="min-[450px]:ml-auto shrink-0 min-[450px]:max-w-[50%]">
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
