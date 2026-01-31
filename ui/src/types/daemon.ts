import type { DaemonCommandType } from './generated'

export type ElectronOnlyCommand = 'get_install_status' | 'install_app' | 'uninstall_app'
export type CommandType = DaemonCommandType | ElectronOnlyCommand

export type SystemID =
  | 'snes'
  | 'psx'
  | 'ps2'
  | 'ps3'
  | 'psvita'
  | 'psp'
  | 'gba'
  | 'nds'
  | '3ds'
  | 'dreamcast'
  | 'gamecube'
  | 'wii'
  | 'wiiu'
  | 'switch'
  | 'e2e-test'

export type EmulatorID =
  | 'retroarch'
  | 'retroarch:bsnes'
  | 'duckstation'
  | 'pcsx2'
  | 'rpcs3'
  | 'vita3k'
  | 'ppsspp'
  | 'mgba'
  | 'melonds'
  | 'flycast'
  | 'cemu'
  | 'azahar'
  | 'dolphin'
  | 'eden'
  | 'e2e-test'

export type ProvisionKind = 'bios' | 'keys' | 'firmware'
export type ProvisionStatus = 'found' | 'missing' | 'invalid' | 'optional'

export interface EmulatorRef {
  readonly id: EmulatorID
  readonly name: string
  readonly defaultVersion?: string
  readonly availableVersions?: readonly string[]
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

export interface SystemConfigEntry {
  readonly emulator: EmulatorID
  readonly pinnedVersion?: string
}

export interface ConfigResponse {
  readonly userStore: string
  readonly systems: Readonly<Partial<Record<SystemID, SystemConfigEntry>>>
}

export interface SetConfigRequest {
  readonly userStore: string
  readonly systems: Readonly<Partial<Record<SystemID, string>>> // "emulator" or "emulator@version"
}

export type DoctorResponse = Readonly<Partial<Record<SystemID, readonly ProvisionResult[]>>>

export interface ProgressEvent {
  readonly step: string
  readonly message: string
  readonly output?: string
}

export interface ApplyResult {
  readonly success: boolean
  readonly storePath: string
}

export interface InstallStatus {
  readonly installed: boolean
  readonly appPath?: string
}

export type SyncMode = 'primary' | 'secondary'
export type SyncState = 'disabled' | 'synced' | 'syncing' | 'disconnected' | 'conflict' | 'error'

export interface SyncDevice {
  readonly id: string
  readonly name: string
  readonly connected: boolean
}

export interface SyncStatusResponse {
  readonly enabled: boolean
  readonly mode?: SyncMode
  readonly running?: boolean
  readonly deviceId?: string
  readonly guiURL?: string
  readonly state?: SyncState
  readonly devices?: readonly SyncDevice[]
}

export interface SyncAddDeviceRequest {
  readonly deviceId: string
  readonly name?: string
}

export interface SyncAddDeviceResponse {
  readonly success: boolean
  readonly deviceId: string
  readonly name: string
}

export interface SyncRemoveDeviceRequest {
  readonly deviceId: string
}

export interface SyncRemoveDeviceResponse {
  readonly success: boolean
  readonly deviceId: string
  readonly name: string
}

export interface UninstallPreviewResponse {
  readonly stateDir: string
  readonly stateDirExists: boolean
  readonly desktopFiles: readonly string[]
  readonly iconFiles: readonly string[]
  readonly configFiles: readonly string[]
  readonly preserved: {
    readonly userStore: string
    readonly configDir: string
  }
}
