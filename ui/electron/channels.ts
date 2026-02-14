// Daemon commands exposed via IPC. Not all CommandType values are exposed -
// some like 'uninstall' and 'install_kyaraben' are internal.
// Type safety is verified in src/lib/channel-check.ts against generated types.
const DAEMON_CHANNELS = [
  'get_systems',
  'get_frontends',
  'get_config',
  'set_config',
  'status',
  'doctor',
  'preflight',
  'apply',
  'cancel_apply',
  'sync_status',
  'sync_add_device',
  'sync_remove_device',
  'sync_start_pairing',
  'sync_join_primary',
  'sync_cancel_pairing',
  'sync_pause',
  'sync_resume',
  'sync_pending',
  'sync_enable',
  'uninstall_preview',
  'refresh_icon_caches',
] as const

// Electron-only commands handled by main.ts, not the daemon.
const ELECTRON_CHANNELS = [
  'get_install_status',
  'install_app',
  'open_path',
  'path_exists',
  'read_file',
  'get_bug_report_info',
  'launch_emulator',
  'open_log_tail',
  'launch_cli_uninstall',
  'check_for_updates',
  'download_update',
  'apply_update',
] as const

export const INVOKE_CHANNELS = [...DAEMON_CHANNELS, ...ELECTRON_CHANNELS] as const

export type InvokeChannel = (typeof INVOKE_CHANNELS)[number]

export const EVENT_CHANNELS = [
  'apply:progress',
  'pairing:progress',
  'sync_enable:progress',
  'update:progress',
] as const

export type EventChannel = (typeof EVENT_CHANNELS)[number]

export interface UpdateProgressEvent {
  percent: number
}
