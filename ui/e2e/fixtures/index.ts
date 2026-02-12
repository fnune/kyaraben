import * as fs from 'node:fs'
import type * as http from 'node:http'
import * as os from 'node:os'
import * as path from 'node:path'
import { startFakeReleasesServer } from './fake-releases-server'

export const SystemIDSNES = 'snes' as const
export const SystemIDGBA = 'gba' as const
export const SystemIDPSX = 'psx' as const
export const EmulatorIDRetroArchBsnes = 'retroarch:bsnes' as const
export const EmulatorIDMGBA = 'mgba' as const
export const EmulatorIDDuckStation = 'duckstation' as const
export const FrontendIDESDE = 'es-de' as const

export type SystemID = typeof SystemIDSNES | typeof SystemIDGBA | typeof SystemIDPSX | string
export type EmulatorID =
  | typeof EmulatorIDRetroArchBsnes
  | typeof EmulatorIDMGBA
  | typeof EmulatorIDDuckStation
  | string
export type FrontendID = typeof FrontendIDESDE | string

export interface TestFixture {
  configDir: string
  stateDir: string
  userStore: string
  env: Record<string, string>
  cleanup: () => void
  releasesServer?: http.Server
}

export interface ConfigFixture {
  userStore?: string
  systems?: Partial<Record<SystemID, EmulatorID[]>>
  emulators?: Partial<Record<EmulatorID, { version?: string }>>
  frontends?: Partial<Record<FrontendID, { enabled: boolean; version?: string }>>
  sync?: {
    enabled?: boolean
    mode?: 'primary' | 'secondary'
    devices?: Array<{ id: string; name: string }>
  }
}

export interface InstalledEmulatorFixture {
  id: EmulatorID
  version: string
  storePath: string
  installed: string
}

export interface ManifestFixture {
  version?: number
  kyarabenVersion?: string
  lastApplied?: string
  installedEmulators?: Partial<Record<EmulatorID, InstalledEmulatorFixture>>
  managedConfigs?: Array<{
    emulatorId: EmulatorID
    target: { type: string; path?: string }
    baselineHash: string
    lastModified: string
    managedKeys: Array<{ path: string[]; value: string }>
  }>
  desktopFiles?: string[]
  iconFiles?: string[]
}

export function createFixture(config?: ConfigFixture, manifest?: ManifestFixture): TestFixture {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'kyaraben-e2e-'))

  const configDir = path.join(tmpDir, 'config')
  const stateDir = path.join(tmpDir, 'state')
  const userStore = path.join(tmpDir, 'Emulation')

  fs.mkdirSync(path.join(configDir, 'kyaraben'), { recursive: true })
  fs.mkdirSync(path.join(stateDir, 'kyaraben', 'build'), { recursive: true })
  fs.mkdirSync(path.join(stateDir, 'kyaraben', 'bin'), { recursive: true })
  fs.mkdirSync(userStore, { recursive: true })

  if (config) {
    const toml = generateConfigToml(config, userStore)
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
    fs.writeFileSync(
      path.join(stateDir, 'kyaraben', 'build', 'manifest.json'),
      JSON.stringify(manifestJson, null, 2),
    )
  }

  const dataDir = path.join(tmpDir, 'data')
  fs.mkdirSync(dataDir, { recursive: true })

  const env: Record<string, string> = {
    XDG_CONFIG_HOME: configDir,
    XDG_STATE_HOME: stateDir,
    XDG_DATA_HOME: dataDir,
    HOME: tmpDir,
    KYARABEN_E2E_FAKE_INSTALLER: '1',
  }

  return {
    configDir,
    stateDir,
    userStore,
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

function generateConfigToml(config: ConfigFixture, defaultUserStore: string): string {
  const lines: string[] = []

  lines.push('[global]')
  lines.push(`user_store = "${config.userStore ?? defaultUserStore}"`)
  lines.push('')

  if (config.sync) {
    lines.push('[sync]')
    lines.push(`enabled = ${config.sync.enabled ?? false}`)
    if (config.sync.mode) {
      lines.push(`mode = "${config.sync.mode}"`)
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
  const biosDir = path.join(fixture.userStore, 'bios', systemId)
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
    manifest: undefined,
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
        [SystemIDGBA]: [EmulatorIDMGBA],
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
        [EmulatorIDMGBA]: {
          id: EmulatorIDMGBA,
          version: '0.10.3',
          storePath: '/tmp/kyaraben-packages/mgba',
          installed: new Date().toISOString(),
        },
      },
      managedConfigs: [
        {
          emulatorId: EmulatorIDRetroArchBsnes,
          target: { type: 'xdg_config', path: 'bsnes/settings.bml' },
          baselineHash: 'abc123',
          lastModified: new Date().toISOString(),
          managedKeys: [],
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
        mode: 'primary',
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
    appImagePath: options.appImagePath,
  })

  fixture.releasesServer = server
  fixture.env.KYARABEN_RELEASES_URL = `http://localhost:${port}/releases/latest`

  const originalCleanup = fixture.cleanup
  fixture.cleanup = () => {
    server.close()
    originalCleanup()
  }
}
