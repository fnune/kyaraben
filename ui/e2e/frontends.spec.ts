import { type ElectronApplication, expect, type Page, test } from '@playwright/test'
import { createFixture, launchElectron, presets, type TestFixture } from './fixtures'

test.describe('Frontend installation', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    const preset = presets.freshInstall()
    fixture = createFixture(preset.config, preset.manifest)
    const result = await launchElectron(fixture)
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('can enable and install EmulationStation DE frontend', async () => {
    const esdeSection = page
      .locator('div')
      .filter({ hasText: /^EmulationStation DE/ })
      .first()
    const toggle = esdeSection.getByRole('switch')
    await expect(toggle).toHaveAttribute('aria-checked', 'false')

    await toggle.click()
    await expect(toggle).toHaveAttribute('aria-checked', 'true')

    const applyButton = page.getByRole('button', { name: 'Apply' })
    await expect(applyButton).toBeVisible()

    await applyButton.click()
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible({ timeout: 30000 })

    await page.getByRole('button', { name: 'Done' }).click()
    await expect(page.getByText('Collection')).toBeVisible()

    const esdeToggleAfter = page
      .locator('div')
      .filter({ hasText: /^EmulationStation DE/ })
      .first()
      .getByRole('switch')
    await expect(esdeToggleAfter).toHaveAttribute('aria-checked', 'true')

    await expect(page.getByRole('button', { name: 'Apply' })).not.toBeVisible()
  })
})

test.describe('Frontend already enabled', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    const preset = presets.frontendEnabled()
    fixture = createFixture(preset.config, preset.manifest)
    const result = await launchElectron(fixture)
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('shows EmulationStation DE as enabled when configured', async () => {
    const esdeSection = page
      .locator('div')
      .filter({ hasText: /^EmulationStation DE/ })
      .first()
    const toggle = esdeSection.getByRole('switch')
    await expect(toggle).toHaveAttribute('aria-checked', 'true')
  })
})
