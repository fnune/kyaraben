import type { ProgressStep, ProgressStepStatus } from '@/types/ui'

export interface ProgressDisplayProps {
  readonly steps: readonly ProgressStep[]
  readonly error?: string
}

function StepIcon({ status }: { readonly status: ProgressStepStatus }) {
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

export function ProgressDisplay({ steps, error }: ProgressDisplayProps) {
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
                {step.speed && (
                  <span
                    className="text-sm text-blue-600 ml-auto"
                    title="System-wide network activity"
                  >
                    {step.speed}
                  </span>
                )}
              </div>
              {step.outputLines && step.outputLines.length > 0 && (
                <div className="mt-1 ml-6 overflow-hidden">
                  <pre className="p-2 text-xs font-mono bg-gray-800 text-gray-200 rounded max-h-32 overflow-y-auto whitespace-pre overflow-x-hidden">
                    {step.outputLines.join('\n')}
                  </pre>
                </div>
              )}
            </li>
          ))}
        </ol>
      )}

      {error && (
        <div className="mt-4 p-3 bg-red-50 text-red-700 rounded border border-red-200">
          <strong>Error:</strong> {error}
        </div>
      )}
    </div>
  )
}
