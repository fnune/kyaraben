import { useEffect, useRef } from 'react'

export interface Step {
  id: string
  label: string
  status: 'pending' | 'in_progress' | 'completed' | 'error' | 'cancelled'
  message?: string
  output?: readonly string[]
  buildPhase?: string
  packageName?: string
  progressPercent?: number
}

export interface ProgressStepsProps {
  steps: readonly Step[]
  error?: string
  cancelled?: boolean
}

function formatBuildProgress(step: Step): string | null {
  switch (step.buildPhase) {
    case 'evaluating':
      return 'Preparing...'
    case 'installing':
      return step.packageName ? `Installing ${step.packageName}...` : null
    default:
      return null
  }
}

function Shimmer() {
  return (
    <div className="h-full w-1/3 bg-linear-to-r from-transparent via-accent to-transparent animate-[shimmer_1.5s_infinite]" />
  )
}

function ProgressBar({ percent }: { readonly percent: number }) {
  return (
    <>
      <div
        className="absolute inset-y-0 left-0 bg-accent rounded-full transition-all duration-300"
        style={{ width: `${percent}%` }}
      />
      <div className="absolute inset-0 overflow-hidden">
        <div className="h-full w-1/3 bg-linear-to-r from-transparent via-white/20 to-transparent animate-[shimmer_1.5s_infinite]" />
      </div>
    </>
  )
}

function StepIcon({ status }: { readonly status: Step['status'] }) {
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

function OutputPre({
  output,
  isInProgress,
  progressPercent,
}: {
  readonly output: readonly string[]
  readonly isInProgress: boolean
  readonly progressPercent?: number
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
        {output.join('\n')}
      </pre>
      {isInProgress && (
        <div className="absolute bottom-0 left-0 right-0 h-1 bg-surface rounded-b overflow-hidden pointer-events-none">
          {hasProgress ? <ProgressBar percent={progressPercent} /> : <Shimmer />}
        </div>
      )}
    </div>
  )
}

export function ProgressSteps({ steps, error, cancelled }: ProgressStepsProps) {
  if (steps.length === 0 && !error && !cancelled) {
    return null
  }

  return (
    <div className="mt-6 p-4 bg-surface-alt rounded-card">
      {steps.length > 0 && (
        <ol className="space-y-2">
          {steps.map((step) => {
            const buildProgress =
              step.id === 'build' && step.status === 'in_progress'
                ? formatBuildProgress(step)
                : null

            return (
              <li key={step.id}>
                <div
                  className="flex items-center gap-2 min-w-0"
                  title={step.message ? `${step.label} ${step.message}` : step.label}
                >
                  <StepIcon status={step.status} />
                  <span className="text-on-surface-secondary truncate">
                    <span className="font-medium">{step.label}</span>
                    {step.id === 'build' && step.status === 'completed' && (
                      <span className="text-sm"> - Done</span>
                    )}
                    {step.id === 'build' && step.status === 'in_progress' && buildProgress && (
                      <span className="text-sm"> - {buildProgress}</span>
                    )}
                    {step.id === 'build' &&
                      step.status === 'in_progress' &&
                      !buildProgress &&
                      step.message && <span className="text-sm"> {step.message}</span>}
                    {step.id !== 'build' && step.message && (
                      <span className="text-sm"> {step.message}</span>
                    )}
                  </span>
                </div>
                {step.output && step.output.length > 0 && (
                  <div className="mt-1 ml-6">
                    <OutputPre
                      output={step.output}
                      isInProgress={step.status === 'in_progress'}
                      {...(step.id === 'build' && step.progressPercent !== undefined
                        ? { progressPercent: step.progressPercent }
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
