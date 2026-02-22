import * as os from 'node:os'
import { contextBridge, ipcRenderer } from 'electron'

contextBridge.exposeInMainWorld('electron', {
  homeDir: os.homedir(),

  invoke: (channel: string, ...args: unknown[]) => ipcRenderer.invoke(channel, ...args),

  on: (channel: string, callback: (data: unknown) => void) => {
    ipcRenderer.removeAllListeners(channel)
    ipcRenderer.on(channel, (_event, data) => callback(data))
    return () => ipcRenderer.removeAllListeners(channel)
  },

  off: (channel: string) => ipcRenderer.removeAllListeners(channel),
})
