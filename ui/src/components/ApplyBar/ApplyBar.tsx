import { useEffect, useState } from 'react'
import { SpeedBadge } from '@/components/SpeedBadge/SpeedBadge'
import { useApply } from '@/lib/ApplyContext'
import { BOTTOM_BAR_HEIGHT, BottomBar } from '@/lib/BottomBar'
import { BottomBarPortal } from '@/lib/BottomBarSlot'
import { Button } from '@/lib/Button'
import { useConfig } from '@/lib/ConfigContext'
import { CHANGE_CONFIG, formatBytes, getChangeGroups } from '@/lib/changeUtils'
import { useHomeDir } from '@/lib/HomeDirContext'
import { collapsePathsInText } from '@/lib/paths'
import { getDownloadSpeedBytes, getStepSubtitle } from '@/lib/progressUtils'
import { ProgressBar, ProgressRail, Shimmer } from '@/lib/progressWidgets'
import { useOpenLog } from '@/lib/useOpenLog'

function hasUserConflicts(
  diffs: readonly {
    userModified?: boolean
    hasChanges?: boolean
    userChanges?: readonly unknown[]
  }[],
): boolean {
  return diffs.some((d) => d.userModified && d.hasChanges && (d.userChanges?.length ?? 0) > 0)
}

function formatConfigChanges(changes: readonly string[]): string {
  if (changes.length === 0) return ''
  const lower = changes.map((c) => c.toLowerCase())
  const first = lower[0] ?? ''
  if (lower.length === 1) return `${capitalize(first)} changed.`
  const last = lower[lower.length - 1] ?? ''
  const rest = lower.slice(0, -1)
  return `${capitalize(rest.join(', '))} and ${last} changed.`
}

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

