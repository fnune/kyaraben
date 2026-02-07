import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'
import { createFixture, type TestFixture } from './fixtures'

let fixture: TestFixture
let electronApp: ElectronApplication
let page: Page

test.beforeAll(async () => {
  const appImagePath = process.env.KYARABEN_APPIMAGE
  if (!appImagePath) {
    throw new Error(
      'KYARABEN_APPIMAGE environment variable must be set to the path of the Electron executable',
    )
  }

  fixture = createFixture({}, undefined)

  console.log(`Testing: ${appImagePath}`)
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

test.describe('Kyaraben App', () => {
  test('displays the app title', async () => {
    await expect(page.getByRole('heading', { level: 1, name: 'Kyaraben' })).toBeVisible()
  })

  test('shows navigation tabs', async () => {
    await expect(page.getByRole('button', { name: 'Systems' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Installation' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Sync' })).toBeVisible()
  })

  test('displays manufacturer groupings', async () => {
    await expect(page.getByRole('heading', { level: 2, name: 'Nintendo' })).toBeVisible({
      timeout: 10000,
    })
    await expect(page.getByRole('heading', { level: 2, name: 'Sony' })).toBeVisible()
  })

  test('displays system cards with emulators', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard).toBeVisible({ timeout: 10000 })
    await expect(snesCard.getByRole('switch')).toBeVisible()
  })

  test('can toggle emulator selection', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    const toggle = snesCard.getByRole('switch').first()

    const wasChecked = (await toggle.getAttribute('aria-checked')) === 'true'
    await toggle.click()
    const isChecked = (await toggle.getAttribute('aria-checked')) === 'true'

    expect(isChecked).toBe(!wasChecked)
  })

  test('shows sticky action bar when changes are made', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    const toggle = snesCard.getByRole('switch').first()

    const wasChecked = (await toggle.getAttribute('aria-checked')) === 'true'
    if (!wasChecked) {
      await toggle.click()
    }

    await expect(page.getByRole('button', { name: 'Apply' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Discard' })).toBeVisible()
  })

  test('can discard changes', async () => {
    const discardButton = page.getByRole('button', { name: 'Discard' })
    if (await discardButton.isVisible()) {
      await discardButton.click()
      await expect(discardButton).not.toBeVisible()
    }
  })

  test('shows emulation folder setting', async () => {
    await expect(page.getByText('Emulation folder')).toBeVisible()
    const input = page.getByPlaceholder('~/Emulation')
    await expect(input).toBeVisible()
  })

  test('can change emulation folder path', async () => {
    const input = page.getByPlaceholder('~/Emulation')
    await input.clear()
    await input.fill('~/TestEmulation')
    await expect(input).toHaveValue('~/TestEmulation')

    await input.clear()
    await input.fill('~/Emulation')
  })
})
