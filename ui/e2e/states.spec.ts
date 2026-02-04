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
  setupFakeNixPortable,
  type TestFixture,
} from './fixtures'

async function launchWithFixture(
  fixture: TestFixture,
  appImagePath: string,
): Promise<{ app: ElectronApplication; page: Page }> {
  setupFakeNixPortable(fixture)

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
    const psxCard = page.getByRole('article').filter({ hasText: 'PlayStationSony · 1994' }).first()
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
    await expect(page.getByText(/\.local\/state\/kyaraben|kyaraben-e2e/)).toBeVisible()
  })

  test('shows preserved paths section', async () => {
    await expect(page.getByText('Preserved on uninstall')).toBeVisible()
  })

  test('shows uninstall section', async () => {
    await expect(page.getByRole('heading', { name: 'Uninstall' })).toBeVisible()
    await expect(page.getByText(/kyaraben uninstall/)).toBeVisible()
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

    const versionSelect = snesCard.locator('select')
    await expect(versionSelect).toBeVisible()
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

    await page.getByRole('button', { name: 'Systems' }).click()
    await expect(page.getByText('Emulation folder')).toBeVisible()
  })

  test('can switch to Sync and back', async () => {
    await page.getByRole('button', { name: 'Sync' }).click()
    await expect(page.getByRole('heading', { name: /Sync/ })).toBeVisible()

    await page.getByRole('button', { name: 'Systems' }).click()
    await expect(page.getByText('Emulation folder')).toBeVisible()
  })
})
