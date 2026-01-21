import { test, expect } from '@playwright/test'

// Mock data for Tauri API calls
const mockSystems = [
  {
    id: 'tic80',
    name: 'TIC-80',
    description: 'Fantasy console',
    emulators: [{ id: 'tic80', name: 'TIC-80' }],
  },
  {
    id: 'snes',
    name: 'Super Nintendo',
    description: 'Super Nintendo Entertainment System',
    emulators: [{ id: 'retroarch-bsnes', name: 'RetroArch (bsnes)' }],
  },
  {
    id: 'psx',
    name: 'PlayStation',
    description: 'Sony PlayStation',
    emulators: [{ id: 'duckstation', name: 'DuckStation' }],
  },
]

const mockConfig = {
  userStore: '~/Emulation',
  systems: {},
}

// Setup mock Tauri API before each test
test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    window.__TAURI_INTERNALS__ = {
      invoke: async (cmd, args) => {
        switch (cmd) {
          case 'get_systems':
            return window.__mockSystems
          case 'get_config':
            return window.__mockConfig
          case 'set_config':
            window.__mockConfig = {
              userStore: args.userStore,
              systems: args.systems,
            }
            return { success: true }
          case 'status':
            return {
              userStore: window.__mockConfig.userStore,
              enabledSystems: Object.keys(window.__mockConfig.systems),
              installedEmulators: [],
              lastApplied: null,
            }
          case 'doctor':
            return {
              tic80: [],
              psx: [
                {
                  filename: 'scph5501.bin',
                  description: 'USA BIOS',
                  required: true,
                  status: 'missing',
                },
              ],
            }
          case 'apply':
            return ['Creating directories...', 'Building emulators...', 'Done!']
          default:
            throw new Error(`Unknown command: ${cmd}`)
        }
      },
    }
  })

  // Inject mock data
  await page.addInitScript(
    (data) => {
      window.__mockSystems = data.systems
      window.__mockConfig = data.config
    },
    { systems: mockSystems, config: mockConfig }
  )
})

test.describe('Kyaraben UI', () => {
  test('loads and displays title', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h1')).toHaveText('Kyaraben')
  })

  test('displays available systems', async ({ page }) => {
    await page.goto('/')

    // Wait for systems to load
    await expect(page.locator('#system-list')).not.toContainText('Loading...')

    // Check systems are rendered
    await expect(page.locator('#system-list')).toContainText('TIC-80')
    await expect(page.locator('#system-list')).toContainText('Super Nintendo')
    await expect(page.locator('#system-list')).toContainText('PlayStation')
  })

  test('can select and deselect systems', async ({ page }) => {
    await page.goto('/')

    // Wait for systems to load
    await expect(page.locator('#system-list')).not.toContainText('Loading...')

    // Find TIC-80 checkbox
    const tic80Checkbox = page.locator('input[value="tic80"]')
    await expect(tic80Checkbox).toBeVisible()

    // Select it
    await tic80Checkbox.check()
    await expect(tic80Checkbox).toBeChecked()

    // Deselect it
    await tic80Checkbox.uncheck()
    await expect(tic80Checkbox).not.toBeChecked()
  })

  test('shows settings when expanded', async ({ page }) => {
    await page.goto('/')

    // Settings should be in a details element
    const settings = page.locator('details')
    const userStoreInput = page.locator('#user-store')

    // Initially closed (input not visible)
    await expect(userStoreInput).not.toBeVisible()

    // Expand settings
    await settings.locator('summary').click()

    // Now input should be visible
    await expect(userStoreInput).toBeVisible()
    await expect(userStoreInput).toHaveValue('~/Emulation')
  })

  test('status button shows current status', async ({ page }) => {
    await page.goto('/')

    // Wait for systems to load
    await expect(page.locator('#system-list')).not.toContainText('Loading...')

    // Click status button
    await page.locator('#btn-status').click()

    // Output section should be visible
    await expect(page.locator('#output-section')).toBeVisible()

    // Check log contains status info
    const log = page.locator('#log')
    await expect(log).toContainText('Emulation folder')
  })

  test('doctor button shows provisions', async ({ page }) => {
    await page.goto('/')

    // Wait for systems to load
    await expect(page.locator('#system-list')).not.toContainText('Loading...')

    // Click doctor button
    await page.locator('#btn-doctor').click()

    // Provisions section should be visible
    await expect(page.locator('#provisions-section')).toBeVisible()

    // Should show PSX BIOS requirement
    const provisionsList = page.locator('#provisions-list')
    await expect(provisionsList).toContainText('scph5501.bin')
    await expect(provisionsList).toContainText('MISSING')
  })

  test('action buttons are visible', async ({ page }) => {
    await page.goto('/')

    await expect(page.locator('#btn-apply')).toBeVisible()
    await expect(page.locator('#btn-apply')).toHaveText('Apply')

    await expect(page.locator('#btn-doctor')).toBeVisible()
    await expect(page.locator('#btn-doctor')).toHaveText('Check provisions')

    await expect(page.locator('#btn-status')).toBeVisible()
    await expect(page.locator('#btn-status')).toHaveText('Status')
  })
})
