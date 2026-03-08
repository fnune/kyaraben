import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'
import {
  buildEnv,
  getElectronArgs,
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
    args: getElectronArgs(),
    env: buildEnv(fixture),
  })

  const page = await app.firstWindow()
  await page.getByRole('img', { name: 'Kyaraben' }).waitFor({ timeout: 30000 })

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
        sync: { enabled: true },
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
    await expect(ctx.page.getByText('All synchronized')).toBeVisible()
  })
})

test.describe('Sync view showing remote device completion', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true },
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
        sync: { enabled: true },
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
        sync: { enabled: true },
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
        sync: { enabled: true },
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

test.describe('Sync view with folder in error state', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true },
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
        c.setFolderState('saves', 'error', 'folder path missing')
      },
    })
  })

  test.afterAll(async () => {
    await cleanupSyncTest(ctx)
  })

  test('shows error state in activity card with folder name and error message', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText('Synchronization error')).toBeVisible()
    await expect(ctx.page.getByText('saves')).toBeVisible()
    await expect(ctx.page.getByText('folder path missing')).toBeVisible()
  })

  test('shows error indicator in folders list', async () => {
    await ctx.page.getByRole('button', { name: /Folders/ }).click()
    await expect(ctx.page.getByText('Error', { exact: true })).toBeVisible()
  })
})

test.describe('Sync view with local changes on joiner', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true },
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

test.describe('Sync view pairing UI when no devices paired', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true },
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

  test('shows unified pairing UI with generate code and enter code options', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText('Pair a device')).toBeVisible()
    await expect(ctx.page.getByRole('button', { name: 'Generate pairing code' })).toBeVisible()
    await expect(ctx.page.getByPlaceholder('ABC123')).toBeVisible()
    await expect(ctx.page.getByText('Use device ID instead')).toBeVisible()
  })

  test('can switch to device ID mode', async () => {
    await ctx.page.getByText('Use device ID instead').click()
    await expect(
      ctx.page.getByPlaceholder('ABC1234-DEF5678-GHI9012-JKL3456-MNO7890-PQR1234-STU5678-VWX9012'),
    ).toBeVisible()
  })
})

test.describe('Sync view settings section', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true },
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
        sync: { enabled: true },
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
        sync: { enabled: true },
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
    await expect(dialog.getByText('This device will no longer synchronize with you')).toBeVisible()
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

