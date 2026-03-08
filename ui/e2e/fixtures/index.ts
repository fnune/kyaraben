import * as fs from 'node:fs'
import type * as http from 'node:http'
import * as os from 'node:os'
import * as path from 'node:path'
import { type ElectronApplication, _electron as electron, type Page } from '@playwright/test'
import { startFakeReleasesServer } from './fake-releases-server'
import {
  type Device,
  FakeSyncthingController,
  type Folder,
  startFakeSyncthingServer,
} from './fake-syncthing-server'
import { type RelayServer, startRelayServer } from './relay-server'

export { type RelayServer, startRelayServer }

function buildEnv(fixture: TestFixture): Record<string, string> {
  const env: Record<string, string> = {}
  for (const [key, value] of Object.entries(process.env)) {
    if (value !== undefined) {
      env[key] = value
    }
  }
  return { ...env, ...fixture.env }
}

function getElectronArgs(): string[] {
  const args: string[] = []
  if (process.env.DISPLAY?.startsWith(':')) {
    args.push('--ozone-platform=x11')
  }
  if (process.env.CI) {
    args.push('--no-sandbox')
  }
  return args
}

export function getAppImagePath(): string {
  const appImagePath = process.env.KYARABEN_APPIMAGE
  if (!appImagePath) {
    throw new Error('KYARABEN_APPIMAGE environment variable must be set')
  }
  return appImagePath
}

export async function launchElectron(
  fixture: TestFixture,
): Promise<{ app: ElectronApplication; page: Page }> {
  const app = await electron.launch({
    executablePath: getAppImagePath(),
    args: getElectronArgs(),
    env: buildEnv(fixture),
  })

  const page = await app.firstWindow()
  await page.getByRole('img', { name: 'Kyaraben' }).waitFor({ timeout: 30000 })

  const dismissButton = page.getByRole('button', { name: 'Dismiss' }).first()
  if (await dismissButton.isVisible({ timeout: 1000 }).catch(() => false)) {
    await dismissButton.click()
  }

  return { app, page }
}

export const SystemIDSNES = 'snes' as const
export const SystemIDGBA = 'gba' as const
export const SystemIDPSX = 'psx' as const
export const EmulatorIDRetroArchBsnes = 'retroarch:bsnes' as const
export const EmulatorIDRetroArchMGBA = 'retroarch:mgba' as const
export const EmulatorIDDuckStation = 'duckstation' as const
export const FrontendIDESDE = 'esde' as const

export type SystemID = typeof SystemIDSNES | typeof SystemIDGBA | typeof SystemIDPSX | string
export type EmulatorID =
  | typeof EmulatorIDRetroArchBsnes
  | typeof EmulatorIDRetroArchMGBA
  | typeof EmulatorIDDuckStation
  | string
export type FrontendID = typeof FrontendIDESDE | string

export interface TestFixtureEnv extends Record<string, string> {
  XDG_CONFIG_HOME: string
  XDG_STATE_HOME: string
  XDG_DATA_HOME: string
  HOME: string
  KYARABEN_E2E_FAKE_INSTALLER: string
  KYARABEN_RELAY_HEALTH_RETRIES: string
}

export interface TestFixture {
  configDir: string
  stateDir: string
  collection: string
  env: TestFixtureEnv
  cleanup: () => void
  releasesServer?: http.Server
  syncthingServer?: http.Server
}

export interface ConfigFixture {
  collection?: string
  systems?: Partial<Record<SystemID, EmulatorID[]>>
  emulators?: Partial<Record<EmulatorID, { version?: string; shaders?: boolean | null }>>
  frontends?: Partial<Record<FrontendID, { enabled: boolean; version?: string }>>
  sync?: {
    enabled?: boolean
    relayUrl?: string
    devices?: Array<{ id: string; name: string }>
  }
}

export interface InstalledEmulatorFixture {
  id: EmulatorID
  version: string
  storePath: string
  installed: string
}

export interface SyncthingInstallFixture {
  version?: string
  configSchemaVersion?: number
  binaryPath?: string
  configDir?: string
  dataDir?: string
  systemdUnitPath?: string
}

