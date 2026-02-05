import esde from '@/assets/frontends/esde.svg'
import type { FrontendID } from '@/types/model.gen'

const FRONTEND_LOGOS: Record<FrontendID, string> = {
  'es-de': esde,
}

export function getFrontendLogo(frontendId: FrontendID): string | undefined {
  return FRONTEND_LOGOS[frontendId]
}

export interface FrontendLogoProps {
  readonly frontendId: FrontendID
  readonly frontendName?: string
  readonly className?: string
}

export function FrontendLogo({ frontendId, frontendName, className = '' }: FrontendLogoProps) {
  const logo = FRONTEND_LOGOS[frontendId]

  if (!logo) {
    return null
  }

  return <img src={logo} alt={frontendName ?? frontendId} className={className} />
}
