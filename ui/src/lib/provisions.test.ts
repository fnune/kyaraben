import { describe, expect, it } from 'vitest'
import type { DoctorResponse, ProvisionResult } from '@/types/daemon'
import { getNewlyFoundProvisions } from './provisions'

function provision(
  filename: string,
  status: 'found' | 'missing',
  description = '',
): ProvisionResult {
  return {
    filename,
    kind: 'bios',
    description,
    status,
    groupRequired: true,
    groupSatisfied: status === 'found',
    groupSize: 1,
  }
}

describe('getNewlyFoundProvisions', () => {
  it('returns empty array when no provisions changed', () => {
    const provisions: DoctorResponse = {
      'retroarch:bsnes': [provision('bios.bin', 'missing', 'USA')],
    }

    expect(getNewlyFoundProvisions(provisions, provisions)).toEqual([])
  })

  it('returns filenames of newly found provisions', () => {
    const oldProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [
        provision('sega_101.bin', 'missing', 'JP'),
        provision('mpr-17933.bin', 'missing', 'NA/EU'),
      ],
    }

    const newProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [
        provision('sega_101.bin', 'found', 'JP'),
        provision('mpr-17933.bin', 'missing', 'NA/EU'),
      ],
    }

    expect(getNewlyFoundProvisions(oldProvisions, newProvisions)).toEqual(['sega_101.bin'])
  })

  it('returns multiple filenames when multiple provisions found', () => {
    const oldProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [
        provision('sega_101.bin', 'missing', 'JP'),
        provision('mpr-17933.bin', 'missing', 'NA/EU'),
      ],
    }

    const newProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [
        provision('sega_101.bin', 'found', 'JP'),
        provision('mpr-17933.bin', 'found', 'NA/EU'),
      ],
    }

    expect(getNewlyFoundProvisions(oldProvisions, newProvisions)).toEqual([
      'sega_101.bin',
      'mpr-17933.bin',
    ])
  })

  it('ignores provisions that were already found', () => {
    const oldProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [provision('sega_101.bin', 'found', 'JP')],
    }

    const newProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [provision('sega_101.bin', 'found', 'JP')],
    }

    expect(getNewlyFoundProvisions(oldProvisions, newProvisions)).toEqual([])
  })

  it('handles new emulators appearing in provisions', () => {
    const oldProvisions: DoctorResponse = {}

    const newProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [provision('sega_101.bin', 'found', 'JP')],
    }

    expect(getNewlyFoundProvisions(oldProvisions, newProvisions)).toEqual(['sega_101.bin'])
  })

  it('handles provisions across multiple emulators', () => {
    const oldProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [provision('sega_101.bin', 'missing', 'JP')],
      pcsx2: [provision('ps2-bios.bin', 'missing', 'USA')],
    }

    const newProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [provision('sega_101.bin', 'found', 'JP')],
      pcsx2: [provision('ps2-bios.bin', 'found', 'USA')],
    }

    expect(getNewlyFoundProvisions(oldProvisions, newProvisions)).toEqual([
      'sega_101.bin',
      'ps2-bios.bin',
    ])
  })
})
