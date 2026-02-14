import { describe, expect, it } from 'vitest'
import {
  addChange,
  type ChangeType,
  calculateEmulatorSizes,
  type EmulatorChangeInput,
  emptyChangeSummary,
  getChangeGroups,
} from './changeUtils'

function emulator(
  id: string,
  packageName: string,
  changeType: ChangeType,
  isInstalled: boolean,
  downloadBytes = 172_000_000,
  coreBytes = 4_000_000,
): EmulatorChangeInput {
  return { id, packageName, downloadBytes, coreBytes, changeType, isInstalled }
}

describe('calculateEmulatorSizes', () => {
  describe('installs with shared packages', () => {
    it('includes package size when installing first core of a package', () => {
      const inputs = [emulator('bsnes', 'retroarch', 'install', false)]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('bsnes')).toBe(172_000_000 + 4_000_000)
    })

    it('excludes package size when another core is already installed', () => {
      const inputs = [
        emulator('bsnes', 'retroarch', null, true),
        emulator('snes9x', 'retroarch', 'install', false),
      ]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('snes9x')).toBe(4_000_000)
    })

    it('includes package size only for first install in batch', () => {
      const inputs = [
        emulator('bsnes', 'retroarch', 'install', false),
        emulator('snes9x', 'retroarch', 'install', false),
      ]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('bsnes')).toBe(172_000_000 + 4_000_000)
      expect(sizes.get('snes9x')).toBe(4_000_000)
    })
  })

  describe('removals with shared packages', () => {
    it('includes package size when removing last core of a package', () => {
      const inputs = [emulator('bsnes', 'retroarch', 'remove', true)]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('bsnes')).toBe(172_000_000 + 4_000_000)
    })

    it('excludes package size when another core remains installed', () => {
      const inputs = [
        emulator('bsnes', 'retroarch', null, true),
        emulator('snes9x', 'retroarch', 'remove', true),
      ]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('snes9x')).toBe(4_000_000)
    })

    it('includes package size only for first removal in batch when removing all cores', () => {
      const inputs = [
        emulator('bsnes', 'retroarch', 'remove', true),
        emulator('snes9x', 'retroarch', 'remove', true),
      ]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('bsnes')).toBe(172_000_000 + 4_000_000)
      expect(sizes.get('snes9x')).toBe(4_000_000)
    })
  })

  describe('mixed scenarios', () => {
    it('handles adding one core while another is being removed from same package', () => {
      const inputs = [
        emulator('bsnes', 'retroarch', 'install', false),
        emulator('snes9x', 'retroarch', 'remove', true),
      ]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('bsnes')).toBe(4_000_000)
      expect(sizes.get('snes9x')).toBe(4_000_000)
    })

    it('handles multiple packages independently', () => {
      const inputs = [
        emulator('bsnes', 'retroarch', 'install', false),
        emulator('pcsx2', 'pcsx2', 'install', false, 50_000_000, 0),
      ]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('bsnes')).toBe(172_000_000 + 4_000_000)
      expect(sizes.get('pcsx2')).toBe(50_000_000)
    })

    it('returns zero for emulators with no change', () => {
      const inputs = [emulator('bsnes', 'retroarch', null, true)]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('bsnes')).toBe(4_000_000)
    })
  })

  describe('upgrades', () => {
    it('treats upgrades like installs for package size calculation', () => {
      const inputs = [emulator('bsnes', 'retroarch', 'upgrade', true)]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('bsnes')).toBe(4_000_000)
    })

    it('includes package size only if no cores currently installed', () => {
      const inputs = [emulator('bsnes', 'retroarch', 'upgrade', false)]
      const sizes = calculateEmulatorSizes(inputs)

      expect(sizes.get('bsnes')).toBe(172_000_000 + 4_000_000)
    })
  })
})

describe('addChange', () => {
  it('counts items with zero download size', () => {
    let summary = emptyChangeSummary()
    summary = addChange(summary, 'install', 0, { id: 'esde', name: 'ES-DE' })

    expect(summary.installs).toBe(1)
    expect(summary.total).toBe(1)
    expect(summary.installItems).toEqual([{ id: 'esde', name: 'ES-DE' }])
    expect(summary.downloadBytes).toBe(0)
  })

  it('counts items without size parameter', () => {
    let summary = emptyChangeSummary()
    summary = addChange(summary, 'install', undefined, { id: 'esde', name: 'ES-DE' })

    expect(summary.installs).toBe(1)
    expect(summary.installItems).toEqual([{ id: 'esde', name: 'ES-DE' }])
  })

  it('does not count items with null change type', () => {
    let summary = emptyChangeSummary()
    summary = addChange(summary, null, 100, { id: 'esde', name: 'ES-DE' })

    expect(summary.total).toBe(0)
    expect(summary.installItems).toEqual([])
  })

  it('accumulates multiple items of same type', () => {
    let summary = emptyChangeSummary()
    summary = addChange(summary, 'install', 100, { id: 'pcsx2', name: 'PCSX2' })
    summary = addChange(summary, 'install', 0, { id: 'esde', name: 'ES-DE' })

    expect(summary.installs).toBe(2)
    expect(summary.installItems).toEqual([
      { id: 'pcsx2', name: 'PCSX2' },
      { id: 'esde', name: 'ES-DE' },
    ])
    expect(summary.downloadBytes).toBe(100)
  })

  it('tracks removes with freed bytes', () => {
    let summary = emptyChangeSummary()
    summary = addChange(summary, 'remove', 50_000_000, { id: 'dolphin', name: 'Dolphin' })

    expect(summary.removes).toBe(1)
    expect(summary.removeItems).toEqual([{ id: 'dolphin', name: 'Dolphin' }])
    expect(summary.freeBytes).toBe(50_000_000)
  })
})

describe('getChangeGroups', () => {
  it('returns groups for items with zero size', () => {
    let summary = emptyChangeSummary()
    summary = addChange(summary, 'install', 0, { id: 'esde', name: 'ES-DE' })

    const groups = getChangeGroups(summary)

    expect(groups).toHaveLength(1)
    expect(groups[0]).toEqual({
      type: 'install',
      items: [{ id: 'esde', name: 'ES-DE' }],
    })
  })

  it('returns multiple groups when different change types exist', () => {
    let summary = emptyChangeSummary()
    summary = addChange(summary, 'install', 0, { id: 'esde', name: 'ES-DE' })
    summary = addChange(summary, 'remove', 100, { id: 'dolphin', name: 'Dolphin' })

    const groups = getChangeGroups(summary)

    expect(groups).toHaveLength(2)
    expect(groups.find((g) => g.type === 'install')?.items).toEqual([{ id: 'esde', name: 'ES-DE' }])
    expect(groups.find((g) => g.type === 'remove')?.items).toEqual([
      { id: 'dolphin', name: 'Dolphin' },
    ])
  })
})
