import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'
import { buildEnv,
  getElectronArgs, createFixture, type TestFixture } from './fixtures'

const SystemIDSNES = 'snes'
const EmulatorIDRetroArchBsnes = 'retroarch:bsnes'
const EmulatorIDRetroArchSnes9x = 'retroarch:snes9x'

async function launchWithFixture(
  fixture: TestFixture,
  appImagePath: string,
): Promise<{ app: ElectronApplication; page: Page }> {
  const app = await electron.launch({
    executablePath: appImagePath,
    args: getElectronArgs(),
    env: buildEnv(fixture),
  })

  const page = await app.firstWindow()
  await page.getByRole('img', { name: 'Kyaraben' }).waitFor({ timeout: 30000 })

  return { app, page }
}

function getAppImagePath(): string {
  const appImagePath = process.env.KYARABEN_APPIMAGE
  if (!appImagePath) {
    throw new Error('KYARABEN_APPIMAGE environment variable must be set')
  }
  return appImagePath
}

test.describe('Default emulator indicator', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    fixture = createFixture(
      {
        systems: {
          [SystemIDSNES]: [EmulatorIDRetroArchBsnes, EmulatorIDRetroArchSnes9x],
        },
      },
      {
        installedEmulators: {},
      },
    )
    const result = await launchWithFixture(fixture, getAppImagePath())
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('shows filled star on default emulator when alternatives exist', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard).toBeVisible()

    const bsnesDefaultButton = snesCard.getByRole('button', { name: 'Default emulator' })
    await expect(bsnesDefaultButton).toBeVisible()
  })

  test('shows outline star on non-default emulator', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    const snes9xDefaultButton = snesCard.getByRole('button', { name: 'Set as default' })
    await expect(snes9xDefaultButton).toBeVisible()
  })

  test('clicking star on non-default emulator makes it default', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })

    const snes9xSetDefaultButton = snesCard.getByRole('button', { name: 'Set as default' })
    await snes9xSetDefaultButton.click()

    await expect(snesCard.getByRole('button', { name: 'Default emulator' })).toBeVisible()
    await expect(snesCard.getByRole('button', { name: 'Set as default' })).toBeVisible()
  })
})

test.describe('Default emulator with single emulator', () => {
  let fixture: TestFixture
  let app: ElectronApplication
  let page: Page

  test.beforeAll(async () => {
    fixture = createFixture(
      {
        systems: {
          [SystemIDSNES]: [EmulatorIDRetroArchBsnes],
        },
      },
      {
        installedEmulators: {},
      },
    )
    const result = await launchWithFixture(fixture, getAppImagePath())
    app = result.app
    page = result.page
  })

  test.afterAll(async () => {
    await app?.close()
    fixture?.cleanup()
  })

  test('does not show star when only one emulator is enabled', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard).toBeVisible()

    await expect(snesCard.getByRole('button', { name: 'Default emulator' })).not.toBeVisible()
    await expect(snesCard.getByRole('button', { name: 'Set as default' })).not.toBeVisible()
  })
})

test.describe('Default emulator not shown when disabled', () => {
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

  test('does not show star on disabled emulator', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard).toBeVisible()

    await expect(snesCard.getByRole('button', { name: 'Default emulator' })).not.toBeVisible()
    await expect(snesCard.getByRole('button', { name: 'Set as default' })).not.toBeVisible()
  })
})
