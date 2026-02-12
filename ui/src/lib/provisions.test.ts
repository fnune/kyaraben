import { describe, expect, it } from 'vitest'
import type { DoctorResponse, ProvisionResult } from '@/types/daemon'
import { getNewlyFoundProvisions } from './provisions'

function provision(
  filename: string,
  status: 'found' | 'missing',
  description = '',
  expectedPath?: string,
  foundPath?: string,
): ProvisionResult {
  const displayName = filename
  const baseResult: ProvisionResult = {
    filename,
    kind: 'bios',
    description,
    status,
    groupRequired: true,
    groupSatisfied: status === 'found',
    groupSize: 1,
    displayName,
    instructions: `Place ${filename} in this directory`,
  }
  if (expectedPath) {
    baseResult.expectedPath = expectedPath
  }
  if (foundPath) {
    baseResult.foundPath = foundPath
  }
  return baseResult
}

describe('getNewlyFoundProvisions', () => {
  it('returns empty array when no provisions changed', () => {
    const provisions: DoctorResponse = {
      'retroarch:bsnes': [provision('bios.bin', 'missing', 'USA')],
    }

    expect(getNewlyFoundProvisions(provisions, provisions)).toEqual([])
  })

  it('returns metadata for newly found provisions', () => {
    const oldProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [
        provision('sega_101.bin', 'missing', 'JP'),
        provision('mpr-17933.bin', 'missing', 'NA/EU'),
      ],
    }

    const newProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [
        provision(
          'sega_101.bin',
          'found',
          'JP',
          '/home/fausto/Emulation/saves/gamecube/EUR',
          '/home/fausto/Emulation/saves/gamecube/EUR/sega_101.bin',
        ),
        provision('mpr-17933.bin', 'missing', 'NA/EU'),
      ],
    }

    expect(getNewlyFoundProvisions(oldProvisions, newProvisions)).toEqual([
      {
        id: 'retroarch:beetle-saturn:sega_101.bin:JP:/home/fausto/Emulation/saves/gamecube/EUR',
        emulatorId: 'retroarch:beetle-saturn',
        filename: 'sega_101.bin',
        displayName: 'sega_101.bin',
        description: 'JP',
        expectedPath: '/home/fausto/Emulation/saves/gamecube/EUR',
        foundPath: '/home/fausto/Emulation/saves/gamecube/EUR/sega_101.bin',
      },
    ])
  })

  it('returns multiple entries when several provisions found across emulators', () => {
    const oldProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [
        provision('sega_101.bin', 'missing', 'JP'),
        provision('mpr-17933.bin', 'missing', 'NA/EU'),
      ],
    }

    const newProvisions: DoctorResponse = {
      'retroarch:beetle-saturn': [
        provision(
          'sega_101.bin',
          'found',
          'JP',
          '/home/fausto/Emulation/saves/gamecube/EUR',
          '/home/fausto/Emulation/saves/gamecube/EUR/sega_101.bin',
        ),
      ],
      pcsx2: [
        provision(
          'ps2-bios.bin',
          'found',
          'PS2 BIOS',
          '/home/fausto/Emulation/bios/ps2',
          '/home/fausto/Emulation/bios/ps2/ps2-bios.bin',
        ),
      ],
    }

    expect(getNewlyFoundProvisions(oldProvisions, newProvisions)).toEqual([
      {
        id: 'retroarch:beetle-saturn:sega_101.bin:JP:/home/fausto/Emulation/saves/gamecube/EUR',
        emulatorId: 'retroarch:beetle-saturn',
        filename: 'sega_101.bin',
        displayName: 'sega_101.bin',
        description: 'JP',
        expectedPath: '/home/fausto/Emulation/saves/gamecube/EUR',
        foundPath: '/home/fausto/Emulation/saves/gamecube/EUR/sega_101.bin',
      },
      {
        id: 'pcsx2:ps2-bios.bin:PS2 BIOS:/home/fausto/Emulation/bios/ps2',
        emulatorId: 'pcsx2',
        filename: 'ps2-bios.bin',
        displayName: 'ps2-bios.bin',
        description: 'PS2 BIOS',
        expectedPath: '/home/fausto/Emulation/bios/ps2',
        foundPath: '/home/fausto/Emulation/bios/ps2/ps2-bios.bin',
      },
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

    expect(getNewlyFoundProvisions(oldProvisions, newProvisions)).toEqual([
      {
        id: 'retroarch:beetle-saturn:sega_101.bin:JP:',
        emulatorId: 'retroarch:beetle-saturn',
        filename: 'sega_101.bin',
        displayName: 'sega_101.bin',
        description: 'JP',
        expectedPath: undefined,
        foundPath: undefined,
      },
    ])
  })
})
