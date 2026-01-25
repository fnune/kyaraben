import * as path from 'node:path'
import { type ElectronApplication, type Page, _electron as electron } from '@playwright/test'
import { expect, test } from '@playwright/test'

let electronApp: ElectronApplication
let page: Page

test.beforeAll(async () => {
  const mainPath = path.join(__dirname, '..', 'dist-electron', 'main.js')

  // --no-sandbox required for Chromium in Docker: https://playwright.dev/docs/ci#docker
  electronApp = await electron.launch({
    args: [mainPath, '--no-sandbox'],
    cwd: path.join(__dirname, '..'),
  })

  page = await electronApp.firstWindow()
  await page.waitForSelector('h1')
})

test.afterAll(async () => {
  if (electronApp) {
    await electronApp.close()
  }
})

test.describe('Kyaraben App', () => {
  test('displays the main title', async () => {
    const title = page.locator('h1')
    await expect(title).toHaveText('Kyaraben')
  })

  test('loads and displays available systems', async () => {
    const systemList = page.locator('#system-list')
    await expect(systemList).not.toContainText('Loading...', { timeout: 10000 })
    await expect(systemList).toContainText('TIC-80')
  })

  test('can toggle system selection', async () => {
    const systemList = page.locator('#system-list')
    await expect(systemList).not.toContainText('Loading...', { timeout: 10000 })

    const tic80Checkbox = page.locator('input[value="tic80"]')
    const wasChecked = await tic80Checkbox.isChecked()
    await tic80Checkbox.click()
    const isChecked = await tic80Checkbox.isChecked()

    expect(isChecked).toBe(!wasChecked)
  })

  test('shows status when clicking Status button', async () => {
    const statusBtn = page.locator('#btn-status')
    await statusBtn.click()

    const outputSection = page.locator('#output-section')
    await expect(outputSection).toBeVisible()

    const log = page.locator('#log')
    await expect(log).toContainText('Emulation folder')
  })

  test('shows provisions when clicking Check provisions button', async () => {
    const doctorBtn = page.locator('#btn-doctor')
    await doctorBtn.click()

    const provisionsSection = page.locator('#provisions-section')
    await expect(provisionsSection).toBeVisible()
  })

  test('can expand settings', async () => {
    const details = page.locator('details')
    const summary = details.locator('summary')
    await summary.click()

    const userStoreInput = page.locator('#user-store')
    await expect(userStoreInput).toBeVisible()

    const value = await userStoreInput.inputValue()
    expect(value).toContain('Emulation')
  })

  test('can change user store path', async () => {
    const details = page.locator('details')
    const isOpen = await details.getAttribute('open')
    if (isOpen === null) {
      const summary = details.locator('summary')
      await summary.click()
    }

    const userStoreInput = page.locator('#user-store')
    await userStoreInput.clear()
    await userStoreInput.fill('~/TestEmulation')

    const value = await userStoreInput.inputValue()
    expect(value).toBe('~/TestEmulation')
  })
})

test.describe('Kyaraben Apply (requires Nix)', () => {
  test('can apply configuration with e2e-test system', async () => {
    const e2eCheckbox = page.locator('input[value="e2e-test"]')
    if (!(await e2eCheckbox.isChecked())) {
      await e2eCheckbox.click()
    }

    const applyBtn = page.locator('#btn-apply')
    await applyBtn.click()

    const outputSection = page.locator('#output-section')
    await expect(outputSection).toBeVisible()

    // Nix builds can take 10+ minutes on first run (no cache)
    const log = page.locator('#log')
    await expect(log).toContainText(/Done!|Error/, { timeout: 600000 })

    const logText = await log.textContent()
    expect(logText).toContain('Done!')
  })
})
