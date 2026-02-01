import type { ReactNode } from 'react'

export interface ButtonProps {
  children: ReactNode
  variant?: 'primary' | 'secondary'
  disabled?: boolean
  onClick?: () => void
}

export function Button({ children, variant = 'primary', disabled, onClick }: ButtonProps) {
  const baseClasses = 'px-4 py-2 rounded-md disabled:opacity-50 disabled:cursor-not-allowed'
  const variantClasses =
    variant === 'primary'
      ? 'bg-blue-600 text-white hover:bg-blue-700'
      : 'bg-gray-700 text-gray-300 hover:bg-gray-600'

  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className={`${baseClasses} ${variantClasses}`}
    >
      {children}
    </button>
  )
}
