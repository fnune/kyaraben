import * as fs from 'node:fs'
import * as os from 'node:os'
import * as path from 'node:path'
import { type ElectronApplication, expect, type Page, test } from '@playwright/test'
import { createFixture, launchElectron, setupFakeReleasesApi, type TestFixture } from './fixtures'

function createFakeAppImage(): string {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'kyaraben-test-'))
  const fakePath = path.join(tempDir, 'fake.AppImage')
  fs.writeFileSync(fakePath, Buffer.alloc(1024 * 100, 'x'))
  return fakePath
}

let fixture: TestFixture
let electronApp: ElectronApplication
let page: Page

test.describe('Update checking', () => {
  test.beforeAll(async () => {
    fixture = createFixture({})
    setupFakeReleasesApi(fixture, { latestVersion: '99.0.0' })
    const result = await launchElectron(fixture)
    electronApp = result.app
    page = result.page
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
    await expect(page.getByText('A new version of Kyaraben is available: 99.0.0.')).toBeVisible({
      timeout: 10000,
    })
    await expect(page.getByText("See what's new.")).toBeVisible()
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
    fixture = createFixture({})
    setupFakeReleasesApi(fixture, { latestVersion: '0.0.1' })
    const result = await launchElectron(fixture)
    electronApp = result.app
    page = result.page
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
    await expect(page.getByText("See what's new.")).not.toBeVisible()
  })
})

test.describe('Version mismatch detection', () => {
  test.beforeAll(async () => {
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
    const result = await launchElectron(fixture)
    electronApp = result.app
    page = result.page
  })

  test.afterAll(async () => {
    if (electronApp) {
      await electronApp.close()
    }
    fixture?.cleanup()
  })

  test('shows upgrade message in action bar when manifest version differs from running version', async () => {
    await expect(
      page.getByText('Kyaraben was updated. Apply to get the latest emulator configs.'),
    ).toBeVisible({ timeout: 10000 })
    await expect(page.getByRole('button', { name: 'Apply' })).toBeVisible()
  })
})

test.describe('Update download with redirect', () => {
  let fakeAppImagePath: string

  test.beforeAll(async () => {
    fakeAppImagePath = createFakeAppImage()
    fixture = createFixture({})
    setupFakeReleasesApi(fixture, {
      latestVersion: '99.0.0',
      appImagePath: fakeAppImagePath,
      simulateRedirect: true,
    })
    const result = await launchElectron(fixture)
    electronApp = result.app
    page = result.page
  })

  test.afterAll(async () => {
    if (electronApp) {
      await electronApp.close()
    }
    fixture?.cleanup()
    if (fakeAppImagePath) {
      fs.rmSync(path.dirname(fakeAppImagePath), { recursive: true, force: true })
    }
  })

  test('handles redirect during download without immediate failure', async () => {
    await page.getByRole('button', { name: 'Installation' }).click()
    await page.getByRole('button', { name: 'Check' }).click()
    await expect(page.getByText('New version available: 99.0.0')).toBeVisible({ timeout: 10000 })

    await page.locator('#main-content').getByRole('button', { name: 'Update now' }).click()
    await expect(page.getByText(/Downloading/)).toBeVisible({ timeout: 5000 })
    await expect(page.getByText(/Downloading.*[1-9]\d*%/)).toBeVisible({ timeout: 10000 })
  })
})
