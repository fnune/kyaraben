import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'
import { createFixture, setupFakeReleasesApi, type TestFixture } from './fixtures'

let fixture: TestFixture
let electronApp: ElectronApplication
let page: Page

test.describe('Update checking', () => {
  test.beforeAll(async () => {
    const appImagePath = process.env.KYARABEN_APPIMAGE
    if (!appImagePath) {
      throw new Error('KYARABEN_APPIMAGE environment variable must be set')
    }

    fixture = createFixture({}, undefined)
    setupFakeReleasesApi(fixture, { latestVersion: '99.0.0' })

    electronApp = await electron.launch({
      executablePath: appImagePath,
      args: ['--no-sandbox'],
      env: {
        ...process.env,
        ...fixture.env,
      },
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
    await expect(page.getByText('A new version of Kyaraben is available: 99.0.0')).toBeVisible({
      timeout: 15000,
    })
    await page.waitForTimeout(3000)
  })

  test('can dismiss update banner', async () => {
    const laterButton = page.getByRole('button', { name: 'Dismiss' })
    await page.waitForTimeout(3000)
    await laterButton.click()
    await expect(page.getByText('A new version of Kyaraben is available')).not.toBeVisible()
    await page.waitForTimeout(3000)
  })

  test('Installation view shows check for updates button', async () => {
    await page.getByRole('button', { name: 'Installation' }).click()
    await page.waitForTimeout(3000)
    await expect(page.getByRole('button', { name: 'Check' })).toBeVisible()
  })

  test('can check for updates from Installation view', async () => {
    const checkButton = page.getByRole('button', { name: 'Check' })
    await page.waitForTimeout(3000)
    await checkButton.click()
    await expect(page.getByText('New version available: 99.0.0')).toBeVisible({ timeout: 10000 })
    await page.waitForTimeout(3000)
  })
})

test.describe('No update available', () => {
  test.beforeAll(async () => {
    const appImagePath = process.env.KYARABEN_APPIMAGE
    if (!appImagePath) {
      throw new Error('KYARABEN_APPIMAGE environment variable must be set')
    }

    fixture = createFixture({}, undefined)
    setupFakeReleasesApi(fixture, { latestVersion: '0.0.1' })

    electronApp = await electron.launch({
      executablePath: appImagePath,
      args: ['--no-sandbox'],
      env: {
        ...process.env,
        ...fixture.env,
      },
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
    await page.waitForTimeout(7000)
    await expect(page.getByText('A new version of Kyaraben is available')).not.toBeVisible()
    await page.waitForTimeout(3000)
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
      env: {
        ...process.env,
        ...fixture.env,
      },
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
    await page.waitForTimeout(3000)
  })

  test('can dismiss apply banner', async () => {
    const laterButton = page.getByRole('button', { name: 'Dismiss' })
    await page.waitForTimeout(3000)
    await laterButton.click()
    await expect(
      page.getByText('Kyaraben was updated. Run Apply to get the latest emulator configs.'),
    ).not.toBeVisible()
    await page.waitForTimeout(3000)
  })
})
