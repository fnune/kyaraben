import { contextBridge, ipcRenderer } from 'electron'

// Expose IPC methods to the renderer process
contextBridge.exposeInMainWorld('electron', {
  invoke: (channel: string, ...args: unknown[]) => {
    const validChannels = [
      'get_systems',
      'get_config',
      'set_config',
      'status',
      'doctor',
      'apply',
      'cancel_apply',
      'get_install_status',
      'install_app',
      'uninstall_app',
      'sync_status',
      'sync_add_device',
      'sync_remove_device',
      'uninstall_preview',
      'open_path',
      'path_exists',
    ]

    if (validChannels.includes(channel)) {
      return ipcRenderer.invoke(channel, ...args)
    }

    throw new Error(`Invalid IPC channel: ${channel}`)
  },

  on: (channel: string, callback: (...args: unknown[]) => void) => {
    const validChannels = ['apply:progress']
    if (validChannels.includes(channel)) {
      ipcRenderer.on(channel, (_event, ...args) => callback(...args))
    }
  },

  off: (channel: string, callback: (...args: unknown[]) => void) => {
    const validChannels = ['apply:progress']
    if (validChannels.includes(channel)) {
      ipcRenderer.removeListener(channel, callback)
    }
  },
})
