import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { System } from '@/types/daemon'
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
        selectedEmulator="retroarch:bsnes"
        pinnedVersion={null}
        installedVersion={null}
        provisions={[]}
        enabled={false}
        onToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    expect(screen.getByText('Super Nintendo')).toBeInTheDocument()
    expect(screen.getByText('RetroArch (bsnes)')).toBeInTheDocument()
  })

  it('shows checkbox as checked when enabled', () => {
    render(
      <SystemCard
        system={mockSystem}
        selectedEmulator="retroarch:bsnes"
        pinnedVersion={null}
        installedVersion={null}
        provisions={[]}
        enabled={true}
        onToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    const checkbox = screen.getByRole('checkbox')
    expect(checkbox).toBeChecked()
  })

  it('calls onToggle when checkbox is clicked', async () => {
    const user = userEvent.setup()
    const onToggle = vi.fn()

    render(
      <SystemCard
        system={mockSystem}
        selectedEmulator="retroarch:bsnes"
        pinnedVersion={null}
        installedVersion={null}
        provisions={[]}
        enabled={false}
        onToggle={onToggle}
        onVersionChange={vi.fn()}
      />,
    )

    await user.click(screen.getByRole('checkbox'))
    expect(onToggle).toHaveBeenCalledWith('snes', true)
  })

  it('renders provision badges when provisions exist', () => {
    render(
      <SystemCard
        system={mockSystem}
        selectedEmulator="retroarch:bsnes"
        pinnedVersion={null}
        installedVersion={null}
        provisions={[
          {
            filename: 'bios.bin',
            description: 'System BIOS',
            required: true,
            status: 'missing',
          },
        ]}
        enabled={false}
        onToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    expect(screen.getByText(/bios\.bin/)).toBeInTheDocument()
    expect(screen.getByText('Requires files')).toBeInTheDocument()
  })

  it('shows optional badge for non-required missing provisions', () => {
    render(
      <SystemCard
        system={mockSystem}
        selectedEmulator="retroarch:bsnes"
        pinnedVersion={null}
        installedVersion={null}
        provisions={[
          {
            filename: 'optional.bin',
            description: 'Optional file',
            required: false,
            status: 'missing',
          },
        ]}
        enabled={false}
        onToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    expect(screen.getByText(/optional\.bin \(optional\)/)).toBeInTheDocument()
  })
})
