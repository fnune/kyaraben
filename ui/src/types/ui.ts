import type { EmulatorID, ProvisionResult, System, SystemID } from './daemon'

export type Manufacturer = 'Nintendo' | 'Sony' | 'Other'

export const SYSTEM_MANUFACTURERS: Readonly<Record<SystemID, Manufacturer>> = {
  snes: 'Nintendo',
  gba: 'Nintendo',
  nds: 'Nintendo',
  switch: 'Nintendo',
  psx: 'Sony',
  psp: 'Sony',
  tic80: 'Other',
  'e2e-test': 'Other',
}

export const MANUFACTURER_ORDER: readonly Manufacturer[] = ['Nintendo', 'Sony', 'Other']

export interface SystemWithMetadata extends System {
  readonly manufacturer: Manufacturer
}

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

export type ProgressStepStatus = 'pending' | 'in_progress' | 'completed' | 'error'

export interface ProgressStep {
  readonly id: string
  readonly label: string
  readonly status: ProgressStepStatus
  readonly message?: string
  readonly outputLines?: readonly string[]
}

export type ApplyStatus = 'idle' | 'applying' | 'success' | 'error'

export interface ApplyState {
  readonly status: ApplyStatus
  readonly steps: readonly ProgressStep[]
  readonly error?: string
}
