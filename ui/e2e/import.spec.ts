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
  type ConfigFixture,
  createFixture,
  EmulatorIDRetroArchBsnes,
  type ManifestFixture,
  SystemIDSNES,
  type TestFixture,
} from './fixtures'

function getAppImagePath(): string {
  const appImagePath = process.env.KYARABEN_APPIMAGE
  if (!appImagePath) {
    throw new Error('KYARABEN_APPIMAGE environment variable must be set')
  }
  return appImagePath
}

interface ImportTestContext {
  app: ElectronApplication
  page: Page
  fixture: TestFixture
  sourceCollection: string
}

function setupSourceCollection(fixture: TestFixture): string {
  const sourceDir = path.join(path.dirname(fixture.collection), 'ExistingCollection')
  fs.mkdirSync(sourceDir, { recursive: true })

  const romsDir = path.join(sourceDir, 'roms', 'snes')
  fs.mkdirSync(romsDir, { recursive: true })
  fs.writeFileSync(path.join(romsDir, 'game1.sfc'), Buffer.alloc(1024, 'A'))
  fs.writeFileSync(path.join(romsDir, 'game2.sfc'), Buffer.alloc(2048, 'B'))

  const savesDir = path.join(sourceDir, 'saves', 'snes')
  fs.mkdirSync(savesDir, { recursive: true })
  fs.writeFileSync(path.join(savesDir, 'game1.srm'), Buffer.alloc(512, 'C'))

  const biosDir = path.join(sourceDir, 'bios', 'psx')
  fs.mkdirSync(biosDir, { recursive: true })
  fs.writeFileSync(path.join(biosDir, 'scph1001.bin'), Buffer.alloc(4096, 'D'))

  return sourceDir
}

async function setupImportTest(options: {
  config: ConfigFixture
  manifest: ManifestFixture
  setupSource?: boolean
}): Promise<ImportTestContext> {
  const fixture = createFixture(options.config, options.manifest)

  const sourceCollection = options.setupSource !== false ? setupSourceCollection(fixture) : ''

  const app = await electron.launch({
    executablePath: getAppImagePath(),
    args: ['--no-sandbox'],
    env: buildEnv(fixture),
  })

  const page = await app.firstWindow()
  await page.getByRole('img', { name: 'Kyaraben' }).waitFor({ timeout: 30000 })

  return { app, page, fixture, sourceCollection }
}

async function cleanupImportTest(ctx: ImportTestContext): Promise<void> {
  await ctx.app?.close()
  ctx.fixture?.cleanup()
}

async function navigateToImport(page: Page): Promise<void> {
  const importButton = page.getByRole('button', { name: 'Import' })
  await expect(importButton).toBeVisible()
  await importButton.click()
  await expect(page.getByRole('heading', { name: 'Import' })).toBeVisible()
}

test.describe('Import view initial state', () => {
  let ctx: ImportTestContext

  test.beforeAll(async () => {
    ctx = await setupImportTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
      },
      manifest: { installedEmulators: {} },
      setupSource: true,
    })
  })

  test.afterAll(async () => {
    await cleanupImportTest(ctx)
  })

  test('shows input for source path', async () => {
    await navigateToImport(ctx.page)
    await expect(ctx.page.getByPlaceholder('~/Emulation')).toBeVisible()
    await ctx.page.waitForTimeout(2000)
  })

  test('shows input for ES-DE path', async () => {
    await expect(ctx.page.getByPlaceholder('~/ES-DE')).toBeVisible()
  })

  test('shows scan button disabled without source path', async () => {
    const scanButton = ctx.page.getByRole('button', { name: 'Scan' })
    await expect(scanButton).toBeVisible()
    await expect(scanButton).toBeDisabled()
  })
})

