import { useEffect, useRef } from 'react'

export interface Step {
  id: string
  label: string
  status: 'pending' | 'in_progress' | 'completed' | 'error' | 'cancelled'
  message?: string
  output?: readonly string[]
}

export interface ProgressStepsProps {
  steps: readonly Step[]
  error?: string
  cancelled?: boolean
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
  showShimmer,
}: {
  readonly output: readonly string[]
  readonly showShimmer: boolean
}) {
  const preRef = useRef<HTMLPreElement>(null)

  // biome-ignore lint/correctness/useExhaustiveDependencies: scroll when output changes
  useEffect(() => {
    if (preRef.current) {
      preRef.current.scrollTop = preRef.current.scrollHeight
    }
  }, [output])

  return (
    <div className="relative">
      <pre
        ref={preRef}
        className="p-2 text-xs font-mono bg-surface text-on-surface-secondary rounded-sm max-h-32 overflow-y-auto whitespace-pre overflow-x-hidden"
      >
        {output.join('\n')}
      </pre>
      {showShimmer && (
        <div className="absolute bottom-0 left-0 right-0 h-1 bg-surface rounded-b overflow-hidden pointer-events-none">
          <div className="h-full w-1/3 bg-linear-to-r from-transparent via-accent to-transparent animate-[shimmer_1.5s_infinite]" />
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
          {steps.map((step) => (
            <li key={step.id}>
              <div
                className="flex items-center gap-2 min-w-0"
                title={step.message ? `${step.label} ${step.message}` : step.label}
              >
                <StepIcon status={step.status} />
                <span className="text-on-surface-secondary truncate">
                  <span className="font-medium">{step.label}</span>
                  {step.message && <span className="text-sm"> {step.message}</span>}
                </span>
              </div>
              {step.output && step.output.length > 0 && (
                <div className="mt-1 ml-6">
                  <OutputPre output={step.output} showShimmer={step.status === 'in_progress'} />
                </div>
              )}
            </li>
          ))}
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
