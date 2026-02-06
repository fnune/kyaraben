import { describe, expect, it } from 'vitest'
import type { EmulatorID, System, SystemID } from '@/types/daemon'

function getPackageName(emulatorId: EmulatorID): string {
  if (emulatorId.includes(':')) {
    return emulatorId.split(':')[0] ?? emulatorId
  }
  return emulatorId
}

function getEmulatorSharingInfo(
  systems: readonly System[],
  selections: ReadonlyMap<SystemID, EmulatorID>,
  installedVersions: ReadonlyMap<EmulatorID, string>,
  currentSystemId: SystemID,
  currentEmulatorId: EmulatorID | null,
): { sharedWith: string[]; installedFor: string[] } {
  if (!currentEmulatorId) {
    return { sharedWith: [], installedFor: [] }
  }

  const currentPackage = getPackageName(currentEmulatorId)
  const sharedWith: string[] = []
  const installedFor: string[] = []

  for (const system of systems) {
    if (system.id === currentSystemId) continue

    const selectedEmulator = selections.get(system.id)
    if (!selectedEmulator) continue

    const selectedPackage = getPackageName(selectedEmulator)
    if (selectedPackage === currentPackage) {
      sharedWith.push(system.label)
      for (const [installedId] of installedVersions) {
        if (getPackageName(installedId) === currentPackage) {
          installedFor.push(system.label)
          break
        }
      }
    }
  }

  return { sharedWith, installedFor }
}

const mockSystems: System[] = [
  {
    id: 'nes',
    name: 'Nintendo Entertainment System',
    description: '',
    manufacturer: 'Nintendo',
    label: 'NES',
    defaultEmulatorId: 'retroarch:mesen',
    emulators: [
      {
        id: 'retroarch:mesen',
        name: 'RetroArch (Mesen)',
        defaultVersion: '1.22.2',
        availableVersions: [],
      },
    ],
  },
  {
    id: 'snes',
    name: 'Super Nintendo',
    description: '',
    manufacturer: 'Nintendo',
    label: 'SNES',
    defaultEmulatorId: 'retroarch:bsnes',
    emulators: [
      {
        id: 'retroarch:bsnes',
        name: 'RetroArch (bsnes)',
        defaultVersion: '1.22.2',
        availableVersions: [],
      },
    ],
  },
  {
    id: 'genesis',
    name: 'Sega Genesis',
    description: '',
    manufacturer: 'Sega',
    label: 'Genesis',
    defaultEmulatorId: 'retroarch:genesis_plus_gx',
    emulators: [
      {
        id: 'retroarch:genesis_plus_gx',
        name: 'RetroArch (Genesis Plus GX)',
        defaultVersion: '1.22.2',
        availableVersions: [],
      },
    ],
  },
  {
    id: 'gb',
    name: 'Game Boy',
    description: '',
    manufacturer: 'Nintendo',
    label: 'GB',
    defaultEmulatorId: 'mgba',
    emulators: [{ id: 'mgba', name: 'mGBA', defaultVersion: '0.10.5', availableVersions: [] }],
  },
  {
    id: 'gbc',
    name: 'Game Boy Color',
    description: '',
    manufacturer: 'Nintendo',
    label: 'GBC',
    defaultEmulatorId: 'mgba',
    emulators: [{ id: 'mgba', name: 'mGBA', defaultVersion: '0.10.5', availableVersions: [] }],
  },
  {
    id: 'gba',
    name: 'Game Boy Advance',
    description: '',
    manufacturer: 'Nintendo',
    label: 'GBA',
    defaultEmulatorId: 'mgba',
    emulators: [{ id: 'mgba', name: 'mGBA', defaultVersion: '0.10.5', availableVersions: [] }],
  },
]