export interface ManifestFixture {
  version?: number
  kyarabenVersion?: string
  lastApplied?: string
  installedEmulators?: Partial<Record<EmulatorID, InstalledEmulatorFixture>>
  managedConfigs?: Array<{
    emulatorId: EmulatorID
    target: { type: string; path?: string }
    writtenEntries?: Record<string, string>
    configInputsWhenWritten?: Record<string, string>
    lastModified: string
  }>
  desktopFiles?: string[]
  iconFiles?: string[]
  syncthingInstall?: SyncthingInstallFixture
}

export function createFixture(config?: ConfigFixture, manifest?: ManifestFixture): TestFixture {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'kyaraben-e2e-'))

  const configDir = path.join(tmpDir, 'config')
  const stateDir = path.join(tmpDir, 'state')
  const collection = path.join(tmpDir, 'Emulation')

  fs.mkdirSync(path.join(configDir, 'kyaraben'), { recursive: true })
  fs.mkdirSync(path.join(stateDir, 'kyaraben', 'build'), { recursive: true })
  fs.mkdirSync(path.join(stateDir, 'kyaraben', 'bin'), { recursive: true })
  fs.mkdirSync(collection, { recursive: true })

  if (config) {
    const toml = generateConfigToml(config, collection)
    fs.writeFileSync(path.join(configDir, 'kyaraben', 'config.toml'), toml)
  }

  if (manifest) {
    const manifestJson: Record<string, unknown> = {
      version: manifest.version ?? 1,
      last_applied: manifest.lastApplied ?? new Date().toISOString(),
      installed_emulators: manifest.installedEmulators ?? {},
      managed_configs: manifest.managedConfigs ?? [],
      desktop_files: manifest.desktopFiles ?? [],
      icon_files: manifest.iconFiles ?? [],
    }
    if (manifest.kyarabenVersion) {
      manifestJson.kyaraben_version = manifest.kyarabenVersion
    }
    if (manifest.syncthingInstall) {
      manifestJson.syncthing_install = {
        version: manifest.syncthingInstall.version ?? '1.27.0',
        config_schema_version: manifest.syncthingInstall.configSchemaVersion ?? 1,
        binary_path: manifest.syncthingInstall.binaryPath,
        config_dir: manifest.syncthingInstall.configDir,
        data_dir: manifest.syncthingInstall.dataDir,
        systemd_unit_path: manifest.syncthingInstall.systemdUnitPath,
      }
    }
    fs.writeFileSync(
      path.join(stateDir, 'kyaraben', 'build', 'manifest.json'),
      JSON.stringify(manifestJson, null, 2),
    )
  }

  const dataDir = path.join(tmpDir, 'data')
  fs.mkdirSync(dataDir, { recursive: true })

  const env: TestFixtureEnv = {
    XDG_CONFIG_HOME: configDir,
    XDG_STATE_HOME: stateDir,
    XDG_DATA_HOME: dataDir,
    HOME: tmpDir,
    KYARABEN_E2E_FAKE_INSTALLER: '1',
    KYARABEN_RELAY_HEALTH_RETRIES: '1',
  }

  return {
    configDir,
    stateDir,
    collection,
    env,
    cleanup: () => {
      const logPath = path.join(stateDir, 'kyaraben', 'kyaraben.log')
      if (fs.existsSync(logPath)) {
        const content = fs.readFileSync(logPath, 'utf-8').trim()
        if (content) {
          const indent = '        '
          const indented = content
            .split('\n')
            .map((line) => indent + line)
            .join('\n')
          console.log(`\n${indent}--- daemon log (${tmpDir}) ---`)
          console.log(indented)
          console.log(`${indent}--- end daemon log ---\n`)
        }
      }
      try {
        fs.rmSync(tmpDir, { recursive: true, force: true })
      } catch {
        // Ignore cleanup errors
      }
    },
  }
}

