import { INPUT_BASE_CLASSES } from './inputStyles'

export interface InputProps {
  readonly value: string
  readonly onChange: (value: string) => void
  readonly placeholder?: string
  readonly disabled?: boolean
  readonly className?: string
  readonly inputMode?: 'none' | 'text' | 'decimal' | 'numeric' | 'tel' | 'search' | 'email' | 'url'
  readonly enterKeyHint?: 'enter' | 'done' | 'go' | 'next' | 'previous' | 'search' | 'send'
}

export function Input({
  value,
  onChange,
  placeholder,
  disabled,
  className,
  inputMode = 'text',
  enterKeyHint,
}: InputProps) {
  const baseClasses = `block w-full px-3 py-2 font-mono tabular-nums ${INPUT_BASE_CLASSES}`
  const classes = className ? `${baseClasses} ${className}` : baseClasses

  return (
    <input
      type="text"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      disabled={disabled}
      inputMode={inputMode}
      enterKeyHint={enterKeyHint}
      className={classes}
    />
  )
}