describe('getEmulatorSharingInfo', () => {
  it('returns empty arrays when emulator is not shared', () => {
    const selections = new Map<SystemID, EmulatorID>([['nes', 'retroarch:mesen']])
    const installedVersions = new Map<EmulatorID, string>()

    const result = getEmulatorSharingInfo(
      mockSystems,
      selections,
      installedVersions,
      'nes',
      'retroarch:mesen',
    )

    expect(result.sharedWith).toEqual([])
    expect(result.installedFor).toEqual([])
  })

  it('returns shared systems when emulator is used by multiple enabled systems', () => {
    const selections = new Map<SystemID, EmulatorID>([
      ['gb', 'mgba'],
      ['gbc', 'mgba'],
      ['gba', 'mgba'],
    ])
    const installedVersions = new Map<EmulatorID, string>()

    const result = getEmulatorSharingInfo(mockSystems, selections, installedVersions, 'gb', 'mgba')

    expect(result.sharedWith).toEqual(['GBC', 'GBA'])
    expect(result.installedFor).toEqual([])
  })

  it('returns installed systems when emulator is already installed for other systems', () => {
    const selections = new Map<SystemID, EmulatorID>([
      ['gb', 'mgba'],
      ['gbc', 'mgba'],
      ['gba', 'mgba'],
    ])
    const installedVersions = new Map<EmulatorID, string>([['mgba', '0.10.5']])

    const result = getEmulatorSharingInfo(mockSystems, selections, installedVersions, 'gb', 'mgba')

    expect(result.sharedWith).toEqual(['GBC', 'GBA'])
    expect(result.installedFor).toEqual(['GBC', 'GBA'])
  })

  it('excludes the current system from results', () => {
    const selections = new Map<SystemID, EmulatorID>([
      ['gb', 'mgba'],
      ['gbc', 'mgba'],
    ])
    const installedVersions = new Map<EmulatorID, string>([['mgba', '0.10.5']])

    const result = getEmulatorSharingInfo(mockSystems, selections, installedVersions, 'gb', 'mgba')

    expect(result.sharedWith).not.toContain('GB')
    expect(result.installedFor).not.toContain('GB')
  })

  it('returns empty arrays when currentEmulatorId is null', () => {
    const selections = new Map<SystemID, EmulatorID>([['gb', 'mgba']])
    const installedVersions = new Map<EmulatorID, string>([['mgba', '0.10.5']])

    const result = getEmulatorSharingInfo(mockSystems, selections, installedVersions, 'nes', null)

    expect(result.sharedWith).toEqual([])
    expect(result.installedFor).toEqual([])
  })

  it('only includes enabled systems in shared list', () => {
    const selections = new Map<SystemID, EmulatorID>([
      ['gb', 'mgba'],
      ['gba', 'mgba'],
    ])
    const installedVersions = new Map<EmulatorID, string>()

    const result = getEmulatorSharingInfo(mockSystems, selections, installedVersions, 'gb', 'mgba')

    expect(result.sharedWith).toEqual(['GBA'])
    expect(result.sharedWith).not.toContain('GBC')
  })

  it('treats RetroArch cores as sharing the same package', () => {
    const selections = new Map<SystemID, EmulatorID>([
      ['nes', 'retroarch:mesen'],
      ['snes', 'retroarch:bsnes'],
      ['genesis', 'retroarch:genesis_plus_gx'],
    ])
    const installedVersions = new Map<EmulatorID, string>()

    const result = getEmulatorSharingInfo(
      mockSystems,
      selections,
      installedVersions,
      'nes',
      'retroarch:mesen',
    )

    expect(result.sharedWith).toEqual(['SNES', 'Genesis'])
  })

  it('detects installed retroarch from different core', () => {
    const selections = new Map<SystemID, EmulatorID>([
      ['nes', 'retroarch:mesen'],
      ['snes', 'retroarch:bsnes'],
    ])
    const installedVersions = new Map<EmulatorID, string>([['retroarch:bsnes', '1.22.2']])

    const result = getEmulatorSharingInfo(
      mockSystems,
      selections,
      installedVersions,
      'nes',
      'retroarch:mesen',
    )

    expect(result.sharedWith).toEqual(['SNES'])
    expect(result.installedFor).toEqual(['SNES'])
  })

  it('does not mix up different standalone emulators', () => {
    const selections = new Map<SystemID, EmulatorID>([
      ['gb', 'mgba'],
      ['nes', 'retroarch:mesen'],
    ])
    const installedVersions = new Map<EmulatorID, string>()

    const result = getEmulatorSharingInfo(mockSystems, selections, installedVersions, 'gb', 'mgba')

    expect(result.sharedWith).toEqual([])
    expect(result.installedFor).toEqual([])
  })
})

describe('getPackageName', () => {
  it('extracts base name from colon-separated emulator ID', () => {
    expect(getPackageName('retroarch:mesen')).toBe('retroarch')
    expect(getPackageName('retroarch:bsnes')).toBe('retroarch')
  })

  it('returns the ID unchanged for standalone emulators', () => {
    expect(getPackageName('mgba')).toBe('mgba')
    expect(getPackageName('dolphin')).toBe('dolphin')
  })
})
