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
  SystemIDGBA,
  type TestFixture,
} from './fixtures'

let fixture: TestFixture
let electronApp: ElectronApplication
let page: Page

test.beforeAll(async () => {
  const appImagePath = process.env.KYARABEN_APPIMAGE
  if (!appImagePath) {
    throw new Error(
      'KYARABEN_APPIMAGE environment variable must be set to the path of the Electron executable',
    )
  }

  fixture = createFixture(
    {
      systems: {
        [SystemIDGBA]: [EmulatorIDMGBA],
      },
    },
    undefined,
  )

  electronApp = await electron.launch({
    executablePath: appImagePath,
    args: ['--no-sandbox'],
    env: {
      ...process.env,
      ...fixture.env,
    },
  })

  page = await electronApp.firstWindow()
  await page.getByRole('heading', { level: 1 }).waitFor()
})

test.afterAll(async () => {
  if (electronApp) {
    await electronApp.close()
  }
  fixture?.cleanup()
})

test.describe('Preflight config review', () => {
  test('preflight IPC returns a valid response', async () => {
    const result = await page.evaluate(async () => {
      return window.electron.invoke('preflight')
    })

    expect(result).toBeDefined()
    expect(result).toHaveProperty('diffs')
    expect(Array.isArray(result.diffs)).toBe(true)
    expect(result).toHaveProperty('filesToBackup')
  })

  test('preflight response diffs have expected shape', async () => {
    const result = (await page.evaluate(async () => {
      return window.electron.invoke('preflight')
    })) as { diffs: Array<Record<string, unknown>> }

    for (const diff of result.diffs) {
      expect(diff).toHaveProperty('path')
      expect(diff).toHaveProperty('isNewFile')
      expect(diff).toHaveProperty('hasChanges')
      expect(diff).toHaveProperty('userModified')
      expect(typeof diff.path).toBe('string')
      expect(typeof diff.isNewFile).toBe('boolean')
      expect(typeof diff.hasChanges).toBe('boolean')
      expect(typeof diff.userModified).toBe('boolean')
    }
  })
})