test.describe('Import scan with existing collection', () => {
  let ctx: ImportTestContext

  test.beforeAll(async () => {
    ctx = await setupImportTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
      },
      manifest: { installedEmulators: {} },
      setupSource: true,
    })
  })

  test.afterAll(async () => {
    await cleanupImportTest(ctx)
  })

  test('can enter source path and scan', async () => {
    await navigateToImport(ctx.page)
    await ctx.page.getByPlaceholder('~/Emulation').fill(ctx.sourceCollection)
    const scanButton = ctx.page.getByRole('button', { name: 'Scan' })
    await expect(scanButton).toBeEnabled()
    await scanButton.click()
    await expect(ctx.page.getByText('Back up your collection')).toBeVisible({ timeout: 10000 })
    await ctx.page.waitForTimeout(3000)
  })

  test('shows system sections with found data', async () => {
    await expect(ctx.page.getByText('Super Nintendo')).toBeVisible()
  })

  test('shows data type labels', async () => {
    await expect(ctx.page.getByText('ROMs', { exact: true })).toBeVisible()
    await expect(ctx.page.getByText('Saves', { exact: true })).toBeVisible()
  })

  test('shows file counts and sizes', async () => {
    await expect(ctx.page.getByText(/2 files/).first()).toBeVisible()
  })

  test('can change source path to go back', async () => {
    await ctx.page.getByRole('button', { name: 'Change' }).click()
    await expect(ctx.page.getByPlaceholder('~/Emulation')).toBeVisible()
    await ctx.page.waitForTimeout(2000)
  })
})

test.describe('Import scan with disabled systems', () => {
  let ctx: ImportTestContext

  test.beforeAll(async () => {
    ctx = await setupImportTest({
      config: {},
      manifest: { installedEmulators: {} },
      setupSource: true,
    })
  })

  test.afterAll(async () => {
    await cleanupImportTest(ctx)
  })

  test('shows not enabled indicator for disabled systems', async () => {
    await navigateToImport(ctx.page)
    await ctx.page.getByPlaceholder('~/Emulation').fill(ctx.sourceCollection)
    await ctx.page.getByRole('button', { name: 'Scan' }).click()
    await expect(ctx.page.getByText('Back up your collection')).toBeVisible({ timeout: 10000 })
    await expect(ctx.page.getByText('not enabled').first()).toBeVisible()
    await ctx.page.waitForTimeout(3000)
  })
})

test.describe('Import scan error handling', () => {
  let ctx: ImportTestContext

  test.beforeAll(async () => {
    ctx = await setupImportTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
      },
      manifest: { installedEmulators: {} },
      setupSource: false,
    })
  })

  test.afterAll(async () => {
    await cleanupImportTest(ctx)
  })

  test('shows warning for non-existent path', async () => {
    await navigateToImport(ctx.page)
    await ctx.page.getByPlaceholder('~/Emulation').fill('/nonexistent/path/that/does/not/exist')
    await ctx.page.getByRole('button', { name: 'Scan' }).click()
    await expect(ctx.page.getByText('That folder does not exist')).toBeVisible()
    await ctx.page.waitForTimeout(3000)
  })
})

test.describe('Import scan with rescan', () => {
  let ctx: ImportTestContext

  test.beforeAll(async () => {
    ctx = await setupImportTest({
      config: {
        systems: { [SystemIDSNES]: [EmulatorIDRetroArchBsnes] },
      },
      manifest: { installedEmulators: {} },
      setupSource: true,
    })
  })

  test.afterAll(async () => {
    await cleanupImportTest(ctx)
  })

  test('can rescan after initial scan', async () => {
    await navigateToImport(ctx.page)
    await ctx.page.getByPlaceholder('~/Emulation').fill(ctx.sourceCollection)
    await ctx.page.getByRole('button', { name: 'Scan' }).click()
    await expect(ctx.page.getByText('Back up your collection')).toBeVisible({ timeout: 10000 })
    await ctx.page.getByRole('button', { name: 'Rescan' }).click()
    await expect(ctx.page.getByText('Back up your collection')).toBeVisible({ timeout: 10000 })
    await ctx.page.waitForTimeout(2000)
  })
})
