import type { LogEntry, LogLevel } from '@shared/logging.gen'
import type { ProgressStep } from '@shared/ui'
import { useEffect, useRef } from 'react'
import { SpeedBadge } from '@/components/SpeedBadge/SpeedBadge'
import { useHomeDir } from '@/lib/HomeDirContext'
import { collapsePathsInText } from '@/lib/paths'
import { getDownloadSpeedBytes, getStepSubtitle } from '@/lib/progressUtils'
import { ProgressBar, ProgressRail, Shimmer } from '@/lib/progressWidgets'

export interface ProgressStepsProps {
  steps: readonly ProgressStep[]
  error?: string
  cancelled?: boolean
}

function StepIcon({ status }: { readonly status: ProgressStep['status'] }) {
  switch (status) {
    case 'completed':
      return <span className="text-status-ok">✓</span>
    case 'in_progress':
      return <span className="text-accent animate-pulse">●</span>
    case 'error':
      return <span className="text-status-error">✗</span>
    case 'cancelled':
      return <span className="text-status-warning">⊘</span>
    default:
      return <span className="text-on-surface-dim">○</span>
  }
}

function StepSubtitle({ children }: { readonly children: React.ReactNode }) {
  return <span className="text-sm ml-1">{children}</span>
}

function getLogLevelClass(level: LogLevel): string {
  switch (level) {
    case 'error':
      return 'text-status-error'
    case 'warn':
      return 'text-status-warning'
    case 'debug':
      return 'text-on-surface-dim'
    default:
      return 'text-on-surface-secondary'
  }
}

function LogEntriesPre({
  logEntries,
  isInProgress,
  progressPercent,
  homeDir,
}: {
  readonly logEntries: readonly LogEntry[]
  readonly isInProgress: boolean
  readonly progressPercent?: number
  readonly homeDir: string
}) {
  const preRef = useRef<HTMLPreElement>(null)

  // biome-ignore lint/correctness/useExhaustiveDependencies: scroll when logEntries changes
  useEffect(() => {
    if (preRef.current) {
      preRef.current.scrollTop = preRef.current.scrollHeight
    }
  }, [logEntries])

  const hasProgress = progressPercent !== undefined && progressPercent > 0

  return (
    <div className="relative">
      <pre
        ref={preRef}
        className="p-2 text-xs font-mono bg-surface rounded-sm max-h-32 overflow-y-auto whitespace-pre overflow-x-hidden"
      >
        {logEntries.map((entry) => (
          <span
            key={`${entry.timestamp}-${entry.message}`}
            className={getLogLevelClass(entry.level)}
          >
            {collapsePathsInText(entry.message, homeDir)}
            {'\n'}
          </span>
        ))}
      </pre>
      {isInProgress && (
        <ProgressRail className="absolute bottom-0 left-0 right-0 h-1 bg-surface rounded-b pointer-events-none">
          {hasProgress ? <ProgressBar percent={progressPercent} /> : <Shimmer />}
        </ProgressRail>
      )}
    </div>
  )
}

function OutputPre({
  output,
  isInProgress,
  progressPercent,
  homeDir,
}: {
  readonly output: readonly string[]
  readonly isInProgress: boolean
  readonly progressPercent?: number
  readonly homeDir: string
}) {
  const preRef = useRef<HTMLPreElement>(null)

  // biome-ignore lint/correctness/useExhaustiveDependencies: scroll when output changes
  useEffect(() => {
    if (preRef.current) {
      preRef.current.scrollTop = preRef.current.scrollHeight
    }
  }, [output])

  const hasProgress = progressPercent !== undefined && progressPercent > 0

  return (
    <div className="relative">
      <pre
        ref={preRef}
        className="p-2 text-xs font-mono bg-surface text-on-surface-secondary rounded-sm max-h-32 overflow-y-auto whitespace-pre overflow-x-hidden"
      >
        {output.map((line) => collapsePathsInText(line, homeDir)).join('\n')}
      </pre>
      {isInProgress && (
        <ProgressRail className="absolute bottom-0 left-0 right-0 h-1 bg-surface rounded-b pointer-events-none">
          {hasProgress ? <ProgressBar percent={progressPercent} /> : <Shimmer />}
        </ProgressRail>
      )}
    </div>
  )
}

export function ProgressSteps({ steps, error, cancelled }: ProgressStepsProps) {
  const homeDir = useHomeDir()

  if (steps.length === 0 && !error && !cancelled) {
    return null
  }

  return (
    <div className="p-4 bg-surface-alt rounded-card">
      {steps.length > 0 && (
        <ol className="space-y-2">
          {steps.map((step) => {
            const subtitle = getStepSubtitle(step)
            const displaySubtitle = subtitle ? collapsePathsInText(subtitle, homeDir) : null
            const computedPercent =
              step.bytesTotal && step.bytesTotal > 0
                ? Math.min(100, Math.floor(((step.bytesDownloaded ?? 0) * 100) / step.bytesTotal))
                : undefined
            const progressPercent = step.progressPercent ?? computedPercent
            const showSpeed = step.id === 'build' && step.status === 'in_progress'
            const downloadSpeedBytes = getDownloadSpeedBytes(step)

            return (
              <li key={step.id}>
                <div
                  className="grid grid-cols-[18px_1fr] items-center gap-2 min-w-0"
                  title={step.message ? `${step.label} ${step.message}` : step.label}
                >
                  <span className="flex items-center justify-center w-[18px]">
                    <StepIcon status={step.status} />
                  </span>
                  <div className="flex items-center justify-between gap-3 min-w-0 flex-1">
                    <span className="text-on-surface-secondary truncate">
                      <span className="font-medium">{step.label}</span>
                      {displaySubtitle && <StepSubtitle>{displaySubtitle}</StepSubtitle>}
                    </span>
                    <SpeedBadge speedBytes={downloadSpeedBytes} show={showSpeed} />
                  </div>
                </div>
                {step.logEntries && step.logEntries.length > 0 && (
                  <div className="mt-1 ml-6">
                    <LogEntriesPre
                      logEntries={step.logEntries}
                      isInProgress={step.status === 'in_progress'}
                      homeDir={homeDir}
                      {...(step.id === 'build' && progressPercent !== undefined
                        ? { progressPercent }
                        : {})}
                    />
                  </div>
                )}
                {!(step.logEntries && step.logEntries.length > 0) &&
                  step.output &&
                  step.output.length > 0 && (
                    <div className="mt-1 ml-6">
                      <OutputPre
                        output={step.output}
                        isInProgress={step.status === 'in_progress'}
                        homeDir={homeDir}
                        {...(step.id === 'build' && progressPercent !== undefined
                          ? { progressPercent }
                          : {})}
                      />
                    </div>
                  )}
              </li>
            )
          })}
        </ol>
      )}

      {cancelled && (
        <div className="mt-4 p-3 bg-status-warning/10 text-status-warning rounded-sm border border-status-warning/30">
          Installation cancelled
        </div>
      )}

      {error && !cancelled && (
        <div className="mt-4 p-3 bg-status-error/10 text-status-error rounded-sm border border-status-error/30">
          {error}
        </div>
      )}
    </div>
  )
}
