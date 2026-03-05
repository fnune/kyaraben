import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { EmulatorSettingsModal } from './EmulatorSettingsModal'

describe('EmulatorSettingsModal', () => {
  const defaultProps = {
    open: true,
    onClose: vi.fn(),
    emulatorId: 'retroarch:bsnes' as const,
    emulatorName: 'RetroArch (bsnes)',
    systemId: 'snes' as const,
    supportsPreset: true,
    preset: null as string | null,
    graphics: { preset: '' },
    onPresetChange: vi.fn(),
    supportsResume: false,
    resume: null as string | null,
    savestate: { resume: '' },
    onResumeChange: vi.fn(),
  }

  it('renders modal with emulator name in title', () => {
    render(<EmulatorSettingsModal {...defaultProps} />)
    expect(screen.getByText('RetroArch (bsnes) settings')).toBeInTheDocument()
  })

  it('shows preset controls when supportsPreset is true', () => {
    render(<EmulatorSettingsModal {...defaultProps} />)
    expect(screen.getByText('Display preset')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Modern pixels' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Upscaled' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Pseudo-authentic' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Manual' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Default' })).toBeInTheDocument()
  })

  it('does not show preset controls when supportsPreset is false', () => {
    render(<EmulatorSettingsModal {...defaultProps} supportsPreset={false} />)
    expect(screen.queryByText('Display preset')).not.toBeInTheDocument()
  })

  it('shows Default as selected when preset is null', () => {
    render(<EmulatorSettingsModal {...defaultProps} preset={null} />)
    const defaultButton = screen.getByRole('button', { name: 'Default' })
    expect(defaultButton).toHaveClass('bg-accent')
  })

  it('shows Modern pixels as selected when preset is "modern-pixels"', () => {
    render(<EmulatorSettingsModal {...defaultProps} preset="modern-pixels" />)
    const button = screen.getByRole('button', { name: 'Modern pixels' })
    expect(button).toHaveClass('bg-accent')
  })

  it('shows Upscaled as selected when preset is "upscaled"', () => {
    render(<EmulatorSettingsModal {...defaultProps} preset="upscaled" />)
    const button = screen.getByRole('button', { name: 'Upscaled' })
    expect(button).toHaveClass('bg-accent')
  })

  it('shows Pseudo-authentic as selected when preset is "pseudo-authentic"', () => {
    render(<EmulatorSettingsModal {...defaultProps} preset="pseudo-authentic" />)
    const button = screen.getByRole('button', { name: 'Pseudo-authentic' })
    expect(button).toHaveClass('bg-accent')
  })

  it('shows Manual as selected when preset is "manual"', () => {
    render(<EmulatorSettingsModal {...defaultProps} preset="manual" />)
    const manualButton = screen.getByRole('button', { name: 'Manual' })
    expect(manualButton).toHaveClass('bg-accent')
  })

  it('calls onPresetChange with "modern-pixels" when Modern pixels is clicked', async () => {
    const onPresetChange = vi.fn()
    const user = userEvent.setup()

    render(
      <EmulatorSettingsModal {...defaultProps} preset={null} onPresetChange={onPresetChange} />,
    )
    await user.click(screen.getByRole('button', { name: 'Modern pixels' }))

    expect(onPresetChange).toHaveBeenCalledWith('modern-pixels')
  })

  it('calls onPresetChange with "upscaled" when Upscaled is clicked', async () => {
    const onPresetChange = vi.fn()
    const user = userEvent.setup()

    render(
      <EmulatorSettingsModal {...defaultProps} preset={null} onPresetChange={onPresetChange} />,
    )
    await user.click(screen.getByRole('button', { name: 'Upscaled' }))

    expect(onPresetChange).toHaveBeenCalledWith('upscaled')
  })

  it('calls onPresetChange with null when Default is clicked', async () => {
    const onPresetChange = vi.fn()
    const user = userEvent.setup()

    render(
      <EmulatorSettingsModal
        {...defaultProps}
        preset="modern-pixels"
        onPresetChange={onPresetChange}
      />,
    )
    await user.click(screen.getByRole('button', { name: 'Default' }))

    expect(onPresetChange).toHaveBeenCalledWith(null)
  })

  it('calls onPresetChange with "manual" when Manual is clicked', async () => {
    const onPresetChange = vi.fn()
    const user = userEvent.setup()

    render(
      <EmulatorSettingsModal {...defaultProps} preset={null} onPresetChange={onPresetChange} />,
    )
    await user.click(screen.getByRole('button', { name: 'Manual' }))

    expect(onPresetChange).toHaveBeenCalledWith('manual')
  })

  it('shows manual message when Manual is explicitly selected', () => {
    render(
      <EmulatorSettingsModal
        {...defaultProps}
        preset="manual"
        graphics={{ preset: 'modern-pixels' }}
      />,
    )
    expect(screen.getByText('Kyaraben will not modify display settings.')).toBeInTheDocument()
  })

  it('shows manual message when Default is selected with no global default', () => {
    render(<EmulatorSettingsModal {...defaultProps} preset={null} graphics={{ preset: '' }} />)
    expect(screen.getByText('Kyaraben will not modify display settings.')).toBeInTheDocument()
  })

  it('shows preset info when Default is selected with global preset', () => {
    render(
      <EmulatorSettingsModal
        {...defaultProps}
        preset={null}
        graphics={{ preset: 'modern-pixels' }}
      />,
    )
    expect(screen.getByText('Kyaraben will apply the modern pixels preset.')).toBeInTheDocument()
  })

  it('shows Default (modern pixels) label when global default is modern-pixels', () => {
    render(
      <EmulatorSettingsModal
        {...defaultProps}
        preset={null}
        graphics={{ preset: 'modern-pixels' }}
      />,
    )
    expect(screen.getByRole('button', { name: 'Default (modern pixels)' })).toBeInTheDocument()
  })
})