export function ApplyBar() {
  const {
    status,
    loading,
    progressSteps,
    cancel,
    logPosition,
    preflightData,
    confirmApply,
    reset,
  } = useApply()
  const { changes, apply, reapply, discard, upgradeAvailable } = useConfig()
  const homeDir = useHomeDir()
  const openLog = useOpenLog()

  const [confirmingDiscard, setConfirmingDiscard] = useState(false)
  const [confirmingCancel, setConfirmingCancel] = useState(false)
  const [cancelling, setCancelling] = useState(false)

  const isApplying = status === 'applying'
  const hasChanges = changes.total > 0 || changes.hasConfigChanges
  const upgradeOnly = upgradeAvailable && !hasChanges

  useEffect(() => {
    if (status !== 'applying') {
      setCancelling(false)
      setConfirmingCancel(false)
    }
  }, [status])

  useEffect(() => {
    if (!confirmingDiscard) return
    const timer = setTimeout(() => setConfirmingDiscard(false), 3000)
    return () => clearTimeout(timer)
  }, [confirmingDiscard])

  useEffect(() => {
    if (!confirmingCancel) return
    const timer = setTimeout(() => setConfirmingCancel(false), 3000)
    return () => clearTimeout(timer)
  }, [confirmingCancel])

  const handleDiscard = () => {
    if (confirmingDiscard) {
      discard()
      setConfirmingDiscard(false)
    } else {
      setConfirmingDiscard(true)
    }
  }

  const handleCancel = () => {
    if (cancelling) return
    if (confirmingCancel) {
      setCancelling(true)
      setConfirmingCancel(false)
      cancel()
    } else {
      setConfirmingCancel(true)
    }
  }

  if (status === 'reviewing' && preflightData) {
    const showOverride = hasUserConflicts(preflightData.diffs ?? [])
    const buttonLabel = loading
      ? 'Applying...'
      : showOverride
        ? 'Continue and override'
        : 'Continue'
    return (
      <BottomBar>
        <button
          type="button"
          onClick={reset}
          disabled={loading}
          className="text-accent hover:text-accent-hover hover:underline disabled:text-on-surface-dim disabled:cursor-not-allowed disabled:no-underline"
        >
          Cancel
        </button>
        <Button onClick={confirmApply} disabled={loading}>
          {buttonLabel}
        </Button>
      </BottomBar>
    )
  }

  if (status === 'confirming_sync') {
    return null
  }

  const isDone = status === 'success' || status === 'error' || status === 'cancelled'
  if (isDone) {
    return (
      <BottomBar>
        <span />
        <div className="flex items-center gap-4">
          <button
            type="button"
            onClick={() => openLog(logPosition ?? undefined)}
            className="text-on-surface-muted hover:text-on-surface-secondary hover:underline text-sm"
          >
            Open log in terminal
          </button>
          <Button onClick={reset}>Done</Button>
        </div>
      </BottomBar>
    )
  }

  if (!hasChanges && !upgradeAvailable && !isApplying) {
    return null
  }

  if (isApplying) {
    const currentStep = [...progressSteps].reverse().find((s) => s.status === 'in_progress')
    const label = currentStep?.label ?? 'Installing...'
    const rawSubtitle = currentStep ? getStepSubtitle(currentStep) : null
    const subtitle = rawSubtitle ? collapsePathsInText(rawSubtitle, homeDir) : null
    const showSpeed = currentStep?.id === 'build' && currentStep.status === 'in_progress'
    const downloadSpeedBytes = currentStep ? getDownloadSpeedBytes(currentStep) : 0
    const computedPercent =
      currentStep?.bytesTotal && currentStep.bytesTotal > 0
        ? Math.min(
            100,
            Math.floor(((currentStep.bytesDownloaded ?? 0) * 100) / currentStep.bytesTotal),
          )
        : undefined
    const progressPercent = currentStep?.progressPercent ?? computedPercent

    return (
      <BottomBarPortal>
        <div
          className={`relative bg-surface-alt/95 backdrop-blur-sm border-t border-outline px-6 ${BOTTOM_BAR_HEIGHT} flex items-center`}
        >
          <ProgressRail className="absolute left-0 right-0 bottom-0 h-1 pointer-events-none">
            {progressPercent !== undefined ? (
              <ProgressBar percent={progressPercent} />
            ) : (
              <Shimmer />
            )}
          </ProgressRail>
          <div className="flex-1 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-4 h-4 border-2 border-accent border-t-transparent rounded-full animate-spin" />
              <span className="text-sm text-on-surface-secondary truncate max-w-md">
                <span className="font-medium">{label}</span>
                {subtitle && <span className="text-on-surface-dim ml-1">{subtitle}</span>}
              </span>
            </div>
            <div className="flex items-center gap-4">
              <SpeedBadge speedBytes={downloadSpeedBytes} show={showSpeed} />
              <button
                type="button"
                onClick={() => openLog(logPosition ?? undefined)}
                className="text-on-surface-muted hover:text-on-surface-secondary hover:underline text-sm"
              >
                Open log in terminal
              </button>
              <button
                type="button"
                onClick={handleCancel}
                disabled={cancelling}
                className={`text-sm ${cancelling ? 'text-on-surface-dim cursor-not-allowed' : confirmingCancel ? 'text-status-error hover:text-status-error hover:underline' : 'text-on-surface-muted hover:text-on-surface-secondary hover:underline'}`}
              >
                {cancelling ? 'Canceling...' : confirmingCancel ? 'Confirm cancel' : 'Cancel'}
              </button>
            </div>
          </div>
        </div>
      </BottomBarPortal>
    )
  }

  const netBytes = changes.downloadBytes - changes.freeBytes
  const hasSize = changes.downloadBytes > 0 || changes.freeBytes > 0
  const changeGroups = getChangeGroups(changes)

  return (
    <BottomBar>
      <div className="flex items-center gap-4 text-sm min-w-0 overflow-hidden">
        {upgradeOnly ? (
          <span className="text-on-surface-secondary">
            Kyaraben was updated. Apply to get the latest emulator configs.
          </span>
        ) : (
          <>
            {changeGroups.length > 0 && (
              <div className="flex items-center gap-3 min-w-0">
                {changeGroups.map(({ type, items }) => {
                  const config = CHANGE_CONFIG[type]
                  const names = items.map((i) => i.name).join(', ')
                  return (
                    <span
                      key={type}
                      className={`flex items-center gap-1.5 min-w-0 ${config.textColor}`}
                    >
                      <span className="shrink-0">{config.icon}</span>
                      <span className="truncate">{names}</span>
                    </span>
                  )
                })}
              </div>
            )}
            {changeGroups.length === 0 && changes.configChanges.length > 0 && (
              <span className="text-on-surface-secondary">
                {formatConfigChanges(changes.configChanges)}
              </span>
            )}
            {hasSize && (
              <div className="flex items-center gap-2 font-mono shrink-0">
                {changes.downloadBytes > 0 && (
                  <span className="text-status-ok">+{formatBytes(changes.downloadBytes)}</span>
                )}
                {changes.freeBytes > 0 && (
                  <span className="text-status-error">-{formatBytes(changes.freeBytes)}</span>
                )}
                {changes.downloadBytes > 0 && changes.freeBytes > 0 && (
                  <span className="text-on-surface-dim">
                    ({netBytes >= 0 ? '+' : '-'}
                    {formatBytes(netBytes)})
                  </span>
                )}
              </div>
            )}
          </>
        )}
      </div>

      <div className="flex items-center gap-4 shrink-0 ml-4">
        {changes.hasConfigChanges && (
          <button
            type="button"
            onClick={handleDiscard}
            disabled={loading}
            className="text-sm text-accent hover:text-accent-hover hover:underline disabled:text-on-surface-dim disabled:cursor-not-allowed disabled:no-underline"
          >
            {confirmingDiscard ? 'Click again to confirm' : 'Discard changes'}
          </button>
        )}
        <Button onClick={upgradeOnly ? reapply : apply} disabled={loading}>
          {loading ? 'Applying...' : 'Apply'}
        </Button>
      </div>
    </BottomBar>
  )
}
