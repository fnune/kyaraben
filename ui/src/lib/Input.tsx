export interface InputProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  disabled?: boolean
}

export function Input({ value, onChange, placeholder, disabled }: InputProps) {
  return (
    <input
      type="text"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      disabled={disabled}
      className="block w-full rounded-control border-outline-strong bg-surface-raised text-on-surface placeholder-on-surface-dim shadow-xs focus:border-accent focus:ring-accent px-3 py-2 border font-mono tabular-nums"
    />
  )
}
