import * as path from 'node:path'
import { type ElectronApplication, type Page, _electron as electron } from '@playwright/test'
import { expect, test } from '@playwright/test'

let electronApp: ElectronApplication
let page: Page

test.beforeAll(async () => {
  // Point directly to the compiled main.js instead of directory
  const mainPath = path.join(__dirname, '..', 'dist-electron', 'main.js')
  console.log('[test] Main path:', mainPath)
  console.log('[test] DISPLAY:', process.env.DISPLAY)

  electronApp = await electron.launch({
    args: [
      mainPath,
      '--no-sandbox', // Required for running in Docker/CI
      '--disable-gpu',
      '--disable-dev-shm-usage', // Helps with Docker memory issues
    ],
    cwd: path.join(__dirname, '..'), // Set working directory to ui/
    env: {
      ...process.env,
      NODE_ENV: 'test',
      ELECTRON_DISABLE_SECURITY_WARNINGS: 'true',
    },
    timeout: 30000, // 30 second timeout for launch
  })
  console.log('[test] Electron launched successfully')

  page = await electronApp.firstWindow()
  console.log('[test] Got first window')

  // Wait for the app to be ready
  await page.waitForSelector('h1')
  console.log('[test] Page ready')
})

test.afterAll(async () => {
  if (electronApp) {
    await electronApp.close()
  }
})

test.describe('Kyaraben App', () => {
  test('displays the main title', async () => {
    const title = await page.locator('h1')
    await expect(title).toHaveText('Kyaraben')
  })

  test('loads and displays available systems', async () => {
    const systemList = page.locator('#system-list')

    // Wait for systems to load (no longer shows "Loading...")
    await expect(systemList).not.toContainText('Loading...', { timeout: 10000 })

    // Should show TIC-80 (always available, no BIOS needed)
    await expect(systemList).toContainText('TIC-80')
  })

  test('can toggle system selection', async () => {
    const systemList = page.locator('#system-list')
    await expect(systemList).not.toContainText('Loading...', { timeout: 10000 })

    // Find TIC-80 checkbox
    const tic80Checkbox = page.locator('input[value="tic80"]')
    const wasChecked = await tic80Checkbox.isChecked()

    // Toggle it
    await tic80Checkbox.click()
    const isChecked = await tic80Checkbox.isChecked()

    expect(isChecked).toBe(!wasChecked)
  })

  test('shows status when clicking Status button', async () => {
    const statusBtn = page.locator('#btn-status')
    await statusBtn.click()

    // Output section should become visible
    const outputSection = page.locator('#output-section')
    await expect(outputSection).toBeVisible()

    // Log should contain status info
    const log = page.locator('#log')
    await expect(log).toContainText('Emulation folder')
  })

  test('shows provisions when clicking Check provisions button', async () => {
    const doctorBtn = page.locator('#btn-doctor')
    await doctorBtn.click()

    // Provisions section should become visible
    const provisionsSection = page.locator('#provisions-section')
    await expect(provisionsSection).toBeVisible()
  })

  test('can expand settings', async () => {
    const details = page.locator('details')
    const summary = details.locator('summary')
    await summary.click()

    const userStoreInput = page.locator('#user-store')
    await expect(userStoreInput).toBeVisible()

    // Default value should be set
    const value = await userStoreInput.inputValue()
    expect(value).toContain('Emulation')
  })

  test('can change user store path', async () => {
    // Expand settings if not already
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
  test('can apply configuration with TIC-80', async () => {
    // This test requires Nix to be available
    // It tests the full stack: UI -> Electron -> Go daemon -> Nix

    // Select TIC-80 (no BIOS required)
    const tic80Checkbox = page.locator('input[value="tic80"]')
    if (!(await tic80Checkbox.isChecked())) {
      await tic80Checkbox.click()
    }

    // Deselect any systems that require BIOS
    const psxCheckbox = page.locator('input[value="psx"]')
    if (await psxCheckbox.isChecked()) {
      await psxCheckbox.click()
    }

    const snesCheckbox = page.locator('input[value="snes"]')
    if (await snesCheckbox.isChecked()) {
      await snesCheckbox.click()
    }

    // Click Apply
    const applyBtn = page.locator('#btn-apply')
    await applyBtn.click()

    // Wait for output section
    const outputSection = page.locator('#output-section')
    await expect(outputSection).toBeVisible()

    // Wait for completion (this can take a while with Nix)
    const log = page.locator('#log')
    await expect(log).toContainText(/Done!|Error/, { timeout: 840000 })

    const logText = await log.textContent()
    expect(logText).toContain('Done!')
  })
})
