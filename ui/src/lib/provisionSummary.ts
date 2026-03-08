import type { ProvisionResult } from '@shared/daemon'
import {
  OPTIONAL_PROVISION_COLOR,
  OPTIONAL_PROVISION_ICON,
  PROVISION_FOUND_COLOR,
  PROVISION_FOUND_ICON,
  PROVISION_MISSING_COLOR,
  PROVISION_MISSING_ICON,
  PROVISION_NOT_NEEDED_COLOR,
  PROVISION_NOT_NEEDED_ICON,
} from '@/lib/provisionStatus'

export type ProvisionSummaryPartKind = 'ready' | 'required' | 'optional' | 'none'

export interface ProvisionSummaryPart {
  kind: ProvisionSummaryPartKind
  label: string
}

export interface ProvisionSummaryIconState {
  icon: string
  iconColor: string
}

export interface ProvisionSummaryCounts {
  foundCount: number
  missingRequiredCount: number
  missingOptionalCount: number
}

export function getProvisionSummaryCounts(
  provisions: readonly ProvisionResult[],
): ProvisionSummaryCounts {
  const foundCount = provisions.filter((p) => p.status === 'found').length
  const missingRequiredCount = provisions.filter(
    (p) => p.status !== 'found' && p.groupRequired && !p.groupSatisfied,
  ).length
  const missingOptionalCount = provisions.filter(
    (p) => p.status !== 'found' && !p.groupRequired,
  ).length

  return { foundCount, missingRequiredCount, missingOptionalCount }
}

export function getProvisionSummaryParts(
  provisions: readonly ProvisionResult[],
): ProvisionSummaryPart[] {
  if (provisions.length === 0) {
    return [{ kind: 'none', label: 'No provisions' }]
  }

  const { foundCount, missingRequiredCount, missingOptionalCount } =
    getProvisionSummaryCounts(provisions)
  const parts: ProvisionSummaryPart[] = []
  const hasCombinedSummary =
    foundCount > 0 && (missingRequiredCount > 0 || missingOptionalCount > 0)

  if (foundCount > 0) {
    parts.push({
      kind: 'ready',
      label: `${foundCount} provision${foundCount > 1 ? 's' : ''} ready`,
    })
  }

  if (missingRequiredCount > 0) {
    const label = `${missingRequiredCount} required provision${missingRequiredCount > 1 ? 's' : ''}`
    parts.push({ kind: 'required', label })
  }

  if (missingOptionalCount > 0) {
    const label = hasCombinedSummary
      ? `${missingOptionalCount} optional`
      : `${missingOptionalCount} optional provision${missingOptionalCount > 1 ? 's' : ''}`
    parts.push({ kind: 'optional', label })
  }

  return parts
}

export function getProvisionSummaryIconState(
  provisions: readonly ProvisionResult[],
): ProvisionSummaryIconState {
  if (provisions.length === 0) {
    return {
      icon: PROVISION_NOT_NEEDED_ICON,
      iconColor: PROVISION_NOT_NEEDED_COLOR,
    }
  }

  const { foundCount, missingRequiredCount } = getProvisionSummaryCounts(provisions)

  if (missingRequiredCount > 0) {
    return {
      icon: PROVISION_MISSING_ICON,
      iconColor: PROVISION_MISSING_COLOR,
    }
  }

  if (foundCount > 0) {
    return {
      icon: PROVISION_FOUND_ICON,
      iconColor: PROVISION_FOUND_COLOR,
    }
  }

  return {
    icon: OPTIONAL_PROVISION_ICON,
    iconColor: OPTIONAL_PROVISION_COLOR,
  }
}
