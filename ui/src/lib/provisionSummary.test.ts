import { describe, expect, it } from 'vitest'
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
import {
  getProvisionSummaryCounts,
  getProvisionSummaryIconState,
  getProvisionSummaryParts,
} from '@/lib/provisionSummary'
import type { ProvisionResult } from '@/types/daemon'

function provisionResult({
  status,
  groupRequired,
  groupSatisfied,
  description = '',
  kind = 'bios',
  filename = 'bios.bin',
  displayName = 'bios.bin',
  groupSize = 1,
}: {
  status: string
  groupRequired: boolean
  groupSatisfied: boolean
  description?: string
  kind?: string
  filename?: string
  displayName?: string
  groupSize?: number
}): ProvisionResult {
  return {
    filename,
    kind,
    description,
    status,
    groupRequired,
    groupSatisfied,
    groupSize,
    displayName,
  }
}

describe('getProvisionSummaryCounts', () => {
  it('counts found, required missing, and optional missing', () => {
    const provisions = [
      provisionResult({ status: 'found', groupRequired: true, groupSatisfied: true }),
      provisionResult({
        status: 'found',
        groupRequired: true,
        groupSatisfied: true,
        filename: 'b.bin',
      }),
      provisionResult({
        status: 'missing',
        groupRequired: true,
        groupSatisfied: false,
        filename: 'c.bin',
      }),
      provisionResult({
        status: 'missing',
        groupRequired: false,
        groupSatisfied: false,
        filename: 'd.bin',
      }),
    ]

    const counts = getProvisionSummaryCounts(provisions)

    expect(counts).toEqual({
      foundCount: 2,
      missingRequiredCount: 1,
      missingOptionalCount: 1,
    })
  })
})

describe('getProvisionSummaryParts', () => {
  it('returns combined labels for found plus optional', () => {
    const provisions = [
      provisionResult({ status: 'found', groupRequired: true, groupSatisfied: true }),
      provisionResult({
        status: 'missing',
        groupRequired: false,
        groupSatisfied: false,
        filename: 'd.bin',
      }),
    ]

    const parts = getProvisionSummaryParts(provisions)

    expect(parts.map((part) => part.label)).toEqual(['1 provision ready', '1 optional'])
  })

  it('returns required labels when required provisions are missing', () => {
    const provisions = [
      provisionResult({ status: 'missing', groupRequired: true, groupSatisfied: false }),
    ]

    const parts = getProvisionSummaryParts(provisions)

    expect(parts.map((part) => part.label)).toEqual(['1 required provision'])
  })
})

describe('getProvisionSummaryIconState', () => {
  it('returns not needed for empty provisions', () => {
    const state = getProvisionSummaryIconState([])

    expect(state).toEqual({
      icon: PROVISION_NOT_NEEDED_ICON,
      iconColor: PROVISION_NOT_NEEDED_COLOR,
    })
  })

  it('returns missing for required unsatisfied provisions', () => {
    const provisions = [
      provisionResult({ status: 'missing', groupRequired: true, groupSatisfied: false }),
    ]

    const state = getProvisionSummaryIconState(provisions)

    expect(state).toEqual({
      icon: PROVISION_MISSING_ICON,
      iconColor: PROVISION_MISSING_COLOR,
    })
  })

  it('returns found when at least one is found and none are required missing', () => {
    const provisions = [
      provisionResult({ status: 'found', groupRequired: true, groupSatisfied: true }),
      provisionResult({
        status: 'missing',
        groupRequired: false,
        groupSatisfied: false,
        filename: 'd.bin',
      }),
    ]

    const state = getProvisionSummaryIconState(provisions)

    expect(state).toEqual({
      icon: PROVISION_FOUND_ICON,
      iconColor: PROVISION_FOUND_COLOR,
    })
  })

  it('returns optional when only optional provisions are missing', () => {
    const provisions = [
      provisionResult({ status: 'missing', groupRequired: false, groupSatisfied: false }),
    ]

    const state = getProvisionSummaryIconState(provisions)

    expect(state).toEqual({
      icon: OPTIONAL_PROVISION_ICON,
      iconColor: OPTIONAL_PROVISION_COLOR,
    })
  })
})
