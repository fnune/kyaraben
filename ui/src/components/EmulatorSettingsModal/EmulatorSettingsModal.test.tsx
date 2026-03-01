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
    supportsShaders: true,
    shaders: null as string | null,
    graphics: { shaders: '' },
    onShaderChange: vi.fn(),
  }

  it('renders modal with emulator name in title', () => {
    render(<EmulatorSettingsModal {...defaultProps} />)
    expect(screen.getByText('RetroArch (bsnes) Settings')).toBeInTheDocument()
  })

  it('shows shader controls when supportsShaders is true', () => {
    render(<EmulatorSettingsModal {...defaultProps} />)
    expect(screen.getByText('Shaders')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'On' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Off' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Manual' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Default' })).toBeInTheDocument()
  })

  it('does not show shader controls when supportsShaders is false', () => {
    render(<EmulatorSettingsModal {...defaultProps} supportsShaders={false} />)
    expect(screen.queryByText('Shaders')).not.toBeInTheDocument()
  })

  it('shows Default as selected when shaders is null', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders={null} />)
    const defaultButton = screen.getByRole('button', { name: 'Default' })
    expect(defaultButton).toHaveClass('bg-accent')
  })

  it('shows On as selected when shaders is "on"', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders="on" />)
    const onButton = screen.getByRole('button', { name: 'On' })
    expect(onButton).toHaveClass('bg-accent')
  })

  it('shows Off as selected when shaders is "off"', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders="off" />)
    const offButton = screen.getByRole('button', { name: 'Off' })
    expect(offButton).toHaveClass('bg-accent')
  })

  it('shows Manual as selected when shaders is "manual"', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders="manual" />)
    const manualButton = screen.getByRole('button', { name: 'Manual' })
    expect(manualButton).toHaveClass('bg-accent')
  })

  it('calls onShaderChange with "on" when On is clicked', async () => {
    const onShaderChange = vi.fn()
    const user = userEvent.setup()

    render(
      <EmulatorSettingsModal {...defaultProps} shaders={null} onShaderChange={onShaderChange} />,
    )
    await user.click(screen.getByRole('button', { name: 'On' }))

    expect(onShaderChange).toHaveBeenCalledWith('on')
  })

  it('calls onShaderChange with "off" when Off is clicked', async () => {
    const onShaderChange = vi.fn()
    const user = userEvent.setup()

    render(
      <EmulatorSettingsModal {...defaultProps} shaders={null} onShaderChange={onShaderChange} />,
    )
    await user.click(screen.getByRole('button', { name: 'Off' }))

    expect(onShaderChange).toHaveBeenCalledWith('off')
  })

  it('calls onShaderChange with null when Default is clicked', async () => {
    const onShaderChange = vi.fn()
    const user = userEvent.setup()

    render(<EmulatorSettingsModal {...defaultProps} shaders="on" onShaderChange={onShaderChange} />)
    await user.click(screen.getByRole('button', { name: 'Default' }))

    expect(onShaderChange).toHaveBeenCalledWith(null)
  })

  it('calls onShaderChange with "manual" when Manual is clicked', async () => {
    const onShaderChange = vi.fn()
    const user = userEvent.setup()

    render(
      <EmulatorSettingsModal {...defaultProps} shaders={null} onShaderChange={onShaderChange} />,
    )
    await user.click(screen.getByRole('button', { name: 'Manual' }))

    expect(onShaderChange).toHaveBeenCalledWith('manual')
  })

  it('shows CRT shader info for CRT systems when On is selected', () => {
    render(<EmulatorSettingsModal {...defaultProps} systemId="snes" shaders="on" />)
    expect(screen.getByText(/CRT shader \(crt-mattias\)\./)).toBeInTheDocument()
  })

  it('shows LCD shader info for LCD systems when On is selected', () => {
    render(<EmulatorSettingsModal {...defaultProps} systemId="gba" shaders="on" />)
    expect(screen.getByText(/LCD shader \(lcd-grid-v2\)\./)).toBeInTheDocument()
  })

  it('shows Dolphin-specific shader info when Dolphin emulator', () => {
    render(<EmulatorSettingsModal {...defaultProps} emulatorId="dolphin" shaders="on" />)
    expect(screen.getByText(/CRT shader \(crt-lottes-fast\)\./)).toBeInTheDocument()
  })

  it('shows manual message when Manual is explicitly selected', () => {
    render(
      <EmulatorSettingsModal
        {...defaultProps}
        shaders="manual"
        graphics={{ shaders: 'recommended' }}
      />,
    )
    expect(screen.getByText('Kyaraben will not modify shader settings.')).toBeInTheDocument()
  })

  it('shows manual message when Default is selected with no global default', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders={null} graphics={{ shaders: '' }} />)
    expect(screen.getByText('Kyaraben will not modify shader settings.')).toBeInTheDocument()
  })

  it('shows resolved shader info when Default is selected with global recommended', () => {
    render(
      <EmulatorSettingsModal
        {...defaultProps}
        shaders={null}
        graphics={{ shaders: 'recommended' }}
      />,
    )
    expect(screen.getByText(/CRT shader \(crt-mattias\)\./)).toBeInTheDocument()
  })

  it('shows Default (recommended) label when global default is recommended', () => {
    render(
      <EmulatorSettingsModal
        {...defaultProps}
        shaders={null}
        graphics={{ shaders: 'recommended' }}
      />,
    )
    expect(screen.getByRole('button', { name: 'Default (recommended)' })).toBeInTheDocument()
  })

  it('shows disable message when Off is selected', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders="off" />)
    expect(screen.getByText('Kyaraben will disable shaders.')).toBeInTheDocument()
  })
})
