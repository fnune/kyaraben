import { Settings } from '@/components/Settings/Settings'
import { SystemGrid } from '@/components/SystemGrid/SystemGrid'
import { Button } from '@/lib/Button'
import { ProgressSteps } from '@/lib/ProgressSteps'
import type { DoctorResponse, EmulatorID, System, SystemID } from '@/types/daemon'
import type { ApplyStatus, ProgressStep } from '@/types/ui'

export interface SystemsViewProps {
  readonly systems: readonly System[]
  readonly selections: Map<SystemID, EmulatorID>
  readonly provisions: DoctorResponse
  readonly userStore: string
  readonly onUserStoreChange: (value: string) => void
  readonly onToggle: (systemId: SystemID, enabled: boolean) => void
  readonly onApply: () => void
  readonly onCancel: () => void
  readonly onError: (message: string) => void
  readonly applyStatus: ApplyStatus
  readonly progressSteps: readonly ProgressStep[]
  readonly error: string | null
  readonly onReset: () => void
}

export function SystemsView({
  systems,
  selections,
  provisions,
  userStore,
  onUserStoreChange,
  onToggle,
  onApply,
  onCancel,
  onError,
  applyStatus,
  progressSteps,
  error,
  onReset,
}: SystemsViewProps) {
  const isApplying = applyStatus === 'applying'
  const showProgress = applyStatus !== 'idle'

  if (showProgress) {
    const errorMessage = applyStatus === 'error' && error ? error : undefined

    return (
      <div className="p-6">
        <ProgressSteps
          steps={progressSteps}
          {...(errorMessage && { error: errorMessage })}
          {...(applyStatus === 'cancelled' && { cancelled: true })}
        />
        <div className="flex gap-2">
          {isApplying && (
            <Button onClick={onCancel} variant="secondary">
              Cancel
            </Button>
          )}
          {!isApplying && <Button onClick={onReset}>Done</Button>}
        </div>
      </div>
    )
  }

  return (
    <div className="p-6">
      <Settings userStore={userStore} onUserStoreChange={onUserStoreChange} onError={onError} />

      <SystemGrid
        systems={systems}
        selections={selections}
        provisions={provisions}
        onToggle={onToggle}
      />

      <div className="mt-6">
        <Button onClick={onApply} disabled={selections.size === 0}>
          Apply
        </Button>
      </div>
    </div>
  )
}
