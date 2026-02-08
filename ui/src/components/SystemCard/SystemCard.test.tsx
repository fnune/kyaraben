import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ToastProvider } from '@/lib/ToastContext'
import type { DoctorResponse, EmulatorID, System } from '@/types/daemon'
import { SystemCard } from './SystemCard'

function renderWithProviders(ui: React.ReactElement) {
  return render(<ToastProvider>{ui}</ToastProvider>)
}

const mockSystem: System = {
  id: 'snes',
  name: 'Super Nintendo',
  description: '16-bit home console by Nintendo (1990)',
  manufacturer: 'Nintendo',
  label: 'SNES',
  defaultEmulatorId: 'retroarch:bsnes',
  emulators: [{ id: 'retroarch:bsnes', name: 'RetroArch (bsnes)' }],
}

describe('SystemCard', () => {
  it('renders system name and emulator', () => {
    renderWithProviders(
      <SystemCard
        system={mockSystem}
        enabledEmulators={new Set<EmulatorID>()}
        emulatorVersions={new Map()}
        installedVersions={new Map()}
        installedExecLines={new Map()}
        managedConfigs={new Map()}
        installedPaths={new Map()}
        provisions={{}}
        sharedPackages={new Set()}
        onEmulatorToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    expect(screen.getByText('Super Nintendo')).toBeInTheDocument()
    expect(screen.getByText('RetroArch (bsnes)')).toBeInTheDocument()
  })

  it('shows toggle as enabled when emulator is enabled', () => {
    renderWithProviders(
      <SystemCard
        system={mockSystem}
        enabledEmulators={new Set<EmulatorID>(['retroarch:bsnes'])}
        emulatorVersions={new Map()}
        installedVersions={new Map()}
        installedExecLines={new Map()}
        managedConfigs={new Map()}
        installedPaths={new Map()}
        provisions={{}}
        sharedPackages={new Set()}
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

    renderWithProviders(
      <SystemCard
        system={mockSystem}
        enabledEmulators={new Set<EmulatorID>()}
        emulatorVersions={new Map()}
        installedVersions={new Map()}
        installedExecLines={new Map()}
        managedConfigs={new Map()}
        installedPaths={new Map()}
        provisions={{}}
        sharedPackages={new Set()}
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
          kind: 'bios',
          description: 'USA',
          status: 'missing',
          groupRequired: true,
          groupSatisfied: false,
          groupSize: 1,
        },
      ],
    }

    renderWithProviders(
      <SystemCard
        system={mockSystem}
        enabledEmulators={new Set<EmulatorID>()}
        emulatorVersions={new Map()}
        installedVersions={new Map()}
        installedExecLines={new Map()}
        managedConfigs={new Map()}
        installedPaths={new Map()}
        provisions={provisions}
        sharedPackages={new Set()}
        onEmulatorToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    expect(screen.getByText(/BIOS \(USA\)/)).toBeInTheDocument()
  })

  it('shows manufacturer and year in header', () => {
    renderWithProviders(
      <SystemCard
        system={mockSystem}
        enabledEmulators={new Set<EmulatorID>()}
        emulatorVersions={new Map()}
        installedVersions={new Map()}
        installedExecLines={new Map()}
        managedConfigs={new Map()}
        installedPaths={new Map()}
        provisions={{}}
        sharedPackages={new Set()}
        onEmulatorToggle={vi.fn()}
        onVersionChange={vi.fn()}
      />,
    )

    expect(screen.getByText(/Nintendo · 1990/)).toBeInTheDocument()
  })
})
