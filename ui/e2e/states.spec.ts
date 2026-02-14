import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'
import {
  createFixture,
  EmulatorIDRetroArchBsnes,
  presets,
  SystemIDSNES,
  type TestFixture,
} from './fixtures'

async function launchWithFixture(
  fixture: TestFixture,
  appImagePath: string,
): Promise<{ app: ElectronApplication; page: Page }> {
  const app = await electron.launch({
    executablePath: appImagePath,
    args: ['--no-sandbox'],
    env: {
      ...process.env,
      ...fixture.env,
    },
  })

  const page = await app.firstWindow()
  await page.getByRole('heading', { level: 1 }).waitFor({ timeout: 30000 })

  return { app, page }
}

function getAppImagePath(): string {
  const appImagePath = process.env.KYARABEN_APPIMAGE
  if (!appImagePath) {
    throw new Error('KYARABEN_APPIMAGE environment variable must be set')
  }
  return appImagePath
}

test.describe('Fresh install state', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    const preset = presets.freshInstall()
    fixture = createFixture(preset.config, preset.manifest)
    const result = await launchWithFixture(fixture, getAppImagePath())
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('shows default emulation folder path', async () => {
    const input = page.getByPlaceholder('~/Emulation')
    await expect(input).toBeVisible()
  })

  test('shows no systems enabled', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard).toBeVisible()

    const toggle = snesCard.getByRole('switch').first()
    const isChecked = await toggle.getAttribute('aria-checked')
    expect(isChecked).toBe('false')
  })

  test('action bar is not visible without changes', async () => {
    await expect(page.getByRole('button', { name: 'Apply' })).not.toBeVisible()
  })
})

test.describe('Systems enabled but not installed', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    const preset = presets.systemsEnabledNotInstalled()
    fixture = createFixture(preset.config, preset.manifest)
    const result = await launchWithFixture(fixture, getAppImagePath())
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('shows enabled systems with toggle on', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard).toBeVisible()

    const toggle = snesCard.getByRole('switch').first()
    const isChecked = await toggle.getAttribute('aria-checked')
    expect(isChecked).toBe('true')
  })

  test('shows PSX system card', async () => {
    const psxCard = page
      .getByRole('article')
      .filter({ hasText: /PlayStation.*Sony · 1994/ })
      .first()
    await expect(psxCard).toBeVisible()
  })

  test('does not show Launch button for non-installed emulator', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard.getByText('Launch')).not.toBeVisible()
  })

  test('shows download size for non-installed emulator', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard.getByText(/\d+(\.\d+)?\s*(MB|GB|KB)/)).toBeVisible()
  })
})

test.describe('Emulators installed', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    const preset = presets.emulatorsInstalled()
    fixture = createFixture(preset.config, preset.manifest)
    const result = await launchWithFixture(fixture, getAppImagePath())
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('shows Launch button for installed emulator', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard).toBeVisible()
    await expect(snesCard.getByText('Launch')).toBeVisible()
  })

  test('shows Paths button for installed emulator', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard.getByText('Paths')).toBeVisible()
  })
})

test.describe('Sync disabled', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    fixture = createFixture(
      { systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] } },
      { installedEmulators: {} },
    )
    const result = await launchWithFixture(fixture, getAppImagePath())
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('shows sync disabled message', async () => {
    await page.getByRole('button', { name: 'Sync' }).click()
    await expect(page.getByText('Sync is not enabled')).toBeVisible()
    await expect(page.getByText('enabled = true')).toBeVisible()
  })
})

test.describe('Installation tab', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    const preset = presets.emulatorsInstalled()
    fixture = createFixture(preset.config, preset.manifest)
    const result = await launchWithFixture(fixture, getAppImagePath())
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('can navigate to Installation tab', async () => {
    await page.getByRole('button', { name: 'Installation' }).click()
    await expect(page.getByText('State directory')).toBeVisible()
  })

  test('shows state directory path', async () => {
    await expect(page.getByText(/state\/kyaraben/).first()).toBeVisible()
  })

  test('shows actions section with uninstall button', async () => {
    await expect(page.getByText('Uninstall Kyaraben')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Uninstall' })).toBeVisible()
  })

  test('shows configuration section', async () => {
    await expect(page.getByRole('heading', { name: 'Configuration' })).toBeVisible()
  })

  test('shows open config button', async () => {
    await expect(page.getByRole('button', { name: 'Open' })).toBeVisible()
  })
})

