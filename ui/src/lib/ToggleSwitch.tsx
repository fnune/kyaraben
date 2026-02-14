export interface ToggleSwitchProps {
  readonly enabled: boolean
  readonly onChange: (enabled: boolean) => void
  readonly disabled?: boolean
}

export function ToggleSwitch({ enabled, onChange, disabled = false }: ToggleSwitchProps) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={enabled}
      disabled={disabled}
      onClick={() => onChange(!enabled)}
      className={`
        relative w-11 h-6 shrink-0 transition-colors
        ${enabled ? 'bg-accent' : 'bg-surface-raised'}
        ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
      `}
      style={{ borderRadius: '2px' }}
    >
      <div
        className={`
          absolute top-0.5 w-5 h-5 bg-white shadow transition-all
          ${enabled ? 'left-5' : 'left-0.5'}
        `}
        style={{ borderRadius: '2px' }}
      />
    </button>
  )
}
