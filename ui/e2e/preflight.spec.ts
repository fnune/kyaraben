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
  buildEnv,
  createFixture,
  EmulatorIDDuckStation,
  EmulatorIDRetroArchMGBA,
  SystemIDGBA,
  SystemIDPSX,
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
          [SystemIDGBA]: [EmulatorIDRetroArchMGBA],
        },
      },
      undefined,
    )

    // Create retroarch config file with a user-modified roms directory value
    const retroarchConfigDir = path.join(fixture.configDir, 'retroarch')
    fs.mkdirSync(retroarchConfigDir, { recursive: true })

    const configContent = [
      `libretro_directory = "${fixture.collection}/cores"`,
      'rgui_browser_directory = "/user/custom/roms/path"',
      'sort_savefiles_enable = "true"',
      'sort_savestates_enable = "true"',
    ].join('\n')

    fs.writeFileSync(path.join(retroarchConfigDir, 'retroarch.cfg'), configContent)

    // Write manifest that records what kyaraben wrote (before user modified the file)
    const manifest = {
      version: 1,
      last_applied: new Date().toISOString(),
      installed_emulators: {
        [EmulatorIDRetroArchMGBA]: {
          id: EmulatorIDRetroArchMGBA,
          version: '1.19.1',
          package_path: path.join(fixture.stateDir, 'kyaraben', 'packages', 'retroarch'),
          installed: new Date().toISOString(),
        },
      },
      managed_configs: [
        {
          emulator_ids: [EmulatorIDRetroArchMGBA],
          target: {
            RelPath: 'retroarch/retroarch.cfg',
            Format: 'cfg',
            BaseDir: 'user_config',
          },
          written_entries: {
            rgui_browser_directory: `"${fixture.collection}/roms"`,
          },
          last_modified: new Date().toISOString(),
          managed_regions: [{ type: 'file' }],
        },
      ],
      desktop_files: [],
      icon_files: [],
    }

    fs.writeFileSync(
      path.join(fixture.stateDir, 'kyaraben', 'build', 'manifest.json'),
      JSON.stringify(manifest, null, 2),
    )

    app = await electron.launch({
      executablePath: getAppImagePath(),
      args: ['--no-sandbox'],
      env: buildEnv(fixture),
    })

    page = await app.firstWindow()
    await page.getByRole('img', { name: 'Kyaraben' }).waitFor({ timeout: 30000 })
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('shows conflict review with details and Cancel returns to catalog view', async () => {
    // GBA is already enabled. Enable SNES to create a change that shows the Apply button.
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await snesCard.getByRole('switch').first().click()

    await page.getByRole('button', { name: 'Apply' }).click()
    await expect(page.getByText('Config conflicts detected')).toBeVisible({ timeout: 10000 })

    // Review screen shows conflict details
    await expect(
      page.getByText('You modified settings managed by kyaraben (will be overwritten):'),
    ).toBeVisible()
    await expect(page.getByText('retroarch/retroarch.cfg')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Open file' }).first()).toBeVisible()
    await expect(page.getByRole('button', { name: 'Continue and override' })).toBeVisible()

    // Cancel returns to systems view
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(page.getByText('Collection')).toBeVisible()
    await expect(page.getByText('Config conflicts detected')).not.toBeVisible()
  })

  test('Continue and override completes the apply', async () => {
    // Re-trigger apply — SNES toggle is still on from earlier
    await page.getByRole('button', { name: 'Apply' }).click()
    await expect(page.getByText('Config conflicts detected')).toBeVisible({ timeout: 10000 })

    await page.getByRole('button', { name: 'Continue and override' }).click()

    await expect(
      page.getByText(/Installing|Applying configuration|Setting up/).first(),
    ).toBeVisible({ timeout: 5000 })

    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible({ timeout: 30000 })

    await page.getByRole('button', { name: 'Done' }).click()
    await expect(page.getByText('Collection')).toBeVisible()
  })
})

