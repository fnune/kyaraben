export interface PreferencesSummaryProps {
  readonly preset: string
  readonly resume: string
  readonly nintendoConfirm: string
  readonly onNavigate: () => void
}

function formatPreset(value: string): string {
  if (value === 'modern-pixels') return 'Modern pixels'
  if (value === 'upscaled') return 'Upscaled'
  if (value === 'pseudo-authentic') return 'Pseudo-authentic'
  return 'Manual'
}

function formatResume(value: string): string {
  if (value === 'recommended') return 'Recommended'
  if (value === 'off') return 'Off'
  return 'Manual'
}

function formatConfirm(value: string): string {
  return value === 'south' ? 'South' : 'East'
}

export function PreferencesSummary({
  preset,
  resume,
  nintendoConfirm,
  onNavigate,
}: PreferencesSummaryProps) {
  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest">
          Preferences
        </span>
        <button
          type="button"
          onClick={onNavigate}
          className="text-sm text-accent hover:text-accent-hover"
        >
          View preferences
        </button>
      </div>
      <div className="flex flex-wrap items-center gap-x-6 gap-y-1 px-4 py-3 bg-surface-alt rounded-card border border-outline text-sm">
        <span className="text-on-surface-muted">
          Display: <span className="text-on-surface">{formatPreset(preset)}</span>
        </span>
        <span className="text-on-surface-muted">
          Resume: <span className="text-on-surface">{formatResume(resume)}</span>
        </span>
        <span className="text-on-surface-muted">
          Confirm: <span className="text-on-surface">{formatConfirm(nintendoConfirm)}</span>
        </span>
      </div>
    </div>
  )
}
