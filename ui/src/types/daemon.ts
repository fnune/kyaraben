import type { CommandType } from './daemon.gen'

// Re-export SystemWithEmulators as System since that's what the daemon returns
// and what the UI uses. The model.gen System is just the basic type without emulators.
export type {
  ApplyResult,
  CancelledResponse,
  Command,
  CommandType,
  ConfigChangeDetail,
  ConfigFileDiff,
  ConfigResponse,
  DoctorResponse,
  EmulatorPaths,
  EmulatorRef,
  ErrorResponse,
  Event,
  EventType,
  FrontendConfRequest,
  FrontendConfResponse,
  FrontendRef,
  GetFrontendsResponse,
  GetSystemsResponse,
  InstalledEmulator,
  ManagedConfigInfo,
  ManagedKeyInfo,
  PreflightResponse,
  PreservedPaths,
  ProgressEvent,
  ProvisionResult,
  SetConfigCommand,
  SetConfigRequest,
  SetConfigResponse,
  StatusResponse,
  SyncDevice,
  SyncEnableProgressEvent,
  SyncEnableRequest,
  SyncEnableResponse,
  SyncFolder,
  SyncJoinPrimaryRequest,
  SyncJoinPrimaryResponse,
  SyncPendingResponse,
  SyncProgress,
  SyncRemoveDeviceCommand,
  SyncRemoveDeviceRequest,
  SyncRemoveDeviceResponse,
  SyncState,
  SyncStatusResponse,
  SystemWithEmulators,
  SystemWithEmulators as System,
  UninstallPreviewResponse,
  UninstallResponse,
  UserChangeDetail,
} from './daemon.gen'
export {
  CommandTypeApply,
  CommandTypeCancelApply,
  CommandTypeDoctor,
  CommandTypeGetConfig,
  CommandTypeGetFrontends,
  CommandTypeGetSystems,
  CommandTypePreflight,
  CommandTypeSetConfig,
  CommandTypeStatus,
  CommandTypeSyncEnable,
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
export type { EmulatorID, FrontendID, Manufacturer, SystemID } from './model.gen'
export { FrontendIDESDE } from './model.gen'

export type ElectronOnlyCommand = 'get_install_status' | 'install_app'
export type DaemonCommandType = CommandType | ElectronOnlyCommand

export type ProvisionKind = 'bios' | 'keys' | 'firmware'
export type ProvisionStatus = 'found' | 'missing' | 'invalid' | 'optional'

export interface InstallStatus {
  readonly installed: boolean
  readonly appPath?: string
  readonly desktopPath?: string
  readonly cliPath?: string
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