test.describe('Version pinning', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    const preset = presets.versionPinned()
    fixture = createFixture(preset.config, preset.manifest)
    const result = await launchWithFixture(fixture, getAppImagePath())
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('shows version selector with pinned version', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard).toBeVisible()

    const versionSelect = snesCard.getByRole('combobox')
    await expect(versionSelect).toBeVisible()
  })
})

test.describe('Apply flow', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    const preset = presets.freshInstall()
    fixture = createFixture(preset.config, preset.manifest)

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

  test('emulator starts disabled', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    const toggle = snesCard.getByRole('switch').first()
    await expect(toggle).toHaveAttribute('aria-checked', 'false')
  })

  test('can enable emulator', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    const toggle = snesCard.getByRole('switch').first()

    await toggle.click()
    await expect(toggle).toHaveAttribute('aria-checked', 'true')
  })

  test('shows download size for enabled emulator', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard.getByText(/\d+(\.\d+)?\s*(MB|GB|KB)/)).toBeVisible()
  })

  test('shows action bar with Apply button', async () => {
    await expect(page.getByRole('button', { name: 'Apply' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Discard' })).toBeVisible()
  })

  test('clicking Apply shows progress', async () => {
    await page.getByRole('button', { name: 'Apply' }).click()

    await expect(
      page.getByText(/Applying configuration|Installing emulators|Setting up/).first(),
    ).toBeVisible({ timeout: 5000 })
  })

  test('progress completes and shows Done button', async () => {
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible({ timeout: 30000 })
  })

  test('clicking Done returns to systems view', async () => {
    await page.getByRole('button', { name: 'Done' }).click()

    await expect(page.getByRole('button', { name: 'Apply' })).not.toBeVisible()
    await expect(page.getByText('Emulation folder')).toBeVisible()
  })

  test('emulator now shows Launch button', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard.getByText('Launch')).toBeVisible({ timeout: 5000 })
  })
})

test.describe('Enable all flow', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    const preset = presets.freshInstall()
    fixture = createFixture(preset.config, preset.manifest)

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

  test('Enable all shows toast and enables systems', async () => {
    await page.getByRole('button', { name: 'Enable all systems' }).click()
    await expect(page.getByText('All systems enabled')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Apply' })).toBeVisible()
  })

  test('Apply installs all emulators', async () => {
    await page.getByRole('button', { name: 'Apply' }).click()
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible({ timeout: 60000 })
  })

  test('Multiple emulators show Launch button after install', async () => {
    await page.getByRole('button', { name: 'Done' }).click()

    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard.getByText('Launch')).toBeVisible({ timeout: 5000 })

    const nesCard = page.getByRole('article').filter({ hasText: 'Nintendo Entertainment System' })
    await expect(nesCard.getByText('Launch')).toBeVisible()

    const psxCard = page
      .getByRole('article')
      .filter({ hasText: /PlayStation.*Sony · 1994/ })
      .first()
    await expect(psxCard.getByText('Launch')).toBeVisible()
  })
})

test.describe('Tab navigation', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    fixture = createFixture({}, undefined)
    const result = await launchWithFixture(fixture, getAppImagePath())
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('starts on Systems tab', async () => {
    await expect(page.getByText('Emulation folder')).toBeVisible()
    await expect(page.getByRole('heading', { level: 2, name: 'Nintendo' })).toBeVisible()
  })

  test('can switch to Installation and back', async () => {
    await page.getByRole('button', { name: 'Installation' }).click()
    await expect(page.getByText('State directory')).toBeVisible()

    await page.getByRole('button', { name: 'Systems', exact: true }).click()
    await expect(page.getByText('Emulation folder')).toBeVisible()
  })

  test('can switch to Sync and back', async () => {
    await page.getByRole('button', { name: 'Sync' }).click()
    await expect(page.getByRole('heading', { name: /Sync/ })).toBeVisible()

    await page.getByRole('button', { name: 'Systems', exact: true }).click()
    await expect(page.getByText('Emulation folder')).toBeVisible()
  })
})
