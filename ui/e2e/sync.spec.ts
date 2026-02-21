import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'
import {
  buildEnv,
  type ConfigFixture,
  createFixture,
  EmulatorIDRetroArchBsnes,
  type FakeSyncthingOptions,
  type ManifestFixture,
  type RelayServer,
  SystemIDSNES,
  setupFakeSyncthingApi,
  startRelayServer,
  type TestFixture,
} from './fixtures'
import type { FakeSyncthingController } from './fixtures/fake-syncthing-server'

function getAppImagePath(): string {
  const appImagePath = process.env.KYARABEN_APPIMAGE
  if (!appImagePath) {
    throw new Error('KYARABEN_APPIMAGE environment variable must be set')
  }
  return appImagePath
}

interface SyncTestContext {
  app: ElectronApplication
  page: Page
  fixture: TestFixture
  controller: FakeSyncthingController
}

async function setupSyncTest(options: {
  config: ConfigFixture
  manifest: ManifestFixture
  syncthing: FakeSyncthingOptions
  setup?: (controller: FakeSyncthingController) => void
}): Promise<SyncTestContext> {
  const fixture = createFixture(options.config, options.manifest)
  const controller = setupFakeSyncthingApi(fixture, options.syncthing)
  options.setup?.(controller)

  const app = await electron.launch({
    executablePath: getAppImagePath(),
    args: ['--no-sandbox'],
    env: buildEnv(fixture),
  })

  const page = await app.firstWindow()
  await page.getByRole('heading', { level: 1 }).waitFor({ timeout: 30000 })

  return { app, page, fixture, controller }
}

async function cleanupSyncTest(ctx: SyncTestContext): Promise<void> {
  await ctx.app?.close()
  ctx.fixture?.cleanup()
}

async function navigateToSync(page: Page): Promise<void> {
  await page.getByRole('button', { name: 'Sync' }).click()
  await expect(page.getByText('running')).toBeVisible()
}

test.describe('Sync view with connected device showing synced status', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [{ deviceID: 'REMOTE-DEVICE-1234567890ABCDEF', name: 'Steam Deck' }],
        folders: [
          { id: 'saves', path: '/home/test/Emulation/saves' },
          { id: 'states', path: '/home/test/Emulation/states' },
        ],
      },
      setup: (c) => c.setConnected('REMOTE-DEVICE-1234567890ABCDEF', true),
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows running state with connected device', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText('synced', { exact: true })).toBeVisible()
    await expect(ctx.page.getByText(/Steam Deck/)).toBeVisible()
    await expect(ctx.page.getByText('All synced')).toBeVisible()
  })
})

test.describe('Sync view primary showing remote device completion', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [{ deviceID: 'REMOTE-DEVICE-1234567890ABCDEF', name: 'Steam Deck' }],
        folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
      },
      setup: (c) => {
        c.setConnected('REMOTE-DEVICE-1234567890ABCDEF', true)
        c.setDeviceCompletion('REMOTE-DEVICE-1234567890ABCDEF', 75, 25_000_000, 100_000_000)
      },
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows device completion percentage when syncing', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText(/Steam Deck/)).toBeVisible()
    await expect(ctx.page.getByText('75% synced')).toBeVisible()
  })
})

test.describe('Sync view showing sync in progress', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'secondary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [{ deviceID: 'PRIMARY-DEVICE-1234567890ABCDEF', name: 'Desktop PC' }],
        folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
      },
      setup: (c) => {
        c.setConnected('PRIMARY-DEVICE-1234567890ABCDEF', true)
        c.setFolderProgress('saves', 50_000_000, 100_000_000)
      },
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows syncing progress with remaining bytes', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText(/remaining/)).toBeVisible()
  })
})

test.describe('Sync view showing scanning progress', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [{ deviceID: 'REMOTE-DEVICE-1234567890ABCDEF', name: 'Steam Deck' }],
        folders: [
          { id: 'saves', path: '/home/test/Emulation/saves' },
          { id: 'states', path: '/home/test/Emulation/states' },
        ],
      },
      setup: (c) => {
        c.setConnected('REMOTE-DEVICE-1234567890ABCDEF', true)
        c.setFolderState('saves', 'scanning')
        c.setFolderState('states', 'scanning')
      },
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows scanning progress with folder name and queue', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText('saves')).toBeVisible()
    await expect(ctx.page.getByText(/Scanning/)).toBeVisible()
    await expect(ctx.page.getByText('+1 folder in queue')).toBeVisible()
  })
})

