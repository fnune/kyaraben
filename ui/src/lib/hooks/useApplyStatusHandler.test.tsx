import type { ApplyStatus, View } from '@shared/ui'
import { VIEW_CATALOG, VIEW_PREFERENCES } from '@shared/ui'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { ToastProvider } from '@/lib/ToastContext'
import { useApplyStatusHandler } from './useApplyStatusHandler'

function TestComponent({
  initialStatus,
  currentView,
  onNavigateToCatalog,
}: {
  initialStatus: ApplyStatus
  currentView: View
  onNavigateToCatalog: () => void
}) {
  const [status, setStatus] = useState<ApplyStatus>(initialStatus)
  useApplyStatusHandler(status, currentView, onNavigateToCatalog)

  return (
    <div>
      <button type="button" onClick={() => setStatus('success')}>
        Set success
      </button>
      <button type="button" onClick={() => setStatus('error')}>
        Set error
      </button>
      <button type="button" onClick={() => setStatus('applying')}>
        Set applying
      </button>
      <button type="button" onClick={() => setStatus('reviewing')}>
        Set reviewing
      </button>
      <button type="button" onClick={() => setStatus('confirming_sync')}>
        Set confirming_sync
      </button>
    </div>
  )
}

function renderWithProviders(
  initialStatus: ApplyStatus,
  currentView: View,
  onNavigateToCatalog = vi.fn(),
) {
  return {
    onNavigateToCatalog,
    ...render(
      <ToastProvider>
        <TestComponent
          initialStatus={initialStatus}
          currentView={currentView}
          onNavigateToCatalog={onNavigateToCatalog}
        />
      </ToastProvider>,
    ),
  }
}

describe('useApplyStatusHandler', () => {
  it('shows persistent toast with navigation link when not in catalog view', async () => {
    const user = userEvent.setup()
    renderWithProviders('applying', VIEW_PREFERENCES)

    await user.click(screen.getByRole('button', { name: 'Set success' }))

    expect(screen.getByText(/Installation complete/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Go to catalog/ })).toBeInTheDocument()
  })

  it('calls onNavigateToCatalog when navigation link is clicked', async () => {
    const user = userEvent.setup()
    const onNavigateToCatalog = vi.fn()
    renderWithProviders('applying', VIEW_PREFERENCES, onNavigateToCatalog)

    await user.click(screen.getByRole('button', { name: 'Set success' }))
    await user.click(screen.getByRole('button', { name: /Go to catalog/ }))

    expect(onNavigateToCatalog).toHaveBeenCalled()
  })

  it('shows simple toast when in catalog view', async () => {
    const user = userEvent.setup()
    renderWithProviders('applying', VIEW_CATALOG)

    await user.click(screen.getByRole('button', { name: 'Set success' }))

    expect(screen.getByText('Installation complete.')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Go to catalog/ })).not.toBeInTheDocument()
  })

  it('shows error toast with navigation link when error occurs outside catalog view', async () => {
    const user = userEvent.setup()
    renderWithProviders('idle', VIEW_PREFERENCES)

    await user.click(screen.getByRole('button', { name: 'Set error' }))

    expect(screen.getByText(/Installation failed/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Go to catalog/ })).toBeInTheDocument()
  })

  it('does not show error toast when error occurs in catalog view', async () => {
    const user = userEvent.setup()
    renderWithProviders('idle', VIEW_CATALOG)

    await user.click(screen.getByRole('button', { name: 'Set error' }))

    expect(screen.queryByText(/Installation failed/)).not.toBeInTheDocument()
  })

  it('does not show duplicate toast when status unchanged', async () => {
    const user = userEvent.setup()
    renderWithProviders('applying', VIEW_PREFERENCES)

    await user.click(screen.getByRole('button', { name: 'Set success' }))
    const toasts = screen.getAllByText(/Installation complete/)
    expect(toasts).toHaveLength(1)

    await user.click(screen.getByRole('button', { name: 'Set applying' }))
    await user.click(screen.getByRole('button', { name: 'Set success' }))

    expect(screen.getAllByText(/Installation complete/)).toHaveLength(2)
  })

  it('navigates to catalog when status changes to reviewing and not in catalog', async () => {
    const user = userEvent.setup()
    const onNavigateToCatalog = vi.fn()
    renderWithProviders('applying', VIEW_PREFERENCES, onNavigateToCatalog)

    await user.click(screen.getByRole('button', { name: 'Set reviewing' }))

    expect(onNavigateToCatalog).toHaveBeenCalled()
  })

  it('does not navigate when status changes to reviewing and already in catalog', async () => {
    const user = userEvent.setup()
    const onNavigateToCatalog = vi.fn()
    renderWithProviders('applying', VIEW_CATALOG, onNavigateToCatalog)

    await user.click(screen.getByRole('button', { name: 'Set reviewing' }))

    expect(onNavigateToCatalog).not.toHaveBeenCalled()
  })

  it('navigates to catalog when status changes to confirming_sync and not in catalog', async () => {
    const user = userEvent.setup()
    const onNavigateToCatalog = vi.fn()
    renderWithProviders('applying', VIEW_PREFERENCES, onNavigateToCatalog)

    await user.click(screen.getByRole('button', { name: 'Set confirming_sync' }))

    expect(onNavigateToCatalog).toHaveBeenCalled()
  })

  it('does not navigate when status changes to confirming_sync and already in catalog', async () => {
    const user = userEvent.setup()
    const onNavigateToCatalog = vi.fn()
    renderWithProviders('applying', VIEW_CATALOG, onNavigateToCatalog)

    await user.click(screen.getByRole('button', { name: 'Set confirming_sync' }))

    expect(onNavigateToCatalog).not.toHaveBeenCalled()
  })
})
