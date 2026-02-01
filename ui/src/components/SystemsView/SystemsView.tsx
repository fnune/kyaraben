import { Settings } from '@/components/Settings/Settings'
import { SystemList } from '@/components/SystemList/SystemList'
import { Button } from '@/lib/Button'
import { ProgressSteps } from '@/lib/ProgressSteps'
import type { DoctorResponse, EmulatorID, System, SystemID } from '@/types/daemon'
import type { ApplyStatus, ProgressStep } from '@/types/ui'

export interface SystemsViewProps {
  readonly systems: readonly System[]
  readonly selections: Map<SystemID, EmulatorID>
  readonly versionSelections: Map<SystemID, string | null>
  readonly installedVersions: Map<EmulatorID, string>
  readonly provisions: DoctorResponse
  readonly userStore: string
  readonly onUserStoreChange: (value: string) => void
  readonly onToggle: (systemId: SystemID, enabled: boolean) => void
  readonly onVersionChange: (systemId: SystemID, version: string | null) => void
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
  versionSelections,
  installedVersions,
  provisions,
  userStore,
  onUserStoreChange,
  onToggle,
  onVersionChange,
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

      <SystemList
        systems={systems}
        selections={selections}
        versionSelections={versionSelections}
        installedVersions={installedVersions}
        provisions={provisions}
        userStore={userStore}
        onToggle={onToggle}
        onEmulatorChange={() => undefined}
        onVersionChange={onVersionChange}
      />

      <div className="mt-6">
        <Button onClick={onApply} disabled={selections.size === 0}>
          Apply
        </Button>
      </div>
    </div>
  )
}
