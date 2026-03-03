import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { ToastProvider } from '@/lib/ToastContext'
import type { ApplyStatus, View } from '@/types/ui'
import { VIEW_CATALOG, VIEW_PREFERENCES } from '@/types/ui'
import { useApplyCompletionToast } from './useApplyCompletionToast'

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
  useApplyCompletionToast(status, currentView, onNavigateToCatalog)

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

describe('useApplyCompletionToast', () => {
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

  it('does not show toast for non-success status changes', async () => {
    const user = userEvent.setup()
    renderWithProviders('idle', VIEW_PREFERENCES)

    await user.click(screen.getByRole('button', { name: 'Set error' }))

    expect(screen.queryByText(/Installation complete/)).not.toBeInTheDocument()
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
})
