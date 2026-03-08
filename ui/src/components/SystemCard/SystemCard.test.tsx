import type { DoctorResponse, EmulatorID, System } from '@shared/daemon'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ToastProvider } from '@/lib/ToastContext'
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
  emulators: [{ id: 'retroarch:bsnes', name: 'RetroArch (bsnes)', supportedSettings: ['preset'] }],
}

const defaultProps = {
  system: mockSystem,
  systemEnabledEmulators: new Set<EmulatorID>(),
  globalEnabledEmulators: new Set<EmulatorID>(),
  defaultEmulatorId: null,
  emulatorVersions: new Map(),
  emulatorPresets: new Map(),
  emulatorResume: new Map(),
  graphics: { preset: '' },
  savestate: { resume: '' },
  installedVersions: new Map(),
  installedExecLines: new Map(),
  managedConfigs: new Map(),
  installedPaths: new Map(),
  provisions: {},
  sharedPackages: new Set<string>(),
  onEmulatorToggle: vi.fn(),
  onSetDefaultEmulator: vi.fn(),
  onVersionChange: vi.fn(),
  onPresetChange: vi.fn(),
  onResumeChange: vi.fn(),
}

describe('SystemCard', () => {
  it('renders system name and emulator', () => {
    renderWithProviders(<SystemCard {...defaultProps} />)

    expect(screen.getByText('Super Nintendo')).toBeInTheDocument()
    expect(screen.getByText('RetroArch (bsnes)')).toBeInTheDocument()
  })

  it('shows toggle as enabled when emulator is enabled', () => {
    renderWithProviders(
      <SystemCard
        {...defaultProps}
        systemEnabledEmulators={new Set<EmulatorID>(['retroarch:bsnes'])}
        globalEnabledEmulators={new Set<EmulatorID>(['retroarch:bsnes'])}
        defaultEmulatorId={'retroarch:bsnes'}
      />,
    )

    const toggle = screen.getByRole('switch')
    expect(toggle).toHaveAttribute('aria-checked', 'true')
  })

  it('calls onEmulatorToggle when toggle is clicked', async () => {
    const user = userEvent.setup()
    const onEmulatorToggle = vi.fn()

    renderWithProviders(<SystemCard {...defaultProps} onEmulatorToggle={onEmulatorToggle} />)

    await user.click(screen.getByRole('switch'))
    expect(onEmulatorToggle).toHaveBeenCalledWith('snes', 'retroarch:bsnes', true)
  })

  it('renders provision status when provisions exist', () => {
    const provisions: DoctorResponse = {
      'snes:retroarch:bsnes': [
        {
          filename: 'bios.bin',
          kind: 'bios',
          description: 'USA',
          status: 'missing',
          groupRequired: true,
          groupSatisfied: false,
          groupSize: 1,
          displayName: 'bios.bin',
          instructions: 'Place bios.bin in this directory',
        },
      ],
    }

    renderWithProviders(<SystemCard {...defaultProps} provisions={provisions} />)

    expect(screen.getByText(/BIOS \(USA\)/)).toBeInTheDocument()
  })

  it('shows manufacturer and year in header', () => {
    renderWithProviders(<SystemCard {...defaultProps} />)

    expect(screen.getByText(/Nintendo · 1990/)).toBeInTheDocument()
  })
})
