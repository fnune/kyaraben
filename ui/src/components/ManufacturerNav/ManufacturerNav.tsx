import type { Manufacturer } from '@shared/ui'

export interface ManufacturerNavProps {
  readonly manufacturers: readonly Manufacturer[]
  readonly enabledManufacturers: ReadonlySet<Manufacturer>
  readonly onManufacturerClick: (manufacturer: Manufacturer) => void
}

export function ManufacturerNav({
  manufacturers,
  enabledManufacturers,
  onManufacturerClick,
}: ManufacturerNavProps) {
  return (
    <nav
      className="flex items-center gap-2 text-xs text-on-surface-muted"
      aria-label="Jump to manufacturer"
    >
      {manufacturers.map((manufacturer, index) => {
        const enabled = enabledManufacturers.has(manufacturer)
        return (
          <span key={manufacturer} className="flex items-center gap-2">
            {index > 0 && <span className="text-on-surface-faint">·</span>}
            <button
              type="button"
              onClick={() => onManufacturerClick(manufacturer)}
              disabled={!enabled}
              className={enabled ? 'hover:text-accent' : 'text-on-surface-faint cursor-default'}
            >
              {manufacturer}
            </button>
          </span>
        )
      })}
    </nav>
  )
}