test.describe('Sync view with device disconnected', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [{ deviceID: 'REMOTE-DEVICE-OFFLINE-1234567890', name: 'Offline Device' }],
        folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
      },
      setup: (c) => c.setConnected('REMOTE-DEVICE-OFFLINE-1234567890', false),
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows offline status for disconnected device', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText('offline', { exact: true })).toBeVisible()
    await expect(ctx.page.getByText(/Offline Device/)).toBeVisible()
  })
})

test.describe('Sync view with local changes on secondary', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'secondary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [{ deviceID: 'PRIMARY-DEVICE-1234567890ABCDEF', name: 'Desktop PC' }],
        folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
      },
      setup: (c) => {
        c.setConnected('PRIMARY-DEVICE-1234567890ABCDEF', true)
        c.addLocalChanges('saves', [
          {
            action: 'changed',
            type: 'file',
            name: 'game1.sav',
            modified: new Date().toISOString(),
            size: 8192,
          },
          {
            action: 'changed',
            type: 'file',
            name: 'game2.sav',
            modified: new Date().toISOString(),
            size: 4096,
          },
        ])
      },
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows local changes indicator and revert action', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText(/local changes/)).toBeVisible()
    await ctx.page.getByRole('button', { name: /Folders/ }).click()
    await expect(ctx.page.getByText('Revert...')).toBeVisible()
  })

  test('can open revert modal and see file list', async () => {
    await ctx.page.getByText('Revert...').click()
    await expect(ctx.page.getByText(/Revert local changes/)).toBeVisible()
    await expect(ctx.page.getByText('game1.sav')).toBeVisible()
    await expect(ctx.page.getByText('game2.sav')).toBeVisible()
    await ctx.page.getByRole('button', { name: 'Cancel' }).click()
  })
})

test.describe('Sync view pairing UI when no devices paired (primary)', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [],
        folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
      },
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows pairing UI for primary with no devices', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText('Pair a device')).toBeVisible()
    await expect(ctx.page.getByRole('button', { name: 'Start pairing' })).toBeVisible()
    await expect(ctx.page.getByText('Waiting for device connection')).toBeVisible()
  })
})

test.describe('Sync view discovery UI when no devices paired (secondary)', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'secondary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [],
        folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
      },
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows pairing code input for secondary with no devices', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText('Connect to primary')).toBeVisible()
    await expect(ctx.page.getByPlaceholder('ABC123')).toBeVisible()
    await expect(ctx.page.getByText('Advanced options')).toBeVisible()
  })

  test('can expand advanced options to see device ID input', async () => {
    await ctx.page.getByText('Advanced options').click()
    await expect(ctx.page.getByPlaceholder('XXXXXXX-XXXXXXX-XXXXXXX-...')).toBeVisible()
  })
})

test.describe('Sync view settings section', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [{ deviceID: 'REMOTE-DEVICE-1234567890ABCDEF', name: 'Steam Deck' }],
        folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
      },
      setup: (c) => c.setConnected('REMOTE-DEVICE-1234567890ABCDEF', true),
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('can expand settings to see reset option', async () => {
    await navigateToSync(ctx.page)
    await ctx.page.getByRole('button', { name: /Settings/ }).click()
    await expect(ctx.page.getByRole('button', { name: 'Reset sync' })).toBeVisible()
    await expect(ctx.page.getByText('Open Syncthing web interface')).toBeVisible()
  })
})

