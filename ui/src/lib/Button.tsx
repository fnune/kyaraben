import type { ReactNode } from 'react'

export interface ButtonProps {
  children: ReactNode
  variant?: 'primary' | 'secondary' | 'danger'
  size?: 'sm' | 'md'
  disabled?: boolean
  onClick?: () => void
}

const VARIANT_CLASSES: Record<NonNullable<ButtonProps['variant']>, string> = {
  primary: 'bg-accent text-white hover:bg-accent-hover',
  secondary: 'bg-surface-raised text-on-surface-secondary hover:bg-outline-strong',
  danger: 'bg-transparent text-status-error hover:text-status-error hover:bg-status-error/10',
}

const SIZE_CLASSES: Record<NonNullable<ButtonProps['size']>, string> = {
  sm: 'px-2.5 py-1 text-xs',
  md: 'px-4 py-2 text-sm',
}

export function Button({
  children,
  variant = 'primary',
  size = 'md',
  disabled,
  onClick,
}: ButtonProps) {
  const baseClasses =
    'rounded-control disabled:opacity-50 disabled:cursor-not-allowed tracking-wide'
  const variantClasses = VARIANT_CLASSES[variant]
  const sizeClasses = SIZE_CLASSES[size]

  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className={`${baseClasses} ${variantClasses} ${sizeClasses}`}
    >
      {children}
    </button>
  )
}
