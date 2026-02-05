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
      className="block w-full rounded-md border-gray-600 bg-gray-700 text-gray-100 placeholder-gray-500 shadow-xs focus:border-blue-500 focus:ring-blue-500 px-3 py-2 border"
    />
  )
}
