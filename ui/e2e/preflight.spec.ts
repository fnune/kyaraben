import * as fs from 'node:fs'
import * as path from 'node:path'
import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'
import {
  createFixture,
  EmulatorIDMGBA,
  setupFakeNixPortable,
  SystemIDGBA,
  type TestFixture,
} from './fixtures'

function getAppImagePath(): string {
  const appImagePath = process.env.KYARABEN_APPIMAGE
  if (!appImagePath) {
    throw new Error('KYARABEN_APPIMAGE environment variable must be set')
  }
  return appImagePath
}

test.describe('Config conflict review', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    fixture = createFixture(
      {
        systems: {
          [SystemIDGBA]: [EmulatorIDMGBA],
        },
      },
      undefined,
    )

    // Create mgba config file with a user-modified bios value
    const mgbaConfigDir = path.join(fixture.configDir, 'mgba')
    fs.mkdirSync(mgbaConfigDir, { recursive: true })

    const originalBiosValue = `${fixture.userStore}/bios/gba/gba_bios.bin`
    const configContent = [
      '; Configuration managed by kyaraben',
      '',
      '[ports.qt]',
      'bios = /user/custom/path/gba_bios.bin',
      `savegamePath = ${fixture.userStore}/saves/gba`,
      `savestatePath = ${fixture.userStore}/states/mgba`,
      `screenshotPath = ${fixture.userStore}/screenshots/gba`,
    ].join('\n')

    fs.writeFileSync(path.join(mgbaConfigDir, 'config.ini'), configContent)

    // Write manifest that records the original baseline (before user modified the file)
    const manifest = {
      version: 1,
      last_applied: new Date().toISOString(),
      installed_emulators: {
        [EmulatorIDMGBA]: {
          id: EmulatorIDMGBA,
          version: '0.10.3',
          store_path: '/nix/store/fake-hash-mgba',
          installed: new Date().toISOString(),
        },
      },
      managed_configs: [
        {
          emulator_id: EmulatorIDMGBA,
          target: {
            RelPath: 'mgba/config.ini',
            Format: 'ini',
            BaseDir: 'user_config',
          },
          baseline_hash: 'hash-before-user-modified-the-file',
          last_modified: new Date().toISOString(),
          managed_keys: [
            { path: ['ports.qt', 'bios'], value: originalBiosValue },
            { path: ['ports.qt', 'savegamePath'], value: `${fixture.userStore}/saves/gba` },
            { path: ['ports.qt', 'savestatePath'], value: `${fixture.userStore}/states/mgba` },
            { path: ['ports.qt', 'screenshotPath'], value: `${fixture.userStore}/screenshots/gba` },
          ],
        },
      ],
      desktop_files: [],
      icon_files: [],
    }

    fs.writeFileSync(
      path.join(fixture.stateDir, 'kyaraben', 'build', 'manifest.json'),
      JSON.stringify(manifest, null, 2),
    )

    setupFakeNixPortable(fixture)

    app = await electron.launch({
      executablePath: getAppImagePath(),
      args: ['--no-sandbox'],
      env: {
        ...process.env,
        ...fixture.env,
      },
    })

    page = await app.firstWindow()
    await page.getByRole('heading', { level: 1 }).waitFor({ timeout: 30000 })
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('clicking Apply shows conflict review when user modified configs', async () => {
    // GBA is already enabled. Enable SNES to create a change that shows the Apply button.
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    const toggle = snesCard.getByRole('switch').first()
    await toggle.click()

    const applyButton = page.getByRole('button', { name: 'Apply' })
    await expect(applyButton).toBeVisible()

    await applyButton.click()

    await expect(page.getByText('Config conflicts detected')).toBeVisible({ timeout: 10000 })
  })

  test('shows conflict details for user-modified file', async () => {
    await expect(page.getByText('You modified keys managed by kyaraben')).toBeVisible()
    await expect(page.getByText('mgba/config.ini')).toBeVisible()
    await expect(page.getByText('Open file')).toBeVisible()
  })

  test('shows action buttons', async () => {
    await expect(page.getByRole('button', { name: 'Continue and override' })).toBeVisible()
    await expect(page.getByText('Cancel')).toBeVisible()
  })

  test('clicking Cancel returns to systems view', async () => {
    await page.getByText('Cancel').click()

    await expect(page.getByText('Emulation folder')).toBeVisible()
    await expect(page.getByText('Config conflicts detected')).not.toBeVisible()
  })

  test('clicking Continue and override proceeds with apply', async () => {
    // Re-trigger apply — SNES is still toggled on from the earlier test
    await page.getByRole('button', { name: 'Apply' }).click()
    await expect(page.getByText('Config conflicts detected')).toBeVisible({ timeout: 10000 })

    await page.getByRole('button', { name: 'Continue and override' }).click()

    await expect(
      page.getByText(/Applying configuration|Installing emulators|Setting up/).first(),
    ).toBeVisible({ timeout: 5000 })
  })

  test('apply completes and shows Done button', async () => {
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible({ timeout: 30000 })
  })

  test('clicking Done returns to systems view', async () => {
    await page.getByRole('button', { name: 'Done' }).click()

    await expect(page.getByText('Emulation folder')).toBeVisible()
  })
})