test.describe('Relay pairing - generate pairing code', () => {
  let ctx: SyncTestContext
  let relay: RelayServer

  test.beforeAll(async () => {
    relay = await startRelayServer()
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, relayUrl: relay.url },
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

  test('shows 6-character pairing code when generating', async () => {
    await navigateToSync(ctx.page)
    await ctx.page.getByRole('button', { name: 'Generate pairing code' }).click()
    const codeElement = ctx.page.locator('code').filter({ hasText: /^[A-Z0-9]{6}$/ })
    await expect(codeElement).toBeVisible({ timeout: 5000 })
    await expect(ctx.page.getByText('Enter this code on the other device')).toBeVisible()
  })

  test('can reveal device ID in pairing mode', async () => {
    await ctx.page.getByText('Use device ID instead').click()
    await expect(ctx.page.locator('code').filter({ hasText: /-/ })).toBeVisible()
  })

  test('can cancel pairing', async () => {
    await ctx.page.getByRole('button', { name: 'Cancel' }).click()
    await expect(ctx.page.getByRole('button', { name: 'Generate pairing code' })).toBeVisible()
  })
})

test.describe('Relay pairing - enter invalid code', () => {
  let ctx: SyncTestContext
  let relay: RelayServer

  test.beforeAll(async () => {
    relay = await startRelayServer()
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, relayUrl: relay.url },
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

test.describe('Device ID pairing - pending device confirmation', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, relayUrl: 'http://localhost:1' },
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

  test('shows pending device confirmation when device ID is used', async () => {
    await navigateToSync(ctx.page)

    await ctx.page.getByRole('button', { name: 'Generate pairing code' }).click()
    await expect(ctx.page.getByText('Use device ID instead')).toBeVisible({ timeout: 5000 })

    ctx.controller.addPendingDevice({
      deviceID: 'PENDING-DEVICE-1234567890ABCDEF',
      name: 'Steam Deck',
      address: '192.168.1.50:22000',
      time: new Date().toISOString(),
    })

    await expect(ctx.page.getByText('Device wants to connect')).toBeVisible({ timeout: 10000 })
    await expect(ctx.page.getByText('Steam Deck')).toBeVisible()
    await expect(ctx.page.getByRole('button', { name: 'Accept' })).toBeVisible()
    await expect(ctx.page.getByRole('button', { name: 'Reject' })).toBeVisible()
  })

  test('can accept pending device', async () => {
    await ctx.page.getByRole('button', { name: 'Accept' }).click()
    await expect(ctx.page.getByText('Device wants to connect')).not.toBeVisible({ timeout: 5000 })
  })
})

test.describe('Device ID pairing - reject pending device', () => {
  let ctx: SyncTestContext

  test.beforeAll(async () => {
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, relayUrl: 'http://localhost:1' },
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

  test('can reject pending device', async () => {
    await navigateToSync(ctx.page)

    await ctx.page.getByRole('button', { name: 'Generate pairing code' }).click()
    await expect(ctx.page.getByText('Use device ID instead')).toBeVisible({ timeout: 5000 })

    ctx.controller.addPendingDevice({
      deviceID: 'REJECTED-DEVICE-1234567890ABC',
      name: 'Unknown Device',
      address: '192.168.1.100:22000',
      time: new Date().toISOString(),
    })

    await expect(ctx.page.getByText('Device wants to connect')).toBeVisible({ timeout: 10000 })
    await ctx.page.getByRole('button', { name: 'Reject' }).click()
    await expect(ctx.page.getByText('Device wants to connect')).not.toBeVisible({ timeout: 5000 })
  })
})

test.describe('Relay pairing - pending device via manual device ID', () => {
  let ctx: SyncTestContext
  let relay: RelayServer

  test.beforeAll(async () => {
    relay = await startRelayServer()
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, relayUrl: relay.url },
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

  test('shows pending device confirmation when device connects via manually shared device ID', async () => {
    await navigateToSync(ctx.page)

    await ctx.page.getByRole('button', { name: 'Generate pairing code' }).click()
    const codeElement = ctx.page.locator('code').filter({ hasText: /^[A-Z0-9]{6}$/ })
    await expect(codeElement).toBeVisible({ timeout: 5000 })

    ctx.controller.addPendingDevice({
      deviceID: 'MANUAL-DEVICE-ID-1234567890AB',
      name: 'Phone',
      address: '192.168.1.200:22000',
      time: new Date().toISOString(),
    })

    await expect(ctx.page.getByText('Device wants to connect')).toBeVisible({ timeout: 10000 })
    await expect(ctx.page.getByText('Phone')).toBeVisible()
    await ctx.page.getByRole('button', { name: 'Accept' }).click()
    await expect(ctx.page.getByText('Device wants to connect')).not.toBeVisible({ timeout: 5000 })
  })
})

test.describe('Relay pairing - host UI clears when guest connects', () => {
  let ctx: SyncTestContext
  let relay: RelayServer
  const guestDeviceID = 'GUEST01-DEVICE2-ABCDEF3-1234567-HIJKLMN-OPQRSTU-VWXYZ12-3456789'

  test.beforeAll(async () => {
    relay = await startRelayServer()
    ctx = await setupSyncTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, relayUrl: relay.url },
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

  test('clears pairing UI when guest device connects via relay', async () => {
    await navigateToSync(ctx.page)
    await expect(ctx.page.getByText('Pair a device')).toBeVisible()

    await ctx.page.getByRole('button', { name: 'Generate pairing code' }).click()
    const codeElement = ctx.page.locator('code').filter({ hasText: /^[A-Z0-9]{6}$/ })
    await expect(codeElement).toBeVisible({ timeout: 5000 })
    await expect(ctx.page.getByText('Waiting for device to connect')).toBeVisible()

    const code = await codeElement.textContent()
    if (!code) throw new Error('Failed to get pairing code')

    const response = await fetch(`${relay.url}/pair/${code}/response`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ deviceId: guestDeviceID }),
    })
    if (!response.ok) throw new Error(`Failed to submit relay response: ${response.status}`)

    ctx.controller.addDevice({ deviceID: guestDeviceID, name: 'Steam Deck' })
    ctx.controller.setConnected(guestDeviceID, true)

    await expect(ctx.page.getByText('Waiting for device to connect')).not.toBeVisible({
      timeout: 10000,
    })
    await expect(ctx.page.getByText('Steam Deck', { exact: true })).toBeVisible()
    await expect(ctx.page.getByText('synced', { exact: true })).toBeVisible()
  })
})
