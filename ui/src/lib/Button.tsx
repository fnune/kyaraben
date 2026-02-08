import type { ReactNode } from 'react'

export interface ButtonProps {
  children: ReactNode
  variant?: 'primary' | 'secondary' | 'danger'
  disabled?: boolean
  onClick?: () => void
}

const VARIANT_CLASSES: Record<NonNullable<ButtonProps['variant']>, string> = {
  primary: 'bg-accent text-white hover:bg-accent-hover',
  secondary: 'bg-surface-raised text-on-surface-secondary hover:bg-outline-strong',
  danger: 'bg-transparent text-red-400 hover:text-red-300 hover:bg-red-500/10',
}

export function Button({ children, variant = 'primary', disabled, onClick }: ButtonProps) {
  const baseClasses =
    'px-4 py-2 rounded-control disabled:opacity-50 disabled:cursor-not-allowed text-sm tracking-wide'
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
