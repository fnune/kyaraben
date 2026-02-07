/// <reference types="vite/client" />

type CommandType =
  | 'status'
  | 'doctor'
  | 'apply'
  | 'get_systems'
  | 'get_config'
  | 'set_config'
  | 'get_install_status'
  | 'install_app'
  | 'sync_status'
  | 'sync_add_device'
  | 'sync_remove_device'

type EventChannel = 'apply:progress'

interface ElectronAPI {
  invoke<T>(command: CommandType, data?: unknown): Promise<T>
  on(channel: EventChannel, callback: (...args: unknown[]) => void): void
  off(channel: EventChannel): void
}

interface Window {
  electron: ElectronAPI
}
