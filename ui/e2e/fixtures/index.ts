import * as fs from 'node:fs'
import * as os from 'node:os'
import * as path from 'node:path'
import type { EmulatorID, SystemID } from '../../src/types/model.gen'
import {
  EmulatorIDDuckStation,
  EmulatorIDMGBA,
  EmulatorIDRetroArchBsnes,
  SystemIDGBA,
  SystemIDPSX,
  SystemIDSNES,
} from '../../src/types/model.gen'

export interface TestFixture {
  configDir: string
  stateDir: string
  userStore: string
  env: Record<string, string>
  cleanup: () => void
}

export interface ConfigFixture {
  userStore?: string
  systems?: Partial<Record<SystemID, EmulatorID[]>>
  emulators?: Partial<Record<EmulatorID, { version?: string }>>
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
    const manifestJson = {
      version: manifest.version ?? 1,
      last_applied: manifest.lastApplied ?? new Date().toISOString(),
      installed_emulators: manifest.installedEmulators ?? {},
      managed_configs: manifest.managedConfigs ?? [],
      desktop_files: manifest.desktopFiles ?? [],
      icon_files: manifest.iconFiles ?? [],
    }
    fs.writeFileSync(
      path.join(stateDir, 'kyaraben', 'build', 'manifest.json'),
      JSON.stringify(manifestJson, null, 2),
    )
  }

  const env: Record<string, string> = {
    XDG_CONFIG_HOME: configDir,
    XDG_STATE_HOME: stateDir,
    HOME: tmpDir,
  }

  return {
    configDir,
    stateDir,
    userStore,
    env,
    cleanup: () => {
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

  return lines.join('\n')
}

export function setupFakeNixPortable(fixture: TestFixture): void {
  const fakeNixSrc = path.join(__dirname, 'fake-nix-portable')
  const fakeNixDest = path.join(fixture.stateDir, 'fake-nix-portable')

  fs.copyFileSync(fakeNixSrc, fakeNixDest)
  fs.chmodSync(fakeNixDest, 0o755)

  const fakeStore = path.join(
    fixture.stateDir,
    'kyaraben',
    'build',
    'nix',
    '.nix-portable',
    'nix',
    'store',
  )
  fs.mkdirSync(fakeStore, { recursive: true })

  fixture.env.KYARABEN_NIX_PORTABLE_PATH = fakeNixDest
  fixture.env.FAKE_NIX_STORE = fakeStore
  fixture.env.FAKE_NIX_PROGRESS = '1'
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
          storePath: '/nix/store/fake-hash-bsnes',
          installed: new Date().toISOString(),
        },
        [EmulatorIDMGBA]: {
          id: EmulatorIDMGBA,
          version: '0.10.3',
          storePath: '/nix/store/fake-hash-mgba',
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
          storePath: '/nix/store/fake-hash-bsnes-old',
          installed: new Date(Date.now() - 86400000).toISOString(),
        },
      },
    },
  }),
}