function generateConfigToml(config: ConfigFixture, defaultCollection: string): string {
  const lines: string[] = []

  lines.push('[global]')
  lines.push(`collection = "${config.collection ?? defaultCollection}"`)
  lines.push('')

  if (config.sync) {
    lines.push('[sync]')
    lines.push(`enabled = ${config.sync.enabled ?? false}`)
    if (config.sync.relayUrl) {
      lines.push(`relays = ["${config.sync.relayUrl}"]`)
    }
    lines.push('')

    if (config.sync.devices && config.sync.devices.length > 0) {
      for (const device of config.sync.devices) {
        lines.push('[[sync.devices]]')
        lines.push(`id = "${device.id}"`)
        lines.push(`name = "${device.name}"`)
        lines.push('')
      }
    }
  }

  if (config.systems) {
    lines.push('[systems]')
    for (const [system, emulators] of Object.entries(config.systems)) {
      if (emulators) {
        const emuList = emulators.map((e) => `"${e}"`).join(', ')
        lines.push(`${system} = [${emuList}]`)
      }
    }
    lines.push('')
  }

  if (config.emulators) {
    for (const [emuId, emuConfig] of Object.entries(config.emulators)) {
      if (emuConfig) {
        lines.push(`[emulators."${emuId}"]`)
        if (emuConfig.version) {
          lines.push(`version = "${emuConfig.version}"`)
        }
        if (emuConfig.shaders !== undefined && emuConfig.shaders !== null) {
          lines.push(`shaders = ${emuConfig.shaders}`)
        }
        lines.push('')
      }
    }
  }

  if (config.frontends) {
    for (const [feId, feConfig] of Object.entries(config.frontends)) {
      if (feConfig) {
        lines.push(`[frontends."${feId}"]`)
        lines.push(`enabled = ${feConfig.enabled}`)
        if (feConfig.version) {
          lines.push(`version = "${feConfig.version}"`)
        }
        lines.push('')
      }
    }
  }

  return lines.join('\n')
}

export function createBiosDirectory(fixture: TestFixture, systemId: SystemID): string {
  const biosDir = path.join(fixture.collection, 'bios', systemId)
  fs.mkdirSync(biosDir, { recursive: true })
  return biosDir
}

export function createFakeBiosFile(biosDir: string, filename: string, content?: Buffer): void {
  const filePath = path.join(biosDir, filename)
  fs.writeFileSync(filePath, content ?? Buffer.from('fake bios content'))
}

export const presets = {
  freshInstall: (): { config: ConfigFixture; manifest?: ManifestFixture } => ({
    config: {},
  }),

  systemsEnabledNotInstalled: (): { config: ConfigFixture; manifest: ManifestFixture } => ({
    config: {
      systems: {
        [SystemIDSNES]: [EmulatorIDRetroArchBsnes],
        [SystemIDPSX]: [EmulatorIDDuckStation],
      },
    },
    manifest: {
      installedEmulators: {},
    },
  }),

  emulatorsInstalled: (): { config: ConfigFixture; manifest: ManifestFixture } => ({
    config: {
      systems: {
        [SystemIDSNES]: [EmulatorIDRetroArchBsnes],
        [SystemIDGBA]: [EmulatorIDRetroArchMGBA],
      },
    },
    manifest: {
      lastApplied: new Date().toISOString(),
      installedEmulators: {
        [EmulatorIDRetroArchBsnes]: {
          id: EmulatorIDRetroArchBsnes,
          version: '115.0.0',
          storePath: '/tmp/kyaraben-packages/bsnes',
          installed: new Date().toISOString(),
        },
        [EmulatorIDRetroArchMGBA]: {
          id: EmulatorIDRetroArchMGBA,
          version: '0.10.3',
          storePath: '/tmp/kyaraben-packages/mgba',
          installed: new Date().toISOString(),
        },
      },
      managedConfigs: [
        {
          emulatorId: EmulatorIDRetroArchBsnes,
          target: { type: 'xdg_config', path: 'bsnes/settings.bml' },
          writtenEntries: {},
          lastModified: new Date().toISOString(),
        },
      ],
    },
  }),

  syncEnabled: (): { config: ConfigFixture; manifest: ManifestFixture } => ({
    config: {
      systems: {
        [SystemIDSNES]: [EmulatorIDRetroArchBsnes],
      },
      sync: {
        enabled: true,
        devices: [{ id: 'DEVICE-ID-123', name: 'My Steam Deck' }],
      },
    },
    manifest: {
      installedEmulators: {},
    },
  }),

  versionPinned: (): { config: ConfigFixture; manifest: ManifestFixture } => ({
    config: {
      systems: {
        [SystemIDSNES]: [EmulatorIDRetroArchBsnes],
      },
      emulators: {
        [EmulatorIDRetroArchBsnes]: { version: '115.0.0' },
      },
    },
    manifest: {
      installedEmulators: {
        [EmulatorIDRetroArchBsnes]: {
          id: EmulatorIDRetroArchBsnes,
          version: '114.0.0',
          storePath: '/tmp/kyaraben-packages/bsnes-old',
          installed: new Date(Date.now() - 86400000).toISOString(),
        },
      },
    },
  }),

  frontendEnabled: (): { config: ConfigFixture; manifest: ManifestFixture } => ({
    config: {
      systems: {
        [SystemIDSNES]: [EmulatorIDRetroArchBsnes],
      },
      frontends: {
        [FrontendIDESDE]: { enabled: true },
      },
    },
    manifest: {
      installedEmulators: {},
    },
  }),
}

