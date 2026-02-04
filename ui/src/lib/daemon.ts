import type {
  BugReportInfo,
  ConfigResponse,
  DoctorResponse,
  InstallStatus,
  SetConfigRequest,
  StatusResponse,
  SyncAddDeviceRequest,
  SyncAddDeviceResponse,
  SyncRemoveDeviceRequest,
  SyncRemoveDeviceResponse,
  SyncStatusResponse,
  System,
  UninstallPreviewResponse,
  UninstallResponse,
} from '@/types/daemon'

type DaemonError = {
  readonly message: string
}

export type DaemonResult<T> =
  | { readonly ok: true; readonly data: T }
  | { readonly ok: false; readonly error: DaemonError }

async function invoke<T>(command: string, data?: unknown): Promise<DaemonResult<T>> {
  try {
    const result = (await window.electron.invoke(command, data)) as T
    return { ok: true, data: result }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    return { ok: false, error: { message } }
  }
}

export const getSystems = () => invoke<readonly System[]>('get_systems')

export const getConfig = () => invoke<ConfigResponse>('get_config')

export const setConfig = (config: SetConfigRequest) =>
  invoke<{ success: boolean }>('set_config', config)

export const getStatus = () => invoke<StatusResponse>('status')

export const runDoctor = () => invoke<DoctorResponse>('doctor')

export const apply = () => invoke<{ messages: readonly string[]; cancelled: boolean }>('apply')

export const cancelApply = () => invoke<{ cancelled: boolean }>('cancel_apply')

export const getInstallStatus = () => invoke<InstallStatus>('get_install_status')

export const installApp = () => invoke<{ success: boolean }>('install_app')

export const getSyncStatus = () => invoke<SyncStatusResponse>('sync_status')

export const addSyncDevice = (req: SyncAddDeviceRequest) =>
  invoke<SyncAddDeviceResponse>('sync_add_device', req)

export const removeSyncDevice = (req: SyncRemoveDeviceRequest) =>
  invoke<SyncRemoveDeviceResponse>('sync_remove_device', req)

export const getUninstallPreview = () => invoke<UninstallPreviewResponse>('uninstall_preview')

export const uninstall = () => invoke<UninstallResponse>('uninstall')

export const getBugReportInfo = () => invoke<BugReportInfo>('get_bug_report_info')

export const launchEmulator = (execLine: string) =>
  invoke<{ success: boolean }>('launch_emulator', execLine)

export const openPath = (path: string) => invoke<string>('open_path', path)

export const readFile = (path: string) => invoke<string>('read_file', path)
