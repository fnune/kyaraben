import type { CommandType } from './daemon.gen'

export type { SystemWithEmulators as System } from './daemon.gen'
export * from './daemon.gen'
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
