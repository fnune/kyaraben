import type { ReactNode } from 'react'
import { Spinner } from './Spinner'

export interface IconButtonProps {
  icon: ReactNode
  label: string
  loading?: boolean
  disabled?: boolean
  onClick?: () => void
}

export function IconButton({ icon, label, loading, disabled, onClick }: IconButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled || loading}
      className="px-3 py-2 border border-outline-strong rounded-control hover:bg-surface-raised disabled:opacity-50"
      title={label}
      aria-label={label}
    >
      {loading ? <Spinner size="md" /> : icon}
    </button>
  )
}
