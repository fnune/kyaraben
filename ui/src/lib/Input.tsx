export interface InputProps {
  readonly value: string
  readonly onChange: (value: string) => void
  readonly placeholder?: string
  readonly disabled?: boolean
  readonly inputMode?: 'none' | 'text' | 'decimal' | 'numeric' | 'tel' | 'search' | 'email' | 'url'
  readonly enterKeyHint?: 'enter' | 'done' | 'go' | 'next' | 'previous' | 'search' | 'send'
}

export function Input({
  value,
  onChange,
  placeholder,
  disabled,
  inputMode = 'text',
  enterKeyHint,
}: InputProps) {
  return (
    <input
      type="text"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      disabled={disabled}
      inputMode={inputMode}
      enterKeyHint={enterKeyHint}
      className="block w-full rounded-control border-outline-strong bg-surface-raised text-on-surface placeholder-on-surface-dim shadow-xs focus:border-accent focus:ring-accent px-3 py-2 border font-mono tabular-nums"
    />
  )
}