test.describe('UI-driven config change', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    fixture = createFixture(
      {
        systems: {
          [SystemIDGBA]: [EmulatorIDRetroArchMGBA],
        },
      },
      undefined,
    )

    app = await electron.launch({
      executablePath: getAppImagePath(),
      args: ['--no-sandbox'],
      env: buildEnv(fixture),
    })

    page = await app.firstWindow()
    await page.getByRole('img', { name: 'Kyaraben' }).waitFor({ timeout: 30000 })
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('changing nintendo_confirm does not trigger conflict review', async () => {
    await page.getByRole('button', { name: 'Apply' }).click()

    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible({ timeout: 30000 })

    await page.getByRole('button', { name: 'Done' }).click()
    await expect(page.getByText('Collection')).toBeVisible()

    await page.getByText('View preferences').click()

    const southButton = page.getByRole('button', { name: /South button confirms/ })
    await expect(southButton).toBeVisible()
    await southButton.click()

    await expect(page.getByRole('button', { name: 'Apply' })).toBeVisible()
    await page.getByRole('button', { name: 'Apply' }).click()

    await expect(page.getByText('Config conflicts detected')).not.toBeVisible({ timeout: 2000 })
    await expect(page.getByText('Kyaraben has updated its defaults')).not.toBeVisible()

    await expect(page.getByText('Installation complete')).toBeVisible({ timeout: 30000 })
    await expect(page.getByRole('heading', { name: 'Display' })).toBeVisible()
  })
})

test.describe('Version upgrade review', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    fixture = createFixture(
      {
        systems: {
          [SystemIDPSX]: [EmulatorIDDuckStation],
        },
      },
      undefined,
    )

    const dataDir = fixture.env.XDG_DATA_HOME
    const duckstationDataDir = path.join(dataDir, 'duckstation')
    fs.mkdirSync(duckstationDataDir, { recursive: true })

    const configContent = [
      '[Main]',
      'SettingsVersion = 3',
      `GamePaths = ${fixture.collection}/roms/psx`,
      '[AutoUpdater]',
      'CheckAtStartup = true',
    ].join('\n')

    fs.writeFileSync(path.join(duckstationDataDir, 'settings.ini'), configContent)

    const manifest = {
      version: 1,
      last_applied: new Date().toISOString(),
      installed_emulators: {
        [EmulatorIDDuckStation]: {
          id: EmulatorIDDuckStation,
          version: '0.1.0',
          package_path: path.join(fixture.stateDir, 'kyaraben', 'packages', 'duckstation'),
          installed: new Date().toISOString(),
        },
      },
      managed_configs: [
        {
          emulator_ids: [EmulatorIDDuckStation],
          target: {
            RelPath: 'duckstation/settings.ini',
            Format: 'ini',
            BaseDir: 'user_data',
          },
          written_entries: {
            'AutoUpdater.CheckAtStartup': 'true',
          },
          config_inputs_when_written: {},
          last_modified: new Date().toISOString(),
          managed_regions: [{ type: 'file' }],
        },
      ],
      desktop_files: [],
      icon_files: [],
    }

    fs.writeFileSync(
      path.join(fixture.stateDir, 'kyaraben', 'build', 'manifest.json'),
      JSON.stringify(manifest, null, 2),
    )

    app = await electron.launch({
      executablePath: getAppImagePath(),
      args: ['--no-sandbox'],
      env: buildEnv(fixture),
    })

    page = await app.firstWindow()
    await page.getByRole('img', { name: 'Kyaraben' }).waitFor({ timeout: 30000 })
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('shows version upgrade review when kyaraben defaults changed', async () => {
    await page.getByRole('button', { name: 'Apply' }).click()

    await expect(page.getByText('Kyaraben has updated its defaults')).toBeVisible({
      timeout: 10000,
    })
    await expect(page.getByText('duckstation/settings.ini')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Continue' })).toBeVisible()

    await page.getByRole('button', { name: 'Continue' }).click()

    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible({ timeout: 30000 })

    await page.getByRole('button', { name: 'Done' }).click()
    await expect(page.getByText('Collection')).toBeVisible()
  })
})
