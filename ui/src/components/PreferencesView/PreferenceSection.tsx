import type { ReactNode } from 'react'

interface PreferenceSectionProps {
  title: string
  intro: ReactNode
  controls: ReactNode
  support?: ReactNode
  headerAction?: ReactNode
}

export function PreferenceSection({
  title,
  intro,
  controls,
  support,
  headerAction,
}: PreferenceSectionProps) {
  return (
    <section className="mb-10">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-semibold text-on-surface-dim uppercase tracking-widest">
          {title}
        </h2>
        {headerAction}
      </div>

      <div className="bg-surface-alt rounded-card border border-outline overflow-hidden">
        <div className="p-4 border-b border-outline">{intro}</div>
        <div className="p-4 space-y-3">{controls}</div>
        {support && <div className="p-4 border-t border-outline">{support}</div>}
      </div>
    </section>
  )
}
