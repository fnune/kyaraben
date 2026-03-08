import type { ProvisionResult } from '@shared/daemon'
import type { ReactNode } from 'react'
import { getProvisionSummaryIconState, getProvisionSummaryParts } from '@/lib/provisionSummary'

type ProvisionSummarySize = 'xs' | 'sm'
export interface ProvisionSummaryProps {
  readonly provisions: readonly ProvisionResult[]
  readonly size?: ProvisionSummarySize
  readonly className?: string
  readonly overrideLabel?: ReactNode
}

export function ProvisionSummary({
  provisions,
  size = 'xs',
  className,
  overrideLabel,
}: ProvisionSummaryProps) {
  const { icon, iconColor } = getProvisionSummaryIconState(provisions)

  let parts: ReactNode = null
  if (!overrideLabel) {
    const summaryParts = getProvisionSummaryParts(provisions)
    parts = summaryParts.map((part, index) => {
      const textColor = part.kind === 'optional' ? 'text-on-surface-dim' : 'text-on-surface-muted'
      return (
        <span key={`${part.kind}-${part.label}`} className={textColor}>
          {index > 0 ? ', ' : ''}
          {part.label}
        </span>
      )
    })
  }

  const sizeClass = size === 'sm' ? 'text-sm' : 'text-xs'

  const wrapperClass = `flex items-center gap-2 ${sizeClass}${className ? ` ${className}` : ''}`

  return (
    <div className={wrapperClass}>
      <span className={`${iconColor} w-3 text-center shrink-0`}>{icon}</span>
      <span className="min-w-0">{overrideLabel ?? parts}</span>
    </div>
  )
}
