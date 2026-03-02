import * as fs from 'node:fs'
import * as path from 'node:path'
import {
  type ElectronApplication,
  _electron as electron,
  expect,
  type Page,
  test,
} from '@playwright/test'
import { buildEnv, createFixture, type TestFixture } from './fixtures'

let fixture: TestFixture
let electronApp: ElectronApplication
let page: Page

test.describe('Config warnings', () => {
  test.beforeAll(async () => {
    const appImagePath = process.env.KYARABEN_APPIMAGE
    if (!appImagePath) {
      throw new Error(
        'KYARABEN_APPIMAGE environment variable must be set to the path of the Electron executable',
      )
    }

    fixture = createFixture({}, undefined)

    const configPath = path.join(fixture.configDir, 'kyaraben', 'config.toml')
    const configWithInvalid = `[global]
collection = "${fixture.collection}"

[controller]
nintendo_confirm = "invalid_value"

[systems]
snes = ["retroarch:bsnes"]
unknown_system = ["unknown_emulator"]

[frontends.unknown_frontend]
enabled = true
`
    fs.writeFileSync(configPath, configWithInvalid)

    console.log(`Testing: ${appImagePath}`)
    electronApp = await electron.launch({
      executablePath: appImagePath,
      args: ['--no-sandbox'],
      env: buildEnv(fixture),
    })

    page = await electronApp.firstWindow()
    await page.getByRole('img', { name: 'Kyaraben' }).waitFor()
  })

  test.afterAll(async () => {
    if (electronApp) {
      await electronApp.close()
    }
    fixture?.cleanup()
  })

  test('shows warning toast for invalid config values', async () => {
    await expect(page.getByText(/Config issues found/i)).toBeVisible({ timeout: 15000 })
  })

  test('app still loads and functions despite invalid config', async () => {
    await expect(page.getByRole('heading', { level: 2, name: 'Nintendo' })).toBeVisible({
      timeout: 10000,
    })
    await expect(page.getByRole('heading', { level: 2, name: 'Sony' })).toBeVisible()
  })

  test('valid system is still enabled after filtering invalid ones', async () => {
    const snesCard = page.getByRole('article').filter({ hasText: 'Super Nintendo' })
    await expect(snesCard).toBeVisible({ timeout: 10000 })
    const toggle = snesCard.getByText(/bsnes/).locator('..').getByRole('switch')
    await expect(toggle).toHaveAttribute('aria-checked', 'true')
  })
})
