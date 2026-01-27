export type SystemID = 'snes' | 'psx' | 'tic80' | 'gba' | 'nds' | 'psp' | 'switch' | 'e2e-test'

export type EmulatorID =
  | 'retroarch:bsnes'
  | 'duckstation'
  | 'tic80'
  | 'retroarch:mgba'
  | 'retroarch:melonds'
  | 'retroarch:ppsspp'
  | 'eden'
  | 'e2e-test'

export type ProvisionKind = 'bios' | 'keys' | 'firmware'
export type ProvisionStatus = 'found' | 'missing' | 'invalid' | 'optional'

export interface EmulatorRef {
  readonly id: EmulatorID
  readonly name: string
}

export interface System {
  readonly id: SystemID
  readonly name: string
  readonly description: string
  readonly emulators: readonly EmulatorRef[]
}

export interface ProvisionResult {
  readonly filename: string
  readonly description: string
  readonly required: boolean
  readonly status: ProvisionStatus
  readonly foundPath?: string
  readonly kind?: ProvisionKind
}

export interface StatusResponse {
  readonly userStore: string
  readonly enabledSystems: readonly SystemID[]
  readonly installedEmulators: readonly {
    readonly id: EmulatorID
    readonly version: string
  }[]
  readonly lastApplied: string
}

export interface ConfigResponse {
  readonly userStore: string
  readonly systems: Readonly<Partial<Record<SystemID, EmulatorID>>>
}

export interface SetConfigRequest {
  readonly userStore: string
  readonly systems: Readonly<Partial<Record<SystemID, EmulatorID>>>
}

export type DoctorResponse = Readonly<Partial<Record<SystemID, readonly ProvisionResult[]>>>

export interface ProgressEvent {
  readonly step: string
  readonly message: string
}

export interface ApplyResult {
  readonly success: boolean
  readonly storePath: string
}

export interface InstallStatus {
  readonly installed: boolean
  readonly appPath?: string
}

export type CommandType =
  | 'status'
  | 'doctor'
  | 'apply'
  | 'get_systems'
  | 'get_config'
  | 'set_config'
  | 'get_install_status'
  | 'install_app'
  | 'uninstall_app'
