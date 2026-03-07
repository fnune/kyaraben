import type { ReactNode } from 'react'

export interface ButtonProps {
  children: ReactNode
  variant?: 'primary' | 'secondary' | 'danger'
  size?: 'sm' | 'md'
  type?: 'button' | 'submit'
  disabled?: boolean
  onClick?: () => void
  className?: string
}

const VARIANT_CLASSES: Record<NonNullable<ButtonProps['variant']>, string> = {
  primary:
    'bg-accent text-white hover:bg-accent-hover disabled:bg-outline disabled:text-on-surface-dim',
  secondary:
    'bg-surface-raised text-on-surface-secondary hover:bg-outline-strong disabled:bg-surface disabled:text-on-surface-dim',
  danger:
    'bg-transparent text-status-error hover:text-status-error hover:bg-status-error/10 disabled:text-on-surface-dim disabled:bg-transparent',
}

const SIZE_CLASSES: Record<NonNullable<ButtonProps['size']>, string> = {
  sm: 'px-2.5 py-1 text-xs',
  md: 'px-4 py-2 text-sm',
}

export function Button({
  children,
  variant = 'primary',
  size = 'md',
  type = 'button',
  disabled,
  onClick,
  className,
}: ButtonProps) {
  const baseClasses = 'rounded-control disabled:cursor-not-allowed tracking-wide'
  const variantClasses = VARIANT_CLASSES[variant]
  const sizeClasses = SIZE_CLASSES[size]
  const allClasses = [baseClasses, variantClasses, sizeClasses, className].filter(Boolean).join(' ')

  return (
    <button type={type} disabled={disabled} onClick={onClick} className={allClasses}>
      {children}
    </button>
  )
}
