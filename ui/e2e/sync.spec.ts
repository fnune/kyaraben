import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'
import {
  buildEnv,
  createFixture,
  EmulatorIDRetroArchBsnes,
  SystemIDSNES,
  setupFakeSyncthingApi,
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

async function launchWithFixture(
  fixture: TestFixture,
): Promise<{ app: ElectronApplication; page: Page }> {
  const app = await electron.launch({
    executablePath: getAppImagePath(),
    args: ['--no-sandbox'],
    env: buildEnv(fixture),
  })

  const page = await app.firstWindow()
  await page.getByRole('heading', { level: 1 }).waitFor({ timeout: 30000 })

  return { app, page }
}

test.describe('Sync view with connected device showing synced status', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page
  let controller: FakeSyncthingController

  test.beforeAll(async () => {
    fixture = createFixture(
      {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary' },
      },
      { installedEmulators: {} },
    )

    controller = setupFakeSyncthingApi(fixture, {
      devices: [{ deviceID: 'REMOTE-DEVICE-1234567890ABCDEF', name: 'Steam Deck' }],
      folders: [
        { id: 'saves', path: '/home/test/Emulation/saves' },
        { id: 'states', path: '/home/test/Emulation/states' },
      ],
    })

    controller.setConnected('REMOTE-DEVICE-1234567890ABCDEF', true)

    const result = await launchWithFixture(fixture)
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('navigates to sync view and shows running state', async () => {
    await page.getByRole('button', { name: 'Sync' }).click()
    await expect(page.getByText('running')).toBeVisible()
  })

  test('shows connected status badge', async () => {
    await expect(page.getByText('connected', { exact: true })).toBeVisible()
  })

  test('shows paired device name', async () => {
    await expect(page.getByText(/Steam Deck/)).toBeVisible()
  })

  test('shows all synced message', async () => {
    await expect(page.getByText('All synced')).toBeVisible()
  })
})

test.describe('Sync view showing sync in progress', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page
  let controller: FakeSyncthingController

  test.beforeAll(async () => {
    fixture = createFixture(
      {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'secondary' },
      },
      { installedEmulators: {} },
    )

    controller = setupFakeSyncthingApi(fixture, {
      devices: [{ deviceID: 'PRIMARY-DEVICE-1234567890ABCDEF', name: 'Desktop PC' }],
      folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
    })

    controller.setConnected('PRIMARY-DEVICE-1234567890ABCDEF', true)
    controller.setFolderProgress('saves', 50_000_000, 100_000_000)

    const result = await launchWithFixture(fixture)
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('navigates to sync view and shows running state', async () => {
    await page.getByRole('button', { name: 'Sync' }).click()
    await expect(page.getByText('running')).toBeVisible()
  })

  test('shows syncing folder with remaining bytes', async () => {
    await expect(page.getByText(/remaining/)).toBeVisible()
  })
})

test.describe('Sync view with device disconnected', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page
  let controller: FakeSyncthingController

  test.beforeAll(async () => {
    fixture = createFixture(
      {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
        sync: { enabled: true, mode: 'primary' },
      },
      { installedEmulators: {} },
    )

    controller = setupFakeSyncthingApi(fixture, {
      devices: [{ deviceID: 'REMOTE-DEVICE-OFFLINE-1234567890', name: 'Offline Device' }],
      folders: [{ id: 'saves', path: '/home/test/Emulation/saves' }],
    })

    controller.setConnected('REMOTE-DEVICE-OFFLINE-1234567890', false)

    const result = await launchWithFixture(fixture)
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('navigates to sync view and shows running state', async () => {
    await page.getByRole('button', { name: 'Sync' }).click()
    await expect(page.getByText('running')).toBeVisible()
  })

  test('shows offline status for disconnected device', async () => {
    await expect(page.getByText('offline', { exact: true })).toBeVisible()
  })

  test('shows device name', async () => {
    await expect(page.getByText(/Offline Device/)).toBeVisible()
  })
})
