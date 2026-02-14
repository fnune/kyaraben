import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'
import { buildEnv, createFixture, setupFakeReleasesApi, type TestFixture } from './fixtures'

let fixture: TestFixture
let electronApp: ElectronApplication
let page: Page

test.describe('Update checking', () => {
  test.beforeAll(async () => {
    const appImagePath = process.env.KYARABEN_APPIMAGE
    if (!appImagePath) {
      throw new Error('KYARABEN_APPIMAGE environment variable must be set')
    }

    fixture = createFixture({})
    setupFakeReleasesApi(fixture, { latestVersion: '99.0.0' })

    electronApp = await electron.launch({
      executablePath: appImagePath,
      args: ['--no-sandbox'],
      env: buildEnv(fixture),
    })

    page = await electronApp.firstWindow()
    await page.getByRole('heading', { level: 1 }).waitFor()
  })

  test.afterAll(async () => {
    if (electronApp) {
      await electronApp.close()
    }
    fixture?.cleanup()
  })

  test('shows update banner when new version available', async () => {
    await page.getByRole('button', { name: 'Installation' }).click()
    await page.getByRole('button', { name: 'Check' }).click()
    await expect(page.getByText('New version available: 99.0.0')).toBeVisible({ timeout: 10000 })
    await page.getByRole('button', { name: 'Catalog', exact: true }).click()
    await expect(page.getByText('A new version of Kyaraben is available: 99.0.0')).toBeVisible({
      timeout: 10000,
    })
  })

  test('can dismiss update banner', async () => {
    const laterButton = page.getByRole('button', { name: 'Dismiss' })
    await laterButton.click()
    await expect(page.getByText('A new version of Kyaraben is available')).not.toBeVisible()
  })

  test('Installation view shows check for updates button', async () => {
    await page.getByRole('button', { name: 'Installation' }).click()
    await expect(page.getByRole('button', { name: 'Check' })).toBeVisible()
  })

  test('can check for updates from Installation view', async () => {
    const checkButton = page.getByRole('button', { name: 'Check' })
    await checkButton.click()
    await expect(page.getByText('New version available: 99.0.0')).toBeVisible({ timeout: 10000 })
  })
})

test.describe('No update available', () => {
  test.beforeAll(async () => {
    const appImagePath = process.env.KYARABEN_APPIMAGE
    if (!appImagePath) {
      throw new Error('KYARABEN_APPIMAGE environment variable must be set')
    }

    fixture = createFixture({})
    setupFakeReleasesApi(fixture, { latestVersion: '0.0.1' })

    electronApp = await electron.launch({
      executablePath: appImagePath,
      args: ['--no-sandbox'],
      env: buildEnv(fixture),
    })

    page = await electronApp.firstWindow()
    await page.getByRole('heading', { level: 1 }).waitFor()
  })

  test.afterAll(async () => {
    if (electronApp) {
      await electronApp.close()
    }
    fixture?.cleanup()
  })

  test('does not show update banner when on latest version', async () => {
    await page.getByRole('button', { name: 'Installation' }).click()
    await expect(page.getByText('Check for updates')).toBeVisible()
    await page.getByRole('button', { name: 'Check' }).click()
    await expect(page.getByRole('button', { name: 'Check' })).toBeEnabled({ timeout: 10000 })
    await expect(page.getByText(/(You're on the latest version|Current version)/)).toBeVisible()
    await page.getByRole('button', { name: 'Catalog', exact: true }).click()
    await expect(page.getByText('A new version of Kyaraben is available')).not.toBeVisible()
  })
})

test.describe('Version mismatch detection', () => {
  test.beforeAll(async () => {
    const appImagePath = process.env.KYARABEN_APPIMAGE
    if (!appImagePath) {
      throw new Error('KYARABEN_APPIMAGE environment variable must be set')
    }

    fixture = createFixture(
      {},
      {
        version: 1,
        lastApplied: new Date().toISOString(),
        installedEmulators: {},
        kyarabenVersion: '0.0.1',
      },
    )

    fixture.env.KYARABEN_VERSION = '0.1.0'

    electronApp = await electron.launch({
      executablePath: appImagePath,
      args: ['--no-sandbox'],
      env: buildEnv(fixture),
    })

    page = await electronApp.firstWindow()
    await page.getByRole('heading', { level: 1 }).waitFor()
  })

  test.afterAll(async () => {
    if (electronApp) {
      await electronApp.close()
    }
    fixture?.cleanup()
  })

  test('shows apply banner when manifest version differs from running version', async () => {
    await expect(
      page.getByText('Kyaraben was updated. Run Apply to get the latest emulator configs.'),
    ).toBeVisible({ timeout: 10000 })
  })

  test('can dismiss apply banner', async () => {
    const laterButton = page.getByRole('button', { name: 'Dismiss' })
    await laterButton.click()
    await expect(
      page.getByText('Kyaraben was updated. Run Apply to get the latest emulator configs.'),
    ).not.toBeVisible()
  })
})
