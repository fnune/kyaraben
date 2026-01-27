import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'

let electronApp: ElectronApplication
let page: Page

test.beforeAll(async () => {
  const appImagePath = process.env.KYARABEN_APPIMAGE
  if (!appImagePath) {
    throw new Error(
      'KYARABEN_APPIMAGE environment variable must be set to the path of the Electron executable',
    )
  }

  console.log(`Testing: ${appImagePath}`)
  electronApp = await electron.launch({
    executablePath: appImagePath,
    args: ['--no-sandbox'],
  })

  page = await electronApp.firstWindow()
  await page.getByRole('heading', { level: 1 }).waitFor()
})

test.afterAll(async () => {
  if (electronApp) {
    await electronApp.close()
  }
})

test.describe('Kyaraben App', () => {
  test('displays the main title', async () => {
    await expect(page.getByRole('heading', { level: 1, name: 'Kyaraben' })).toBeVisible()
  })

  test('displays the tagline', async () => {
    await expect(page.getByText('Declarative emulation manager')).toBeVisible()
  })

  test('loads and displays available systems', async () => {
    // Wait for systems to load - TIC-80 should appear
    await expect(page.getByRole('heading', { name: 'TIC-80' })).toBeVisible({ timeout: 10000 })
  })

  test('can toggle system selection', async () => {
    // Find the TIC-80 system card and its checkbox
    const tic80Card = page.getByRole('article').filter({ hasText: 'TIC-80' })
    const checkbox = tic80Card.getByRole('checkbox')

    const wasChecked = await checkbox.isChecked()
    await checkbox.click()
    const isChecked = await checkbox.isChecked()

    expect(isChecked).toBe(!wasChecked)
  })

  test('shows emulation folder setting', async () => {
    await expect(page.getByText('Emulation folder')).toBeVisible()
    const input = page.getByLabel('Emulation folder')
    await expect(input).toBeVisible()
    await expect(input).toHaveValue(/Emulation/)
  })

  test('can change emulation folder path', async () => {
    const input = page.getByLabel('Emulation folder')
    await input.clear()
    await input.fill('~/TestEmulation')
    await expect(input).toHaveValue('~/TestEmulation')
  })

  test('displays manufacturer groupings', async () => {
    // Systems should be grouped by manufacturer
    await expect(page.getByRole('heading', { level: 2, name: 'Other' })).toBeVisible({
      timeout: 10000,
    })
  })

  test('shows Apply and Check provisions buttons', async () => {
    await expect(page.getByRole('button', { name: 'Apply' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Check provisions' })).toBeVisible()
  })
})

test.describe('Kyaraben Apply (requires Nix)', () => {
  test.skip(!!process.env.CI, 'Skipped on CI - Nix builds are too slow')
  test('can apply configuration with e2e-test system', async () => {
    // Find and enable the e2e-test system
    const e2eCard = page.getByRole('article').filter({ hasText: 'e2e-test' })
    const checkbox = e2eCard.getByRole('checkbox')

    if (!(await checkbox.isChecked())) {
      await checkbox.click()
    }

    // Click Apply
    await page.getByRole('button', { name: 'Apply' }).click()

    // Wait for completion (Nix builds can take 10+ minutes on first run)
    await expect(page.getByText(/Complete|Error/)).toBeVisible({ timeout: 600000 })

    // Verify success
    await expect(page.getByText('Complete')).toBeVisible()
  })
})
