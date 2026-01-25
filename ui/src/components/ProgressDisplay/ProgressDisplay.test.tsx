import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { ProgressStep } from '@/types/ui'
import { ProgressDisplay } from './ProgressDisplay'

describe('ProgressDisplay', () => {
  it('renders nothing when no steps and no error', () => {
    const { container } = render(<ProgressDisplay steps={[]} />)
    expect(container.firstChild).toBeNull()
  })

  it('renders steps with labels', () => {
    const steps: ProgressStep[] = [
      { id: 'step1', label: 'Building Nix flake', status: 'completed' },
      { id: 'step2', label: 'Writing configs', status: 'in_progress' },
    ]

    render(<ProgressDisplay steps={steps} />)

    expect(screen.getByText('Building Nix flake')).toBeInTheDocument()
    expect(screen.getByText('Writing configs')).toBeInTheDocument()
  })

  it('renders step messages', () => {
    const steps: ProgressStep[] = [
      {
        id: 'step1',
        label: 'Building',
        status: 'completed',
        message: 'Done in 5s',
      },
    ]

    render(<ProgressDisplay steps={steps} />)

    expect(screen.getByText('Done in 5s')).toBeInTheDocument()
  })

  it('renders error message', () => {
    render(<ProgressDisplay steps={[]} error="Build failed" />)

    expect(screen.getByText(/Build failed/)).toBeInTheDocument()
  })

  it('renders both steps and error', () => {
    const steps: ProgressStep[] = [{ id: 'step1', label: 'Building', status: 'error' }]

    render(<ProgressDisplay steps={steps} error="Nix build failed" />)

    expect(screen.getByText('Building')).toBeInTheDocument()
    expect(screen.getByText(/Nix build failed/)).toBeInTheDocument()
  })
})
