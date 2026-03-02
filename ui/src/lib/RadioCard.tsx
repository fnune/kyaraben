export interface RadioCardProps {
  readonly title: string
  readonly description: string
  readonly selected: boolean
  readonly onSelect: () => void
  readonly className?: string
  readonly wrap?: boolean
}

export function RadioCard({
  title,
  description,
  selected,
  onSelect,
  className = '',
  wrap = false,
}: RadioCardProps) {
  const textClass = wrap ? '' : 'truncate'
  return (
    <button
      type="button"
      onClick={onSelect}
      className={`text-left rounded-card border-2 transition-colors ${
        selected
          ? 'border-accent bg-accent/5'
          : 'border-outline bg-surface hover:border-outline-hover'
      } ${className}`}
    >
      <div className="flex items-start gap-3">
        <div
          className={`w-4 h-4 shrink-0 rounded-full border-2 mt-0.5 ${
            selected ? 'border-accent bg-accent' : 'border-outline'
          }`}
          style={selected ? { boxShadow: 'inset 0 0 0 3px var(--color-surface)' } : undefined}
        />
        <div className="min-w-0 flex-1">
          <span
            className={`block font-medium ${textClass} ${selected ? 'text-accent' : 'text-on-surface'}`}
          >
            {title}
          </span>
          <span className={`block text-sm text-on-surface-muted ${textClass}`}>{description}</span>
        </div>
      </div>
    </button>
  )
}
