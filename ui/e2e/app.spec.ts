import { expect } from '@wdio/globals'

describe('Kyaraben App', () => {
  it('displays the main title', async () => {
    const title = await $('h1')
    await expect(title).toHaveText('Kyaraben')
  })

  it('loads and displays available systems', async () => {
    const systemList = await $('#system-list')

    // Wait for systems to load (no longer shows "Loading...")
    await browser.waitUntil(
      async () => {
        const text = await systemList.getText()
        return !text.includes('Loading...')
      },
      { timeout: 10000, timeoutMsg: 'Systems did not load' }
    )

    // Should show TIC-80 (always available, no BIOS needed)
    const text = await systemList.getText()
    expect(text).toContain('TIC-80')
  })

  it('can toggle system selection', async () => {
    // Wait for systems to load
    const systemList = await $('#system-list')
    await browser.waitUntil(
      async () => {
        const text = await systemList.getText()
        return !text.includes('Loading...')
      },
      { timeout: 10000 }
    )

    // Find TIC-80 checkbox
    const tic80Checkbox = await $('input[value="tic80"]')

    // Toggle it
    const wasChecked = await tic80Checkbox.isSelected()
    await tic80Checkbox.click()
    const isChecked = await tic80Checkbox.isSelected()

    expect(isChecked).toBe(!wasChecked)
  })

  it('shows status when clicking Status button', async () => {
    const statusBtn = await $('#btn-status')
    await statusBtn.click()

    // Output section should become visible
    const outputSection = await $('#output-section')
    await expect(outputSection).toBeDisplayed()

    // Log should contain status info
    const log = await $('#log')
    const logText = await log.getText()
    expect(logText).toContain('Emulation folder')
  })

  it('shows provisions when clicking Check provisions button', async () => {
    const doctorBtn = await $('#btn-doctor')
    await doctorBtn.click()

    // Provisions section should become visible
    const provisionsSection = await $('#provisions-section')
    await expect(provisionsSection).toBeDisplayed()
  })

  it('can expand settings', async () => {
    const details = await $('details')
    const summary = await details.$('summary')
    await summary.click()

    const userStoreInput = await $('#user-store')
    await expect(userStoreInput).toBeDisplayed()

    // Default value should be set
    const value = await userStoreInput.getValue()
    expect(value).toContain('Emulation')
  })

  it('can change user store path', async () => {
    // Expand settings if not already
    const details = await $('details')
    const isOpen = await details.getAttribute('open')
    if (!isOpen) {
      const summary = await details.$('summary')
      await summary.click()
    }

    const userStoreInput = await $('#user-store')
    await userStoreInput.clearValue()
    await userStoreInput.setValue('~/TestEmulation')

    const value = await userStoreInput.getValue()
    expect(value).toBe('~/TestEmulation')
  })
})

describe('Kyaraben Apply (requires Nix)', () => {
  it('can apply configuration with TIC-80', async () => {
    // This test requires Nix to be available
    // It tests the full stack: UI -> Tauri -> Go daemon -> Nix

    // Select TIC-80 (no BIOS required)
    const tic80Checkbox = await $('input[value="tic80"]')
    if (!(await tic80Checkbox.isSelected())) {
      await tic80Checkbox.click()
    }

    // Deselect any systems that require BIOS
    const psxCheckbox = await $('input[value="psx"]')
    if (await psxCheckbox.isSelected()) {
      await psxCheckbox.click()
    }

    const snesCheckbox = await $('input[value="snes"]')
    if (await snesCheckbox.isSelected()) {
      await snesCheckbox.click()
    }

    // Click Apply
    const applyBtn = await $('#btn-apply')
    await applyBtn.click()

    // Wait for output section
    const outputSection = await $('#output-section')
    await expect(outputSection).toBeDisplayed()

    // Wait for completion (this can take a while with Nix)
    const log = await $('#log')
    await browser.waitUntil(
      async () => {
        const text = await log.getText()
        return text.includes('Done!') || text.includes('Error')
      },
      { timeout: 300000, timeoutMsg: 'Apply did not complete in 5 minutes' }
    )

    const logText = await log.getText()
    expect(logText).toContain('Done!')
  })
})
