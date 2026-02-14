import { describe, expect, it } from 'vitest'
import { type ChangeType, calculateEmulatorSizes, type EmulatorChangeInput } from './changeUtils'

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
