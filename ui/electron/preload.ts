import { contextBridge, ipcRenderer } from 'electron'

const INVOKE_CHANNELS = [
  'get_systems',
  'get_frontends',
  'get_config',
  'set_config',
  'status',
  'doctor',
  'preflight',
  'apply',
  'cancel_apply',
  'get_install_status',
  'install_app',
  'sync_status',
  'sync_add_device',
  'sync_remove_device',
  'uninstall_preview',
  'refresh_icon_caches',
  'open_path',
  'path_exists',
  'read_file',
  'get_bug_report_info',
  'launch_emulator',
  'open_log_tail',
  'launch_cli_uninstall',
]

const EVENT_CHANNELS = ['apply:progress']

contextBridge.exposeInMainWorld('electron', {
  invoke: (channel: string, ...args: unknown[]) => {
    if (!INVOKE_CHANNELS.includes(channel)) {
      throw new Error(`Invalid IPC channel: ${channel}`)
    }
    return ipcRenderer.invoke(channel, ...args)
  },

  on: (channel: string, callback: (...args: unknown[]) => void) => {
    if (!EVENT_CHANNELS.includes(channel)) return
    ipcRenderer.removeAllListeners(channel)
    ipcRenderer.on(channel, (_event, ...args) => callback(...args))
  },

  off: (channel: string) => {
    if (!EVENT_CHANNELS.includes(channel)) return
    ipcRenderer.removeAllListeners(channel)
  },
})
