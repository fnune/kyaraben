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
    shaders: null as boolean | null,
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
  })

  it('does not show shader controls when supportsShaders is false', () => {
    render(<EmulatorSettingsModal {...defaultProps} supportsShaders={false} />)
    expect(screen.queryByText('Shaders')).not.toBeInTheDocument()
  })

  it('shows Manual as selected when shaders is null', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders={null} />)
    const manualButton = screen.getByRole('button', { name: 'Manual' })
    expect(manualButton).toHaveClass('bg-accent')
  })

  it('shows On as selected when shaders is true', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders={true} />)
    const onButton = screen.getByRole('button', { name: 'On' })
    expect(onButton).toHaveClass('bg-accent')
  })

  it('shows Off as selected when shaders is false', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders={false} />)
    const offButton = screen.getByRole('button', { name: 'Off' })
    expect(offButton).toHaveClass('bg-accent')
  })

  it('calls onShaderChange with true when On is clicked', async () => {
    const onShaderChange = vi.fn()
    const user = userEvent.setup()

    render(
      <EmulatorSettingsModal {...defaultProps} shaders={null} onShaderChange={onShaderChange} />,
    )
    await user.click(screen.getByRole('button', { name: 'On' }))

    expect(onShaderChange).toHaveBeenCalledWith(true)
  })

  it('calls onShaderChange with false when Off is clicked', async () => {
    const onShaderChange = vi.fn()
    const user = userEvent.setup()

    render(
      <EmulatorSettingsModal {...defaultProps} shaders={null} onShaderChange={onShaderChange} />,
    )
    await user.click(screen.getByRole('button', { name: 'Off' }))

    expect(onShaderChange).toHaveBeenCalledWith(false)
  })

  it('calls onShaderChange with null when Manual is clicked', async () => {
    const onShaderChange = vi.fn()
    const user = userEvent.setup()

    render(
      <EmulatorSettingsModal {...defaultProps} shaders={true} onShaderChange={onShaderChange} />,
    )
    await user.click(screen.getByRole('button', { name: 'Manual' }))

    expect(onShaderChange).toHaveBeenCalledWith(null)
  })

  it('shows CRT shader info for CRT systems when On is selected', () => {
    render(<EmulatorSettingsModal {...defaultProps} systemId="snes" shaders={true} />)
    expect(screen.getByText(/CRT shader \(crt-mattias\)\./)).toBeInTheDocument()
  })

  it('shows LCD shader info for LCD systems when On is selected', () => {
    render(<EmulatorSettingsModal {...defaultProps} systemId="gba" shaders={true} />)
    expect(screen.getByText(/LCD shader \(lcd-grid-v2\)\./)).toBeInTheDocument()
  })

  it('shows Dolphin-specific shader info when Dolphin emulator', () => {
    render(<EmulatorSettingsModal {...defaultProps} emulatorId="dolphin" shaders={true} />)
    expect(screen.getByText(/CRT shader \(crt-lottes-fast\)\./)).toBeInTheDocument()
  })

  it('shows manual message when Manual is selected', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders={null} />)
    expect(screen.getByText('Kyaraben will not modify shader settings.')).toBeInTheDocument()
  })

  it('shows disable message when Off is selected', () => {
    render(<EmulatorSettingsModal {...defaultProps} shaders={false} />)
    expect(screen.getByText('Kyaraben will disable shaders.')).toBeInTheDocument()
  })
})