export interface FakeReleasesOptions {
  latestVersion: string
  appImagePath?: string
}

let nextPort = 19500

export function setupFakeReleasesApi(fixture: TestFixture, options: FakeReleasesOptions): void {
  const port = nextPort++
  const server = startFakeReleasesServer(port, {
    version: options.latestVersion,
    ...(options.appImagePath !== undefined && { appImagePath: options.appImagePath }),
  })

  fixture.releasesServer = server
  fixture.env.KYARABEN_RELEASES_URL = `http://localhost:${port}/releases/latest`

  const originalCleanup = fixture.cleanup
  fixture.cleanup = () => {
    server.close()
    originalCleanup()
  }
}

export interface FakeSyncthingOptions {
  myID?: string
  devices?: Device[]
  folders?: Folder[]
}

export function setupFakeSyncthingApi(
  fixture: TestFixture,
  options: FakeSyncthingOptions = {},
): FakeSyncthingController {
  const port = nextPort++
  const myID = options.myID ?? 'LOCAL01-DEVICE2-IDFAKE3-1234567-890ABCD-EFGHIJK-LMNOPQR-STUVWXY'

  const controller = new FakeSyncthingController(myID)

  if (options.devices) {
    for (const device of options.devices) {
      controller.addDevice(device)
    }
  }

  if (options.folders) {
    for (const folder of options.folders) {
      controller.addFolder(folder)
    }
  }

  const server = startFakeSyncthingServer(port, controller)
  fixture.syncthingServer = server

  const configPath = path.join(fixture.configDir, 'kyaraben', 'config.toml')
  if (fs.existsSync(configPath)) {
    let content = fs.readFileSync(configPath, 'utf-8')
    if (content.includes('[sync.syncthing]')) {
      content = content.replace(
        /\[sync\.syncthing\]/,
        `[sync.syncthing]\nbase_url = "http://localhost:${port}"`,
      )
    } else {
      content += `\n[sync.syncthing]\nbase_url = "http://localhost:${port}"\n`
    }
    fs.writeFileSync(configPath, content)
  }

  const syncthingDir = path.join(fixture.stateDir, 'kyaraben', 'syncthing-bin')
  fs.mkdirSync(syncthingDir, { recursive: true })
  const fakeBinaryPath = path.join(syncthingDir, 'syncthing')
  fs.writeFileSync(fakeBinaryPath, '#!/bin/sh\necho "fake syncthing"')
  fs.chmodSync(fakeBinaryPath, 0o755)

  const syncthingConfigDir = path.join(fixture.stateDir, 'kyaraben', 'syncthing', 'config')
  fs.mkdirSync(syncthingConfigDir, { recursive: true })
  fs.writeFileSync(path.join(syncthingConfigDir, '.apikey'), 'fake-api-key-for-e2e-tests')

  const manifestPath = path.join(fixture.stateDir, 'kyaraben', 'build', 'manifest.json')
  if (fs.existsSync(manifestPath)) {
    const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf-8'))
    manifest.syncthing_install = {
      version: '1.27.0',
      config_schema_version: 1,
      binary_path: fakeBinaryPath,
      config_dir: path.join(fixture.stateDir, 'kyaraben', 'syncthing'),
      data_dir: path.join(fixture.stateDir, 'kyaraben', 'syncthing-data'),
    }
    fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2))
  }

  const originalCleanup = fixture.cleanup
  fixture.cleanup = () => {
    server.close()
    originalCleanup()
  }

  return controller
}
