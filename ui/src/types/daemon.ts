import type { CommandType } from './daemon.gen'

// Re-export SystemWithEmulators as System since that's what the daemon returns
// and what the UI uses. The model.gen System is just the basic type without emulators.
export type {
  ApplyResult,
  CancelledResponse,
  Command,
  CommandType,
  ConfigResponse,
  DoctorResponse,
  EmulatorRef,
  ErrorResponse,
  Event,
  EventType,
  GetSystemsResponse,
  InstalledEmulator,
  PreservedPaths,
  ProgressEvent,
  ProvisionResult,
  SetConfigCommand,
  SetConfigRequest,
  SetConfigResponse,
  StatusResponse,
  SyncAddDeviceCommand,
  SyncAddDeviceRequest,
  SyncAddDeviceResponse,
  SyncDevice,
  SyncRemoveDeviceCommand,
  SyncRemoveDeviceRequest,
  SyncRemoveDeviceResponse,
  SyncState,
  SyncStatusResponse,
  SystemConf,
  SystemWithEmulators,
  SystemWithEmulators as System,
  UninstallPreviewResponse,
} from './daemon.gen'
export {
  CommandTypeApply,
  CommandTypeCancelApply,
  CommandTypeDoctor,
  CommandTypeGetConfig,
  CommandTypeGetSystems,
  CommandTypeSetConfig,
  CommandTypeStatus,
  CommandTypeSyncAddDevice,
  CommandTypeSyncRemoveDevice,
  CommandTypeSyncStatus,
  CommandTypeUninstallPreview,
  EventTypeCancelled,
  EventTypeError,
  EventTypeProgress,
  EventTypeReady,
  EventTypeResult,
  SyncStateConflict,
  SyncStateDisabled,
  SyncStateDisconnected,
  SyncStateError,
  SyncStateSynced,
  SyncStateSyncing,
} from './daemon.gen'
export type { EmulatorID, Manufacturer, SystemID } from './model.gen'

export type ElectronOnlyCommand = 'get_install_status' | 'install_app' | 'uninstall_app'
export type DaemonCommandType = CommandType | ElectronOnlyCommand

export type ProvisionKind = 'bios' | 'keys' | 'firmware'
export type ProvisionStatus = 'found' | 'missing' | 'invalid' | 'optional'

export interface InstallStatus {
  readonly installed: boolean
  readonly appPath?: string
}

export type SyncMode = 'primary' | 'secondary'

export interface StateDirInfo {
  readonly exists: boolean
  readonly manifestExists: boolean
  readonly flakeExists: boolean
  readonly brokenSymlinks: readonly string[]
}

export interface BugReportInfo {
  readonly appVersion: string
  readonly platform: string
  readonly arch: string
  readonly osRelease: string
  readonly stateDir: StateDirInfo
}
