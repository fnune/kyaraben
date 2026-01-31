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
      className="px-3 py-2 border border-gray-300 rounded-md hover:bg-gray-100 disabled:opacity-50"
      title={label}
      aria-label={label}
    >
      {loading ? <Spinner size="md" /> : icon}
    </button>
  )
}
