import type { EmulatorID, ProvisionResult, SystemID } from './daemon'
import type { Manufacturer } from './model.gen'

export type View = 'systems' | 'installation' | 'sync'

export const MANUFACTURER_ORDER: readonly Manufacturer[] = ['Nintendo', 'Sony', 'Sega', 'Other']

export interface EmulatorProvisions {
  readonly emulatorId: EmulatorID
  readonly emulatorName: string
  readonly provisions: readonly ProvisionResult[]
  readonly hasRequiredMissing: boolean
  readonly hasOptionalMissing: boolean
}

export interface SystemSelection {
  readonly systemId: SystemID
  readonly emulatorId: EmulatorID
  readonly enabled: boolean
}

export type ProgressStepStatus = 'pending' | 'in_progress' | 'completed' | 'error' | 'cancelled'

export interface ProgressStep {
  readonly id: string
  readonly label: string
  readonly status: ProgressStepStatus
  readonly message?: string
  readonly output?: readonly string[]
}

export type ApplyStatus = 'idle' | 'reviewing' | 'applying' | 'success' | 'error' | 'cancelled'

export interface ApplyState {
  readonly status: ApplyStatus
  readonly steps: readonly ProgressStep[]
  readonly error?: string
}
