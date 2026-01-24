import { contextBridge, ipcRenderer } from 'electron'

// Expose IPC methods to the renderer process
contextBridge.exposeInMainWorld('electron', {
  invoke: (channel: string, ...args: unknown[]) => {
    // Whitelist of allowed channels
    const validChannels = [
      'get_systems',
      'get_config',
      'set_config',
      'status',
      'doctor',
      'apply',
      'get_install_status',
      'install_app',
      'uninstall_app',
    ]

    if (validChannels.includes(channel)) {
      return ipcRenderer.invoke(channel, ...args)
    }

    throw new Error(`Invalid IPC channel: ${channel}`)
  },
})