test.describe('Sync view with multiple folders', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [{ deviceID: 'REMOTE-DEVICE-1234567890ABCDEF', name: 'Steam Deck' }],
        folders: [
          { id: 'saves', path: '/home/test/Emulation/saves' },
          { id: 'states', path: '/home/test/Emulation/states' },
          { id: 'screenshots', path: '/home/test/Emulation/screenshots' },
        ],
      },
      setup: (c) => c.setConnected('REMOTE-DEVICE-1234567890ABCDEF', true),
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows folder count and can expand to see all folders', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText(/Folders \(3\)/)).toBeVisible()
    await ctx.page.getByRole('button', { name: /Folders/ }).click()
    await expect(ctx.page.getByText('saves')).toBeVisible()
    await expect(ctx.page.getByText('states')).toBeVisible()
    await expect(ctx.page.getByText('screenshots')).toBeVisible()
  })
})

test.describe('Sync view remove device flow', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary' },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [{ deviceID: 'DEVICE-TO-REMOVE-1234567890AB', name: 'Old Device' }],
        folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
      },
      setup: (c) => c.setConnected('DEVICE-TO-REMOVE-1234567890AB', true),
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows confirmation dialog with device name and explanation', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText(/Old Device/)).toBeVisible()
    await ctx.page.getByRole('button', { name: 'Remove device' }).click()

    const dialog = ctx.page.getByRole('dialog')
    await expect(dialog).toBeVisible()
    await expect(dialog.getByText('Are you sure you want to remove')).toBeVisible()
    await expect(dialog.getByText('Old Device')).toBeVisible()
    await expect(dialog.getByText('This device will no longer sync with you')).toBeVisible()
    await expect(dialog.getByRole('button', { name: 'Cancel' })).toBeVisible()
    await expect(dialog.getByRole('button', { name: 'Remove device' })).toBeVisible()

    await dialog.getByRole('button', { name: 'Cancel' }).click()
    await expect(dialog).not.toBeVisible()
  })

  test('removes device after confirmation', async () => {
    await ctx.page.getByRole('button', { name: 'Remove device' }).click()
    await ctx.page.getByRole('dialog').getByRole('button', { name: 'Remove device' }).click()
    await expect(ctx.page.getByRole('dialog')).not.toBeVisible()
    await expect(ctx.page.getByText(/Old Device/)).not.toBeVisible()
    await expect(ctx.page.getByText('Pair a device')).toBeVisible()
  })
})

test.describe('Relay pairing - primary shows pairing code', () => {
  let ctx: SyncTestContext
  let relay: RelayServer

  test.beforeAll(async () => {
    relay = await startRelayServer()
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary', relayUrl: relay.url },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [],
        folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
      },
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
    relay.close()
  })

  test('shows 6-character pairing code when starting pairing', async () => {
    await navigateToSync(ctx.page)
    await ctx.page.getByRole('button', { name: 'Start pairing' }).click()
    const codeElement = ctx.page.locator('code').filter({ hasText: /^[A-Z0-9]{6}$/ })
    await expect(codeElement).toBeVisible({ timeout: 5000 })
    await expect(ctx.page.getByText('Enter this code on your secondary device')).toBeVisible()
  })

  test('can reveal device ID in pairing mode', async () => {
    await ctx.page.getByText('Show device ID').click()
    await expect(ctx.page.locator('code').filter({ hasText: /-/ })).toBeVisible()
  })

  test('can stop pairing', async () => {
    await ctx.page.getByRole('button', { name: 'Stop pairing' }).click()
    await expect(ctx.page.getByRole('button', { name: 'Start pairing' })).toBeVisible()
  })
})

test.describe('Relay pairing - secondary enters invalid code', () => {
  let ctx: SyncTestContext
  let relay: RelayServer

  test.beforeAll(async () => {
    relay = await startRelayServer()
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'secondary', relayUrl: relay.url },
      },
      manifest: { installedEmulators: {} },
      syncthing: {
        devices: [],
        folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
      },
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
    relay.close()
  })

  test('shows error for invalid pairing code', async () => {
    await navigateToSync(ctx.page)
    await ctx.page.getByPlaceholder('ABC123').fill('XXXXXX')
    await ctx.page.getByRole('button', { name: 'Connect' }).click()
    await expect(ctx.page.getByText(/invalid pairing code|session not found/i).first()).toBeVisible(
      {
        timeout: 5000,
      },
    )
  })
})
