import type { InvokeChannel } from '@electron/channels'
import type {
  BugReportInfo,
  ConfigResponse,
  DoctorResponse,
  FrontendRef,
  InstallStatus,
  PreflightResponse,
  SetConfigRequest,
  StatusResponse,
  StorageDevicesResponse,
  SyncDiscoveredDevicesResponse,
  SyncEnableRequest,
  SyncEnableResponse,
  SyncJoinPrimaryRequest,
  SyncJoinPrimaryResponse,
  SyncLocalChangesRequest,
  SyncLocalChangesResponse,
  SyncPendingResponse,
  SyncRemoveDeviceRequest,
  SyncRemoveDeviceResponse,
  SyncResetResponse,
  SyncRevertFolderRequest,
  SyncRevertFolderResponse,
  SyncSetSettingsRequest,
  SyncSetSettingsResponse,
  SyncStartPairingResponse,
  SyncStatusResponse,
  System,
  UninstallPreviewResponse,
} from '@/types/daemon'

type DaemonError = {
  readonly message: string
}

export type DaemonResult<T> =
  | { readonly ok: true; readonly data: T }
  | { readonly ok: false; readonly error: DaemonError }

async function invoke<T>(command: InvokeChannel, data?: unknown): Promise<DaemonResult<T>> {
  try {
    const result = (await window.electron.invoke(command, data)) as T
    return { ok: true, data: result }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    return { ok: false, error: { message } }
  }
}

export const getSystems = () => invoke<readonly System[]>('get_systems')

export const getFrontends = () => invoke<readonly FrontendRef[]>('get_frontends')

export const getConfig = () => invoke<ConfigResponse>('get_config')

export const setConfig = (config: SetConfigRequest) =>
  invoke<{ success: boolean }>('set_config', config)

export const getStatus = () => invoke<StatusResponse>('status')

export const runDoctor = () => invoke<DoctorResponse>('doctor')

export const preflight = () => invoke<PreflightResponse>('preflight')

export const apply = () => invoke<{ messages: readonly string[]; cancelled: boolean }>('apply')

export const cancelApply = () => invoke<{ cancelled: boolean }>('cancel_apply')

export const getInstallStatus = () => invoke<InstallStatus>('get_install_status')

export const installApp = () => invoke<{ success: boolean }>('install_app')

export const getSyncStatus = () => invoke<SyncStatusResponse>('sync_status')

export const removeSyncDevice = (req: SyncRemoveDeviceRequest) =>
  invoke<SyncRemoveDeviceResponse>('sync_remove_device', req)

export const startSyncPairing = () => invoke<SyncStartPairingResponse>('sync_start_pairing')

export const joinSyncPrimary = (req: SyncJoinPrimaryRequest) =>
  invoke<SyncJoinPrimaryResponse>('sync_join_primary', req)

export const cancelSyncPairing = () => invoke<{ cancelled: boolean }>('sync_cancel_pairing')

export const getSyncPending = () => invoke<SyncPendingResponse>('sync_pending')

export const enableSync = (req: SyncEnableRequest) => invoke<SyncEnableResponse>('sync_enable', req)

export const revertSyncFolder = (req: SyncRevertFolderRequest) =>
  invoke<SyncRevertFolderResponse>('sync_revert_folder', req)

export const getSyncLocalChanges = (req: SyncLocalChangesRequest) =>
  invoke<SyncLocalChangesResponse>('sync_local_changes', req)

export const resetSync = () => invoke<SyncResetResponse>('sync_reset')

export const getDiscoveredDevices = () =>
  invoke<SyncDiscoveredDevicesResponse>('sync_discovered_devices')

export const setSyncSettings = (req: SyncSetSettingsRequest) =>
  invoke<SyncSetSettingsResponse>('sync_set_settings', req)

export const getUninstallPreview = () => invoke<UninstallPreviewResponse>('uninstall_preview')

export const getStorageDevices = () => invoke<StorageDevicesResponse>('get_storage_devices')

export const selectDirectory = () => invoke<string | null>('select_directory')

export const getBugReportInfo = () => invoke<BugReportInfo>('get_bug_report_info')

export const launchEmulator = (execLine: string) =>
  invoke<{ success: boolean }>('launch_emulator', execLine)

export const openPath = (path: string) => invoke<string>('open_path', path)

export const openUrl = (url: string) => invoke<string>('open_url', url)

export const readFile = (path: string) => invoke<string>('read_file', path)

export const openLogTail = (position?: number) =>
  invoke<{ success: boolean; error?: string; command?: string }>('open_log_tail', position)

export const launchCliUninstall = () =>
  invoke<{ success: boolean; error?: string }>('launch_cli_uninstall')

export interface UpdateInfo {
  available: boolean
  currentVersion: string
  latestVersion: string
  downloadUrl: string
  releaseNotes?: string
}

export const checkForUpdates = () => invoke<UpdateInfo>('check_for_updates')

export const downloadUpdate = (url: string) =>
  invoke<{ success: boolean; path?: string; error?: string }>('download_update', url)

export const applyUpdate = (tempPath: string) =>
  invoke<{ success: boolean; error?: string }>('apply_update', tempPath)
