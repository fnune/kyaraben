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
      return <span className="text-green-400">✓</span>
    case 'in_progress':
      return <span className="text-blue-400 animate-pulse">●</span>
    case 'error':
      return <span className="text-red-400">✗</span>
    case 'cancelled':
      return <span className="text-amber-400">⊘</span>
    default:
      return <span className="text-gray-500">○</span>
  }
}

export function ProgressSteps({ steps, error, cancelled }: ProgressStepsProps) {
  if (steps.length === 0 && !error && !cancelled) {
    return null
  }

  return (
    <div className="mt-6 p-4 bg-gray-800 rounded-lg">
      {steps.length > 0 && (
        <ol className="space-y-2">
          {steps.map((step) => (
            <li key={step.id}>
              <div className="flex items-center gap-2">
                <StepIcon status={step.status} />
                <span className="font-medium text-gray-300">{step.label}</span>
                {step.message && <span className="text-sm text-gray-500">{step.message}</span>}
              </div>
              {step.output && step.output.length > 0 && (
                <div className="mt-1 ml-6 relative">
                  <pre className="p-2 text-xs font-mono bg-gray-900 text-gray-300 rounded max-h-32 overflow-y-auto whitespace-pre overflow-x-hidden">
                    {step.output.join('\n')}
                  </pre>
                  {step.status === 'in_progress' && (
                    <div className="absolute bottom-0 left-0 right-0 h-1 bg-gray-900 rounded-b overflow-hidden">
                      <div className="h-full w-1/3 bg-gradient-to-r from-transparent via-blue-500 to-transparent animate-[shimmer_1.5s_infinite]" />
                    </div>
                  )}
                </div>
              )}
            </li>
          ))}
        </ol>
      )}

      {cancelled && (
        <div className="mt-4 p-3 bg-amber-500/10 text-amber-400 rounded border border-amber-500/30">
          Installation cancelled
        </div>
      )}

      {error && !cancelled && (
        <div className="mt-4 p-3 bg-red-500/10 text-red-400 rounded border border-red-500/30">
          {error}
        </div>
      )}
    </div>
  )
}
