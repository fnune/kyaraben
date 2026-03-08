import { type ElectronApplication, expect, type Page, test } from '@playwright/test'
import { createFixture, launchElectron, type TestFixture } from './fixtures'

let fixture: TestFixture
let electronApp: ElectronApplication
let page: Page

test.beforeAll(async () => {
  fixture = createFixture({}, undefined)
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

test.describe('Kyaraben App', () => {
  test('displays the app title', async () => {
    await expect(page.getByRole('img', { name: 'Kyaraben' })).toBeVisible()
  })

  test('shows navigation tabs', async () => {
    await expect(page.getByRole('button', { name: 'Catalog', exact: true })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Synchronization' })).toBeVisible()
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
    await expect(snesCard.getByRole('switch').first()).toBeVisible()
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

  test('shows collection setting', async () => {
    await expect(page.getByText('Collection')).toBeVisible()
    const input = page.getByPlaceholder('~/Emulation')
    await expect(input).toBeVisible()
  })

  test('can change collection path', async () => {
    const input = page.getByPlaceholder('~/Emulation')
    await input.clear()
    await input.fill('~/EmulationTest')
    await expect(input).toHaveValue('~/EmulationTest')

    await input.clear()
    await input.fill('~/Emulation')
  })

  test('shows storage selector with internal storage card', async () => {
    await expect(page.getByRole('button', { name: /Internal storage/i })).toBeVisible()
  })

  test('shows custom folder option', async () => {
    await expect(page.getByRole('button', { name: /Custom folder/i })).toBeVisible()
  })

  test('shows storage info message', async () => {
    await expect(page.getByText(/Changing storage will not move existing files/i)).toBeVisible()
  })

  test('internal storage card shows free space info', async () => {
    const storageCard = page.getByRole('button', { name: /Internal storage/i })
    await expect(storageCard.getByText(/free of/i)).toBeVisible()
  })

  test('selecting internal storage updates path input', async () => {
    const internalCard = page.getByRole('button', { name: /Internal storage/i })
    await internalCard.click()

    const input = page.getByPlaceholder('~/Emulation')
    await expect(input).toHaveValue(/\/Emulation$/)
  })
})
