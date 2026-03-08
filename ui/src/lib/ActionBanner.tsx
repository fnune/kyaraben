import type { ReactNode } from 'react'

export interface ActionBannerProps {
  readonly title: ReactNode
  readonly description?: ReactNode
  readonly actions: ReactNode
  readonly variant?: 'accent' | 'success' | 'warning'
}

const variantStyles = {
  accent: 'bg-accent-muted border-accent/30',
  success: 'bg-status-success/10 border-status-success/30',
  warning: 'bg-status-warning/10 border-status-warning/30',
}

export function ActionBanner({
  title,
  description,
  actions,
  variant = 'accent',
}: ActionBannerProps) {
  return (
    <div className={`border-b px-4 py-3 ${variantStyles[variant]}`}>
      <div className="flex items-center justify-between gap-4">
        <div className="flex-1 min-w-0">
          <p className="text-sm text-on-surface">{title}</p>
          {description}
        </div>
        <div className="flex items-center gap-2 shrink-0">{actions}</div>
      </div>
    </div>
  )
}
