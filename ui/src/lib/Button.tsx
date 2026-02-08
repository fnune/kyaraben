import type { ReactNode } from 'react'

export interface ButtonProps {
  children: ReactNode
  variant?: 'primary' | 'secondary' | 'danger'
  disabled?: boolean
  onClick?: () => void
}

const VARIANT_CLASSES: Record<NonNullable<ButtonProps['variant']>, string> = {
  primary: 'bg-blue-600 text-white hover:bg-blue-700',
  secondary: 'bg-gray-700 text-gray-300 hover:bg-gray-600',
  danger: 'bg-transparent text-red-400 hover:text-red-300 hover:bg-red-500/10',
}

export function Button({ children, variant = 'primary', disabled, onClick }: ButtonProps) {
  const baseClasses = 'px-4 py-2 rounded-md disabled:opacity-50 disabled:cursor-not-allowed'
  const variantClasses = VARIANT_CLASSES[variant]

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
