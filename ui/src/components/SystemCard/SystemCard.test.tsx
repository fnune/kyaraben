import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { DoctorResponse, EmulatorID, System } from '@/types/daemon'
import { SystemCard } from './SystemCard'

const mockSystem: System = {
  id: 'snes',
  name: 'Super Nintendo',
  description: '16-bit home console by Nintendo (1990)',
  manufacturer: 'Nintendo',
  label: 'SNES',
  emulators: [{ id: 'retroarch:bsnes', name: 'RetroArch (bsnes)' }],
}

describe('SystemCard', () => {
  it('renders system name and emulator', () => {
    render(
      <SystemCard
        system={mockSystem}
        enabledEmulators={new Set<EmulatorID>()}
        emulatorVersions={new Map()}
        installedVersions={new Map()}
        provisions={{}}
        userStore="~/Emulation"
        onEmulatorToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    expect(screen.getByText('Super Nintendo')).toBeInTheDocument()
    expect(screen.getByText('RetroArch (bsnes)')).toBeInTheDocument()
  })

  it('shows toggle as enabled when emulator is enabled', () => {
    render(
      <SystemCard
        system={mockSystem}
        enabledEmulators={new Set<EmulatorID>(['retroarch:bsnes'])}
        emulatorVersions={new Map()}
        installedVersions={new Map()}
        provisions={{}}
        userStore="~/Emulation"
        onEmulatorToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    const toggle = screen.getByRole('switch')
    expect(toggle).toHaveAttribute('aria-checked', 'true')
  })

  it('calls onEmulatorToggle when toggle is clicked', async () => {
    const user = userEvent.setup()
    const onEmulatorToggle = vi.fn()

    render(
      <SystemCard
        system={mockSystem}
        enabledEmulators={new Set<EmulatorID>()}
        emulatorVersions={new Map()}
        installedVersions={new Map()}
        provisions={{}}
        userStore="~/Emulation"
        onEmulatorToggle={onEmulatorToggle}
        onVersionChange={vi.fn()}
      />,
    )

    await user.click(screen.getByRole('switch'))
    expect(onEmulatorToggle).toHaveBeenCalledWith('retroarch:bsnes', true)
  })

  it('renders provision status when provisions exist', () => {
    const provisions: DoctorResponse = {
      'retroarch:bsnes': [
        {
          filename: 'bios.bin',
          description: 'System BIOS',
          required: true,
          status: 'missing',
        },
      ],
    }

    render(
      <SystemCard
        system={mockSystem}
        enabledEmulators={new Set<EmulatorID>()}
        emulatorVersions={new Map()}
        installedVersions={new Map()}
        provisions={provisions}
        userStore="~/Emulation"
        onEmulatorToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    expect(screen.getByText('System BIOS')).toBeInTheDocument()
  })

  it('shows manufacturer and year in header', () => {
    render(
      <SystemCard
        system={mockSystem}
        enabledEmulators={new Set<EmulatorID>()}
        emulatorVersions={new Map()}
        installedVersions={new Map()}
        provisions={{}}
        userStore="~/Emulation"
        onEmulatorToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    expect(screen.getByText(/Nintendo · 1990/)).toBeInTheDocument()
  })
})
