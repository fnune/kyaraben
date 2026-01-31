import { describe, expect, it } from 'vitest'

type ActionType =
  | 'will-install'
  | 'will-update'
  | 'will-uninstall'
  | 'already-installed'
  | 'shared-uninstall'
  | null

function getAction(
  enabled: boolean,
  installedVersion: string | null,
  effectiveVersion: string | null,
  emulatorSharedWith: readonly string[],
  emulatorInstalledFor: readonly string[],
): ActionType {
  if (!enabled) {
    if (!installedVersion) return null
    if (emulatorSharedWith.length > 0) return 'shared-uninstall'
    return 'will-uninstall'
  }

  if (!installedVersion) {
    if (emulatorInstalledFor.length > 0) return 'already-installed'
    return 'will-install'
  }

  if (effectiveVersion && installedVersion !== effectiveVersion) return 'will-update'
  return null
}

describe('getAction', () => {
  describe('when system is disabled', () => {
    it('returns null when emulator was never installed', () => {
      const result = getAction(false, null, null, [], [])
      expect(result).toBeNull()
    })

    it('returns will-uninstall when emulator is installed and not shared', () => {
      const result = getAction(false, '1.0.0', null, [], [])
      expect(result).toBe('will-uninstall')
    })

    it('returns shared-uninstall when emulator is shared with other systems', () => {
      const result = getAction(false, '1.0.0', null, ['SNES', 'Genesis'], [])
      expect(result).toBe('shared-uninstall')
    })
  })

  describe('when system is enabled', () => {
    it('returns will-install when emulator is not installed anywhere', () => {
      const result = getAction(true, null, '1.0.0', [], [])
      expect(result).toBe('will-install')
    })

    it('returns already-installed when emulator is installed for other systems', () => {
      const result = getAction(true, null, '1.0.0', [], ['SNES'])
      expect(result).toBe('already-installed')
    })

    it('returns null when emulator is already installed at correct version', () => {
      const result = getAction(true, '1.0.0', '1.0.0', [], [])
      expect(result).toBeNull()
    })

    it('returns will-update when version differs', () => {
      const result = getAction(true, '1.0.0', '2.0.0', [], [])
      expect(result).toBe('will-update')
    })

    it('returns null when effectiveVersion is null', () => {
      const result = getAction(true, '1.0.0', null, [], [])
      expect(result).toBeNull()
    })
  })

  describe('edge cases', () => {
    it('prioritizes already-installed over will-install even with shared systems', () => {
      const result = getAction(true, null, '1.0.0', ['SNES'], ['SNES'])
      expect(result).toBe('already-installed')
    })

    it('returns will-update even when shared with other systems', () => {
      const result = getAction(true, '1.0.0', '2.0.0', ['SNES'], ['SNES'])
      expect(result).toBe('will-update')
    })
  })
})
