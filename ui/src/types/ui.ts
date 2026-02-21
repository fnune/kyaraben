import type { EmulatorID, ProvisionResult, SystemID } from './daemon'
import type { LogEntry } from './logging.gen'
import type { Manufacturer } from './model.gen'

export const VIEW_CATALOG = 'catalog' as const
export const VIEW_INSTALLATION = 'installation' as const
export const VIEW_SYNC = 'sync' as const
export type View = typeof VIEW_CATALOG | typeof VIEW_INSTALLATION | typeof VIEW_SYNC

export const VIEW_LABELS: Record<View, string> = {
  [VIEW_CATALOG]: 'Catalog',
  [VIEW_INSTALLATION]: 'Installation',
  [VIEW_SYNC]: 'Synchronization',
}

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
  readonly logEntries?: readonly LogEntry[]
  readonly buildPhase?: string
  readonly packageName?: string
  readonly progressPercent?: number
  readonly bytesDownloaded?: number
  readonly bytesTotal?: number
  readonly bytesPerSecond?: number
}

export type ApplyStatus =
  | 'idle'
  | 'reviewing'
  | 'confirming_sync'
  | 'applying'
  | 'success'
  | 'error'
  | 'cancelled'

export interface ApplyState {
  readonly status: ApplyStatus
  readonly steps: readonly ProgressStep[]
  readonly error?: string
}
