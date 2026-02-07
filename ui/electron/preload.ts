import { contextBridge, ipcRenderer } from 'electron'

contextBridge.exposeInMainWorld('electron', {
  invoke: (channel: string, ...args: unknown[]) => ipcRenderer.invoke(channel, ...args),

  on: (channel: string, callback: (...args: unknown[]) => void) => {
    ipcRenderer.removeAllListeners(channel)
    ipcRenderer.on(channel, (_event, ...args) => callback(...args))
  },

  off: (channel: string) => ipcRenderer.removeAllListeners(channel),
})
