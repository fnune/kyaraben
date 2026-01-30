export interface Step {
  id: string
  label: string
  status: 'pending' | 'in_progress' | 'completed' | 'error'
  message?: string
  output?: readonly string[]
}

export interface ProgressStepsProps {
  steps: readonly Step[]
  error?: string | undefined
}

function StepIcon({ status }: { readonly status: Step['status'] }) {
  switch (status) {
    case 'completed':
      return <span className="text-green-600">✓</span>
    case 'in_progress':
      return <span className="text-blue-600 animate-pulse">●</span>
    case 'error':
      return <span className="text-red-600">✗</span>
    default:
      return <span className="text-gray-400">○</span>
  }
}

export function ProgressSteps({ steps, error }: ProgressStepsProps) {
  if (steps.length === 0 && !error) {
    return null
  }

  return (
    <div className="mt-6 p-4 bg-gray-50 rounded-lg">
      {steps.length > 0 && (
        <ol className="space-y-2">
          {steps.map((step) => (
            <li key={step.id}>
              <div className="flex items-center gap-2">
                <StepIcon status={step.status} />
                <span className="font-medium text-gray-700">{step.label}</span>
                {step.message && <span className="text-sm text-gray-500">{step.message}</span>}
              </div>
              {step.output && step.output.length > 0 && (
                <div className="mt-1 ml-6 overflow-hidden">
                  <pre className="p-2 text-xs font-mono bg-gray-800 text-gray-200 rounded max-h-32 overflow-y-auto whitespace-pre overflow-x-hidden">
                    {step.output.join('\n')}
                  </pre>
                </div>
              )}
            </li>
          ))}
        </ol>
      )}

      {error && (
        <div className="mt-4 p-3 bg-red-50 text-red-700 rounded border border-red-200">{error}</div>
      )}
    </div>
  )
}
